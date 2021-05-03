package goyek

import (
	"context"
	"io"
	"time"
)

// runner is used to run a Command.
type runner struct {
	Ctx         context.Context
	TaskName    string
	Output      io.Writer
	ParamValues map[string]ParamValue
}

// runResult contains the results of a Command run.
type runResult struct {
	failed   bool
	skipped  bool
	duration time.Duration
}

// Failed returns true if a command failed.
// Failure can be caused by invocation of Error, Fail or related methods or a panic.
func (r runResult) Failed() bool {
	return r.failed
}

// Skipped returns true if a command was skipped.
// Skip is casused by invocation of Skip or related methods.
func (r runResult) Skipped() bool {
	return r.skipped
}

// Duration returns the durations of the Command.
func (r runResult) Duration() time.Duration {
	return r.duration
}

// Run runs the command.
func (r runner) Run(command func(tf *TF)) runResult {
	finished := make(chan runResult)
	go func() {
		writer := &syncWriter{Writer: r.Output}
		tf := &TF{
			ctx:         r.Ctx,
			name:        r.TaskName,
			writer:      writer,
			paramValues: r.ParamValues,
		}
		from := time.Now()
		defer func() {
			if r := recover(); r != nil {
				tf.Errorf("panic: %v", r)
			}
			result := runResult{
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
