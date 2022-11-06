package goyek_test

import (
	"testing"

	"github.com/goyek/goyek/v2"
)

func TestRunner(t *testing.T) {
	testCases := []struct {
		desc   string
		want   goyek.Result
		action func(*goyek.A)
	}{
		{
			desc:   "pass",
			want:   goyek.Result{Status: goyek.StatusPassed},
			action: func(a *goyek.A) {},
		},
		{
			desc:   "fail",
			want:   goyek.Result{Status: goyek.StatusFailed},
			action: func(a *goyek.A) { a.Fail() },
		},
		{
			desc:   "failnow",
			want:   goyek.Result{Status: goyek.StatusFailed},
			action: func(a *goyek.A) { a.FailNow() },
		},
		{
			desc:   "skip",
			want:   goyek.Result{Status: goyek.StatusSkipped},
			action: func(a *goyek.A) { a.Skip() },
		},
		{
			desc:   "nil",
			want:   goyek.Result{Status: goyek.StatusNotRun},
			action: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			r := goyek.NewRunner(tc.action)
			got := r(goyek.Input{})

			assertEqual(t, got, tc.want, "shoud return proper result")
		})
	}
}

func TestRunner_panic(t *testing.T) {
	payload := "panicked"
	r := goyek.NewRunner(func(a *goyek.A) { panic(payload) })

	got := r(goyek.Input{})

	assertEqual(t, got.Status, goyek.StatusFailed, "shoud return proper status")
	assertEqual(t, got.PanicValue, payload, "shoud return proper panic value")
}
