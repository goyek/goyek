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

func Test_DefaultFlow(t *testing.T) {
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
	assertPass(t, goyek.Execute(context.Background(), nil), "Execute")
}

func Test_Define_empty_name(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)

	act := func() { flow.Define(goyek.Task{}) }

	assertPanics(t, act, "should panic")
}

func Test_Define_same_name(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	task := goyek.Task{Name: "task"}
	flow.Define(task)

	act := func() { flow.Define(task) }

	assertPanics(t, act, "should not be possible to register tasks with same name twice")
}

func Test_Define_bad_dep(t *testing.T) {
	flow := &goyek.Flow{}
	otherFlow := &goyek.Flow{}
	task := otherFlow.Define(goyek.Task{Name: "different-flow"})

	act := func() { flow.Define(goyek.Task{Name: "dep-from-different-flow", Deps: goyek.Deps{task}}) }

	assertPanics(t, act, "should not be possible use dependencies from different flow")
}

func Test_successful(t *testing.T) {
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

func Test_dependency_failure(t *testing.T) {
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

	assertFail(t, err, "should return error from first task")
	assertEqual(t, got(), []int{11, 0, 0}, "should execute task 1")
}

func Test_fail(t *testing.T) {
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

	assertFail(t, err, "should return error")
	assertTrue(t, failed, "a.Failed() should return true")
}

func Test_skip(t *testing.T) {
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

	err := flow.Execute(context.Background(), []string{"task"})

	assertPass(t, err, "should pass")
	assertTrue(t, skipped, "a.Skipped() should return true")
}

func Test_task_panics(t *testing.T) {
	testCases := []struct {
		desc   string
		action func(a *goyek.A)
	}{
		{
			desc:   "regular panic",
			action: func(*goyek.A) { panic("panicked!") },
		},
		{
			desc:   "nil panic",
			action: func(*goyek.A) { panic(nil) },
		},
		{
			desc:   "runtime.Goexit()",
			action: func(*goyek.A) { runtime.Goexit() },
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

			assertFail(t, err, "should return error from first task")
		})
	}
}

func Test_cancelation(t *testing.T) {
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

func Test_cancelation_during_last_task(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	flow.Define(goyek.Task{
		Name: "task",
		Action: func(*goyek.A) {
			cancel()
		},
	})

	err := flow.Execute(ctx, []string{"task"})

	assertPass(t, err, "should pass as the flow completed")
}

func Test_empty_action(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	flow.Define(goyek.Task{
		Name: "task",
	})

	err := flow.Execute(context.Background(), []string{"task"})

	assertPass(t, err, "should pass")
}

func Test_invalid_args(t *testing.T) {
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

			assertInvalid(t, err, "should return error bad args")
		})
	}
}

