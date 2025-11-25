package middleware_test

import (
	"context"
	"strings"
	"testing"

	"github.com/goyek/goyek/v3"
	"github.com/goyek/goyek/v3/middleware"
)

func TestReportFlow(t *testing.T) {
	flow := &goyek.Flow{}
	flow.Define(goyek.Task{Name: "task"})
	flow.Define(goyek.Task{Name: "failing", Action: func(a *goyek.A) { a.Fail() }})
	flow.UseExecutor(middleware.ReportFlow)

	testCases := []struct {
		desc string
		want string
		act  func()
	}{
		{
			desc: "pass",
			want: "ok",
			act:  func() { _ = flow.Execute(context.Background(), []string{"task"}) },
		},
		{
			desc: "fail",
			want: "task failed: failing",
			act:  func() { _ = flow.Execute(context.Background(), []string{"failing"}) },
		},
		{
			desc: "invalid",
			want: "task provided but not defined: bad",
			act:  func() { _ = flow.Execute(context.Background(), []string{"bad"}) },
		},
		{
			desc: "canceled",
			want: "context canceled",
			act: func() {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				_ = flow.Execute(ctx, []string{"task"})
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			b := &strings.Builder{}
			flow.SetOutput(b)
			tc.act()
			if got := b.String(); !strings.Contains(got, tc.want) {
				t.Errorf("got: %s; should contain: %s", got, tc.want)
			}
		})
	}
}
