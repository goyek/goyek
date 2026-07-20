package goyek_test

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/goyek/goyek/v3"
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
			action: func(*goyek.A) {},
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

			assertEqual(t, got, tc.want, "should return proper result")
		})
	}
}

func TestRunner_panic(t *testing.T) {
	payload := "panicked"
	r := goyek.NewRunner(func(*goyek.A) { panic(payload) })

	got := r(goyek.Input{})

	assertEqual(t, got.Status, goyek.StatusFailed, "should return proper status")
	assertEqual(t, got.PanicValue, payload, "should return proper panic value")
}

func TestNewRunner_concurrent_printing(t *testing.T) {
	const (
		goroutines         = 16
		writesPerGoroutine = 25
	)
	start := make(chan struct{})
	runner := goyek.NewRunner(func(a *goyek.A) {
		var wg sync.WaitGroup
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				<-start
				for j := 0; j < writesPerGoroutine; j++ {
					a.Logf("message-%02d-%02d", id, j)
				}
			}(i)
		}
		close(start)
		wg.Wait()
	})

	out := &strings.Builder{}
	gotResult := runner(goyek.Input{Output: out})
	assertEqual(t, gotResult.Status, goyek.StatusPassed, "should return proper status")

	got := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
	want := make([]string, 0, goroutines*writesPerGoroutine)
	for i := 0; i < goroutines; i++ {
		for j := 0; j < writesPerGoroutine; j++ {
			want = append(want, fmt.Sprintf("message-%02d-%02d", i, j))
		}
	}
	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("concurrent output mismatch\ngot:  %q\nwant: %q", got, want)
	}
}
