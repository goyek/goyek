package goyek_test

import (
	"testing"

	"github.com/goyek/goyek/v2"
)

func TestRunner(t *testing.T) {
	testCases := []struct {
		desc   string
		want   goyek.Result
		action func(*goyek.TF)
	}{
		{
			desc:   "pass",
			want:   goyek.Result{Status: goyek.StatusPassed},
			action: func(tf *goyek.TF) {},
		},
		{
			desc:   "fail",
			want:   goyek.Result{Status: goyek.StatusFailed},
			action: func(tf *goyek.TF) { tf.Fail() },
		},
		{
			desc:   "failnow",
			want:   goyek.Result{Status: goyek.StatusFailed},
			action: func(tf *goyek.TF) { tf.FailNow() },
		},
		{
			desc:   "skip",
			want:   goyek.Result{Status: goyek.StatusSkipped},
			action: func(tf *goyek.TF) { tf.Skip() },
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
	r := goyek.NewRunner(func(tf *goyek.TF) { panic(payload) })

	got := r(goyek.Input{})

	assertEqual(t, got.Status, goyek.StatusFailed, "shoud return proper status")
	assertEqual(t, got.PanicValue, payload, "shoud return proper panic value")
}
