package middleware

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSpoolBuffer_inMemory(t *testing.T) {
	buffer := newSpoolBuffer(5)
	t.Cleanup(func() { _ = buffer.Close() })

	if n, err := buffer.Write([]byte("ab")); err != nil || n != 2 {
		t.Fatalf("Write returned %d, %v; want 2, nil", n, err)
	}
	if n, err := buffer.WriteString("cde"); err != nil || n != 3 {
		t.Fatalf("WriteString returned %d, %v; want 3, nil", n, err)
	}
	if buffer.file != nil {
		t.Fatal("output at the memory limit spilled to a file")
	}

	var output bytes.Buffer
	if n, err := buffer.WriteTo(&output); err != nil || n != 5 {
		t.Fatalf("WriteTo returned %d, %v; want 5, nil", n, err)
	}
	if got, want := output.String(), "abcde"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestSpoolBuffer_spillsAndRemovesFile(t *testing.T) {
	tests := []struct {
		name  string
		write func(io.Writer, string) (int, error)
	}{
		{
			name: "Write",
			write: func(w io.Writer, s string) (int, error) {
				return w.Write([]byte(s))
			},
		},
		{
			name:  "WriteString",
			write: io.WriteString,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const (
				limit  = 8
				prefix = "prefix"
			)
			input := strings.Repeat("x", 1<<20)
			buffer := newSpoolBuffer(limit)
			t.Cleanup(func() { _ = buffer.Close() })
			if n, err := buffer.WriteString(prefix); err != nil || n != len(prefix) {
				t.Fatalf("prefix write returned %d, %v; want %d, nil", n, err, len(prefix))
			}

			n, err := tt.write(buffer, input)
			if err != nil {
				t.Fatalf("write returned error: %v", err)
			}
			if n != len(input) {
				t.Fatalf("write returned %d, want %d", n, len(input))
			}
			if buffer.file == nil {
				t.Fatal("large output did not spill to a file")
			}
			path := buffer.file.Name()
			info, err := buffer.file.Stat()
			if err != nil {
				t.Fatalf("stat spill file: %v", err)
			}
			if runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
				t.Fatalf("spill file permissions = %v; group or other access is set", info.Mode().Perm())
			}

			var output bytes.Buffer
			want := prefix + input
			if written, err := buffer.WriteTo(&output); err != nil || written != int64(len(want)) {
				t.Fatalf("WriteTo returned %d, %v; want %d, nil", written, err, len(want))
			}
			if got := output.String(); got != want {
				t.Fatalf("replayed output length = %d, want %d", len(got), len(want))
			}

			if err := buffer.Close(); err != nil {
				t.Fatalf("Close returned error: %v", err)
			}
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatalf("spill file still exists after Close: %v", err)
			}
		})
	}
}

func TestSpoolBuffer_WriteToShortWriter(t *testing.T) {
	tests := []struct {
		name  string
		limit int
	}{
		{name: "in memory", limit: 16},
		{name: "spilled", limit: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := newSpoolBuffer(tt.limit)
			t.Cleanup(func() { _ = buffer.Close() })
			if n, err := buffer.WriteString("message"); err != nil || n != len("message") {
				t.Fatalf("WriteString returned %d, %v; want %d, nil", n, err, len("message"))
			}

			n, err := buffer.WriteTo(&shortWriter{remaining: 2})

			if n != 2 {
				t.Fatalf("WriteTo returned %d, want 2", n)
			}
			if !errors.Is(err, io.ErrShortWrite) {
				t.Fatalf("WriteTo returned error %v, want %v", err, io.ErrShortWrite)
			}
		})
	}
}

func TestSpoolBuffer_Reset(t *testing.T) {
	buffer := newSpoolBuffer(2)
	t.Cleanup(func() { _ = buffer.Close() })
	if _, err := buffer.WriteString("message"); err != nil {
		t.Fatalf("WriteString returned error: %v", err)
	}
	if buffer.file == nil {
		t.Fatal("output did not spill to disk")
	}
	path := buffer.file.Name()

	if err := buffer.Reset(); err != nil {
		t.Fatalf("Reset returned error: %v", err)
	}

	if buffer.Len() != 0 {
		t.Fatalf("length after Reset = %d, want 0", buffer.Len())
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("spill file still exists after Reset: %v", err)
	}
}

