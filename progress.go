package goyek

import (
	"context"
	"fmt"
	"io"
	"runtime"
)

// Progress object is passed to Task's Action function to manage task state.
//
// Progress Task ends when its Action function returns or calls any of the methods
// FailNow, Fatal, Fatalf, SkipNow, Skip, or Skipf.
//
// All methods must be called only from the goroutine running the
// Action function.
type Progress struct {
	ctx         context.Context
	name        string
	writer      io.Writer
	paramValues map[string]ParamValue
	failed      bool
	skipped     bool
}

// Context returns the flows' run context.
func (p *Progress) Context() context.Context {
	return p.ctx
}

// Name returns the name of the running task.
func (p *Progress) Name() string {
	return p.name
}

// Output returns the io.Writer used to print output.
func (p *Progress) Output() io.Writer {
	return p.writer
}

// Log formats its arguments using default formatting, analogous to Println,
// and prints the text to Output. A final newline is added.
// The text will be printed only if the task fails or flow is run in Verbose mode.
func (p *Progress) Log(args ...interface{}) {
	fmt.Fprintln(p.writer, args...)
}

// Logf formats its arguments according to the format, analogous to Printf,
// and prints the text to Output. A final newline is added.
// The text will be printed only if the task fails or flow is run in Verbose mode.
func (p *Progress) Logf(format string, args ...interface{}) {
	fmt.Fprintf(p.writer, format+"\n", args...)
}

// Error is equivalent to Log followed by Fail.
func (p *Progress) Error(args ...interface{}) {
	p.Log(args...)
	p.Fail()
}

// Errorf is equivalent to Logf followed by Fail.
func (p *Progress) Errorf(format string, args ...interface{}) {
	p.Logf(format, args...)
	p.Fail()
}

// Failed reports whether the function has failed.
func (p *Progress) Failed() bool {
	return p.failed
}

// Fail marks the function as having failed but continues execution.
func (p *Progress) Fail() {
	p.failed = true
}

// Fatal is equivalent to Log followed by FailNow.
func (p *Progress) Fatal(args ...interface{}) {
	p.Log(args...)
	p.FailNow()
}

// Fatalf is equivalent to Logf followed by FailNow.
func (p *Progress) Fatalf(format string, args ...interface{}) {
	p.Logf(format, args...)
	p.FailNow()
}

// FailNow marks the function as having failed
// and stops its execution by calling runtime.Goexit
// (which then runs all deferred calls in the current goroutine).
// It finishes the whole flow.
func (p *Progress) FailNow() {
	p.Fail()
	runtime.Goexit()
}

// Skipped reports whether the task was skipped.
func (p *Progress) Skipped() bool {
	return p.skipped
}

// Skip is equivalent to Log followed by SkipNow.
func (p *Progress) Skip(args ...interface{}) {
	p.Log(args...)
	p.SkipNow()
}

// Skipf is equivalent to Logf followed by SkipNow.
func (p *Progress) Skipf(format string, args ...interface{}) {
	p.Logf(format, args...)
	p.SkipNow()
}

// SkipNow marks the task as having been skipped
// and stops its execution by calling runtime.Goexit
// (which then runs all deferred calls in the current goroutine).
// If a test fails (see Error, Errorf, Fail) and is then skipped,
// it is still considered to have failed.
// Flow will continue at the next task.
func (p *Progress) SkipNow() {
	p.skipped = true
	runtime.Goexit()
}
