package goyek

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const skipCount = 3

// TF is a type passed to Task's Action function to manage task state.
//
// A Task ends when its Action function returns or calls any of the methods
// FailNow, Fatal, Fatalf, SkipNow, Skip, or Skipf.
//
// All methods must be called only from the goroutine running the
// Action function.
type TF struct {
	ctx     context.Context
	name    string
	output  io.Writer
	failed  bool
	skipped bool
}

// Context returns the flows' run context.
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
	tf.log(args...)
}

// Logf formats its arguments according to the format, analogous to Printf,
// and prints the text to Output. A final newline is added.
// The text will be printed only if the task fails or flow is run in Verbose mode.
func (tf *TF) Logf(format string, args ...interface{}) {
	tf.logf(format, args...)
}

// Error is equivalent to Log followed by Fail.
func (tf *TF) Error(args ...interface{}) {
	tf.log(args...)
	tf.Fail()
}

// Errorf is equivalent to Logf followed by Fail.
func (tf *TF) Errorf(format string, args ...interface{}) {
	tf.logf(format, args...)
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

// Fatal is equivalent to Log followed by FailNow.
func (tf *TF) Fatal(args ...interface{}) {
	tf.log(args...)
	tf.FailNow()
}

// Fatalf is equivalent to Logf followed by FailNow.
func (tf *TF) Fatalf(format string, args ...interface{}) {
	tf.logf(format, args...)
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
	tf.log(args...)
	tf.SkipNow()
}

// Skipf is equivalent to Logf followed by SkipNow.
func (tf *TF) Skipf(format string, args ...interface{}) {
	tf.logf(format, args...)
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

// runResult contains the results of a Action run.
type runResult struct {
	Failed   bool
	Skipped  bool
	Duration time.Duration
}

// run executes the action in a seperate goroutine to enable
// interuption using runtime.Goexit().
func (tf *TF) run(action func(tf *TF)) runResult {
	finished := make(chan runResult, 1)
	go func() {
		from := time.Now()
		defer func() {
			if r := recover(); r != nil {
				txt := fmt.Sprintf("panic: %v", r)
				const skipUntilPanic = 3
				txt = decorate(txt, skipUntilPanic)
				io.WriteString(tf.output, txt) //nolint // not checking errors when writing to output
				tf.failed = true
			}
			result := runResult{
				Failed:   tf.failed,
				Skipped:  tf.skipped,
				Duration: time.Since(from),
			}
			finished <- result
		}()
		action(tf)
	}()
	return <-finished
}

// log is used internally in order to provide proper prefix.
func (tf *TF) log(args ...interface{}) {
	txt := fmt.Sprint(args...)
	txt = decorate(txt, skipCount)
	io.WriteString(tf.output, txt) //nolint // not checking errors when writing to output
}

// lof is used internally in order to provide proper prefix.
func (tf *TF) logf(format string, args ...interface{}) {
	txt := fmt.Sprintf(format, args...)
	txt = decorate(txt, skipCount)
	io.WriteString(tf.output, txt) //nolint // not checking errors when writing to output
}

// decorate prefixes the string with the file and line of the call site
// and inserts the final newline if needed and indentation spaces for formatting.
func decorate(s string, skip int) string {
	_, file, line, _ := runtime.Caller(skip)
	if file != "" {
		// Truncate file name at last file name separator.
		if index := strings.LastIndex(file, "/"); index >= 0 {
			file = file[index+1:]
		} else if index = strings.LastIndex(file, "\\"); index >= 0 {
			file = file[index+1:]
		}
	} else {
		file = "???"
	}
	if line == 0 {
		line = 1
	}
	buf := &strings.Builder{}
	// Every line is indented at least 6 spaces.
	buf.WriteString("      ")
	fmt.Fprintf(buf, "%s:%d: ", file, line)
	lines := strings.Split(s, "\n")
	if l := len(lines); l > 1 && lines[l-1] == "" {
		lines = lines[:l-1]
	}
	for i, line := range lines {
		if i > 0 {
			// Second and subsequent lines are indented an additional 4 spaces.
			buf.WriteString("\n          ")
		}
		buf.WriteString(line)
	}
	buf.WriteByte('\n')
	return buf.String()
}
