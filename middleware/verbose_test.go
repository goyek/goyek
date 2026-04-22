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

	r(goyek.Input{Output: sb})

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

			r(goyek.Input{Output: sb})

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
	r(goyek.Input{Output: sb})

	if got, want := strings.Count(sb.String(), strings.TrimSpace(message)), goroutines; got != want {
		t.Fatalf("got %d occurrences, want %d", got, want)
	}
}
