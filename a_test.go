package goyek_test

import (
	"context"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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
	type ctxKeyT struct{}
	var ctxKey ctxKeyT

	cases := []struct {
		name       string
		parallel   bool
		throwPanic bool
	}{
		{
			name:       "parallel and panic",
			parallel:   true,
			throwPanic: true,
		},
		{
			name:     "parallel",
			parallel: true,
		},
		{
			name: "sequential",
		},
		{
			name:       "sequential and panic",
			throwPanic: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			flow, task, childTasksResults, _ := prepareFlowAndTask(t, 3, ctxKey, func(*goyek.A, int) {})
			childTasks := prepareTasks(t, flow, len(childTasksResults), c.parallel, task, func(i int) func(a *goyek.A) {
				return func(a *goyek.A) {
					a.Cleanup(onceCall(t, a.Name()+" cleanup"))

					if c.throwPanic {
						a.Cleanup(func() {
							panic("some panic")
						})
					}

					prevCtx := a.Context()
					newA := a.WithContext(context.WithValue(prevCtx, ctxKey, i))
					assertEqual(t, newA.Name(), a.Name(), "name changed for "+a.Name())
					assertEqual(t, a.Context(), prevCtx, "context changed for "+a.Name())

					task.Action()(newA)
				}
			})

			main := flow.Define(goyek.Task{
				Name: "main",
				Deps: childTasks,
			})

			chkFunc := assertPass
			if c.throwPanic {
				chkFunc = assertFail
			}
			chkFunc(t, flow.Execute(context.Background(), []string{main.Name()}), "flow execute")

			for index, result := range childTasksResults {
				chkFunc := assertTrue
				if !c.parallel && c.throwPanic && index != 0 { // only first will be true if not parallel and with panic
					chkFunc = assertFalse
				}

				chkFunc(t, result, "child result "+strconv.Itoa(index)+" is incorrect")
			}
		})
	}
}

func TestA_WithContextFatal(t *testing.T) {
	type ctxKeyT struct{}
	var ctxKey ctxKeyT

	cases := []struct {
		name     string
		parallel bool
	}{
		{
			name:     "parallel",
			parallel: true,
		},
		{
			name: "sequential",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			flow, task, childTasksResults, loggerSpy := prepareFlowAndTask(t, 3, ctxKey, func(a *goyek.A, i int) {
				a.Fatal(i)

				a.Cleanup(func() {
					failed := a.Failed()
					assertTrue(t, failed, "a.Failed() should return true for "+a.Name())
				})
			})
			childTasks := prepareTasks(t, flow, len(childTasksResults), c.parallel, task, func(i int) func(a *goyek.A) {
				return func(a *goyek.A) {
					a.Cleanup(onceCall(t, a.Name()+" cleanup"))
					call := onceCall(t, a.Name())

					prevCtx := a.Context()
					newA := a.WithContext(context.WithValue(prevCtx, ctxKey, i))
					assertEqual(t, newA.Name(), a.Name(), "name changed for "+a.Name())
					assertEqual(t, a.Context(), prevCtx, "context changed for "+a.Name())
					call()
					task.Action()(newA)

					call()
					assertEqual(t, a.Context(), prevCtx, "context changed for "+a.Name()+" after "+task.Name()+" action call")
					a.Cleanup(func() {
						failed := a.Failed()
						assertTrue(t, failed, "a.Failed() should return true for "+a.Name())
					})
				}
			})

			main := flow.Define(goyek.Task{
				Name: "main",
				Deps: childTasks,
			})

			err := flow.Execute(context.Background(), []string{main.Name()})
			assertFail(t, err, "flow execute")
			assertErrorContains(t, err, "task failed", "must be a failed error")

			for index, result := range childTasksResults {
				chkFunc := assertTrue
				if !c.parallel && index != 0 { // only first will be true if not parallel and with panic
					chkFunc = assertFalse
				}

				chkFunc(t, result, "child result "+strconv.Itoa(index)+" is incorrect")
			}

			expectedFatalCalls := 1
			if c.parallel {
				expectedFatalCalls = len(childTasks)
			}

			assertTrue(t, loggerSpy.called, "logger call")
			assertEqual(t, loggerSpy.fatalCallCount, expectedFatalCalls, "fatal calls count")
		})
	}
}

