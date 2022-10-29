package goyek_test

import (
	"context"
	"io"
	"io/ioutil"
	"testing"

	"github.com/goyek/goyek/v2"
)

func TestTF_uses_Logger_dynamic_interface(t *testing.T) {
	testCases := []struct {
		desc   string
		action func(tf *goyek.TF)
	}{
		{
			desc:   "Helper",
			action: func(tf *goyek.TF) { tf.Helper() },
		},
		{
			desc:   "Log",
			action: func(tf *goyek.TF) { tf.Log() },
		},
		{
			desc:   "Logf",
			action: func(tf *goyek.TF) { tf.Logf("") },
		},
		{
			desc:   "Error",
			action: func(tf *goyek.TF) { tf.Error() },
		},
		{
			desc:   "Errorf",
			action: func(tf *goyek.TF) { tf.Errorf("") },
		},
		{
			desc:   "Fatal",
			action: func(tf *goyek.TF) { tf.Fatal() },
		},
		{
			desc:   "Fatalf",
			action: func(tf *goyek.TF) { tf.Fatalf("") },
		},
		{
			desc:   "Helper",
			action: func(tf *goyek.TF) { tf.Skip() },
		},
		{
			desc:   "Skipf",
			action: func(tf *goyek.TF) { tf.Skipf("") },
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			flow := &goyek.Flow{}

			flow.SetOutput(ioutil.Discard)
			loggerSpy := &helperLoggerSpy{}
			flow.SetLogger(loggerSpy)
			flow.Define(goyek.Task{
				Name:   "task",
				Action: tc.action,
			})

			_ = flow.Execute(context.Background(), []string{"task"})

			assertTrue(t, loggerSpy.called, "called logger")
		})
	}
}

type helperLoggerSpy struct {
	called bool
}

func (l *helperLoggerSpy) Log(w io.Writer, args ...interface{}) {
	l.called = true
}

func (l *helperLoggerSpy) Logf(w io.Writer, format string, args ...interface{}) {
	l.called = true
}

func (l *helperLoggerSpy) Error(w io.Writer, args ...interface{}) {
	l.called = true
}

func (l *helperLoggerSpy) Errorf(w io.Writer, format string, args ...interface{}) {
	l.called = true
}

func (l *helperLoggerSpy) Fatal(w io.Writer, args ...interface{}) {
	l.called = true
}

func (l *helperLoggerSpy) Fatalf(w io.Writer, format string, args ...interface{}) {
	l.called = true
}

func (l *helperLoggerSpy) Skip(w io.Writer, args ...interface{}) {
	l.called = true
}

func (l *helperLoggerSpy) Skipf(w io.Writer, format string, args ...interface{}) {
	l.called = true
}

func (l *helperLoggerSpy) Helper() {
	l.called = true
}
