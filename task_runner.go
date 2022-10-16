package goyek

import (
	"context"
	"io"
)

type input struct {
	Context      context.Context
	TaskName     string
	Output       io.Writer
	LogDecorator func(string) string
}

type runner func(input) result

type taskRunner struct {
	action func(tf *TF)
}

func (r taskRunner) run(in input) result {
	tf := &TF{
		ctx:       in.Context,
		name:      in.TaskName,
		output:    &syncWriter{Writer: in.Output},
		decorator: in.LogDecorator,
	}
	return tf.run(r.action)
}

type interceptor func(runner) runner
