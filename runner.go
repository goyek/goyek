package goyek

import (
	"context"
	"io"
	"runtime/debug"
	"time"
)

// runner is used to run a Action.
type runner struct {
	Ctx         context.Context
	TaskName    string
	Output      io.Writer
	ParamValues map[string]ParamValue
}

// runResult contains the results of a Action run.
type runResult struct {
	failed   bool
	skipped  bool
	duration time.Duration
}

// Failed returns true if a action failed.
// Actions are failed by calling Error, Fail or related methods, or a panic.
func (r runResult) Failed() bool {
	return r.failed
}

// Skipped returns true if a action was skipped.
// Actions are skipped by calling Skip or related methods.
func (r runResult) Skipped() bool {
	return r.skipped
}

// Duration returns the durations of the Action.
func (r runResult) Duration() time.Duration {
	return r.duration
}

// Run runs the action.
func (r runner) Run(action func(p *Progress)) runResult {
	finished := make(chan runResult)
	go func() {
		writer := &syncWriter{Writer: r.Output}
		p := &Progress{
			ctx:         r.Ctx,
			name:        r.TaskName,
			writer:      writer,
			paramValues: r.ParamValues,
		}
		from := time.Now()
		defer func() {
			if r := recover(); r != nil {
				p.Errorf("panic: %v", r)
				p.Log(string(debug.Stack()))
			}
			result := runResult{
				failed:   p.failed,
				skipped:  p.skipped,
				duration: time.Since(from),
			}
			finished <- result
		}()
		action(p)
	}()
	return <-finished
}
