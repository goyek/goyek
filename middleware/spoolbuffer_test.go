package middleware

import (
	"bytes"
	"errors"
	"io"
	"os"
	"runtime"
	"testing"
)

func TestSpoolBufferThreshold(t *testing.T) {
	tests := []struct {
		name        string
		size        int
		writeString bool
		wantSpill   bool
	}{
		{name: "Write below threshold", size: maxBufferedOutputBytes - 1},
		{name: "Write at threshold", size: maxBufferedOutputBytes},
		{name: "Write above threshold", size: maxBufferedOutputBytes + 1, wantSpill: true},
		{name: "WriteString below threshold", size: maxBufferedOutputBytes - 1, writeString: true},
		{name: "WriteString at threshold", size: maxBufferedOutputBytes, writeString: true},
		{name: "WriteString above threshold", size: maxBufferedOutputBytes + 1, writeString: true, wantSpill: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := newSpoolBuffer(maxBufferedOutputBytes)
			defer buffer.Close()
			want := bytes.Repeat([]byte("x"), tt.size)

			var n int
			var err error
			if tt.writeString {
				n, err = buffer.WriteString(string(want))
			} else {
				n, err = buffer.Write(want)
			}
			if err != nil {
				t.Fatalf("write error = %v", err)
			}
			if n != len(want) {
				t.Fatalf("write count = %d, want %d", n, len(want))
			}
			if got := buffer.file != nil; got != tt.wantSpill {
				t.Fatalf("spilled = %v, want %v", got, tt.wantSpill)
			}
			if tt.wantSpill && buffer.memory != nil {
				t.Fatal("in-memory buffer was retained after spilling")
			}

			destination := &countingWriter{}
			written, err := buffer.WriteTo(destination)
			if err != nil {
				t.Fatalf("WriteTo() error = %v", err)
			}
			if written != int64(len(want)) {
				t.Fatalf("WriteTo() = %d, want %d", written, len(want))
			}
			if !tt.wantSpill && destination.calls != 1 {
				t.Fatalf("destination Write calls = %d, want 1", destination.calls)
			}
			if !bytes.Equal(destination.Bytes(), want) {
				t.Fatal("WriteTo() changed the buffered data")
			}
		})
	}
}

func TestSpoolBufferSpillsAndPreservesOutput(t *testing.T) {
	buffer := newSpoolBuffer(maxBufferedOutputBytes)
	want := append(bytes.Repeat([]byte("line\n"), maxBufferedOutputBytes/5+1), 0, 0xff)
	want = append(want, []byte("no-final-newline")...)

	n, err := buffer.Write(want[:maxBufferedOutputBytes+1])
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if n != maxBufferedOutputBytes+1 {
		t.Fatalf("Write() = %d, want %d", n, maxBufferedOutputBytes+1)
	}
	n, err = buffer.WriteString(string(want[maxBufferedOutputBytes+1:]))
	if err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}
	if n != len(want)-(maxBufferedOutputBytes+1) {
		t.Fatalf("WriteString() = %d, want %d", n, len(want)-(maxBufferedOutputBytes+1))
	}
	if buffer.file == nil {
		t.Fatal("buffer did not spill above the memory limit")
	}
	if buffer.memory != nil {
		t.Fatal("in-memory buffer was retained after spilling")
	}

	path := buffer.file.Name()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat(%q) error = %v", path, err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("temporary file permissions = %o; group or other access is set", info.Mode().Perm())
	}

	destination := &bytes.Buffer{}
	written, err := buffer.WriteTo(destination)
	if err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	if written != int64(len(want)) {
		t.Fatalf("WriteTo() = %d, want %d", written, len(want))
	}
	if !bytes.Equal(destination.Bytes(), want) {
		t.Fatal("WriteTo() changed spilled binary output")
	}

	if err := buffer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("temporary file still exists after Close(): %v", err)
	}
}

func TestSpoolBufferWriteStringCrossesThreshold(t *testing.T) {
	buffer := newSpoolBuffer(4)
	defer buffer.Close()

	if _, err := buffer.WriteString("abcd"); err != nil {
		t.Fatalf("first WriteString() error = %v", err)
	}
	if buffer.file != nil {
		t.Fatal("buffer spilled at the memory limit")
	}
	if _, err := buffer.WriteString("ef"); err != nil {
		t.Fatalf("second WriteString() error = %v", err)
	}
	if buffer.file == nil {
		t.Fatal("buffer did not spill above the memory limit")
	}

	destination := &bytes.Buffer{}
	if _, err := buffer.WriteTo(destination); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	if got, want := destination.String(), "abcdef"; got != want {
		t.Fatalf("WriteTo() = %q, want %q", got, want)
	}
}

