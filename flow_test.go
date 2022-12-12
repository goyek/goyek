package goyek_test

import (
	"context"
	"io/ioutil"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/goyek/goyek/v2"
)

func TestDefaultFlow(t *testing.T) {
	goyek.SetOutput(ioutil.Discard)
	assertEqual(t, goyek.Output(), ioutil.Discard, "Output")

	goyek.SetLogger(goyek.FmtLogger{})
	assertEqual(t, goyek.GetLogger(), goyek.FmtLogger{}, "Logger")

	goyek.SetUsage(goyek.Print)
	goyek.Usage()()

	task := goyek.Define(goyek.Task{Name: "task"})
	other := goyek.Define(goyek.Task{Name: "other"})
	goyek.Undefine(other)
	assertEqual(t, goyek.Tasks()[0].Name(), "task", "Tasks")

	goyek.SetDefault(task)
	assertEqual(t, goyek.Default(), task, "Default")

	goyek.Use(func(r goyek.Runner) goyek.Runner { return r })
	if err := goyek.Execute(context.Background(), nil); err != nil {
		t.Errorf("should pass, but was: %v", err)
	}
}

func TestDefineEmptyName(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)

	defer func() {
		if r := recover(); r == nil {
			t.Error("should panic when defining an empty name")
		}
	}()
	flow.Define(goyek.Task{})
}

func TestDefineSameName(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	task := goyek.Task{Name: "task"}
	flow.Define(task)

	defer func() {
		if r := recover(); r == nil {
			t.Error("should not be possible to register tasks with same name twice")
		}
	}()
	flow.Define(task)
}

func TestDefineBadDep(t *testing.T) {
	flow := &goyek.Flow{}
	otherFlow := &goyek.Flow{}
	task := otherFlow.Define(goyek.Task{Name: "different-flow"})

	defer func() {
		if r := recover(); r == nil {
			t.Error("should not be possible use dependencies from different flow")
		}
	}()
	flow.Define(goyek.Task{Name: "dep-from-different-flow", Deps: goyek.Deps{task}})
}

func TestExecutePass(t *testing.T) {
	ctx := context.Background()
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	var executed1 int
	task1 := flow.Define(goyek.Task{
		Name: "task-1",
		Action: func(*goyek.A) {
			executed1++
		},
	})
	var executed2 int
	flow.Define(goyek.Task{
		Name: "task-2",
		Action: func(*goyek.A) {
			executed2++
		},
		Deps: goyek.Deps{task1},
	})
	var executed3 int
	flow.Define(goyek.Task{
		Name: "task-3",
		Action: func(*goyek.A) {
			executed3++
		},
		Deps: goyek.Deps{task1},
	})
	got := func() []int {
		return []int{executed1, executed2, executed3}
	}

	err := flow.Execute(ctx, []string{"task-1"})
	requireEqual(t, err, nil, "first execution should pass")
	requireEqual(t, got(), []int{1, 0, 0}, "should execute task 1")

	err = flow.Execute(ctx, []string{"task-2"})
	requireEqual(t, err, nil, "second execution should pass")
	requireEqual(t, got(), []int{2, 1, 0}, "should execute task 1 and 2")

	err = flow.Execute(ctx, []string{"task-1", "task-2", "task-3"})
	requireEqual(t, err, nil, "third execution should pass")
	requireEqual(t, got(), []int{3, 2, 1}, "should execute task 1 and 2 and 3")
}

func TestExecuteDepFail(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	var executed1 int
	task1 := flow.Define(goyek.Task{
		Name: "task-1",
		Action: func(a *goyek.A) {
			executed1++
			a.Error("it still runs")
			executed1 += 10
			a.FailNow()
			executed1 += 100
		},
	})
	var executed2 int
	flow.Define(goyek.Task{
		Name: "task-2",
		Action: func(*goyek.A) {
			executed2++
		},
		Deps: goyek.Deps{task1},
	})
	var executed3 int
	flow.Define(goyek.Task{
		Name: "task-3",
		Action: func(*goyek.A) {
			executed3++
		},
		Deps: goyek.Deps{task1},
	})
	got := func() []int {
		return []int{executed1, executed2, executed3}
	}

	err := flow.Execute(context.Background(), []string{"task-2", "task-3"})

	if _, ok := err.(*goyek.FailError); !ok {
		t.Errorf("should return fail error from first task, but was: %v", err)
	}
	assertEqual(t, got(), []int{11, 0, 0}, "should execute task 1")
}

