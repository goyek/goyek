package goyek

import (
	"context"
	"io"
	"io/ioutil"
	"sync"
)

// Input banana.
type Input struct {
	Context  context.Context
	TaskName string
	Output   io.Writer
	Logger   Logger
}

// Result banana.
type Result struct {
	Status     Status
	PanicValue interface{}
	PanicStack []byte
}

// Status represents the Status of an action run.
type Status uint8

// Statuses banana.
const (
	StatusNotRun Status = iota
	StatusPassed
	StatusFailed
	StatusSkipped
)

// Runner banana.
type Runner func(Input) Result

// NewRunner banana.
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
