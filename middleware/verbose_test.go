package middleware_test

import (
	"strings"
	"testing"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/goyek/v2/middleware"
)

func TestSilentNonFailedForFail(t *testing.T) {
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

func TestSilentNonFailedForNotFailed(t *testing.T) {
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