func TestSpoolBuffer_spillCreateError(t *testing.T) {
	missingTempDir := filepath.Join(t.TempDir(), "missing")
	setProcessTempDir(t, missingTempDir)

	tests := []struct {
		name  string
		write func(*spoolBuffer) (int, error)
	}{
		{
			name: "Write",
			write: func(buffer *spoolBuffer) (int, error) {
				return buffer.Write([]byte("d"))
			},
		},
		{
			name: "WriteString",
			write: func(buffer *spoolBuffer) (int, error) {
				return buffer.WriteString("d")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := newSpoolBuffer(3)
			if n, err := buffer.WriteString("abc"); err != nil || n != 3 {
				t.Fatalf("prefix write returned %d, %v; want 3, nil", n, err)
			}

			n, err := tt.write(buffer)

			if n != 0 {
				t.Fatalf("write returned %d, want 0", n)
			}
			if err == nil {
				t.Fatal("write returned nil error for an unavailable temporary directory")
			}
			if buffer.file != nil {
				t.Fatal("failed spill retained a file")
			}
			if got := buffer.buffer.String(); got != "abc" {
				t.Fatalf("buffered output = %q, want %q", got, "abc")
			}
			if got := buffer.Len(); got != 3 {
				t.Fatalf("length = %d, want 3", got)
			}
		})
	}
}

func TestSpoolBuffer_spillWriteError(t *testing.T) {
	buffer := newSpoolBuffer(3)
	t.Cleanup(func() { _ = buffer.Close() })
	if n, err := buffer.WriteString("abc"); err != nil || n != 3 {
		t.Fatalf("prefix write returned %d, %v; want 3, nil", n, err)
	}

	file, err := os.CreateTemp(t.TempDir(), outputSpillPattern)
	if err != nil {
		t.Fatalf("create spill file: %v", err)
	}
	path := file.Name()
	t.Cleanup(func() { _ = os.Remove(path) })
	if closeErr := file.Close(); closeErr != nil {
		t.Fatalf("close spill file: %v", closeErr)
	}

	err = buffer.spillTo(file)

	if err == nil {
		t.Fatal("spillTo returned nil error for a closed spill file")
	}
	if buffer.file != nil {
		t.Fatal("failed spill retained a file")
	}
	if got := buffer.buffer.String(); got != "abc" {
		t.Fatalf("buffered output = %q, want %q", got, "abc")
	}
	if got := buffer.Len(); got != 3 {
		t.Fatalf("length = %d, want 3", got)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("spill file still exists after failed write: %v", statErr)
	}

	if n, err := buffer.WriteString("d"); err != nil || n != 1 {
		t.Fatalf("retry returned %d, %v; want 1, nil", n, err)
	}
	var output bytes.Buffer
	if n, err := buffer.WriteTo(&output); err != nil || n != 4 {
		t.Fatalf("WriteTo after retry returned %d, %v; want 4, nil", n, err)
	}
	if got, want := output.String(), "abcd"; got != want {
		t.Fatalf("output after retry = %q, want %q", got, want)
	}
}

func TestSpoolBuffer_WriteToClosedFile(t *testing.T) {
	buffer := newSpoolBuffer(0)
	t.Cleanup(func() { _ = buffer.Close() })
	if n, err := buffer.WriteString("message"); err != nil || n != len("message") {
		t.Fatalf("WriteString returned %d, %v; want %d, nil", n, err, len("message"))
	}
	if err := buffer.file.Close(); err != nil {
		t.Fatalf("closing spill file: %v", err)
	}

	n, err := buffer.WriteTo(io.Discard)

	if n != 0 {
		t.Fatalf("WriteTo returned %d, want 0", n)
	}
	if err == nil {
		t.Fatal("WriteTo returned nil error for a closed spill file")
	}
}

func TestSpoolBuffer_CloseClosedFile(t *testing.T) {
	buffer := newSpoolBuffer(0)
	t.Cleanup(func() { _ = buffer.Close() })
	if n, err := buffer.WriteString("message"); err != nil || n != len("message") {
		t.Fatalf("WriteString returned %d, %v; want %d, nil", n, err, len("message"))
	}
	path := buffer.file.Name()
	if err := buffer.file.Close(); err != nil {
		t.Fatalf("closing spill file: %v", err)
	}

	err := buffer.Close()

	if err == nil {
		t.Fatal("Close returned nil error for an already closed spill file")
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("spill file still exists after Close: %v", statErr)
	}
}
