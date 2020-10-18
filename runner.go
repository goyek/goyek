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

func (r Runner) Run() *TF {
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

	tf := &TF{
		ctx:    ctx,
		name:   name,
		writer: writer,
	}
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		from := time.Now()
		r.Command(tf)
		tf.duration = time.Since(from)
	}()
	<-finished
	return tf
}
