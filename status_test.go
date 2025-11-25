package goyek_test

import (
	"testing"

	"github.com/goyek/goyek/v3"
)

func TestStatus_String(t *testing.T) {
	testCases := []struct {
		name string
		s    goyek.Status
		want string
	}{
		{name: "NotRun", s: goyek.StatusNotRun, want: "NOOP"},
		{name: "Passed", s: goyek.StatusPassed, want: "PASS"},
		{name: "Failed", s: goyek.StatusFailed, want: "FAIL"},
		{name: "Skipped", s: goyek.StatusSkipped, want: "SKIP"},
		{name: "Other", s: goyek.Status(123), want: "goyek.Status(123)"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.s.String(); got != tc.want {
				t.Errorf("Status.String() = %v, want %v", got, tc.want)
			}
		})
	}
}
