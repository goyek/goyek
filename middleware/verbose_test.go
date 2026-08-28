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
	const goroutines = 5
	const message = "msg "

	r := func(in goyek.Input) goyek.Result {
		var wg sync.WaitGroup
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				io.WriteString(in.Output, message) //nolint:errcheck // not checking errors when writing to output
			}()
		}
		wg.Wait()
		return goyek.Result{Status: goyek.StatusFailed}
	}
	r = middleware.SilentNonFailed(r)

	sb := &strings.Builder{}
	r(goyek.Input{Output: goyek.SyncWriter(sb)})

	if got, want := strings.Count(sb.String(), strings.TrimSpace(message)), goroutines; got != want {
		t.Fatalf("got %d occurrences, want %d", got, want)
	}
}

func TestSilentNonFailed_nilOutput(t *testing.T) {
	runner := middleware.SilentNonFailed(func(in goyek.Input) goyek.Result {
		message := strings.Repeat("x", (1<<20)+1)
		n, err := io.WriteString(in.Output, message)
		if err != nil {
			t.Fatalf("writing output: %v", err)
		}
		if n != len(message) {
			t.Fatalf("wrote %d bytes, want %d", n, len(message))
		}
		return goyek.Result{Status: goyek.StatusFailed}
	})

	result := runner(goyek.Input{})

	if result.Status != goyek.StatusFailed {
		t.Fatalf("got status %v, want %v", result.Status, goyek.StatusFailed)
	}
}

func TestSilentNonFailed_discardOutput(t *testing.T) {
	runner := middleware.SilentNonFailed(func(in goyek.Input) goyek.Result {
		if in.Output != io.Discard {
			t.Fatalf("output = %T, want io.Discard", in.Output)
		}
		message := strings.Repeat("x", (1<<20)+1)
		n, err := io.WriteString(in.Output, message)
		if err != nil {
			t.Fatalf("writing output: %v", err)
		}
		if n != len(message) {
			t.Fatalf("wrote %d bytes, want %d", n, len(message))
		}
		return goyek.Result{Status: goyek.StatusFailed}
	})

	result := runner(goyek.Input{Output: goyek.SyncWriter(io.Discard)})

	if result.Status != goyek.StatusFailed {
		t.Fatalf("got status %v, want %v", result.Status, goyek.StatusFailed)
	}
}

func TestSilentNonFailed_preservesLargeFailedOutput(t *testing.T) {
	const outputLimit = 1 << 20
	message := strings.Repeat("x", outputLimit+1)
	runner := middleware.SilentNonFailed(func(in goyek.Input) goyek.Result {
		n, err := io.WriteString(in.Output, message)
		if err != nil {
			t.Fatalf("writing output: %v", err)
		}
		if n != len(message) {
			t.Fatalf("wrote %d bytes, want %d", n, len(message))
		}
		return goyek.Result{Status: goyek.StatusFailed}
	})

	out := &strings.Builder{}
	result := runner(goyek.Input{Output: goyek.SyncWriter(out)})

	if result.Status != goyek.StatusFailed {
		t.Fatalf("got status %v, want %v", result.Status, goyek.StatusFailed)
	}
	if got := out.String(); got != message {
		t.Fatalf("output length = %d, want %d", len(got), len(message))
	}
}

func TestSilentNonFailed_uncomparableOutput(t *testing.T) {
	runner := middleware.SilentNonFailed(func(in goyek.Input) goyek.Result {
		_, _ = io.WriteString(in.Output, "message")
		return goyek.Result{Status: goyek.StatusFailed}
	})

	result := runner(goyek.Input{Output: uncomparableWriter(make([]byte, 64))})

	if result.Status != goyek.StatusFailed {
		t.Fatalf("got status %v, want %v", result.Status, goyek.StatusFailed)
	}
}
