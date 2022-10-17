package goyek

import (
	"context"
	"io"
	"io/ioutil"
	"sync"
)

// Input received by the task runner.
type Input struct {
	Context  context.Context
	TaskName string
	Output   io.Writer
	Logger   Logger
}

// Result of a task run.
type Result struct {
	Status     Status
	PanicValue interface{}
	PanicStack []byte
}

// Status represents the status of a task run.
type Status uint8

// Statuses of task run.
const (
	StatusNotRun Status = iota
	StatusPassed
	StatusFailed
	StatusSkipped
)

// Runner represents a task runner function.
type Runner func(Input) Result

// NewRunner returns a task runner used by Flow.
//
// It can be useful when for testing and debugging
// a task action or middleware.
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

func (r taskRunner) run(in Input) Result {
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

	return tf.run(r.action)
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
