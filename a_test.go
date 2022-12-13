package goyek_test

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/goyek/goyek/v2"
)

func TestACleanup(t *testing.T) {
	out := &strings.Builder{}

	got := goyek.NewRunner(func(a *goyek.A) {
		a.Cleanup(func() {
			a.Cleanup(func() {
				a.Log("5")
				panic("second panic")
			})
			a.Cleanup(func() {
				a.Log("4")
			})
			a.Log("3")
			panic("first panic")
		})
		a.Log("1")
		a.Cleanup(func() {
			a.Log("2")
		})
	})(goyek.Input{Logger: &goyek.FmtLogger{}, Output: out})

	if got, want := got.Status, goyek.StatusFailed; got != want {
		msg := "bad status"
		t.Errorf("%s\ngot = %q\nwant = %q", msg, got, want)
	}
	if got, want := got.PanicValue, "first panic"; got != want {
		msg := "wrong panic value"
		t.Errorf("%s\ngot = %q\nwant = %q", msg, got, want)
	}
	if want, got := "1\n2\n3\n4\n5\n", out.String(); !strings.Contains(got, want) {
		msg := "should call cleanup funcs in LIFO order"
		t.Errorf("%s\ngot = %q\nwant substr = %q", msg, got, want)
	}
}

func TestACleanupPanic(t *testing.T) {
	out := &strings.Builder{}

	got := goyek.NewRunner(func(a *goyek.A) {
		a.Cleanup(func() {
			a.Cleanup(func() {
				a.Log("5")
				panic("second panic")
			})
			a.Cleanup(func() {
				a.Log("4")
			})
			a.Log("3")
			panic("first panic")
		})
		a.Log("1")
		a.Cleanup(func() {
			a.Log("2")
		})
		panic("action panic")
	})(goyek.Input{Logger: &goyek.FmtLogger{}, Output: out})

	if got, want := got.Status, goyek.StatusFailed; got != want {
		msg := "bad status"
		t.Errorf("%s\ngot = %q\nwant = %q", msg, got, want)
	}
	if got, want := got.PanicValue, "action panic"; got != want {
		msg := "wrong panic value"
		t.Errorf("%s\ngot = %q\nwant = %q", msg, got, want)
	}
	if want, got := "1\n2\n3\n4\n5\n", out.String(); !strings.Contains(got, want) {
		msg := "should call cleanup funcs in LIFO order"
		t.Errorf("%s\ngot = %q\nwant substr = %q", msg, got, want)
	}
}

func TestACleanupFail(t *testing.T) {
	got := goyek.NewRunner(func(a *goyek.A) {
		a.Cleanup(func() {
			a.Fail()
		})
	})(goyek.Input{})

	if got, want := got.Status, goyek.StatusFailed; got != want {
		msg := "bad status"
		t.Errorf("%s\ngot = %q\nwant = %q", msg, got, want)
	}
}

func TestASetenv(t *testing.T) {
	key := "GOYEK_TEST_ENV"
	val := "1"

	res := goyek.NewRunner(func(a *goyek.A) {
		a.Setenv(key, val)

		if got := os.Getenv(key); got != val {
			msg := "should set the value"
			t.Errorf("%s\ngot = %q\nwant = %q", msg, got, val)
		}
	})(goyek.Input{})

	if got, want := res.Status, goyek.StatusPassed; got != want {
		msg := "bad status"
		t.Errorf("%s\ngot = %q\nwant = %q", msg, got, want)
	}
	if got, want := os.Getenv(key), ""; got != want {
		msg := "should restore the value after the action"
		t.Errorf("%s\ngot = %q\nwant = %q", msg, got, want)
	}
}

func TestASetenvOverride(t *testing.T) {
	key := "GOYEK_TEST_ENV"
	prev := "0"
	val := "1"
	os.Setenv(key, prev)   //nolint:errcheck // should never happen
	defer os.Unsetenv(key) //nolint:errcheck // should never happen

	res := goyek.NewRunner(func(a *goyek.A) {
		a.Setenv(key, val)

		if got := os.Getenv(key); got != val {
			msg := "should set the value"
			t.Errorf("%s\ngot = %q\nwant = %q", msg, got, val)
		}
	})(goyek.Input{})

	if got, want := res.Status, goyek.StatusPassed; got != want {
		msg := "bad status"
		t.Errorf("%s\ngot = %q\nwant = %q", msg, got, want)
	}
	if got, want := os.Getenv(key), prev; got != want {
		msg := "should restore the value after the action"
		t.Errorf("%s\ngot = %q\nwant = %q", msg, got, want)
	}
}

func TestATempDir(t *testing.T) {
	var dir string
	res := goyek.NewRunner(func(a *goyek.A) {
		dir = a.TempDir()

		if _, err := os.Lstat(dir); err != nil {
			t.Errorf("the dir should exist, dir: %v, err: %v", dir, err)
		}
	})(goyek.Input{TaskName: "0!ą😊"})

	if got, want := res.Status, goyek.StatusPassed; got != want {
		msg := "bad status"
		t.Errorf("%s\ngot = %q\nwant = %q", msg, got, want)
	}
	if _, err := os.Lstat(dir); os.IsExist(err) {
		t.Errorf("dir is not removed after the action finished, dir: %v", dir)
	}
}

func TestALogFuncsCallLogger(t *testing.T) {
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

			if !loggerSpy.called {
				t.Errorf("logger not called")
			}
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
