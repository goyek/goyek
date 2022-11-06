package goyek_test

import (
	"context"
	"io"
	"io/ioutil"
	"testing"

	"github.com/goyek/goyek/v2"
)

func TestA_uses_Logger_dynamic_interface(t *testing.T) {
	testCases := []struct {
		desc   string
		action func(a *goyek.A)
	}{
		{
			desc:   "Helper",
			action: func(a *goyek.A) { a.Helper() },
		},
		{
			desc:   "Log",
			action: func(a *goyek.A) { a.Log() },
		},
		{
			desc:   "Logf",
			action: func(a *goyek.A) { a.Logf("") },
		},
		{
			desc:   "Error",
			action: func(a *goyek.A) { a.Error() },
		},
		{
			desc:   "Errorf",
			action: func(a *goyek.A) { a.Errorf("") },
		},
		{
			desc:   "Fatal",
			action: func(a *goyek.A) { a.Fatal() },
		},
		{
			desc:   "Fatalf",
			action: func(a *goyek.A) { a.Fatalf("") },
		},
		{
			desc:   "Helper",
			action: func(a *goyek.A) { a.Skip() },
		},
		{
			desc:   "Skipf",
			action: func(a *goyek.A) { a.Skipf("") },
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
