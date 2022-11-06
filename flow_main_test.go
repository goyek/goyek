package goyek

import (
	"context"
	"io/ioutil"
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
	flow.SetOutput(ioutil.Discard)
	called := false
	flow.SetUsage(func() { called = true })

	flow.main(context.Background(), nil)

	if !called {
		t.Error("usage should be called for invalid input")
	}
}
