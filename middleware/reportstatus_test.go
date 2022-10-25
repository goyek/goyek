package middleware_test

import (
	"strings"
	"testing"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/goyek/v2/middleware"
)

func TestReportStatus(t *testing.T) {
	taskName := "my-task"
	tests := []struct {
		name   string
		status goyek.Status
		want   string
	}{
		{
			name:   "Passed",
			status: goyek.StatusPassed,
			want:   "PASS: " + taskName,
		},
		{
			name:   "Failed",
			status: goyek.StatusFailed,
			want:   "FAIL: " + taskName,
		},
		{
			name:   "Skipped",
			status: goyek.StatusSkipped,
			want:   "SKIP: " + taskName,
		},
		{
			name:   "NotRun",
			status: goyek.StatusNotRun,
			want:   "NOOP: " + taskName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := &strings.Builder{}
			r := goyek.Runner(func(i goyek.Input) goyek.Result { return goyek.Result{Status: tt.status} })
			r = middleware.ReportStatus(r)

			r(goyek.Input{TaskName: taskName, Output: sb})

			if !strings.Contains(sb.String(), tt.want) {
				t.Errorf("got: %q; but should contain: %q", sb.String(), tt.want)
			}
		})
	}

	panicTests := []struct {
		name         string
		panicPayload interface{}
		want         string
	}{
		{
			name:         "Panic",
			panicPayload: "crashed",
			want:         "panic: crashed",
		},
		{
			name:         "NilPanic",
			panicPayload: nil,
			want:         "panic(nil) or runtime.Goexit() called",
		},
	}
	for _, tt := range panicTests {
		t.Run(tt.name, func(t *testing.T) {
			sb := &strings.Builder{}
			r := goyek.Runner(func(i goyek.Input) goyek.Result {
				return goyek.Result{PanicStack: []byte("stacktrace"), PanicValue: tt.panicPayload}
			})
			r = middleware.ReportStatus(r)

			r(goyek.Input{TaskName: taskName, Output: sb})

			if !strings.Contains(sb.String(), tt.want) {
				t.Errorf("got: %q; but should contain: %q", sb.String(), tt.want)
			}

			if !strings.Contains(sb.String(), "stacktrace") {
				t.Errorf("got: %q; but should contain: \"stacktrace\"", sb.String())
			}
		})
	}
}
