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

func Test_main_deadline(t *testing.T) {
	flow := &Flow{}
	flow.SetOutput(io.Discard)
	flow.Define(Task{Name: "task", Action: func(a *A) {
		<-a.Context().Done()
	}})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	got := flow.main(ctx, []string{"task"})
	if got != 1 {
		t.Errorf("expected exit code 1 for deadline exceeded, got %d", got)
	}
}

func TestFailError_Error(t *testing.T) {
	err := &FailError{Task: "my-task"}
	got := err.Error()
	want := "task failed: my-task"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExecute_package(t *testing.T) {
	orig := DefaultFlow
	defer func() { DefaultFlow = orig }()
	DefaultFlow = &Flow{}
	Define(Task{Name: "task"})
	err := Execute(context.Background(), []string{"task"})
	if err != nil {
		t.Errorf("expected nil err, got %v", err)
	}
}

func TestMain_package(t *testing.T) {
	orig := DefaultFlow
	defer func() { DefaultFlow = orig }()
	DefaultFlow = &Flow{}
	Define(Task{Name: "task"})

	restoreOsExit := osExit
	defer func() { osExit = restoreOsExit }()
	osExit = func(code int) {}

	Main([]string{"task"})
}

func TestPrint_package(t *testing.T) {
	orig := DefaultFlow
	defer func() { DefaultFlow = orig }()
	DefaultFlow = &Flow{}
	Define(Task{Name: "task", Usage: "usage"})
	Print()
}

func TestTasks_package(t *testing.T) {
	orig := DefaultFlow
	defer func() { DefaultFlow = orig }()
	DefaultFlow = &Flow{}
	Define(Task{Name: "task"})
	tasks := Tasks()
	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}
}

func TestOutput_package(t *testing.T) {
	orig := DefaultFlow
	defer func() { DefaultFlow = orig }()
	DefaultFlow = &Flow{}
	SetOutput(io.Discard)
	if Output() != io.Discard {
		t.Error("Output() should return io.Discard")
	}
}

func TestLogger_package(t *testing.T) {
	orig := DefaultFlow
	defer func() { DefaultFlow = orig }()
	DefaultFlow = &Flow{}
	l := &FmtLogger{}
	SetLogger(l)
	if GetLogger() != l {
		t.Error("GetLogger() should return l")
	}
}

func TestUsage_package(t *testing.T) {
	orig := DefaultFlow
	defer func() { DefaultFlow = orig }()
	DefaultFlow = &Flow{}
	called := false
	SetUsage(func() { called = true })
	Usage()()
	if !called {
		t.Error("Usage should be called")
	}
}

func TestDefault_package(t *testing.T) {
	orig := DefaultFlow
	defer func() { DefaultFlow = orig }()
	DefaultFlow = &Flow{}
	task := Define(Task{Name: "task"})
	SetDefault(task)
	if Default() != task {
		t.Error("Default() should return task")
	}
}

func TestUse_package(t *testing.T) {
	orig := DefaultFlow
	defer func() { DefaultFlow = orig }()
	DefaultFlow = &Flow{}
	Use(func(r Runner) Runner { return r })
	UseExecutor(func(e Executor) Executor { return e })
}
