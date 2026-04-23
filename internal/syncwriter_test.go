package internal_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/goyek/goyek/v3/internal"
)

const hello = "hello"

func TestSyncWriter_Write(t *testing.T) {
	buf := &bytes.Buffer{}
	sw := internal.SyncWriter(buf)
	p := []byte(hello)
	n, err := sw.Write(p)
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}
	if n != len(p) {
		t.Fatalf("got %d, want %d", n, len(p))
	}
	if buf.String() != hello {
		t.Fatalf("got %q, want %q", buf.String(), hello)
	}
}

func TestSyncWriter_WriteString(t *testing.T) {
	t.Run("StringWriter", func(t *testing.T) {
		sb := &strings.Builder{}
		sw := internal.SyncWriter(sb).(io.StringWriter)
		n, err := sw.WriteString(hello)
		if err != nil {
			t.Fatalf("WriteString error: %v", err)
		}
		if n != len(hello) {
			t.Fatalf("got %d, want %d", n, len(hello))
		}
		if sb.String() != hello {
			t.Fatalf("got %q, want %q", sb.String(), hello)
		}
	})

}

func TestSyncWriter_DoubleWrap(t *testing.T) {
	buf := &bytes.Buffer{}
	sw1 := internal.SyncWriter(buf)
	sw2 := internal.SyncWriter(sw1)

	if sw1 != sw2 {
		t.Error("SyncWriter should not double wrap")
	}
}
