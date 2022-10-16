package goyek

import (
	"context"
	"io"
	"sync"
)

// Input banana.
type Input struct {
	Context      context.Context
	TaskName     string
	Output       io.Writer
	LogDecorator func(string) string
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
	tf := &TF{
		ctx:       in.Context,
		name:      in.TaskName,
		output:    &syncWriter{Writer: in.Output},
		decorator: in.LogDecorator,
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