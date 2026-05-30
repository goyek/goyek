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
	got := err.Error()
	want := "task failed: task"
	if got != want {
		t.Errorf("got: %q; want: %q", got, want)
	}
}

func TestWrappers(t *testing.T) {
	// Just call them to ensure coverage.
	// They use DefaultFlow, which we should probably reset or be careful with.
	Define(Task{Name: "wrapper-task"})
	Tasks()
	Output()
	SetOutput(io.Discard)
	SetLogger(GetLogger())
	SetUsage(Usage())
	SetDefault(Default())
	Undefine(Tasks()[0])
	Use(func(r Runner) Runner { return r })
	UseExecutor(func(e Executor) Executor { return e })
	Execute(context.Background(), nil)
	Print()
}

func Test_main_context_error(t *testing.T) {
	flow := &Flow{}
	flow.SetOutput(io.Discard)
	flow.Define(Task{Name: "task"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if got := flow.main(ctx, []string{"task"}); got != 1 {
		t.Errorf("got: %d; want: 1", got)
	}

	ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
	defer cancel()

	if got := flow.main(ctx, []string{"task"}); got != 1 {
		t.Errorf("got: %d; want: 1", got)
	}
}