func TestSpoolBufferRetriesAfterCreateFailure(t *testing.T) {
	wantErr := errors.New("create failed")
	directory := t.TempDir()
	createCalls := 0
	buffer := newSpoolBufferWithFileOps(
		4,
		func() (spoolFile, error) {
			createCalls++
			if createCalls == 1 {
				return nil, wantErr
			}
			return os.CreateTemp(directory, "retry-")
		},
		os.Remove,
	)
	defer buffer.Close()

	if _, err := buffer.WriteString("base"); err != nil {
		t.Fatalf("initial WriteString() error = %v", err)
	}
	n, err := buffer.WriteString("tail")
	if n != 0 || !errors.Is(err, wantErr) {
		t.Fatalf("failed WriteString() = (%d, %v), want (0, %v)", n, err, wantErr)
	}
	if buffer.file != nil || buffer.memory.String() != "base" {
		t.Fatal("create failure changed buffered data")
	}

	n, err = buffer.WriteString("tail")
	if err != nil || n != len("tail") {
		t.Fatalf("retried WriteString() = (%d, %v), want (%d, nil)", n, err, len("tail"))
	}
	destination := &bytes.Buffer{}
	if _, err := buffer.WriteTo(destination); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	if got, want := destination.String(), "basetail"; got != want {
		t.Fatalf("WriteTo() = %q, want %q", got, want)
	}
}

func TestSpoolBufferRetriesAfterCopyFailure(t *testing.T) {
	wantErr := errors.New("copy failed")
	directory := t.TempDir()
	var paths []string
	createCalls := 0
	buffer := newSpoolBufferWithFileOps(
		4,
		func() (spoolFile, error) {
			file, err := os.CreateTemp(directory, "copy-")
			if err != nil {
				return nil, err
			}
			paths = append(paths, file.Name())
			createCalls++
			if createCalls == 1 {
				return &faultFile{File: file, failOnWrite: 1, partial: 2, err: wantErr}, nil
			}
			return file, nil
		},
		os.Remove,
	)
	defer buffer.Close()

	if _, err := buffer.WriteString("base"); err != nil {
		t.Fatalf("initial WriteString() error = %v", err)
	}
	n, err := buffer.WriteString("tail")
	if n != 0 || !errors.Is(err, wantErr) {
		t.Fatalf("failed WriteString() = (%d, %v), want (0, %v)", n, err, wantErr)
	}
	if buffer.file != nil || buffer.memory.String() != "base" {
		t.Fatal("copy failure changed buffered data")
	}
	if _, err := os.Stat(paths[0]); !os.IsNotExist(err) {
		t.Fatalf("failed spill file still exists: %v", err)
	}

	if _, err := buffer.WriteString("tail"); err != nil {
		t.Fatalf("retried WriteString() error = %v", err)
	}
	destination := &bytes.Buffer{}
	if _, err := buffer.WriteTo(destination); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	if got, want := destination.String(), "basetail"; got != want {
		t.Fatalf("WriteTo() = %q, want %q", got, want)
	}
}

func TestSpoolBufferRetriesAfterPartialSpillWrite(t *testing.T) {
	wantErr := errors.New("write failed")
	directory := t.TempDir()
	buffer := newSpoolBufferWithFileOps(
		3,
		func() (spoolFile, error) {
			file, err := os.CreateTemp(directory, "partial-")
			if err != nil {
				return nil, err
			}
			return &faultFile{File: file, failOnWrite: 2, partial: 2, err: wantErr}, nil
		},
		os.Remove,
	)
	defer buffer.Close()

	if _, err := buffer.WriteString("old"); err != nil {
		t.Fatalf("initial WriteString() error = %v", err)
	}
	n, err := buffer.Write([]byte("ABCDE"))
	if n != 2 || !errors.Is(err, wantErr) {
		t.Fatalf("partial Write() = (%d, %v), want (2, %v)", n, err, wantErr)
	}
	if _, err := buffer.Write([]byte("CDE")); err != nil {
		t.Fatalf("retried Write() error = %v", err)
	}

	destination := &bytes.Buffer{}
	if _, err := buffer.WriteTo(destination); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	if got, want := destination.String(), "oldABCDE"; got != want {
		t.Fatalf("WriteTo() = %q, want %q", got, want)
	}
}

func TestSpoolBufferWriteToRetriesShortWrites(t *testing.T) {
	tests := []struct {
		name  string
		limit int
	}{
		{name: "memory", limit: 100},
		{name: "file", limit: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := newSpoolBuffer(tt.limit)
			defer buffer.Close()
			const want = "0123456789"
			if _, err := buffer.WriteString(want); err != nil {
				t.Fatalf("WriteString() error = %v", err)
			}

			destination := &shortWriter{max: 3}
			var total int64
			for total < int64(len(want)) {
				n, err := buffer.WriteTo(destination)
				total += n
				if total < int64(len(want)) && !errors.Is(err, io.ErrShortWrite) {
					t.Fatalf("WriteTo() error = %v, want %v", err, io.ErrShortWrite)
				}
				if total == int64(len(want)) && err != nil {
					t.Fatalf("final WriteTo() error = %v", err)
				}
				if n == 0 {
					t.Fatal("WriteTo() made no progress")
				}
			}
			if got := destination.String(); got != want {
				t.Fatalf("retried WriteTo() = %q, want %q", got, want)
			}
		})
	}
}

