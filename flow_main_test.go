package goyek

import (
	"context"
	"errors"
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
				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
				defer cancel()
				return flow.main(ctx, []string{"task"})
			},
		},
		{
			desc: "canceled err",
			want: 1,
			act: func() int {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				// Execute will return context.Canceled
				return flow.main(ctx, []string{"task"})
			},
		},
		{
			desc: "Execute returns nil but ctx is canceled",
			want: 1,
			act: func() int {
				ctx, cancel := context.WithCancel(context.Background())
				f := &Flow{}
				f.Define(Task{
					Name: "cancel",
					Action: func(_ *A) {
						cancel()
					},
				})
				return f.main(ctx, []string{"cancel"})
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
	if got := err.Error(); got != "task failed: task" {
		t.Errorf("got: %q; want: \"task failed: task\"", got)
	}
}

func TestExecute_package_level(t *testing.T) {
	if err := Execute(context.Background(), nil); err == nil {
		t.Error("should fail for no tasks")
	}
}

func TestExecute_deadline_exceeded(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	f := &Flow{}
	f.Define(Task{Name: "task"})
	err := f.Execute(ctx, []string{"task"})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("got: %v; want: context deadline exceeded", err)
	}
}

func TestMain_package_level(_ *testing.T) {
	origOsExit := osExit
	defer func() { osExit = origOsExit }()
	osExit = func(int) {}

	DefaultFlow.SetOutput(io.Discard)
	DefaultFlow.tasks = nil
	DefaultFlow.Define(Task{Name: "task"})

	Main([]string{"task"})
}

func TestTasks_package_level(_ *testing.T) {
	Tasks()
}

func TestPrint_package_level(_ *testing.T) {
	Print()
}
