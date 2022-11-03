package goyek

import (
	"context"
	"io"
	"io/ioutil"
	"runtime/debug"
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
		Output   io.Writer
		Logger   Logger
	}

	// Result of a task run.
	Result struct {
		Status     Status
		PanicValue interface{}
		PanicStack []byte
	}
)

// NewRunner returns a task runner used by Flow.
//
// It can be useful when for testing and debugging
// a task action or middleware.
//
// The following defaults are set for Input
// (take notice that they are different than Flow defaults):
//
//	Context = context.Background()
//	Output = ioutil.Discard
//	Logger = FmtLogger{}
//
// It can be also used as a building block for a custom
// workflow runner if you are missing any functionalities
// provided by Flow (like concurrent dependencies execution).
func NewRunner(action func(tf *TF)) Runner {
	r := taskRunner{action: action}
	return r.run
}

type taskRunner struct {
	action func(tf *TF)
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
		out = ioutil.Discard
	}

	logger := in.Logger
	if logger == nil {
		logger = FmtLogger{}
	}

	tf := &TF{
		ctx:    ctx,
		name:   in.TaskName,
		output: &syncWriter{Writer: out},
		logger: logger,
	}

	ch := make(chan Result, 1)
	go func() {
		finished := false
		defer func() {
			res := Result{}
			switch {
			case tf.Failed():
				res.Status = StatusFailed
			case tf.Skipped():
				res.Status = StatusSkipped
			case finished:
				res.Status = StatusPassed
			default:
				res.Status = StatusFailed
				res.PanicValue = recover()
				res.PanicStack = debug.Stack()
			}
			ch <- res
		}()
		r.action(tf)
		finished = true
	}()
	return <-ch
}

type syncWriter struct {
	io.Writer
	mtx sync.Mutex
}

func (w *syncWriter) Write(p []byte) (int, error) {
	defer func() { w.mtx.Unlock() }()
	w.mtx.Lock()
	return w.Writer.Write(p)
}
