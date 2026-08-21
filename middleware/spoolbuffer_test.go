package middleware

import (
	"bytes"
	"errors"
	"io"
	"os"
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
