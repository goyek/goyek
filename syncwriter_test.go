package goyek_test

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/goyek/goyek/v3"
)

func TestSyncWriter_Write_race(t *testing.T) {
	sb := &strings.Builder{}
	sw := goyek.Sync(sb)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			fmt.Fprintf(sw, "log from goroutine %d\n", i)
		}(i)
	}
	wg.Wait()

	got := sb.String()
	for i := 0; i < 100; i++ {
		msg := fmt.Sprintf("log from goroutine %d\n", i)
		if !strings.Contains(got, msg) {
			t.Errorf("missing log message: %q", msg)
		}
	}
}

func TestSyncWriter_WriteString_race(t *testing.T) {
	sb := &strings.Builder{}
	sw := goyek.Sync(sb)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sw.WriteString(fmt.Sprintf("log from goroutine %d\n", i)) //nolint:errcheck // not checking errors when writing to output
		}(i)
	}
	wg.Wait()

	got := sb.String()
	for i := 0; i < 100; i++ {
		msg := fmt.Sprintf("log from goroutine %d\n", i)
		if !strings.Contains(got, msg) {
			t.Errorf("missing log message: %q", msg)
		}
	}
}

func TestSyncWriter_WriteString_fallback(t *testing.T) {
	// A writer that doesn't implement io.StringWriter
	mw := &mockWriter{}
	sw := goyek.Sync(mw)

	n, err := sw.WriteString("hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5 bytes written, got %d", n)
	}
	if mw.sb.String() != "hello" {
		t.Errorf("expected \"hello\", got %q", mw.sb.String())
	}
}

type mockWriter struct {
	sb  strings.Builder
	err error
}

func (m *mockWriter) Write(p []byte) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.sb.Write(p)
}

func TestSyncWriter_Write_Error(t *testing.T) {
	mw := &mockWriter{err: fmt.Errorf("error")}
	sw := goyek.Sync(mw)

	_, err := sw.Write([]byte("hello"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSyncWriter_WriteString_Error(t *testing.T) {
	mw := &mockWriter{err: fmt.Errorf("error")}
	sw := goyek.Sync(mw)

	_, err := sw.WriteString("hello")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSyncWriter_WriteString_StringWriter_Error(t *testing.T) {
	mw := &mockStringWriter{err: fmt.Errorf("error")}
	sw := goyek.Sync(mw)

	_, err := sw.WriteString("hello")
	if err == nil {
		t.Fatal("expected error")
	}
}

type mockStringWriter struct {
	err error
}

func (m *mockStringWriter) Write(p []byte) (int, error) {
	return 0, nil
}

func (m *mockStringWriter) WriteString(s string) (int, error) {
	return 0, m.err
}

func TestSync_Reuse(t *testing.T) {
	sw1 := goyek.Sync(&strings.Builder{})
	sw2 := goyek.Sync(sw1)

	if sw1 != sw2 {
		t.Error("Sync should return the same SyncWriter if already a SyncWriter")
	}
}

func TestSync_Nil(t *testing.T) {
	sw := goyek.Sync(nil)
	if sw.Writer != io.Discard {
		t.Error("Sync(nil) should use io.Discard")
	}
}
