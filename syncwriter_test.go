package goyek_test

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/goyek/goyek/v3"
)

func TestSyncWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	w := goyek.SyncWriter(buf)

	n, err := w.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}
	if n != len("hello") {
		t.Fatalf("Write returned %d, want %d", n, len("hello"))
	}
	if got := buf.String(); got != "hello" {
		t.Fatalf("output = %q, want %q", got, "hello")
	}
}

func TestSyncWriter_WriteString(t *testing.T) {
	buf := &strings.Builder{}
	w := goyek.SyncWriter(buf)
	stringWriter, ok := w.(io.StringWriter)
	if !ok {
		t.Fatal("SyncWriter result does not implement io.StringWriter")
	}

	n, err := stringWriter.WriteString("hello")
	if err != nil {
		t.Fatalf("WriteString error: %v", err)
	}
	if n != len("hello") {
		t.Fatalf("WriteString returned %d, want %d", n, len("hello"))
	}
	if got := buf.String(); got != "hello" {
		t.Fatalf("output = %q, want %q", got, "hello")
	}
}

func TestSyncWriter_WriteStringFallback(t *testing.T) {
	buf := &bytes.Buffer{}
	w := goyek.SyncWriter(writeOnly{Writer: buf})
	stringWriter := w.(io.StringWriter)

	n, err := stringWriter.WriteString("hello")
	if err != nil {
		t.Fatalf("WriteString error: %v", err)
	}
	if n != len("hello") {
		t.Fatalf("WriteString returned %d, want %d", n, len("hello"))
	}
	if got := buf.String(); got != "hello" {
		t.Fatalf("output = %q, want %q", got, "hello")
	}
}

type writeOnly struct {
	io.Writer
}

func TestSyncWriter_concurrentWriteAndWriteString(t *testing.T) {
	const goroutines = 10
	const message = "message"

	buf := &strings.Builder{}
	w := goyek.SyncWriter(buf)
	stringWriter := w.(io.StringWriter)
	start := make(chan struct{})

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(useWriteString bool) {
			defer wg.Done()
			<-start
			if useWriteString {
				_, _ = stringWriter.WriteString(message)
			} else {
				_, _ = w.Write([]byte(message))
			}
		}(i%2 == 0)
	}

	close(start)
	wg.Wait()

	if got, want := strings.Count(buf.String(), message), goroutines; got != want {
		t.Fatalf("got %d occurrences, want %d", got, want)
	}
}

func TestSyncWriter_idempotent(t *testing.T) {
	w := goyek.SyncWriter(&bytes.Buffer{})

	if got := goyek.SyncWriter(w); got != w {
		t.Fatal("wrapping a SyncWriter result changed the writer")
	}
}

func TestSyncWriter_nil(t *testing.T) {
	if got := goyek.SyncWriter(nil); got != nil {
		t.Fatalf("SyncWriter(nil) = %T, want nil", got)
	}
}

func ExampleSyncWriter() {
	var output strings.Builder
	out := goyek.SyncWriter(&output)

	var wg sync.WaitGroup
	for _, message := range []string{"first\n", "second\n"} {
		wg.Add(1)
		go func(message string) {
			defer wg.Done()
			_, _ = io.WriteString(out, message)
		}(message)
	}
	wg.Wait()

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	sort.Strings(lines)
	fmt.Println(lines)
	// Output: [first second]
}
