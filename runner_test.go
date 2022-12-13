package goyek_test

import (
	"reflect"
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

			if got := r(goyek.Input{}); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got = %#v\nwant = %#v", got, tc.want)
			}
		})
	}
}

func TestRunnerPanic(t *testing.T) {
	payload := "panicked"
	r := goyek.NewRunner(func(a *goyek.A) { panic(payload) })

	got := r(goyek.Input{})

	if got, want := got.Status, goyek.StatusFailed; got != want {
		msg := "bad status"
		t.Errorf("%s\ngot = %q\nwant = %q", msg, got, want)
	}
	if got, want := got.PanicValue, payload; got != want {
		msg := "wrong panic value"
		t.Errorf("%s\ngot = %q\nwant = %q", msg, got, want)
	}
}
