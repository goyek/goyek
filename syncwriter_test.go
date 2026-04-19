package goyek

import (
	"bytes"
	"io"
	"testing"
)

const hello = "hello"

func TestSyncWriter_Write(t *testing.T) {
	buf := &bytes.Buffer{}
	sw := &SyncWriter{Writer: buf}
	n, err := sw.Write([]byte(hello))
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Errorf("expected 5 bytes, got %d", n)
	}
	if buf.String() != hello {
		t.Errorf("expected %s, got %q", hello, buf.String())
	}
}

func TestSyncWriter_WriteString(t *testing.T) {
	t.Run("io.StringWriter", func(t *testing.T) {
		buf := &bytes.Buffer{}
		sw := &SyncWriter{Writer: buf}
		n, err := sw.WriteString(hello)
		if err != nil {
			t.Fatal(err)
		}
		if n != 5 {
			t.Errorf("expected 5 bytes, got %d", n)
		}
		if buf.String() != hello {
			t.Errorf("expected %s, got %q", hello, buf.String())
		}
	})

	t.Run("io.Writer only", func(t *testing.T) {
		buf := &bytes.Buffer{}
		// Wrap to hide WriteString method
		sw := &SyncWriter{Writer: struct{ io.Writer }{buf}}
		n, err := sw.WriteString(hello)
		if err != nil {
			t.Fatal(err)
		}
		if n != 5 {
			t.Errorf("expected 5 bytes, got %d", n)
		}
		if buf.String() != hello {
			t.Errorf("expected %s, got %q", hello, buf.String())
		}
	})
}
