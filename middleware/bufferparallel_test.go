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
	for _, want := range []string{
		"=== NAME  task-1\nHello\n",
		"=== NAME  task-1\nFarewell\n",
		"=== NAME  task-2\nHi\n",
		"=== NAME  task-2\nBye\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output does not contain %q\nGOT:\n%s", want, got)
		}
	}
}

func TestBufferParallel_concurrent_printing_standalone(t *testing.T) {
	const goroutines = 5
	const message = "msg "

	runner := middleware.BufferParallel(func(in goyek.Input) goyek.Result {
		var wg sync.WaitGroup
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				io.WriteString(in.Output, message) //nolint:errcheck // not checking errors when writing to output
			}()
		}
		wg.Wait()
		return goyek.Result{Status: goyek.StatusPassed}
	})

	out := &strings.Builder{}
	runner(goyek.Input{
		Parallel: true,
		Output:   goyek.SyncWriter(out),
	})

	if got, want := strings.Count(out.String(), strings.TrimSpace(message)), goroutines; got != want {
		t.Fatalf("got %d occurrences, want %d", got, want)
	}
}

func TestBufferParallel_nilOutput(t *testing.T) {
	runner := middleware.BufferParallel(func(in goyek.Input) goyek.Result {
		if in.Output != io.Discard {
			t.Fatalf("output = %T, want io.Discard", in.Output)
		}
		message := strings.Repeat("x", (1<<20)+1)
		if n, err := io.WriteString(in.Output, message); err != nil || n != len(message) {
			t.Fatalf("WriteString returned %d, %v; want %d, nil", n, err, len(message))
		}
		return goyek.Result{Status: goyek.StatusPassed}
	})

	result := runner(goyek.Input{Parallel: true})

	if result.Status != goyek.StatusPassed {
		t.Fatalf("got status %v, want %v", result.Status, goyek.StatusPassed)
	}
}

func TestBufferParallel_discardOutput(t *testing.T) {
	runner := middleware.BufferParallel(func(in goyek.Input) goyek.Result {
		if in.Output != io.Discard {
			t.Fatalf("output = %T, want io.Discard", in.Output)
		}
		message := strings.Repeat("x", (1<<20)+1)
		if n, err := io.WriteString(in.Output, message); err != nil || n != len(message) {
			t.Fatalf("WriteString returned %d, %v; want %d, nil", n, err, len(message))
		}
		return goyek.Result{Status: goyek.StatusPassed}
	})

	result := runner(goyek.Input{Parallel: true, Output: goyek.SyncWriter(io.Discard)})

	if result.Status != goyek.StatusPassed {
		t.Fatalf("got status %v, want %v", result.Status, goyek.StatusPassed)
	}
}

func TestBufferParallel_nonParallelPassThrough(t *testing.T) {
	output := &strings.Builder{}
	runner := middleware.BufferParallel(func(in goyek.Input) goyek.Result {
		if in.Output != output {
			t.Fatal("non-parallel output was replaced")
		}
		_, _ = io.WriteString(in.Output, "message")
		return goyek.Result{Status: goyek.StatusPassed}
	})

	result := runner(goyek.Input{Output: output})

	if result.Status != goyek.StatusPassed {
		t.Fatalf("got status %v, want %v", result.Status, goyek.StatusPassed)
	}
	if got, want := output.String(), "message"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestBufferParallel_uncomparableOutput(t *testing.T) {
	runner := middleware.BufferParallel(func(in goyek.Input) goyek.Result {
		_, _ = io.WriteString(in.Output, "message\n")
		return goyek.Result{Status: goyek.StatusPassed}
	})

	result := runner(goyek.Input{
		TaskName: "task",
		Parallel: true,
		Output:   uncomparableWriter(make([]byte, 64)),
	})

	if result.Status != goyek.StatusPassed {
		t.Fatalf("got status %v, want %v", result.Status, goyek.StatusPassed)
	}
}

func TestBufferParallel_streamsBeforeTaskReturns(t *testing.T) {
	const message = "message\n"
	written := make(chan struct{})
	release := make(chan struct{})
	writeResult := make(chan struct {
		n   int
		err error
	}, 1)
	runner := middleware.BufferParallel(func(in goyek.Input) goyek.Result {
		n, err := io.WriteString(in.Output, message)
		writeResult <- struct {
			n   int
			err error
		}{n: n, err: err}
		close(written)
		<-release
		return goyek.Result{Status: goyek.StatusPassed}
	})

	out := &strings.Builder{}
	done := make(chan goyek.Result)
	go func() {
		done <- runner(goyek.Input{
			TaskName: "task",
			Parallel: true,
			Output:   goyek.SyncWriter(out),
		})
	}()
	<-written
	got := out.String()
	result := <-writeResult
	close(release)
	taskResult := <-done

	if result.err != nil {
		t.Fatalf("writing output: %v", result.err)
	}
	if result.n != len(message) {
		t.Fatalf("wrote %d bytes, want %d", result.n, len(message))
	}
	if want := "=== NAME  task\n" + message; got != want {
		t.Fatalf("output before task return = %q, want %q", got, want)
	}
	if taskResult.Status != goyek.StatusPassed {
		t.Fatalf("got status %v, want %v", taskResult.Status, goyek.StatusPassed)
	}
}

func TestBufferParallel_composesWithSilentNonFailed(t *testing.T) {
	message := strings.Repeat("x", (1<<20)+1) + "\n"
	action := func(in goyek.Input) goyek.Result {
		if n, err := io.WriteString(in.Output, message); err != nil || n != len(message) {
			t.Fatalf("WriteString returned %d, %v; want %d, nil", n, err, len(message))
		}
		return goyek.Result{Status: goyek.StatusFailed}
	}
	tests := []struct {
		name string
		wrap func(goyek.Runner) goyek.Runner
	}{
		{
			name: "buffer outside silent",
			wrap: func(next goyek.Runner) goyek.Runner {
				return middleware.BufferParallel(middleware.SilentNonFailed(next))
			},
		},
		{
			name: "silent outside buffer",
			wrap: func(next goyek.Runner) goyek.Runner {
				return middleware.SilentNonFailed(middleware.BufferParallel(next))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &strings.Builder{}
			result := tt.wrap(action)(goyek.Input{
				TaskName: "task",
				Parallel: true,
				Output:   goyek.SyncWriter(output),
			})

			if result.Status != goyek.StatusFailed {
				t.Fatalf("got status %v, want %v", result.Status, goyek.StatusFailed)
			}
			want := "=== NAME  task\n" + strings.Repeat("x", 1<<20) + "\n\t... [output truncated]\n"
			if got := output.String(); got != want {
				t.Fatalf("output length = %d, want %d", len(got), len(want))
			}
		})
	}
}

type uncomparableWriter []byte

func (w uncomparableWriter) Write(p []byte) (int, error) {
	copy(w, p)
	return len(p), nil
}