func TestExecuteFail(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	failed := false
	flow.Define(goyek.Task{
		Name: "task",
		Action: func(a *goyek.A) {
			defer func() {
				failed = a.Failed()
			}()
			a.Fatal("failing")
		},
	})

	err := flow.Execute(context.Background(), []string{"task"})

	if _, ok := err.(*goyek.FailError); !ok {
		t.Errorf("should return fail error, but was: %v", err)
	}
	if !failed {
		t.Errorf("a.Failed() should return true")
	}
}

func TestExecuteSkip(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	skipped := false
	flow.Define(goyek.Task{
		Name: "task",
		Action: func(a *goyek.A) {
			defer func() {
				skipped = a.Skipped()
			}()
			a.Skip("skipping")
		},
	})

	if err := flow.Execute(context.Background(), []string{"task"}); err != nil {
		t.Errorf("should pass, but was: %v", err)
	}
	if !skipped {
		t.Errorf("a.Skipped() should return true")
	}
}

func TestExecutePanic(t *testing.T) {
	testCases := []struct {
		desc   string
		action func(a *goyek.A)
	}{
		{
			desc:   "regular panic",
			action: func(a *goyek.A) { panic("panicked!") },
		},
		{
			desc:   "nil panic",
			action: func(a *goyek.A) { panic(nil) },
		},
		{
			desc:   "runtime.Goexit()",
			action: func(a *goyek.A) { runtime.Goexit() },
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			flow := &goyek.Flow{}
			flow.SetOutput(ioutil.Discard)
			flow.Define(goyek.Task{
				Name:   "task",
				Action: tc.action,
			})

			err := flow.Execute(context.Background(), []string{"task"})

			if _, ok := err.(*goyek.FailError); !ok {
				t.Errorf("should return fail error, but was: %v", err)
			}
		})
	}
}

func TestExecuteCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	flow.Define(goyek.Task{
		Name: "task",
	})

	err := flow.Execute(ctx, []string{"task"})

	assertEqual(t, err, context.Canceled, "should return error canceled")
}

func TestExecuteCancelInTask(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	flow.Define(goyek.Task{
		Name: "task",
		Action: func(a *goyek.A) {
			cancel()
		},
	})

	if err := flow.Execute(ctx, []string{"task"}); err != nil {
		t.Errorf("should pass, but was: %v", err)
	}
}

func TestExecuteNoop(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	flow.Define(goyek.Task{
		Name: "task",
	})

	if err := flow.Execute(context.Background(), []string{"task"}); err != nil {
		t.Errorf("should pass, but was: %v", err)
	}
}

func TestExecuteErrorParallel(t *testing.T) {
	timeout := time.NewTimer(10 * time.Second)
	defer timeout.Stop()

	out := &strings.Builder{}
	flow := &goyek.Flow{}
	flow.SetOutput(out)
	flow.Define(goyek.Task{
		Name: "task",
		Action: func(a *goyek.A) {
			go func() {
				a.Fail()
			}()
			for {
				if a.Failed() {
					return
				}
				select {
				case <-timeout.C:
					t.Error("test timeout")
					return
				default:
				}
			}
		},
	})

	err := flow.Execute(context.Background(), []string{"task"})

	if _, ok := err.(*goyek.FailError); !ok {
		t.Errorf("should return fail error, but was: %v", err)
	}
}

