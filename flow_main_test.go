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
