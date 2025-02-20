package goyek_test

import (
	"context"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/goyek/goyek/v2"
)

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

func TestA_WithContext(t *testing.T) {
	t.Parallel()
	type ctxKeyT struct{}
	var ctxKey ctxKeyT

	for _, parallel := range []bool{false, true} {
		parallel := parallel
		for _, throwPanic := range []bool{false, true} {
			throwPanic := throwPanic
			t.Run("Parallel="+strconv.FormatBool(parallel)+" and panic="+strconv.FormatBool(throwPanic), func(t *testing.T) {
				t.Parallel()

				flow := &goyek.Flow{}

				flow.SetOutput(io.Discard)
				loggerSpy := &helperLoggerSpy{}
				flow.SetLogger(loggerSpy)

				depsTask := flow.Define(goyek.Task{
					Name: "deps",
					Action: func(a *goyek.A) {
						a.Cleanup(onceCall(t, a.Name()+" cleanup"))
						onceCall(t, a.Name())()
					},
				})

				childTasks := make(goyek.Deps, 10)
				childTasksResults := make([]bool, len(childTasks))

				task := flow.Define(goyek.Task{
					Name: "task",
					Action: func(a *goyek.A) {
						a.Cleanup(onceCall(t, a.Name()+" cleanup"))
						onceCall(t, a.Name())()

						k, ok := a.Context().Value(ctxKey).(int)
						if !ok {
							return
						}

						childTasksResults[k] = true
					},
					Deps: goyek.Deps{depsTask},
				})

				for i := 0; i < len(childTasks); i++ {
					i := i
					childTasks[i] = flow.Define(goyek.Task{
						Name: "child task " + strconv.Itoa(i),
						Action: func(a *goyek.A) {
							a.Cleanup(onceCall(t, a.Name()+" cleanup"))
							onceCall(t, a.Name())()

							if throwPanic {
								a.Cleanup(func() {
									panic("some panic")
								})
							}

							prevCtx := a.Context()
							newA := a.WithContext(context.WithValue(a.Context(), ctxKey, i))
							assertEqual(t, newA.Name(), a.Name(), "name changed for "+a.Name())
							task.Action()(newA)
							assertEqual(t, a.Context(), prevCtx, "context changed for "+a.Name())
						},
						Deps:     goyek.Deps{depsTask},
						Parallel: parallel,
					})
				}

				main := flow.Define(goyek.Task{
					Name: "main",
					Deps: childTasks,
				})

				chkFunc := assertPass
				if throwPanic {
					chkFunc = assertFail
				}
				chkFunc(t, flow.Execute(context.Background(), []string{main.Name()}), "flow execute")

				for index, result := range childTasksResults {
					chkFunc := assertTrue
					if !parallel && throwPanic && index != 0 { // only first will be true if not parallel and with panic
						chkFunc = assertFalse
					}

					chkFunc(t, result, "child result "+strconv.Itoa(index)+" is false")
				}
			})
		}
	}
}

func onceCall(t *testing.T, name string) func() {
	t.Helper()

	var callsCount int
	var callsCountM sync.Mutex

	t.Cleanup(func() {
		assertEqual(t, callsCount, 1, "unexpected number of calls '"+name+"'")
	})

	return func() {
		callsCountM.Lock()
		defer callsCountM.Unlock()
		callsCount++
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
