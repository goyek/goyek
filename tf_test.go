package goyek_test

import (
	"context"
	"io"
	"io/ioutil"
	"testing"

	"github.com/goyek/goyek/v2"
)

func TestTF_Helper(t *testing.T) {
	flow := &goyek.Flow{}

	flow.SetOutput(ioutil.Discard)
	loggerSpy := &helperLoggerSpy{}
	flow.SetLogger(loggerSpy)
	flow.Define(goyek.Task{
		Name: "task",
		Action: func(tf *goyek.TF) {
			tf.Helper()
		},
	})

	_ = flow.Execute(context.Background(), []string{"task"})

	assertTrue(t, loggerSpy.called, "called helper")
}

type helperLoggerSpy struct {
	called bool
}

func (l *helperLoggerSpy) Log(w io.Writer, args ...interface{}) {
}

func (l *helperLoggerSpy) Logf(w io.Writer, format string, args ...interface{}) {
}

func (l *helperLoggerSpy) Helper() {
	l.called = true
}
