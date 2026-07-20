package internal_test

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
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
}

func TestSyncWriter_ConcurrentWriteAndWriteString(t *testing.T) {
	const (
		goroutines         = 16
		writesPerGoroutine = 25
	)

	buf := &strings.Builder{}
	sw := internal.SyncWriter(buf)
	stringWriter := sw.(io.StringWriter)
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
					n, err = sw.Write([]byte(line))
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

func TestSyncWriter_DoubleWrap(t *testing.T) {
	buf := &bytes.Buffer{}
	sw1 := internal.SyncWriter(buf)
	sw2 := internal.SyncWriter(sw1)

	if sw1 != sw2 {
		t.Error("SyncWriter should not double wrap")
	}
}
