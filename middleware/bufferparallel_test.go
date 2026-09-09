package middleware_test

import (
	"context"
	"io"
	"runtime"
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

	result := runner(goyek.Input{
		Parallel: true,
		Output:   goyek.SyncWriter(io.Discard),
	})

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

func TestBufferParallel_defersOutputUntilTaskReturns(t *testing.T) {
	const message = "message\n"
	written := make(chan struct{})
	release := make(chan struct{})
	runner := middleware.BufferParallel(func(in goyek.Input) goyek.Result {
		if n, err := io.WriteString(in.Output, message); err != nil || n != len(message) {
			t.Errorf("WriteString returned %d, %v; want %d, nil", n, err, len(message))
		}
		close(written)
		<-release
		return goyek.Result{Status: goyek.StatusPassed}
	})

	output := &strings.Builder{}
	done := make(chan goyek.Result, 1)
	go func() {
		done <- runner(goyek.Input{
			TaskName: "task",
			Parallel: true,
			Output:   goyek.SyncWriter(output),
		})
	}()
	<-written
	if got := output.String(); got != "" {
		t.Fatalf("output before task return = %q, want empty", got)
	}
	close(release)

	if result := <-done; result.Status != goyek.StatusPassed {
		t.Fatalf("got status %v, want %v", result.Status, goyek.StatusPassed)
	}
	if got := output.String(); got != message {
		t.Fatalf("output = %q, want %q", got, message)
	}
}

func TestBufferParallel_preservesLargeOutputInContiguousBlocks(t *testing.T) {
	previousProcs := runtime.GOMAXPROCS(1)
	t.Cleanup(func() { runtime.GOMAXPROCS(previousProcs) })

	const outputSize = (1 << 20) + 1024
	outputs := map[string]string{
		"task-a": strings.Repeat("a", outputSize),
		"task-b": strings.Repeat("b", outputSize),
	}
	ready := make(chan struct{}, len(outputs))
	release := make(chan struct{})
	runner := middleware.BufferParallel(func(in goyek.Input) goyek.Result {
		message := outputs[in.TaskName]
		if n, err := io.WriteString(in.Output, message); err != nil || n != len(message) {
			t.Errorf("WriteString returned %d, %v; want %d, nil", n, err, len(message))
		}
		ready <- struct{}{}
		<-release
		return goyek.Result{Status: goyek.StatusPassed}
	})

	output := &yieldingWriter{}
	done := make(chan goyek.Result, len(outputs))
	for taskName := range outputs {
		taskName := taskName
		go func() {
			done <- runner(goyek.Input{
				TaskName: taskName,
				Parallel: true,
				Output:   output,
			})
		}()
	}
	for range outputs {
		<-ready
	}
	close(release)
	for range outputs {
		if result := <-done; result.Status != goyek.StatusPassed {
			t.Fatalf("got status %v, want %v", result.Status, goyek.StatusPassed)
		}
	}

	got := output.String()
	wantAB := outputs["task-a"] + outputs["task-b"]
	wantBA := outputs["task-b"] + outputs["task-a"]
	if got != wantAB && got != wantBA {
		t.Fatalf("parallel output was truncated or interleaved; got length %d, want %d", len(got), len(wantAB))
	}
}

func TestBufferParallel_composesWithSilentNonFailed(t *testing.T) {
	message := strings.Repeat("x", (1<<20)+1)
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
			if got := output.String(); got != message {
				t.Fatalf("output length = %d, want %d", len(got), len(message))
			}
		})
	}
}

func TestBufferParallel_uncomparableOutput(t *testing.T) {
	runner := middleware.BufferParallel(func(in goyek.Input) goyek.Result {
		_, _ = io.WriteString(in.Output, "message")
		return goyek.Result{Status: goyek.StatusPassed}
	})

	result := runner(goyek.Input{
		Parallel: true,
		Output:   uncomparableWriter(make([]byte, 64)),
	})

	if result.Status != goyek.StatusPassed {
		t.Fatalf("got status %v, want %v", result.Status, goyek.StatusPassed)
	}
}

type uncomparableWriter []byte

func (w uncomparableWriter) Write(p []byte) (int, error) {
	copy(w, p)
	return len(p), nil
}

type yieldingWriter struct {
	mu      sync.Mutex
	builder strings.Builder
}

func (w *yieldingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	n, err := w.builder.Write(p)
	w.mu.Unlock()
	runtime.Gosched()
	return n, err
}

func (w *yieldingWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.builder.String()
}
