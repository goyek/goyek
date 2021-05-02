package goyek

import (
	"context"
	"io"
	"io/ioutil"
	"time"
)

// Runner is used to run a Command.
type Runner struct {
	Ctx         context.Context
	TaskName    string
	Output      io.Writer
	ParamValues map[string]ParamValue
}

// RunResult contains the results of a Command run.
type RunResult struct {
	failed   bool
	skipped  bool
	duration time.Duration
}

// Failed returns true if a command failed.
// Failure can be caused by invocation of Error, Fail or related methods or a panic.
func (r RunResult) Failed() bool {
	return r.failed
}

// Skipped returns true if a command was skipped.
// Skip is casused by invocation of Skip or related methods.
func (r RunResult) Skipped() bool {
	return r.skipped
}

// Passed true if a command passed.
// It means that it has not failed, nor skipped.
func (r RunResult) Passed() bool {
	return !r.failed && !r.skipped
}

// Duration returns the durations of the Command.
func (r RunResult) Duration() time.Duration {
	return r.duration
}

// Run runs the command.
func (r Runner) Run(command func(tf *TF)) RunResult {
	ctx := context.Background()
	if r.Ctx != nil {
		ctx = r.Ctx
	}
	name := "no-name"
	if r.TaskName != "" {
		name = r.TaskName
	}
	writer := ioutil.Discard
	if r.Output != nil {
		writer = &syncWriter{Writer: r.Output}
	}

	finished := make(chan RunResult)
	go func() {
		tf := &TF{
			ctx:         ctx,
			name:        name,
			writer:      writer,
			paramValues: r.ParamValues,
		}
		from := time.Now()
		defer func() {
			if r := recover(); r != nil {
				tf.Errorf("panic: %v", r)
			}
			result := RunResult{
				failed:   tf.failed,
				skipped:  tf.skipped,
				duration: time.Since(from),
			}
			finished <- result
		}()
		command(tf)
	}()
	return <-finished
}
