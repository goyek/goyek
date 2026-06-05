package goyek

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"
)

func TestFlow_main(t *testing.T) {
	flow := &Flow{}
	flow.SetOutput(&strings.Builder{})
	flow.Define(Task{Name: "task"})
	flow.Define(Task{Name: "failing", Action: func(a *A) { a.Fail() }})

	testCases := []struct {
		desc string
		want int
		act  func() int
	}{
		{
			desc: "pass",
			want: 0,
			act:  func() int { return flow.main(context.Background(), []string{"task"}) },
		},
		{
			desc: "fail",
			want: 1,
			act:  func() int { return flow.main(context.Background(), []string{"failing"}) },
		},
		{
			desc: "invalid",
			want: 2,
			act:  func() int { return flow.main(context.Background(), []string{"bad"}) },
		},
		{
			desc: "canceled",
			want: 1,
			act: func() int {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return flow.main(ctx, []string{"task"})
			},
		},
		{
			desc: "deadline exceeded",
			want: 1,
			act: func() int {
				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
				defer cancel()
				return flow.main(ctx, []string{"task"})
			},
		},
		{
			desc: "interrupted",
			want: 1,
			act: func() int {
				flow := &Flow{}
				flow.SetOutput(io.Discard)
				flow.Define(Task{
					Name: "task",
					Action: func(a *A) {
						a.WithContext(context.Background()) // just to use some context
					},
				})
				ctx, cancel := context.WithCancel(context.Background())
				flow.Define(Task{
					Name: "interrupter",
					Action: func(a *A) {
						cancel()
					},
				})
				return flow.main(ctx, []string{"interrupter"})
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			if got := tc.act(); got != tc.want {
				t.Errorf("got: %d; want: %d", got, tc.want)
			}
		})
	}
}

func Test_main_usage(t *testing.T) {
	flow := &Flow{}
	flow.SetOutput(io.Discard)
	called := false
	flow.SetUsage(func() { called = true })

	flow.main(context.Background(), nil)

	if !called {
		t.Error("usage should be called for invalid input")
	}
}

func TestFailError_Error(t *testing.T) {
	err := &FailError{Task: "task"}
	want := "task failed: task"
	if got := err.Error(); got != want {
		t.Errorf("got: %q; want: %q", got, want)
	}
}
