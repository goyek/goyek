package goyek

import (
	"context"
	"io"
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
	StatusPanicked
	StatusFailed
	StatusSkipped
)

// Runner banana.
type Runner func(Input) Result

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

// Interceptor banana.
type Interceptor func(Runner) Runner
