package goyek

import (
	"context"
	"fmt"
	"io"
	"runtime"
)

// The A object is passed to Task's Action function to manage task state.
//
// A Task ends when its Action function returns or calls any of the methods
// FailNow, Fatal, Fatalf, SkipNow, Skip, or Skipf.
//
// All methods must be called only from the goroutine running the
// Action function.
type A struct {
	ctx         context.Context
	name        string
	writer      io.Writer
	paramValues map[string]ParamValue
	failed      bool
	skipped     bool
}

// Context returns the flows' run context.
func (a *A) Context() context.Context {
	return a.ctx
}

// Name returns the name of the running task.
func (a *A) Name() string {
	return a.name
}

// Output returns the io.Writer used to print output.
func (a *A) Output() io.Writer {
	return a.writer
}

// Log formats its arguments using default formatting, analogous to Println,
// and prints the text to Output. A final newline is added.
// The text will be printed only if the task fails or flow is run in Verbose mode.
func (a *A) Log(args ...interface{}) {
	fmt.Fprintln(a.writer, args...)
}

// Logf formats its arguments according to the format, analogous to Prina,
// and prints the text to Output. A final newline is added.
// The text will be printed only if the task fails or flow is run in Verbose mode.
func (a *A) Logf(format string, args ...interface{}) {
	fmt.Fprintf(a.writer, format+"\n", args...)
}

// Error is equivalent to Log followed by Fail.
func (a *A) Error(args ...interface{}) {
	a.Log(args...)
	a.Fail()
}

// Errorf is equivalent to Logf followed by Fail.
func (a *A) Errorf(format string, args ...interface{}) {
	a.Logf(format, args...)
	a.Fail()
}

// Failed reports whether the function has failed.
func (a *A) Failed() bool {
	return a.failed
}

// Fail marks the function as having failed but continues execution.
func (a *A) Fail() {
	a.failed = true
}

// Fatal is equivalent to Log followed by FailNow.
func (a *A) Fatal(args ...interface{}) {
	a.Log(args...)
	a.FailNow()
}

// Fatalf is equivalent to Logf followed by FailNow.
func (a *A) Fatalf(format string, args ...interface{}) {
	a.Logf(format, args...)
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
	return a.skipped
}

// Skip is equivalent to Log followed by SkipNow.
func (a *A) Skip(args ...interface{}) {
	a.Log(args...)
	a.SkipNow()
}

// Skipf is equivalent to Logf followed by SkipNow.
func (a *A) Skipf(format string, args ...interface{}) {
	a.Logf(format, args...)
	a.SkipNow()
}

// SkipNow marks the task as having been skipped
// and stops its execution by calling runtime.Goexit
// (which then runs all deferred calls in the current goroutine).
// If a test fails (see Error, Errorf, Fail) and is then skipped,
// it is still considered to have failed.
// Flow will continue at the next task.
func (a *A) SkipNow() {
	a.skipped = true
	runtime.Goexit()
}
