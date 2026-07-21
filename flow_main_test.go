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
			desc: "interrupted success",
			want: 0,
			act: func() int {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				flow := &Flow{}
				flow.SetOutput(&strings.Builder{})
				flow.Define(Task{
					Name: "task",
					Action: func(*A) {
						cancel()
					},
				})
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

func TestFlow_runMain(t *testing.T) {
	flow := &Flow{}
	flow.SetOutput(io.Discard)
	flow.Define(Task{Name: "task"})

	exited := make(chan int, 1)
	done := make(chan int, 1)
	go func() {
		done <- flow.runMain([]string{"task"}, func(code int) {
			exited <- code
		})
	}()

	select {
	case got := <-done:
		if got != exitCodePass {
			t.Fatalf("got exit code %d, want %d", got, exitCodePass)
		}
	case <-time.After(time.Second):
		t.Fatal("runMain did not finish")
	}

	select {
	case code := <-exited:
		t.Fatalf("exit called with code %d", code)
	default:
	}
}

func TestFlow_runMain_sharesSynchronizedOutput(t *testing.T) {
	flow := &Flow{}
	out := io.Discard
	flow.SetOutput(out)
	flow.Define(Task{Name: "task"})

	var gotOutput, gotFlowOutput io.Writer
	flow.UseExecutor(func(next Executor) Executor {
		return func(in ExecuteInput) error {
			gotOutput = in.Output
			gotFlowOutput = flow.Output()
			return next(in)
		}
	})

	code := flow.runMain([]string{"task"}, func(int) {})

	if code != exitCodePass {
		t.Fatalf("got exit code %d, want %d", code, exitCodePass)
	}
	if gotOutput == out {
		t.Fatal("executor middleware received the unsynchronized output")
	}
	if SyncWriter(gotOutput) != gotOutput {
		t.Fatal("Flow.runMain and Flow.Execute did not share one output wrapper")
	}
	if gotFlowOutput != gotOutput {
		t.Fatal("Flow.Output did not expose Flow.Main's synchronized output")
	}
	if flow.Output() != out {
		t.Fatal("Flow.runMain changed the configured output")
	}
}

func TestFlow_runMain_reusesSyncWriter(t *testing.T) {
	flow := &Flow{}
	out := SyncWriter(io.Discard)
	flow.SetOutput(out)
	flow.Define(Task{Name: "task"})

	var gotOutput io.Writer
	flow.UseExecutor(func(next Executor) Executor {
		return func(in ExecuteInput) error {
			gotOutput = in.Output
			return next(in)
		}
	})

	code := flow.runMain([]string{"task"}, func(int) {})

	if code != exitCodePass {
		t.Fatalf("got exit code %d, want %d", code, exitCodePass)
	}
	if gotOutput != out {
		t.Fatal("Flow.runMain wrapped a SyncWriter result again")
	}
	if flow.Output() != out {
		t.Fatal("Flow.runMain changed the configured output")
	}
}
