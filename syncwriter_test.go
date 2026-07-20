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

type writeOnly struct {
	io.Writer
}

func TestSyncWriter_WriteAndWriteStringDoNotOverlap(t *testing.T) {
	underlying := &blockingStringWriter{
		writeStarted:       make(chan struct{}),
		releaseWrite:       make(chan struct{}),
		writeStringEntered: make(chan struct{}),
	}
	w := goyek.SyncWriter(underlying)
	stringWriter := w.(io.StringWriter)

	writeDone := make(chan error, 1)
	go func() {
		_, err := w.Write([]byte("write"))
		writeDone <- err
	}()
	select {
	case <-underlying.writeStarted:
	case <-time.After(time.Second):
		t.Fatal("Write did not reach the underlying writer")
	}

	writeStringAttempted := make(chan struct{})
	writeStringDone := make(chan error, 1)
	go func() {
		close(writeStringAttempted)
		_, err := stringWriter.WriteString("string")
		writeStringDone <- err
	}()
	<-writeStringAttempted

	overlapped := false
	select {
	case <-underlying.writeStringEntered:
		overlapped = true
	case <-time.After(50 * time.Millisecond):
	}
	close(underlying.releaseWrite)

	if err := <-writeDone; err != nil {
		t.Fatalf("Write error: %v", err)
	}
	if err := <-writeStringDone; err != nil {
		t.Fatalf("WriteString error: %v", err)
	}
	if overlapped {
		t.Fatal("WriteString reached the underlying writer before Write returned")
	}
}

type blockingStringWriter struct {
	writeStarted       chan struct{}
	releaseWrite       chan struct{}
	writeStringEntered chan struct{}
}

func (w *blockingStringWriter) Write(p []byte) (int, error) {
	close(w.writeStarted)
	<-w.releaseWrite
	return len(p), nil
}

func (w *blockingStringWriter) WriteString(s string) (int, error) {
	close(w.writeStringEntered)
	return len(s), nil
}

func TestSyncWriter_concurrentWriteAndWriteString(t *testing.T) {
	const (
		goroutines         = 16
		writesPerGoroutine = 25
	)

	buf := &strings.Builder{}
	w := goyek.SyncWriter(buf)
	stringWriter := w.(io.StringWriter)
	start := make(chan struct{})
	errs := make(chan error, goroutines*writesPerGoroutine)
	want := make([]string, 0, goroutines*writesPerGoroutine)

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		messages := make([]string, 0, writesPerGoroutine)
		for j := 0; j < writesPerGoroutine; j++ {
			message := fmt.Sprintf("message-%02d-%02d", i, j)
			messages = append(messages, message)
			want = append(want, message)
		}

		wg.Add(1)
		go func(useWriteString bool, messages []string) {
			defer wg.Done()
			<-start
			for _, message := range messages {
				line := message + "\n"
				var n int
				var err error
				if useWriteString {
					n, err = stringWriter.WriteString(line)
				} else {
					n, err = w.Write([]byte(line))
				}
				if err != nil {
					errs <- err
				} else if n != len(line) {
					errs <- fmt.Errorf("wrote %d bytes, want %d", n, len(line))
				}
			}
		}(i%2 == 0, messages)
	}

	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Errorf("write error: %v", err)
	}

	got := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("concurrent output mismatch\ngot:  %q\nwant: %q", got, want)
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