func TestExecuteInvalidArgs(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	flow.Define(goyek.Task{Name: "task"})
	testCases := []struct {
		desc string
		args []string
		opts []goyek.Option
	}{
		{
			desc: "missing task name",
		},
		{
			desc: "empty task name",
			args: []string{""},
		},
		{
			desc: "not registered task name",
			args: []string{"unknown"},
		},
		{
			desc: "empty skip task name",
			args: []string{"task"},
			opts: []goyek.Option{goyek.Skip("")},
		},
		{
			desc: "not registered skip task name",
			args: []string{"task"},
			opts: []goyek.Option{goyek.Skip("unknown")},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			err := flow.Execute(context.Background(), tc.args, tc.opts...)

			if err == nil {
				t.Errorf("should return a non-nil error")
			} else if _, ok := err.(*goyek.FailError); ok {
				t.Errorf("should NOT return a FailError, but was: %v", err)
			}
		})
	}
}

func TestPrinting(t *testing.T) {
	out := &strings.Builder{}
	flow := &goyek.Flow{}
	flow.SetOutput(out)
	skipped := flow.Define(goyek.Task{
		Name: "skipped",
		Action: func(a *goyek.A) {
			a.Skipf("Skipf %d", 0)
		},
	})
	flow.Define(goyek.Task{
		Name: "failing",
		Deps: goyek.Deps{skipped},
		Action: func(a *goyek.A) {
			a.Log("Log", 1)
			a.Logf("Logf %d", 2)
			a.Error("Error", 3)
			a.Errorf("Errorf %d", 4)
			a.Fatalf("Fatalf %d", 5)
		},
	})

	_ = flow.Execute(context.Background(), []string{"failing"})

	assertContains(t, out, "Skipf 0", "should contain proper output from \"skipped\" task")
	assertContains(t, out, "Fatalf 5", "should contain proper output from \"failing\" task")
}

func TestPrintingParallel(t *testing.T) {
	out := &strings.Builder{}
	flow := &goyek.Flow{}
	flow.SetOutput(out)
	flow.Define(goyek.Task{
		Name: "task",
		Action: func(a *goyek.A) {
			ch := make(chan struct{}, 1)
			go func() {
				defer func() { ch <- struct{}{} }()
				a.Log("from child goroutine\nwith new line")
			}()
			a.Error("from main goroutine")
			<-ch
		},
	})

	err := flow.Execute(context.Background(), []string{"task"})

	if _, ok := err.(*goyek.FailError); !ok {
		t.Errorf("should return fail error, but was: %v", err)
	}
	assertContains(t, out, "from child goroutine", "should contain log from child goroutine")
	assertContains(t, out, "from main goroutine", "should contain log from main goroutine")
}

func TestName(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	taskName := "my-named-task"
	var got string
	flow.Define(goyek.Task{
		Name: taskName,
		Action: func(a *goyek.A) {
			got = a.Name()
		},
	})

	_ = flow.Execute(context.Background(), []string{taskName})

	assertEqual(t, got, taskName, "should return proper Name value")
}

func TestOutput(t *testing.T) {
	out := &strings.Builder{}
	flow := &goyek.Flow{}
	flow.SetOutput(out)
	msg := "hello there"
	flow.Define(goyek.Task{
		Name: "task",
		Action: func(a *goyek.A) {
			a.Output().Write([]byte(msg)) //nolint:errcheck // not checking errors when writing to output
		},
	})

	_ = flow.Execute(context.Background(), []string{"task"})

	assertContains(t, out, msg, "should contain message send via output")
}

func TestSetDefault(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	taskRan := false
	task := flow.Define(goyek.Task{
		Name: "task",
		Action: func(a *goyek.A) {
			taskRan = true
		},
	})
	flow.SetDefault(task)

	if err := flow.Execute(context.Background(), nil); err != nil {
		t.Errorf("should pass, but was: %v", err)
	}
	if !taskRan {
		t.Errorf("task should have run")
	}
}

func TestSetDefaultPanic(t *testing.T) {
	flow := &goyek.Flow{}
	otherFlow := &goyek.Flow{}
	task := otherFlow.Define(goyek.Task{Name: "different-flow"})

	defer func() {
		if r := recover(); r == nil {
			t.Error("should panic when using a task defined in other flow")
		}
	}()
	flow.SetDefault(task)
}