func TestSpoolBufferSpillUsesReaderFrom(t *testing.T) {
	buffer := newSpoolBuffer(1)
	defer buffer.Close()
	if _, err := buffer.WriteString("spilled"); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}

	destination := &readerFromWriter{}
	if _, err := buffer.WriteTo(destination); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	if destination.readFromCalls != 1 {
		t.Fatalf("ReadFrom() calls = %d, want 1", destination.readFromCalls)
	}
	if destination.writeCalls != 0 {
		t.Fatalf("Write() calls = %d, want 0", destination.writeCalls)
	}
	if got, want := destination.String(), "spilled"; got != want {
		t.Fatalf("WriteTo() = %q, want %q", got, want)
	}
}

func TestSpoolBufferClose(t *testing.T) {
	t.Run("memory", func(t *testing.T) {
		buffer := newSpoolBuffer(10)
		if _, err := buffer.WriteString("data"); err != nil {
			t.Fatalf("WriteString() error = %v", err)
		}
		if err := buffer.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
		if buffer.memory != nil {
			t.Fatal("Close() retained in-memory data")
		}
		if err := buffer.Close(); err != nil {
			t.Fatalf("second Close() error = %v", err)
		}
		if _, err := buffer.Write(nil); !errors.Is(err, errSpoolBufferClosed) {
			t.Fatalf("Write() error = %v, want %v", err, errSpoolBufferClosed)
		}
		if _, err := buffer.WriteString(""); !errors.Is(err, errSpoolBufferClosed) {
			t.Fatalf("WriteString() error = %v, want %v", err, errSpoolBufferClosed)
		}
		if _, err := buffer.WriteTo(io.Discard); !errors.Is(err, errSpoolBufferClosed) {
			t.Fatalf("WriteTo() error = %v, want %v", err, errSpoolBufferClosed)
		}
	})

	t.Run("file", func(t *testing.T) {
		buffer := newSpoolBuffer(1)
		if _, err := buffer.WriteString("data"); err != nil {
			t.Fatalf("WriteString() error = %v", err)
		}
		path := buffer.file.Name()
		if err := buffer.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("temporary file still exists after Close(): %v", err)
		}
	})
}

func TestSpoolBufferCloseRemovesFileAfterCloseError(t *testing.T) {
	wantErr := errors.New("close failed")
	directory := t.TempDir()
	var path string
	buffer := newSpoolBufferWithFileOps(
		1,
		func() (spoolFile, error) {
			file, err := os.CreateTemp(directory, "close-")
			if err != nil {
				return nil, err
			}
			path = file.Name()
			return &closeErrorFile{File: file, err: wantErr}, nil
		},
		os.Remove,
	)
	if _, err := buffer.WriteString("data"); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}
	if err := buffer.Close(); !errors.Is(err, wantErr) {
		t.Fatalf("Close() error = %v, want %v", err, wantErr)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("temporary file still exists after Close(): %v", err)
	}
}

type countingWriter struct {
	bytes.Buffer
	calls int
}

func (w *countingWriter) Write(p []byte) (int, error) {
	w.calls++
	return w.Buffer.Write(p)
}

type shortWriter struct {
	buffer bytes.Buffer
	max    int
}

func (w *shortWriter) Write(p []byte) (int, error) {
	if len(p) > w.max {
		p = p[:w.max]
	}
	return w.buffer.Write(p)
}

func (w *shortWriter) String() string {
	return w.buffer.String()
}

type readerFromWriter struct {
	bytes.Buffer
	readFromCalls int
	writeCalls    int
}

func (w *readerFromWriter) Write(p []byte) (int, error) {
	w.writeCalls++
	return w.Buffer.Write(p)
}

func (w *readerFromWriter) ReadFrom(r io.Reader) (int64, error) {
	w.readFromCalls++
	return w.Buffer.ReadFrom(r)
}

type faultFile struct {
	*os.File
	writes      int
	failOnWrite int
	partial     int
	err         error
}

func (f *faultFile) Write(p []byte) (int, error) {
	f.writes++
	if f.writes != f.failOnWrite {
		return f.File.Write(p)
	}
	n := f.partial
	if n > len(p) {
		n = len(p)
	}
	if n > 0 {
		if _, err := f.File.Write(p[:n]); err != nil {
			return 0, err
		}
	}
	return n, f.err
}

type closeErrorFile struct {
	*os.File
	err error
}

func (f *closeErrorFile) Close() error {
	_ = f.File.Close()
	return f.err
}

var (
	_ io.StringWriter = (*spoolBuffer)(nil)
	_ io.WriterTo     = (*spoolBuffer)(nil)
)
