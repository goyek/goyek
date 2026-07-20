package middleware_test

import (
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/goyek/goyek/v3"
	"github.com/goyek/goyek/v3/middleware"
)

func TestSilentNonFailed_failed(t *testing.T) {
	msg := "message"
	sb := &strings.Builder{}
	r := func(i goyek.Input) goyek.Result {
		i.Output.Write([]byte(msg)) //nolint:errcheck // not checking errors when writing to output
		return goyek.Result{Status: goyek.StatusFailed}
	}
	r = middleware.SilentNonFailed(r)

	r(goyek.Input{Output: goyek.SyncWriter(sb)})

	if !strings.Contains(sb.String(), msg) {
		t.Errorf("got: %q; but should contain: %q", sb.String(), msg)
	}
}

func TestSilentNonFailed_notFailed(t *testing.T) {
	tests := []struct {
		name   string
		status goyek.Status
	}{
		{
			name:   "Passed",
			status: goyek.StatusPassed,
		},
		{
			name:   "Skipped",
			status: goyek.StatusSkipped,
		},
		{
			name:   "NotRun",
			status: goyek.StatusNotRun,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := "message"
			sb := &strings.Builder{}
			r := func(i goyek.Input) goyek.Result {
				i.Output.Write([]byte(msg)) //nolint:errcheck // not checking errors when writing to output
				return goyek.Result{Status: tt.status}
			}
			r = middleware.SilentNonFailed(r)

			r(goyek.Input{Output: goyek.SyncWriter(sb)})

			if strings.Contains(sb.String(), msg) {
				t.Errorf("got: %q; but should not contain: %q", sb.String(), msg)
			}
		})
	}
}

func TestSilentNonFailed_concurrent_printing(t *testing.T) {
	const (
		goroutines         = 16
		writesPerGoroutine = 25
		message            = "msg "
	)

	r := func(in goyek.Input) goyek.Result {
		start := make(chan struct{})
		var wg sync.WaitGroup
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				for j := 0; j < writesPerGoroutine; j++ {
					io.WriteString(in.Output, message) //nolint:errcheck // not checking errors when writing to output
				}
			}()
		}
		close(start)
		wg.Wait()
		return goyek.Result{Status: goyek.StatusFailed}
	}
	r = middleware.SilentNonFailed(r)

	sb := &strings.Builder{}
	r(goyek.Input{Output: goyek.SyncWriter(sb)})

	if got, want := sb.String(), strings.Repeat(message, goroutines*writesPerGoroutine); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestSilentNonFailed_nilOutput(t *testing.T) {
	runner := middleware.SilentNonFailed(func(in goyek.Input) goyek.Result {
		_, _ = io.WriteString(in.Output, "discarded")
		return goyek.Result{Status: goyek.StatusFailed}
	})

	result := runner(goyek.Input{})

	if result.Status != goyek.StatusFailed {
		t.Fatalf("got status %v, want %v", result.Status, goyek.StatusFailed)
	}
}

func TestSilentNonFailed_flushesBufferedOutputInOneWrite(t *testing.T) {
	out := &recordingWriter{}
	runner := middleware.SilentNonFailed(func(in goyek.Input) goyek.Result {
		_, _ = io.WriteString(in.Output, "first")
		_, _ = io.WriteString(in.Output, "second")
		return goyek.Result{Status: goyek.StatusFailed}
	})

	result := runner(goyek.Input{Output: out})

	if result.Status != goyek.StatusFailed {
		t.Fatalf("got status %v, want %v", result.Status, goyek.StatusFailed)
	}
	writes := out.records()
	if len(writes) != 1 || writes[0] != "firstsecond" {
		t.Fatalf("writes = %q, want %q", writes, []string{"firstsecond"})
	}
}