func TestTasks(t *testing.T) {
	flow := &goyek.Flow{}
	t1 := flow.Define(goyek.Task{Name: "one"})
	flow.Define(goyek.Task{Name: "two", Usage: "action", Deps: goyek.Deps{t1}})
	flow.Define(goyek.Task{Name: "three"})

	got := flow.Tasks()

	assertEqual(t, len(got), 3, "should return all tasks")
	assertEqual(t, got[0].Name(), "one", "should first return one")
	assertEqual(t, got[1].Name(), "three", "should then return one (sorted)")
	assertEqual(t, got[2].Name(), "two", "should next return two")
	assertEqual(t, got[2].Usage(), "action", "should return usage")
	assertEqual(t, got[2].Deps()[0], t1, "should return dependency")
}

func TestDefault(t *testing.T) {
	flow := &goyek.Flow{}
	task := flow.Define(goyek.Task{Name: "name"})
	flow.SetDefault(task)

	got := flow.Default()

	assertEqual(t, got.Name(), "name", "should return the default task")
}

func TestDefaultEmpty(t *testing.T) {
	flow := &goyek.Flow{}
	got := flow.Default()

	if got != nil {
		t.Errorf("should be nil by default, but got: %v", got)
	}
}

func TestSetDefaultNil(t *testing.T) {
	flow := &goyek.Flow{}
	task := flow.Define(goyek.Task{Name: "name"})
	flow.SetDefault(task)
	flow.SetDefault(nil)

	got := flow.Default()

	if got != nil {
		t.Errorf("should clear the default, but got: %v", got)
	}
}

func TestPrint(t *testing.T) {
	out := &strings.Builder{}
	flow := &goyek.Flow{}
	flow.SetOutput(out)
	task := flow.Define(goyek.Task{Name: "task", Usage: "use it"})
	flow.Define(goyek.Task{Name: "hidden"})
	flow.Define(goyek.Task{Name: "with-dependency", Usage: "print", Deps: goyek.Deps{task}})
	flow.SetDefault(task)

	flow.Print()

	assertContains(t, out, "use it", "should print the usage of the task")
	assertContains(t, out, "Default task: task", "should print the default task")
	assertNotContains(t, out, "hidden", "should not print task with no usage")
	assertContains(t, out, "(depends on: task)", "should print the task dependencies")
}

func TestSetLogger(t *testing.T) {
	out := &strings.Builder{}
	flow := &goyek.Flow{}
	flow.SetOutput(out)
	flow.SetLogger(goyek.FmtLogger{})
	flow.Define(goyek.Task{
		Name: "task",
		Action: func(a *goyek.A) {
			a.Log("first")
			a.Logf("second")
		},
	})

	_ = flow.Execute(context.Background(), []string{"task"})

	assertContains(t, out, "first", "should call Log")
	assertContains(t, out, "second", "should call Logf")
}

func TestLoggerDefault(t *testing.T) {
	flow := &goyek.Flow{}

	assertEqual(t, flow.Logger(), &goyek.CodeLineLogger{}, "should have proper default")
}

func TestOutputDefault(t *testing.T) {
	flow := &goyek.Flow{}

	assertEqual(t, flow.Output(), os.Stdout, "should have proper default")
}

func TestUsageDefault(t *testing.T) {
	getFuncName := func(fn func()) string {
		return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	}
	flow := &goyek.Flow{}

	want := getFuncName(flow.Print)
	got := getFuncName(flow.Usage())

	assertEqual(t, got, want, "should have proper default")
}

func TestUse(t *testing.T) {
	out := &strings.Builder{}
	flow := &goyek.Flow{}
	flow.SetOutput(out)
	flow.Define(goyek.Task{
		Name: "task",
	})
	flow.Use(func(next goyek.Runner) goyek.Runner {
		return func(i goyek.Input) goyek.Result {
			i.Output.Write([]byte("message")) //nolint:errcheck // not checking errors when writing to output
			return goyek.Result{}
		}
	})

	_ = flow.Execute(context.Background(), []string{"task"})

	assertContains(t, out, "message", "should call middleware with proper input")
}

func TestUseNil(t *testing.T) {
	flow := &goyek.Flow{}

	defer func() {
		if r := recover(); r == nil {
			t.Error("should panic on nil middleware")
		}
	}()
	flow.Use(nil)
}

