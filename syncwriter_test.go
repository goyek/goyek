package goyek_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/goyek/goyek/v3"
)

func TestSyncWriter_Write(t *testing.T) {
	const msg = "hello"
	buf := &bytes.Buffer{}
	w := &goyek.SyncWriter{Writer: buf}

	n, err := w.Write([]byte(msg))
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Errorf("got %d, want 5", n)
	}
	if buf.String() != msg {
		t.Errorf("got %q, want %q", buf.String(), msg)
	}
}

func TestSyncWriter_WriteString(t *testing.T) {
	const msg = "hello"
	t.Run("StringWriter", func(t *testing.T) {
		sb := &strings.Builder{}
		w := &goyek.SyncWriter{Writer: sb}

		n, err := w.WriteString(msg)
		if err != nil {
			t.Fatal(err)
		}
		if n != 5 {
			t.Errorf("got %d, want 5", n)
		}
		if sb.String() != msg {
			t.Errorf("got %q, want %q", sb.String(), msg)
		}
	})

	t.Run("io.Writer", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := &goyek.SyncWriter{Writer: struct{ io.Writer }{buf}}

		n, err := w.WriteString(msg)
		if err != nil {
			t.Fatal(err)
		}
		if n != 5 {
			t.Errorf("got %d, want 5", n)
		}
		if buf.String() != msg {
			t.Errorf("got %q, want %q", buf.String(), msg)
		}
	})
}
