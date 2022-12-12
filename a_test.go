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

	assertEqual(t, got.Status, goyek.StatusFailed, "shoud return proper status")
	assertEqual(t, got.PanicValue, "first panic", "shoud return proper panic value")
	if want, got := "1\n2\n3\n4\n5\n", out.String(); !strings.Contains(got, want) {
		t.Errorf("should call cleanup funcs in LIFO order\ngot = %q\nwant substr = %q", got, want)
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

	assertEqual(t, got.Status, goyek.StatusFailed, "shoud return proper status")
	assertEqual(t, got.PanicValue, "action panic", "shoud return proper panic value")
	if want, got := "1\n2\n3\n4\n5\n", out.String(); !strings.Contains(got, want) {
		t.Errorf("should call cleanup funcs in LIFO order\ngot = %q\nwant substr = %q", got, want)
	}
}

func TestACleanupFail(t *testing.T) {
	got := goyek.NewRunner(func(a *goyek.A) {
		a.Cleanup(func() {
			a.Fail()
		})
	})(goyek.Input{})

	assertEqual(t, got.Status, goyek.StatusFailed, "shoud return proper status")
}

func TestASetenv(t *testing.T) {
	key := "GOYEK_TEST_ENV"
	val := "1"

	res := goyek.NewRunner(func(a *goyek.A) {
		a.Setenv(key, val)

		got := os.Getenv(key)
		assertEqual(t, got, val, "should set the value")
	})(goyek.Input{})

	assertEqual(t, res.Status, goyek.StatusPassed, "shoud return proper status")
	got := os.Getenv(key)
	assertEqual(t, got, "", "should restore the value after the action")
}

func TestASetenvOverride(t *testing.T) {
	key := "GOYEK_TEST_ENV"
	prev := "0"
	val := "1"
	os.Setenv(key, prev)   //nolint:errcheck // should never happen
	defer os.Unsetenv(key) //nolint:errcheck // should never happen

	res := goyek.NewRunner(func(a *goyek.A) {
		a.Setenv(key, val)

		got := os.Getenv(key)
		assertEqual(t, got, val, "should set the value")
	})(goyek.Input{})

	assertEqual(t, res.Status, goyek.StatusPassed, "shoud return proper status")
	got := os.Getenv(key)
	assertEqual(t, got, prev, "should restore the value after the action")
}

func TestATempDir(t *testing.T) {
	var dir string
	res := goyek.NewRunner(func(a *goyek.A) {
		dir = a.TempDir()

		_, err := os.Lstat(dir)
		assertEqual(t, err, nil, "the dir should exixt")
	})(goyek.Input{TaskName: "0!ą😊"})

	assertEqual(t, res.Status, goyek.StatusPassed, "shoud return proper status")
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