func Test_printing(t *testing.T) {
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

func Test_concurrent_printing(t *testing.T) {
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

	assertFail(t, err, "should fail")
	assertContains(t, out, "from child goroutine", "should contain log from child goroutine")
	assertContains(t, out, "from main goroutine", "should contain log from main goroutine")
}

func Test_concurrent_error(t *testing.T) {
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

	assertFail(t, err, "should fail")
}

func Test_name(t *testing.T) {
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

func Test_output(t *testing.T) {
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

func Test_SetDefault(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	taskRan := false
	task := flow.Define(goyek.Task{
		Name: "task",
		Action: func(*goyek.A) {
			taskRan = true
		},
	})
	flow.SetDefault(task)

	err := flow.Execute(context.Background(), nil)

	assertPass(t, err, "should pass")
	assertTrue(t, taskRan, "task should have run")
}

func TestFlow_SetDefault_panic(t *testing.T) {
	flow := &goyek.Flow{}
	otherFlow := &goyek.Flow{}
	task := otherFlow.Define(goyek.Task{Name: "different-flow"})

	act := func() {
		flow.SetDefault(task)
	}

	assertPanics(t, act, "should panic when using a task defined in other flo")
}

func TestFlow_Tasks(t *testing.T) {
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

func TestFlow_Default(t *testing.T) {
	flow := &goyek.Flow{}
	task := flow.Define(goyek.Task{Name: "name"})
	flow.SetDefault(task)

	got := flow.Default()

	assertEqual(t, got.Name(), "name", "should return the default task")
}

func TestFlow_Default_empty(t *testing.T) {
	flow := &goyek.Flow{}
	got := flow.Default()

	if got != nil {
		t.Errorf("should be nil by default, but got: %v", got)
	}
}

func TestFlow_Default_nil(t *testing.T) {
	flow := &goyek.Flow{}
	task := flow.Define(goyek.Task{Name: "name"})
	flow.SetDefault(task)
	flow.SetDefault(nil)

	got := flow.Default()

	if got != nil {
		t.Errorf("should clear the default, but got: %v", got)
	}
}

func TestFlow_Print(t *testing.T) {
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

func TestFlow_Logger(t *testing.T) {
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

func TestFlow_Logger_default(t *testing.T) {
	flow := &goyek.Flow{}

	assertEqual(t, flow.Logger(), &goyek.CodeLineLogger{}, "should have proper default")
}

func TestFlow_Output_default(t *testing.T) {
	flow := &goyek.Flow{}

	assertEqual(t, flow.Output(), os.Stdout, "should have proper default")
}

func TestFlow_Usage_default(t *testing.T) {
	getFuncName := func(fn func()) string {
		return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	}
	flow := &goyek.Flow{}

	want := getFuncName(flow.Print)
	got := getFuncName(flow.Usage())

	assertEqual(t, got, want, "should have proper default")
}

func TestFlow_Use(t *testing.T) {
	out := &strings.Builder{}
	flow := &goyek.Flow{}
	flow.SetOutput(out)
	flow.Define(goyek.Task{
		Name: "task",
	})
	flow.Use(func(goyek.Runner) goyek.Runner {
		return func(i goyek.Input) goyek.Result {
			i.Output.Write([]byte("message")) //nolint:errcheck // not checking errors when writing to output
			return goyek.Result{}
		}
	})

	_ = flow.Execute(context.Background(), []string{"task"})

	assertContains(t, out, "message", "should call middleware with proper input")
}

func TestFlow_Use_nil_middleware(t *testing.T) {
	flow := &goyek.Flow{}

	act := func() {
		flow.Use(nil)
	}

	assertPanics(t, act, "should panic on nil middleware")
}

func TestFlow_UseExecutor(t *testing.T) {
	out := &strings.Builder{}
	flow := &goyek.Flow{}
	flow.SetOutput(out)
	flow.Define(goyek.Task{
		Name: "task",
	})
	flow.UseExecutor(func(goyek.Executor) goyek.Executor {
		return func(i goyek.ExecuteInput) error {
			i.Output.Write([]byte("message")) //nolint:errcheck // not checking errors when writing to output
			return nil
		}
	})

	_ = flow.Execute(context.Background(), []string{"task"})

	assertContains(t, out, "message", "should call executor middleware")
}

func TestFlow_UseExecutor_nil_middleware(t *testing.T) {
	flow := &goyek.Flow{}

	act := func() {
		flow.UseExecutor(nil)
	}

	assertPanics(t, act, "should panic on nil middleware")
}

func TestFlow_Undefine(t *testing.T) {
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

func TestFlow_Undefine_bad_task(t *testing.T) {
	flow := &goyek.Flow{}
	otherFlow := &goyek.Flow{}
	task := otherFlow.Define(goyek.Task{Name: "different-flow"})

	act := func() { flow.Undefine(task) }

	assertPanics(t, act, "should not be possible undefine task from different flow")
}

func TestNoDeps(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	depNotRun := true
	dep := flow.Define(goyek.Task{
		Name: "dep",
		Action: func(*goyek.A) {
			depNotRun = false
		},
	})
	flow.Define(goyek.Task{
		Name: "task",
		Deps: goyek.Deps{dep},
	})

	err := flow.Execute(context.Background(), []string{"task"}, goyek.NoDeps())

	assertPass(t, err, "should pass")
	assertTrue(t, depNotRun, "deps should not have run")
}

func TestSkip_dep(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	taskRun := false
	depNotRun := true
	dep := flow.Define(goyek.Task{
		Name: "dep",
		Action: func(*goyek.A) {
			depNotRun = false
		},
	})
	flow.Define(goyek.Task{
		Name: "task",
		Deps: goyek.Deps{dep},
		Action: func(*goyek.A) {
			taskRun = true
		},
	})

	err := flow.Execute(context.Background(), []string{"task"}, goyek.Skip("dep"))

	assertPass(t, err, "should pass")
	assertTrue(t, taskRun, "task should have run")
	assertTrue(t, depNotRun, "dep should not have run")
}

func TestSkip_task(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	taskNotRun := true
	depNotRun := true
	dep := flow.Define(goyek.Task{
		Name: "dep",
		Action: func(*goyek.A) {
			depNotRun = false
		},
	})
	flow.Define(goyek.Task{
		Name: "task",
		Deps: goyek.Deps{dep},
		Action: func(*goyek.A) {
			taskNotRun = false
		},
	})

	err := flow.Execute(context.Background(), []string{"task"}, goyek.Skip("task"))

	assertPass(t, err, "should pass")
	assertTrue(t, taskNotRun, "task should not have run")
	assertTrue(t, depNotRun, "dep should not have run")
}

func TestSkip_shared_dep(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	taskNotRun := true
	otherRun := false
	depRun := false
	dep := flow.Define(goyek.Task{
		Name: "dep",
		Action: func(*goyek.A) {
			depRun = true
		},
	})
	flow.Define(goyek.Task{
		Name: "other",
		Deps: goyek.Deps{dep},
		Action: func(*goyek.A) {
			otherRun = true
		},
	})
	flow.Define(goyek.Task{
		Name: "task",
		Deps: goyek.Deps{dep},
		Action: func(*goyek.A) {
			taskNotRun = false
		},
	})

	err := flow.Execute(context.Background(), []string{"task", "other"}, goyek.Skip("task"))

	assertPass(t, err, "should pass")
	assertTrue(t, taskNotRun, "task should not have run")
	assertTrue(t, otherRun, "other should have run")
	assertTrue(t, depRun, "dep should have run")
}

func TestFlow_Parallel(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	ch := make(chan struct{})
	flow.Define(goyek.Task{
		Name:     "task-1",
		Parallel: true,
		Action: func(*goyek.A) {
			ch <- struct{}{}
		},
	})
	flow.Define(goyek.Task{
		Name:     "task-2",
		Parallel: true,
		Action: func(*goyek.A) {
			<-ch
		},
	})

	err := flow.Execute(context.Background(), []string{"task-1", "task-2"})

	assertPass(t, err, "should pass")
}

func TestFlow_Parallel_complex(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)

	var executed1, executed2, executed3, executed4, executed5 int

	taskSync := flow.Define(goyek.Task{
		Name:   "task-sync-1",
		Action: func(*goyek.A) { executed1++ },
	})
	taskSync2 := flow.Define(goyek.Task{
		Name:   "task-sync-2",
		Action: func(*goyek.A) { executed2++ },
	})
	flow.Define(goyek.Task{
		Name:   "task-sync-3",
		Action: func(*goyek.A) { executed3++ },
	})

	flow.Define(goyek.Task{
		Name:     "task-parallel-4",
		Parallel: true,
		Deps:     goyek.Deps{taskSync},
		Action:   func(*goyek.A) { executed4++ },
	})
	flow.Define(goyek.Task{
		Name:     "task-parallel-5",
		Parallel: true,
		Action: func(a *goyek.A) {
			executed5++
			a.FailNow()
		},
		Deps: goyek.Deps{taskSync, taskSync2},
	})

	err := flow.Execute(context.Background(), []string{"task-parallel-4", "task-parallel-5", "task-sync-1", "task-sync-3", "task-parallel-4"})

	assertFail(t, err, "should return error from task-parallel-5")
	assertEqual(t, executed1, 1, "should execute task-sync-1 only once")
	assertEqual(t, executed2, 1, "should execute task-sync-2 only once")
	assertEqual(t, executed3, 0, "should not execute task-sync-3")
	assertEqual(t, executed4, 1, "should execute task-parallel-4 only once")
	assertEqual(t, executed5, 1, "should execute task-parallel-5 only once")
}

func TestFlow_Parallel_NoDeps(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	depNotRun := true
	dep := flow.Define(goyek.Task{
		Name: "dep",
		Action: func(*goyek.A) {
			depNotRun = false
		},
	})
	flow.Define(goyek.Task{
		Name:     "task",
		Parallel: true,
		Deps:     goyek.Deps{dep},
	})
	flow.Define(goyek.Task{
		Name:     "task-2",
		Parallel: true,
		Deps:     goyek.Deps{dep},
	})

	err := flow.Execute(context.Background(), []string{"task", "task-2"}, goyek.NoDeps())

	assertPass(t, err, "should pass")
	assertTrue(t, depNotRun, "deps should not have run")
}

func Test_Parallel_concurrent_printing(t *testing.T) {
	out := &strings.Builder{}
	flow := &goyek.Flow{}
	flow.SetOutput(out)
	flow.Define(goyek.Task{
		Name:     "task-1",
		Parallel: true,
		Action: func(a *goyek.A) {
			a.Log("from 1")
		},
	})
	flow.Define(goyek.Task{
		Name:     "task-2",
		Parallel: true,
		Action: func(a *goyek.A) {
			a.Log("from 2")
		},
	})

	_ = flow.Execute(context.Background(), []string{"task-1", "task-2"})

	assertContains(t, out, "from 1", "should contain log from task-1")
	assertContains(t, out, "from 2", "should contain log from task-2")
}
