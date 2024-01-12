package middleware_test

import (
	"context"
	"strings"
	"testing"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/goyek/v2/middleware"
)

func TestBufferParallel(t *testing.T) {
	out := &strings.Builder{}
	flow := &goyek.Flow{}
	flow.SetOutput(out)
	flow.SetLogger(goyek.FmtLogger{})
	flow.Use(middleware.BufferParallel)

	flow.Define(goyek.Task{
		Name:     "task-1",
		Parallel: true,
		Action: func(a *goyek.A) {
			a.Log("Hello")
			a.Log("Farewell")
		},
	})
	flow.Define(goyek.Task{
		Name:     "task-2",
		Parallel: true,
		Action: func(a *goyek.A) {
			a.Log("Hi")
			a.Log("Bye")
		},
	})

	_ = flow.Execute(context.Background(), []string{"task-1", "task-2"})

	_ = flow.Execute(context.Background(), []string{"task"})

	got := out.String()
	if !strings.Contains(got, "Hello\nFarewell") {
		t.Fatalf("should have not mixed input from task-1\nGOT:\n%s", got)
	}
	if !strings.Contains(got, "Hi\nBye") {
		t.Fatalf("should have not mixed input from task-2\nGOT:\n%s", got)
	}
}