func TestA_WithContextSkipped(t *testing.T) {
	type ctxKeyT struct{}
	var ctxKey ctxKeyT

	cases := []struct {
		name     string
		parallel bool
	}{
		{
			name:     "parallel",
			parallel: true,
		},
		{
			name: "sequential",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			flow, task, childTasksResults, loggerSpy := prepareFlowAndTask(t, 3, ctxKey, func(a *goyek.A, i int) {
				a.Skip(i)

				a.Cleanup(func() {
					skipped := a.Skipped()
					assertTrue(t, skipped, "a.Skipped() should return true for "+a.Name())
				})
			})
			childTasks := prepareTasks(t, flow, len(childTasksResults), c.parallel, task, func(i int) func(a *goyek.A) {
				return func(a *goyek.A) {
					a.Cleanup(onceCall(t, a.Name()+" cleanup"))
					call := onceCall(t, a.Name())

					call()
					prevCtx := a.Context()
					newA := a.WithContext(context.WithValue(prevCtx, ctxKey, i))
					assertEqual(t, newA.Name(), a.Name(), "name changed for "+a.Name())
					assertEqual(t, a.Context(), prevCtx, "context changed for "+a.Name())

					task.Action()(newA)

					call()
					a.Cleanup(func() {
						skipped := a.Skipped()
						assertTrue(t, skipped, "a.Skipped() should return true for "+a.Name())
					})
				}
			})

			main := flow.Define(goyek.Task{
				Name: "main",
				Deps: childTasks,
			})

			err := flow.Execute(context.Background(), []string{main.Name()})
			assertPass(t, err, "flow execute")

			for index, result := range childTasksResults {
				assertTrue(t, result, "child result "+strconv.Itoa(index)+" is incorrect")
			}

			assertTrue(t, loggerSpy.called, "logger call")
			assertEqual(t, loggerSpy.skipCallCount, len(childTasks), "skip calls count")
		})
	}
}

func prepareFlowAndTask(t *testing.T, childCount int, ctxKey interface{}, additionalAction func(*goyek.A, int)) (*goyek.Flow, *goyek.DefinedTask, []bool, *helperLoggerSpy) {
	t.Helper()

	flow := &goyek.Flow{}

	flow.SetOutput(io.Discard)
	loggerSpy := &helperLoggerSpy{}
	flow.SetLogger(loggerSpy)

	depsTask := flow.Define(goyek.Task{
		Name: "deps",
		Action: func(a *goyek.A) {
			a.Cleanup(onceCall(t, a.Name()+" cleanup"))
		},
	})

	childTasksResults := make([]bool, childCount)
	return flow, flow.Define(goyek.Task{
		Name: "task",
		Action: func(a *goyek.A) {
			a.Cleanup(onceCall(t, a.Name()+" cleanup"))

			k, ok := a.Context().Value(ctxKey).(int)
			if !ok {
				return
			}

			childTasksResults[k] = true

			additionalAction(a, k)
		},
		Deps: goyek.Deps{depsTask},
	}), childTasksResults, loggerSpy
}

func prepareTasks(t *testing.T, flow *goyek.Flow, count int, parallel bool, src *goyek.DefinedTask, actionBuilder func(i int) func(a *goyek.A)) []*goyek.DefinedTask {
	t.Helper()

	childTasks := make(goyek.Deps, count)
	for i := 0; i < count; i++ {
		i := i
		childTasks[i] = flow.Define(goyek.Task{
			Name:     "child task " + strconv.Itoa(i),
			Action:   actionBuilder(i),
			Deps:     goyek.Deps{src},
			Parallel: parallel,
		})
	}

	return childTasks
}

func TestA_WithContextNilCtx(t *testing.T) {
	flow := &goyek.Flow{}

	flow.SetOutput(io.Discard)
	loggerSpy := &helperLoggerSpy{}
	flow.SetLogger(loggerSpy)

	task := flow.Define(goyek.Task{
		Name: "test",
		Action: func(a *goyek.A) {
			a.WithContext(nil)
		},
	})

	assertFail(t, flow.Execute(context.Background(), []string{task.Name()}), "must contain error")
}

func onceCall(t *testing.T, name string) func() {
	t.Helper()

	var callsCount atomic.Int64

	t.Cleanup(func() {
		assertEqual(t, callsCount.Load(), int64(1), "unexpected number of calls '"+name+"'")
	})

	return func() {
		callsCount.Add(1)
	}
}

type helperLoggerSpy struct {
	called         bool
	skipCallCount  int
	fatalCallCount int
	mu             sync.Mutex
}

func (l *helperLoggerSpy) Log(_ io.Writer, _ ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.called = true
}

func (l *helperLoggerSpy) Logf(_ io.Writer, _ string, _ ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.called = true
}

func (l *helperLoggerSpy) Error(_ io.Writer, _ ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.called = true
}

func (l *helperLoggerSpy) Errorf(_ io.Writer, _ string, _ ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.called = true
}

func (l *helperLoggerSpy) Fatal(_ io.Writer, _ ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.called = true
	l.fatalCallCount++
}

func (l *helperLoggerSpy) Fatalf(_ io.Writer, _ string, _ ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.called = true
}

func (l *helperLoggerSpy) Skip(_ io.Writer, _ ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.called = true
	l.skipCallCount++
}

func (l *helperLoggerSpy) Skipf(_ io.Writer, _ string, _ ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.called = true
}

func (l *helperLoggerSpy) Helper() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.called = true
}
