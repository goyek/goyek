package goyek

import (
	"context"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync"
)

// TF is a type passed to task's action function to manage task state.
//
// A task ends when its action function returns or calls any of the methods
// FailNow, Fatal, Fatalf, SkipNow, Skip, or Skipf.
// Those methods must be called only from the goroutine running the action function.
//
// The other reporting methods, such as the variations of Log and Error,
// may be called simultaneously from multiple goroutines.
type TF struct {
	ctx      context.Context
	name     string
	output   io.Writer
	logger   Logger
	failedMu sync.Mutex
	failed   bool
	skipped  bool
}

// Context returns the run context.
func (tf *TF) Context() context.Context {
	return tf.ctx
}

// Name returns the name of the running task.
func (tf *TF) Name() string {
	return tf.name
}

// Output returns the io.Writer used to print output.
func (tf *TF) Output() io.Writer {
	return tf.output
}

// Cmd is like exec.Command, but it assigns tf's context
// and assigns Stdout and Stderr to tf's output,
// and Stdin to os.Stdin.
func (tf *TF) Cmd(name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(tf.Context(), name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = tf.output
	cmd.Stdout = tf.output
	return cmd
}

// Log formats its arguments using default formatting, analogous to Println,
// and prints the text to Output. A final newline is added.
// The text will be printed only if the task fails or flow is run in Verbose mode.
func (tf *TF) Log(args ...interface{}) {
	tf.logger.Log(tf.output, args...)
}

// Logf formats its arguments according to the format, analogous to Printf,
// and prints the text to Output. A final newline is added.
// The text will be printed only if the task fails or flow is run in Verbose mode.
func (tf *TF) Logf(format string, args ...interface{}) {
	tf.logger.Logf(tf.output, format, args...)
}

// Error is equivalent to Log followed by Fail.
func (tf *TF) Error(args ...interface{}) {
	tf.logger.Log(tf.output, args...)
	tf.Fail()
}

// Errorf is equivalent to Logf followed by Fail.
func (tf *TF) Errorf(format string, args ...interface{}) {
	tf.logger.Logf(tf.output, format, args...)
	tf.Fail()
}

// Failed reports whether the function has failed.
func (tf *TF) Failed() bool {
	tf.failedMu.Lock()
	res := tf.failed
	tf.failedMu.Unlock()
	return res
}

// Fail marks the function as having failed but continues execution.
func (tf *TF) Fail() {
	tf.failedMu.Lock()
	tf.failed = true
	tf.failedMu.Unlock()
}

// Fatal is equivalent to Log followed by FailNow.
func (tf *TF) Fatal(args ...interface{}) {
	tf.logger.Log(tf.output, args...)
	tf.FailNow()
}

// Fatalf is equivalent to Logf followed by FailNow.
func (tf *TF) Fatalf(format string, args ...interface{}) {
	tf.logger.Logf(tf.output, format, args...)
	tf.FailNow()
}

// FailNow marks the function as having failed
// and stops its execution by calling runtime.Goexit
// (which then runs all deferred calls in the current goroutine).
// It finishes the whole flow.
func (tf *TF) FailNow() {
	tf.Fail()
	runtime.Goexit()
}

// Skipped reports whether the task was skipped.
func (tf *TF) Skipped() bool {
	return tf.skipped
}

// Skip is equivalent to Log followed by SkipNow.
func (tf *TF) Skip(args ...interface{}) {
	tf.logger.Log(tf.output, args...)
	tf.SkipNow()
}

// Skipf is equivalent to Logf followed by SkipNow.
func (tf *TF) Skipf(format string, args ...interface{}) {
	tf.logger.Logf(tf.output, format, args...)
	tf.SkipNow()
}

// SkipNow marks the task as having been skipped
// and stops its execution by calling runtime.Goexit
// (which then runs all deferred calls in the current goroutine).
// If a test fails (see Error, Errorf, Fail) and is then skipped,
// it is still considered to have failed.
// Flow will continue at the next task.
func (tf *TF) SkipNow() {
	tf.skipped = true
	runtime.Goexit()
}
