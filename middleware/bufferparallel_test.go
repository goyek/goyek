package middleware_test

import (
	"context"
	"io"
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

func TestBufferParallel_concurrent_printing_standalone(t *testing.T) {
	const (
		goroutines         = 16
		writesPerGoroutine = 25
		message            = "msg "
	)

	runner := middleware.BufferParallel(func(in goyek.Input) goyek.Result {
		start := make(chan struct{})
		var wg sync.WaitGroup
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				for j := 0; j < writesPerGoroutine; j++ {
					io.WriteString(in.Output, message) //nolint:errcheck // not checking errors when writing to output
				}
			}()
		}
		close(start)
		wg.Wait()
		return goyek.Result{Status: goyek.StatusPassed}
	})

	out := &strings.Builder{}
	runner(goyek.Input{
		Parallel: true,
		Output:   goyek.SyncWriter(out),
	})

	if got, want := out.String(), strings.Repeat(message, goroutines*writesPerGoroutine); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestBufferParallel_nilOutput(t *testing.T) {
	runner := middleware.BufferParallel(func(in goyek.Input) goyek.Result {
		_, _ = io.WriteString(in.Output, "discarded")
		return goyek.Result{Status: goyek.StatusPassed}
	})

	result := runner(goyek.Input{Parallel: true})

	if result.Status != goyek.StatusPassed {
		t.Fatalf("got status %v, want %v", result.Status, goyek.StatusPassed)
	}
}
