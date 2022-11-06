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
		out = ioutil.Discard
	}

	logger := in.Logger
	if logger == nil {
		logger = FmtLogger{}
	}

	a := &A{
		ctx:    ctx,
		name:   in.TaskName,
		output: &syncWriter{Writer: out},
		logger: logger,
	}

	var (
		ch         = make(chan struct{})
		finished   bool
		panicVal   interface{}
		panicStack []byte
	)
	go func() {
		defer close(ch)
		defer runCleanups(a, &finished, &panicVal, &panicStack)
		defer func() {
			if finished {
				return
			}
			panicVal = recover()
			panicStack = debug.Stack()
		}()
		r.action(a)
		finished = true
	}()
	<-ch

	res := Result{}
	switch {
	case a.Failed():
		res.Status = StatusFailed
	case a.Skipped():
		res.Status = StatusSkipped
	case finished && panicVal == nil:
		res.Status = StatusPassed
	default:
		res.Status = StatusFailed
		res.PanicValue = panicVal
		res.PanicStack = panicStack
	}
	return res
}

func runCleanups(a *A, finished *bool, panicVal *interface{}, panicStack *[]byte) {
	// we capture only the first panic
	cleanupFinished := false
	if *finished {
		defer func() {
			if cleanupFinished {
				return
			}
			*panicVal = recover()
			*panicStack = debug.Stack()
			*finished = false
		}()
	} else {
		defer func() {
			_ = recover() // ignore next panics
		}()
	}

	// Make sure that if a cleanup function panics,
	// we still run the remaining cleanup functions.
	defer func() {
		a.cleanupsMu.Lock()
		recur := len(a.cleanups) > 0
		a.cleanupsMu.Unlock()
		if recur {
			runCleanups(a, finished, panicVal, panicStack)
		}
	}()

	for {
		var cleanup func()
		a.cleanupsMu.Lock()
		if len(a.cleanups) > 0 {
			last := len(a.cleanups) - 1
			cleanup = a.cleanups[last]
			a.cleanups = a.cleanups[:last]
		}
		a.cleanupsMu.Unlock()
		if cleanup == nil {
			cleanupFinished = true
			return
		}
		cleanup()
	}
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