func TestUndefine(t *testing.T) {
	flow := &goyek.Flow{}
	task := flow.Define(goyek.Task{Name: "name"})
	dep := flow.Define(goyek.Task{Name: "dep"})
	pipeline := flow.Define(goyek.Task{Name: "task", Deps: goyek.Deps{dep, task}})

	flow.SetDefault(task)
	flow.Undefine(task)

	got := flow.Tasks()

	assertEqual(t, len(got), 2, "should return only one task")
	assertEqual(t, got[1].Name(), "task", "should first return one")
	assertEqual(t, got[1].Deps(), goyek.Deps{dep}, "should remove dependency")
	assertEqual(t, pipeline.Deps(), goyek.Deps{dep}, "should remove dependency")
	if flow.Default() != nil {
		t.Errorf("should clear the default, but got: %v", got)
	}
}

func TestUndefineBadTask(t *testing.T) {
	flow := &goyek.Flow{}
	otherFlow := &goyek.Flow{}
	task := otherFlow.Define(goyek.Task{Name: "different-flow"})

	defer func() {
		if r := recover(); r == nil {
			t.Error("should not be possible undefine task from different flow")
		}
	}()
	flow.Undefine(task)
}

func TestNoDeps(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	var depRun bool
	dep := flow.Define(goyek.Task{
		Name: "dep",
		Action: func(a *goyek.A) {
			depRun = true
		},
	})
	flow.Define(goyek.Task{
		Name: "task",
		Deps: goyek.Deps{dep},
	})

	if err := flow.Execute(context.Background(), []string{"task"}, goyek.NoDeps()); err != nil {
		t.Errorf("should pass, but was: %v", err)
	}
	if depRun {
		t.Errorf("deps should not have run")
	}
}

func TestSkipDep(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	var taskRun, depRun bool
	dep := flow.Define(goyek.Task{
		Name: "dep",
		Action: func(a *goyek.A) {
			depRun = true
		},
	})
	flow.Define(goyek.Task{
		Name: "task",
		Deps: goyek.Deps{dep},
		Action: func(a *goyek.A) {
			taskRun = true
		},
	})

	if err := flow.Execute(context.Background(), []string{"task"}, goyek.Skip("dep")); err != nil {
		t.Errorf("should pass, but was: %v", err)
	}
	if !taskRun {
		t.Errorf("task should have run")
	}
	if depRun {
		t.Errorf("dep should not have run")
	}
}

func TestSkipTask(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	var taskRun, depRun bool
	dep := flow.Define(goyek.Task{
		Name: "dep",
		Action: func(a *goyek.A) {
			depRun = true
		},
	})
	flow.Define(goyek.Task{
		Name: "task",
		Deps: goyek.Deps{dep},
		Action: func(a *goyek.A) {
			taskRun = true
		},
	})

	if err := flow.Execute(context.Background(), []string{"task"}, goyek.Skip("task")); err != nil {
		t.Errorf("should pass, but was: %v", err)
	}
	if taskRun {
		t.Errorf("task should not have run")
	}
	if depRun {
		t.Errorf("dep should not have run")
	}
}

func TestSkipSharedDep(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	var taskRun, otherRun, depRun bool
	dep := flow.Define(goyek.Task{
		Name: "dep",
		Action: func(a *goyek.A) {
			depRun = true
		},
	})
	flow.Define(goyek.Task{
		Name: "other",
		Deps: goyek.Deps{dep},
		Action: func(a *goyek.A) {
			otherRun = true
		},
	})
	flow.Define(goyek.Task{
		Name: "task",
		Deps: goyek.Deps{dep},
		Action: func(a *goyek.A) {
			taskRun = true
		},
	})

	if err := flow.Execute(context.Background(), []string{"task", "other"}, goyek.Skip("task")); err != nil {
		t.Errorf("should pass, but was: %v", err)
	}
	if taskRun {
		t.Errorf("task should not have run")
	}
	if !otherRun {
		t.Errorf("other should have run")
	}
	if !depRun {
		t.Errorf("dep should have run")
	}
}
