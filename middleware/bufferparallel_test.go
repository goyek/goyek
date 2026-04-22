package middleware_test

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/goyek/goyek/v3"
	"github.com/goyek/goyek/v3/middleware"
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

func TestBufferParallel_NonParallel(t *testing.T) {
	out := &strings.Builder{}
	flow := &goyek.Flow{}
	flow.SetOutput(out)
	flow.Use(middleware.BufferParallel)

	flow.Define(goyek.Task{
		Name:     "task",
		Parallel: false,
		Action: func(a *goyek.A) {
			a.Log("Hello")
		},
	})

	_ = flow.Execute(context.Background(), []string{"task"})

	if !strings.Contains(out.String(), "Hello") {
		t.Errorf("expected \"Hello\", got %q", out.String())
	}
}

func TestBufferParallel_concurrent_printing(t *testing.T) {
	out := &strings.Builder{}
	flow := &goyek.Flow{}
	flow.SetOutput(out)
	flow.Use(middleware.BufferParallel)

	flow.Define(goyek.Task{
		Name:     "task",
		Parallel: true,
		Action: func(a *goyek.A) {
			var wg sync.WaitGroup
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					a.Log("some log message")
				}()
			}
			wg.Wait()
		},
	})

	err := flow.Execute(context.Background(), []string{"task"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if got := strings.Count(out.String(), "some log message"); got != 10 {
		t.Errorf("should synchronize output and keep all log messages, got %d occurrences in: %s", got, out.String())
	}
}
