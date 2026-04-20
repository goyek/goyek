package middleware_test

import (
	"context"
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
	out := &strings.Builder{}
	flow := &goyek.Flow{}
	flow.SetOutput(out)
	flow.Use(middleware.SilentNonFailed)

	flow.Define(goyek.Task{
		Name: "task",
		Action: func(a *goyek.A) {
			var wg sync.WaitGroup
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					a.Log("some log message")
				}()
			}
			wg.Wait()
		},
	})

	err := flow.Execute(context.Background(), []string{"task"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if strings.Contains(out.String(), "some log message") {
		t.Errorf("should not output but got: %s", out.String())
	}
}
