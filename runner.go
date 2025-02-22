package goyek

import (
	"context"
	"io"
	"sync"
)

// Task runner types.
type (
	// Runner represents a task runner function.
	Runner func(Input) Result

	// Input received by the task runner.
	Input struct {
		Context  context.Context
		TaskName string
		Parallel bool
		Output   io.Writer
		Logger   Logger
	}

	// Result of a task run.
	Result struct {
		Status     Status
		PanicValue interface{}
		PanicStack []byte
	}

	// Middleware represents a task runner interceptor.
	Middleware func(Runner) Runner
)

// NewRunner returns a task runner used by Flow.
//
// It can be useful for testing and debugging
// a task action or middleware.
//
// The following defaults are set for Input
// (take notice that they are different than Flow defaults):
//
//	Context = context.Background()
//	Output = io.Discard
//	Logger = FmtLogger{}
//
// It can be also used as a building block for a custom
// workflow runner if you are missing any functionalities
// provided by Flow (like concurrent dependencies execution).
func NewRunner(action func(a *A)) Runner {
	r := taskRunner{action: action}
	return r.run
}

type taskRunner struct {
	action func(a *A)
}

// run executes the action in a separate goroutine to enable
// interuption using runtime.Goexit().
func (r taskRunner) run(in Input) Result {
	if r.action == nil {
		return Result{}
	}

	ctx := in.Context
	if ctx == nil {
		ctx = context.Background()
	}

	out := in.Output
	if out == nil {
		out = io.Discard
	}

	logger := in.Logger
	if logger == nil {
		logger = FmtLogger{}
	}

	var failed, skipped bool
	a := &A{
		mu:       &sync.Mutex{},
		failed:   &failed,
		skipped:  &skipped,
		cleanups: &[]func(){},
		ctx:      ctx,
		name:     in.TaskName,
		output:   out,
		logger:   logger,
	}

	finished, panicVal, panicStack := a.run(r.action)

	res := Result{}
	switch {
	case a.Failed():
		res.Status = StatusFailed
	case a.Skipped():
		res.Status = StatusSkipped
	case finished:
		res.Status = StatusPassed
	default:
		res.Status = StatusFailed
		res.PanicValue = panicVal
		res.PanicStack = panicStack
	}
	return res
}
