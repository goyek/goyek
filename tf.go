package taskflow

import (
	"context"
	"fmt"
	"io"
	"runtime"
)

// TF is a type passed to Task's Command function to manage task state.
//
// A Task ends when its Command function returns or calls any of the methods
// FailNow, Fatal, Fatalf, SkipNow, Skip, or Skipf.
//
// All methods must be called only from the goroutine running the
// Command function.
type TF struct {
	ctx     context.Context
	name    string
	writer  io.Writer
	params  map[string]string
	verbose bool
	failed  bool
	skipped bool
}

// Context returns the taskflows' run context.
func (tf *TF) Context() context.Context {
	return tf.ctx
}

// Name returns the name of the running task.
func (tf *TF) Name() string {
	return tf.name
}

// Verbose returns if verbose mode was set.
func (tf *TF) Verbose() bool {
	return tf.verbose
}

// Params returns the key-value string parameters.
func (tf *TF) Params() map[string]string {
	return tf.params
}

// Output returns the io.Writer used to print output.
func (tf *TF) Output() io.Writer {
	return tf.writer
}

// Logf formats its arguments according to the format, analogous to Printf,
// and prints the text to Output. A final newline is added.
// The text will be printed only if the task fails or taskflow is run in Verbose mode.
func (tf *TF) Logf(format string, args ...interface{}) {
	fmt.Fprintf(tf.writer, format+"\n", args...)
}

// Errorf is equivalent to Logf followed by Fail.
func (tf *TF) Errorf(format string, args ...interface{}) {
	tf.Logf(format, args...)
	tf.Fail()
}

// Failed reports whether the function has failed.
func (tf *TF) Failed() bool {
	return tf.failed
}

// Fail marks the function as having failed but continues execution.
func (tf *TF) Fail() {
	tf.failed = true
}

// Fatalf is equivalent to Logf followed by FailNow.
func (tf *TF) Fatalf(format string, args ...interface{}) {
	tf.Logf(format, args...)
	tf.FailNow()
}

// FailNow marks the function as having failed
// and stops its execution by calling runtime.Goexit
// (which then runs all deferred calls in the current goroutine).
// It finishes the whole taskflow.
func (tf *TF) FailNow() {
	tf.Fail()
	runtime.Goexit()
}

// Skipped reports whether the task was skipped.
func (tf *TF) Skipped() bool {
	return tf.skipped
}

// Skipf is equivalent to Logf followed by SkipNow.
func (tf *TF) Skipf(format string, args ...interface{}) {
	tf.Logf(format, args...)
	tf.SkipNow()
}

// SkipNow marks the task as having been skipped
// and stops its execution by calling runtime.Goexit
// (which then runs all deferred calls in the current goroutine).
// If a test fails (see Error, Errorf, Fail) and is then skipped,
// it is still considered to have failed.
// Taskflow will continue at the next task.
func (tf *TF) SkipNow() {
	tf.skipped = true
	runtime.Goexit()
}
