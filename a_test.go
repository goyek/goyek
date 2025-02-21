package goyek_test

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/goyek/goyek/v2"
)

func TestA_WithContext(t *testing.T) {
	testCases := []struct {
		desc        string
		fn          func(a, a2 *goyek.A)
		wantStatus  goyek.Status
		wantFailed  bool
		wantSkipped bool
	}{
		{
			desc:       "Pass",
			fn:         func(_, _ *goyek.A) {},
			wantStatus: goyek.StatusPassed,
		},
		{
			desc: "Fail",
			fn: func(a, _ *goyek.A) {
				a.FailNow()
			},
			wantStatus: goyek.StatusFailed,
			wantFailed: true,
		},
		{
			desc: "FailDerived",
			fn: func(_, a2 *goyek.A) {
				a2.FailNow()
			},
			wantStatus: goyek.StatusFailed,
			wantFailed: true,
		},
		{
			desc: "Skip",
			fn: func(a, _ *goyek.A) {
				a.SkipNow()
			},
			wantStatus:  goyek.StatusSkipped,
			wantSkipped: true,
		},
		{
			desc: "SkipDerived",
			fn: func(_, a2 *goyek.A) {
				a2.SkipNow()
			},
			wantStatus:  goyek.StatusSkipped,
			wantSkipped: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			sb := &strings.Builder{}

			ctx := context.Background()
			type ctxKey struct{}
			newCtx := context.WithValue(ctx, ctxKey{}, 0)

			var (
				got, got2         context.Context
				failed, skipped   bool
				failed2, skipped2 bool
			)
			res := goyek.NewRunner(func(a *goyek.A) {
				a2 := a.WithContext(newCtx)
				got = a.Context()   // ctx
				got2 = a2.Context() // newCtx
				a2.Log("1")
				a.Cleanup(func() {
					skipped = a.Skipped()
					failed = a.Failed()
					a.Log("3")
				})
				a2.Cleanup(func() {
					skipped2 = a2.Skipped()
					failed2 = a2.Failed()
					a2.Log("2")
				})
				tc.fn(a, a2)
			})(goyek.Input{Context: ctx, Output: sb, Logger: goyek.FmtLogger{}})

			if res.Status != tc.wantStatus {
				t.Errorf("status was %s but want %s", res.Status, tc.wantStatus)
			}
			if got != ctx {
				t.Errorf("original Context returned %v but want %v", got, ctx)
			}
			if got2 != newCtx {
				t.Errorf("derived Context returned %v but want %v", got2, newCtx)
			}
			if out, want := sb.String(), "1\n2\n3\n"; out != want {
				t.Errorf("logging or cleanup failed, out was %q but want %q", out, want)
			}
			if failed != tc.wantFailed {
				t.Errorf("original Failed returned %v but want %v", failed, tc.wantFailed)
			}
			if skipped != tc.wantSkipped {
				t.Errorf("original Skipped returned %v but want %v", skipped, tc.wantSkipped)
			}
			if failed2 != tc.wantFailed {
				t.Errorf("derived Failed returned %v but want %v", failed2, tc.wantFailed)
			}
			if skipped2 != tc.wantSkipped {
				t.Errorf("derived Skipped returned %v but want %v", skipped2, tc.wantSkipped)
			}
		})
	}
}

func TestA_WithContext_nil(t *testing.T) {
	out := &strings.Builder{}
	got := goyek.NewRunner(func(a *goyek.A) {
		a.Log("1")
		a.WithContext(nil) //nolint:staticcheck // panic intentionally
		a.Log("2")
	})(goyek.Input{Logger: &goyek.FmtLogger{}, Output: out})

	assertEqual(t, got.Status, goyek.StatusFailed, "shoud return proper status")
	assertEqual(t, got.PanicValue, "nil context", "shoud return proper panic value")
	assertEqual(t, out.String(), "1\n", "should interrupt execution")
}

func TestA_Cleanup(t *testing.T) {
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
	assertContains(t, out, "1\n2\n3\n4\n5", "should call cleanup funcs in LIFO order")
}

func TestA_Cleanup_when_action_panics(t *testing.T) {
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
	assertContains(t, out, "1\n2\n3\n4\n5", "should call cleanup funcs in LIFO order")
}

func TestA_Cleanup_Fail(t *testing.T) {
	got := goyek.NewRunner(func(a *goyek.A) {
		a.Cleanup(func() {
			a.Fail()
		})
	})(goyek.Input{})

	assertEqual(t, got.Status, goyek.StatusFailed, "shoud return proper status")
}

func TestA_Cleanup_nil(t *testing.T) {
	out := &strings.Builder{}
	got := goyek.NewRunner(func(a *goyek.A) {
		a.Cleanup(func() {
			a.Log("3")
		})
		a.Log("1")
		a.Cleanup(nil) // nil cleanup func is gracefully ignored
		a.Log("2")
	})(goyek.Input{Logger: &goyek.FmtLogger{}, Output: out})

	assertEqual(t, got.Status, goyek.StatusPassed, "shoud return proper status")
	assertEqual(t, out.String(), "1\n2\n3\n", "should continue execution")
}

func TestA_Setenv(t *testing.T) {
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

func TestA_Setenv_restore(t *testing.T) {
	key := "GOYEK_TEST_ENV"
	prev := "0"
	val := "1"
	os.Setenv(key, prev)
	defer os.Unsetenv(key)

	res := goyek.NewRunner(func(a *goyek.A) {
		a.Setenv(key, val)

		got := os.Getenv(key)
		assertEqual(t, got, val, "should set the value")
	})(goyek.Input{})

	assertEqual(t, res.Status, goyek.StatusPassed, "shoud return proper status")
	got := os.Getenv(key)
	assertEqual(t, got, prev, "should restore the value after the action")
}

func TestA_TempDir(t *testing.T) {
	var dir string
	res := goyek.NewRunner(func(a *goyek.A) {
		dir = a.TempDir()

		_, err := os.Lstat(dir)
		assertEqual(t, err, nil, "the dir should exixt")
	})(goyek.Input{TaskName: "0!ą😊"})

	assertEqual(t, res.Status, goyek.StatusPassed, "shoud return proper status")
	_, err := os.Lstat(dir)
	assertTrue(t, os.IsNotExist(err), "should remove the dir after the action")
}

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

			flow.SetOutput(io.Discard)
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

func (l *helperLoggerSpy) Log(_ io.Writer, _ ...interface{}) {
	l.called = true
}

func (l *helperLoggerSpy) Logf(_ io.Writer, _ string, _ ...interface{}) {
	l.called = true
}

func (l *helperLoggerSpy) Error(_ io.Writer, _ ...interface{}) {
	l.called = true
}

func (l *helperLoggerSpy) Errorf(_ io.Writer, _ string, _ ...interface{}) {
	l.called = true
}

func (l *helperLoggerSpy) Fatal(_ io.Writer, _ ...interface{}) {
	l.called = true
}

func (l *helperLoggerSpy) Fatalf(_ io.Writer, _ string, _ ...interface{}) {
	l.called = true
}

func (l *helperLoggerSpy) Skip(_ io.Writer, _ ...interface{}) {
	l.called = true
}

func (l *helperLoggerSpy) Skipf(_ io.Writer, _ string, _ ...interface{}) {
	l.called = true
}

func (l *helperLoggerSpy) Helper() {
	l.called = true
}
