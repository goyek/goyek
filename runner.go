package taskflow

import (
	"context"
	"io"
	"io/ioutil"
	"time"
)

type Runner struct {
	Command func(tf *TF)
	Ctx     context.Context
	Name    string
	Out     io.Writer
}

type RunResult struct {
	failed   bool
	skipped  bool
	duration time.Duration
}

func (r RunResult) Failed() bool {
	return r.failed
}

func (r RunResult) Skipped() bool {
	return r.skipped
}

func (r RunResult) Passed() bool {
	return !r.failed && !r.skipped
}

func (r RunResult) Duration() time.Duration {
	return r.duration
}

func (r Runner) Run() RunResult {
	ctx := context.Background()
	if r.Ctx != nil {
		ctx = r.Ctx
	}
	name := "no-name"
	if r.Name != "" {
		name = r.Name
	}
	writer := ioutil.Discard
	if r.Out != nil {
		writer = r.Out
	}

	finished := make(chan RunResult)
	go func() {
		tf := &TF{
			ctx:    ctx,
			name:   name,
			writer: writer,
		}
		from := time.Now()
		defer func() {
			result := RunResult{
				failed:   tf.failed,
				skipped:  tf.skipped,
				duration: time.Since(from),
			}
			finished <- result
		}()
		r.Command(tf)
	}()
	return <-finished
}
