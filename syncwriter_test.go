package goyek_test

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

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

func TestSyncWriter_ReadFrom(t *testing.T) {
	buf := &strings.Builder{}
	w := goyek.SyncWriter(buf)
	readerFrom, ok := w.(io.ReaderFrom)
	if !ok {
		t.Fatal("SyncWriter result does not implement io.ReaderFrom")
	}

	n, err := readerFrom.ReadFrom(strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("ReadFrom error: %v", err)
	}
	if n != int64(len("hello")) {
		t.Fatalf("ReadFrom returned %d, want %d", n, len("hello"))
	}
	if got := buf.String(); got != "hello" {
		t.Fatalf("output = %q, want %q", got, "hello")
	}
}

func TestSyncWriter_ReadFromIsAtomic(t *testing.T) {
	buf := &strings.Builder{}
	w := goyek.SyncWriter(buf)
	readerFrom := w.(io.ReaderFrom)
	reader := &blockingReader{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}

	readDone := make(chan error, 1)
	go func() {
		_, err := readerFrom.ReadFrom(reader)
		readDone <- err
	}()
	<-reader.started

	writeStarted := make(chan struct{})
	writeDone := make(chan error, 1)
	go func() {
		close(writeStarted)
		_, err := io.WriteString(w, "other")
		writeDone <- err
	}()
	<-writeStarted
	select {
	case err := <-writeDone:
		close(reader.release)
		<-readDone
		t.Fatalf("concurrent WriteString completed during ReadFrom: %v", err)
	case <-time.After(20 * time.Millisecond):
	}

	close(reader.release)
	if err := <-readDone; err != nil {
		t.Fatalf("ReadFrom returned error: %v", err)
	}
	if err := <-writeDone; err != nil {
		t.Fatalf("WriteString returned error: %v", err)
	}
	if got, want := buf.String(), "firstsecondother"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

type blockingReader struct {
	started chan struct{}
	release chan struct{}
	reads   int
}

func (r *blockingReader) Read(p []byte) (int, error) {
	switch r.reads {
	case 0:
		r.reads++
		close(r.started)
		return copy(p, "first"), nil
	case 1:
		r.reads++
		<-r.release
		return copy(p, "second"), nil
	default:
		return 0, io.EOF
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

func TestSyncWriter_discard(t *testing.T) {
	if got := goyek.SyncWriter(io.Discard); got != io.Discard {
		t.Fatalf("SyncWriter(io.Discard) = %T, want io.Discard", got)
	}
}

func TestSyncWriter_uncomparableWriter(t *testing.T) {
	output := uncomparableWriter(make([]byte, len("message")))
	w := goyek.SyncWriter(output)

	if n, err := w.Write([]byte("message")); err != nil || n != len("message") {
		t.Fatalf("Write returned %d, %v; want %d, nil", n, err, len("message"))
	}
}

type uncomparableWriter []byte

func (w uncomparableWriter) Write(p []byte) (int, error) {
	copy(w, p)
	return len(p), nil
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
