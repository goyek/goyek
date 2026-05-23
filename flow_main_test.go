package goyek

import (
	"context"
	"io"
	"time"
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
		{
			desc: "deadline exceeded",
			want: 1,
			act: func() int {
				ctx, cancel := context.WithTimeout(context.Background(), -time.Hour)
				defer cancel()
				return flow.main(ctx, []string{"task"})
			},
		},
		{
			desc: "interrupted during task but Execute returned nil",
			want: 1,
			act: func() int {
				ctx, cancel := context.WithCancel(context.Background())
				f := &Flow{}
				f.Define(Task{
					Name: "task",
					Action: func(_ *A) {
						cancel()
					},
				})
				return f.main(ctx, []string{"task"})
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

func TestFailError_Error(t *testing.T) {
	err := &FailError{Task: "task"}
	want := "task failed: task"
	if got := err.Error(); got != want {
		t.Errorf("got: %q; want: %q", got, want)
	}
}

func TestFlow_main_usage(t *testing.T) {
	flow := &Flow{}
	flow.SetOutput(io.Discard)
	called := false
	flow.SetUsage(func() { called = true })

	flow.main(context.Background(), nil)

	if !called {
		t.Error("usage should be called for invalid input")
	}
}

func TestPackageWrappers(t *testing.T) {
	// Setup
	origDefaultFlow := DefaultFlow
	DefaultFlow = &Flow{}
	defer func() { DefaultFlow = origDefaultFlow }()

	task := Define(Task{Name: "task", Usage: "usage"})

	// Test Execute
	if err := Execute(context.Background(), []string{"task"}); err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	// Test Tasks
	tasks := Tasks()
	if len(tasks) != 1 || tasks[0].Name() != "task" {
		t.Errorf("Tasks returned unexpected result: %v", tasks)
	}

	// Test Print
	Print()

	// Test Default
	if Default() != nil {
		t.Error("Default should be nil")
	}

	SetDefault(task)
	if Default() != task {
		t.Error("Default mismatch after SetDefault")
	}

	// Test Logger
	SetLogger(FmtLogger{})
	if _, ok := GetLogger().(FmtLogger); !ok {
		t.Error("GetLogger mismatch after SetLogger")
	}

	// Test Output
	SetOutput(io.Discard)
	if Output() != io.Discard {
		t.Error("Output mismatch after SetOutput")
	}

	// Test Usage
	called := false
	SetUsage(func() { called = true })
	Usage()()
	if !called {
		t.Error("Usage callback not called")
	}

	// Test Undefine
	Undefine(task)
	if len(Tasks()) != 0 {
		t.Error("Tasks should be empty after Undefine")
	}

	// Test Use/UseExecutor
	Use(func(r Runner) Runner { return r })
	UseExecutor(func(e Executor) Executor { return e })
}
