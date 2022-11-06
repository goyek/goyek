package goyek

import (
	"context"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync"
)

// A is a type passed to task's action function to manage task state.
//
// A task ends when its action function returns or calls any of the methods
// FailNow, Fatal, Fatalf, SkipNow, Skip, or Skipf.
// Those methods must be called only from the goroutine running the action function.
//
// The other reporting methods, such as the variations of Log and Error,
// may be called simultaneously from multiple goroutines.
type A struct {
	ctx      context.Context
	name     string
	output   io.Writer
	logger   Logger
	mu       sync.Mutex
	failed   bool
	skipped  bool
	cleanups []func()
}

// Context returns the run context.
func (a *A) Context() context.Context {
	return a.ctx
}

// Name returns the name of the running task.
func (a *A) Name() string {
	return a.name
}

// Output returns the io.Writer used to print output.
func (a *A) Output() io.Writer {
	return a.output
}

// Cmd is like exec.Command, but it assigns tf's context
// and assigns Stdout and Stderr to tf's output,
// and Stdin to os.Stdin.
func (a *A) Cmd(name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(a.Context(), name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = a.output
	cmd.Stdout = a.output
	return cmd
}

// Log formats its arguments using default formatting, analogous to Println,
// and prints the text to Output. A final newline is added.
// The text will be printed only if the task fails or flow is run in Verbose mode.
func (a *A) Log(args ...interface{}) {
	a.logger.Log(a.output, args...)
}

// Logf formats its arguments according to the format, analogous to Printf,
// and prints the text to Output. A final newline is added.
// The text will be printed only if the task fails or flow is run in Verbose mode.
func (a *A) Logf(format string, args ...interface{}) {
	a.logger.Logf(a.output, format, args...)
}

// Error is equivalent to Log followed by Fail.
func (a *A) Error(args ...interface{}) {
	if l, ok := a.logger.(interface {
		Error(w io.Writer, args ...interface{})
	}); ok {
		l.Error(a.output, args...)
	} else {
		a.logger.Log(a.output, args...)
	}

	a.Fail()
}

// Errorf is equivalent to Logf followed by Fail.
func (a *A) Errorf(format string, args ...interface{}) {
	if l, ok := a.logger.(interface {
		Errorf(w io.Writer, format string, args ...interface{})
	}); ok {
		l.Errorf(a.output, format, args...)
	} else {
		a.logger.Logf(a.output, format, args...)
	}

	a.Fail()
}

// Failed reports whether the function has failed.
func (a *A) Failed() bool {
	a.mu.Lock()
	res := a.failed
	a.mu.Unlock()
	return res
}

// Fail marks the function as having failed but continues execution.
func (a *A) Fail() {
	a.mu.Lock()
	a.failed = true
	a.mu.Unlock()
}

// Fatal is equivalent to Log followed by FailNow.
func (a *A) Fatal(args ...interface{}) {
	if l, ok := a.logger.(interface {
		Fatal(w io.Writer, args ...interface{})
	}); ok {
		l.Fatal(a.output, args...)
	} else {
		a.logger.Log(a.output, args...)
	}

	a.FailNow()
}

// Fatalf is equivalent to Logf followed by FailNow.
func (a *A) Fatalf(format string, args ...interface{}) {
	if l, ok := a.logger.(interface {
		Fatalf(w io.Writer, format string, args ...interface{})
	}); ok {
		l.Fatalf(a.output, format, args...)
	} else {
		a.logger.Logf(a.output, format, args...)
	}

	a.FailNow()
}

// FailNow marks the function as having failed
// and stops its execution by calling runtime.Goexit
// (which then runs all deferred calls in the current goroutine).
// It finishes the whole flow.
func (a *A) FailNow() {
	a.Fail()
	runtime.Goexit()
}

// Skipped reports whether the task was skipped.
func (a *A) Skipped() bool {
	a.mu.Lock()
	res := a.skipped
	a.mu.Unlock()
	return res
}

// Skip is equivalent to Log followed by SkipNow.
func (a *A) Skip(args ...interface{}) {
	if l, ok := a.logger.(interface {
		Skip(w io.Writer, args ...interface{})
	}); ok {
		l.Skip(a.output, args...)
	} else {
		a.logger.Log(a.output, args...)
	}

	a.SkipNow()
}

// Skipf is equivalent to Logf followed by SkipNow.
func (a *A) Skipf(format string, args ...interface{}) {
	if l, ok := a.logger.(interface {
		Skipf(w io.Writer, format string, args ...interface{})
	}); ok {
		l.Skipf(a.output, format, args...)
	} else {
		a.logger.Logf(a.output, format, args...)
	}
	a.SkipNow()
}

// SkipNow marks the task as having been skipped
// and stops its execution by calling runtime.Goexit
// (which then runs all deferred calls in the current goroutine).
// If a test fails (see Error, Errorf, Fail) and is then skipped,
// it is still considered to have failed.
// Flow will continue at the next task.
func (a *A) SkipNow() {
	a.mu.Lock()
	a.skipped = true
	a.mu.Unlock()
	runtime.Goexit()
}

// Cleanup registers a function to be called when task's action function completes.
// Cleanup functions will be called in last added, first called order.
func (a *A) Cleanup(fn func()) {
	a.mu.Lock()
	a.cleanups = append(a.cleanups, fn)
	a.mu.Unlock()
}

// Helper calls logger's Helper method if implemented.
// Is us used to mark the calling function as a helper function.
// By default, when printing file and line information, that function will be skipped.
func (a *A) Helper() {
	if h, ok := a.logger.(interface {
		Helper()
	}); ok {
		h.Helper()
	}
}
