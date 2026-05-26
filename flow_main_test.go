package goyek

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
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
	want := "task failed: task"
	if got := err.Error(); got != want {
		t.Errorf("got: %q; want: %q", got, want)
	}
}

func TestExecute(t *testing.T) {
	task := Define(Task{Name: "task-execute"})
	defer Undefine(task)

	if err := Execute(context.Background(), []string{"task-execute"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test failure returns FailError
	taskFail := Define(Task{Name: "task-fail", Action: func(a *A) { a.Fail() }})
	defer Undefine(taskFail)
	err := Execute(context.Background(), []string{"task-fail"})
	var ferr *FailError
	if !errors.As(err, &ferr) {
		t.Errorf("expected FailError, got %T", err)
	}
}

func TestMain(t *testing.T) {
	origOsExit := osExit
	defer func() { osExit = origOsExit }()
	osExit = func(code int) {}

	task := Define(Task{Name: "task-main"})
	defer Undefine(task)

	Main([]string{"task-main"})
}

func TestPrint(t *testing.T) {
	flow := &Flow{}
	flow.SetOutput(io.Discard)
	flow.Define(Task{Name: "task", Usage: "usage"})
	flow.Print()

	// Test package-level wrapper
	Print()
}

func TestTasks(t *testing.T) {
	flow := &Flow{}
	flow.Define(Task{Name: "b"})
	flow.Define(Task{Name: "a"})
	tasks := flow.Tasks()
	if len(tasks) != 2 || tasks[0].Name() != "a" || tasks[1].Name() != "b" {
		t.Errorf("unexpected tasks: %v", tasks)
	}

	// Test package-level wrapper
	_ = Tasks()
}

func TestDefine_panics(t *testing.T) {
	flow := &Flow{}
	t.Run("empty name", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("expected panic")
			}
		}()
		flow.Define(Task{Name: ""})
	})
	t.Run("duplicate name", func(t *testing.T) {
		flow.Define(Task{Name: "dup"})
		defer func() {
			if recover() == nil {
				t.Error("expected panic")
			}
		}()
		flow.Define(Task{Name: "dup"})
	})
	t.Run("bad dep", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("expected panic")
			}
		}()
		flow.Define(Task{Name: "bad", Deps: []*DefinedTask{{name: "missing"}}})
	})
}

func TestFlow_main_deadline(t *testing.T) {
	flow := &Flow{}
	flow.SetOutput(io.Discard)
	flow.Define(Task{Name: "task"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel the context

	if got := flow.main(ctx, []string{"task"}); got != 1 {
		t.Errorf("got: %d; want: 1", got)
	}
}
