package goyek_test

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/goyek/goyek/v2"
)

func Test_Define_empty_name(t *testing.T) {
	flow := &goyek.Flow{Output: &strings.Builder{}}

	act := func() { flow.Define(goyek.Task{}) }

	assertPanics(t, act, "should panic")
}

func Test_Define_same_name(t *testing.T) {
	flow := &goyek.Flow{Output: &strings.Builder{}}
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
	flow := &goyek.Flow{Output: &strings.Builder{}}
	var executed1 int
	task1 := flow.Define(goyek.Task{
		Name: "task-1",
		Action: func(*goyek.TF) {
			executed1++
		},
	})
	var executed2 int
	flow.Define(goyek.Task{
		Name: "task-2",
		Action: func(*goyek.TF) {
			executed2++
		},
		Deps: goyek.Deps{task1},
	})
	var executed3 int
	flow.Define(goyek.Task{
		Name: "task-3",
		Action: func(*goyek.TF) {
			executed3++
		},
		Deps: goyek.Deps{task1},
	})
	got := func() []int {
		return []int{executed1, executed2, executed3}
	}

	err := flow.Execute(ctx, "task-1")
	requireEqual(t, err, nil, "first execution should pass")
	requireEqual(t, got(), []int{1, 0, 0}, "should execute task 1")

	err = flow.Execute(ctx, "task-2")
	requireEqual(t, err, nil, "second execution should pass")
	requireEqual(t, got(), []int{2, 1, 0}, "should execute task 1 and 2")

	err = flow.Execute(ctx, "task-1", "task-2", "task-3")
	requireEqual(t, err, nil, "third execution should pass")
	requireEqual(t, got(), []int{3, 2, 1}, "should execute task 1 and 2 and 3")
}

func Test_dependency_failure(t *testing.T) {
	flow := &goyek.Flow{Output: &strings.Builder{}}
	var executed1 int
	task1 := flow.Define(goyek.Task{
		Name: "task-1",
		Action: func(tf *goyek.TF) {
			executed1++
			tf.Error("it still runs")
			executed1 += 10
			tf.FailNow()
			executed1 += 100
		},
	})
	var executed2 int
	flow.Define(goyek.Task{
		Name: "task-2",
		Action: func(*goyek.TF) {
			executed2++
		},
		Deps: goyek.Deps{task1},
	})
	var executed3 int
	flow.Define(goyek.Task{
		Name: "task-3",
		Action: func(*goyek.TF) {
			executed3++
		},
		Deps: goyek.Deps{task1},
	})
	got := func() []int {
		return []int{executed1, executed2, executed3}
	}

	err := flow.Execute(context.Background(), "task-2", "task-3")

	assertFail(t, err, "should return error from first task")
	assertEqual(t, got(), []int{11, 0, 0}, "should execute task 1")
}

func Test_fail(t *testing.T) {
	flow := &goyek.Flow{Output: &strings.Builder{}}
	failed := false
	flow.Define(goyek.Task{
		Name: "task",
		Action: func(tf *goyek.TF) {
			defer func() {
				failed = tf.Failed()
			}()
			tf.Fatal("failing")
		},
	})

	err := flow.Execute(context.Background(), "task")

	assertFail(t, err, "should return error")
	assertTrue(t, failed, "tf.Failed() should return true")
}

func Test_skip(t *testing.T) {
	flow := &goyek.Flow{Output: &strings.Builder{}}
	skipped := false
	flow.Define(goyek.Task{
		Name: "task",
		Action: func(tf *goyek.TF) {
			defer func() {
				skipped = tf.Skipped()
			}()
			tf.Skip("skipping")
		},
	})

	err := flow.Execute(context.Background(), "task")

	assertPass(t, err, "should pass")
	assertTrue(t, skipped, "tf.Skipped() should return true")
}

func Test_task_panics(t *testing.T) {
	testCases := []struct {
		desc   string
		action func(tf *goyek.TF)
	}{
		{
			desc:   "regular panic",
			action: func(tf *goyek.TF) { panic("panicked!") },
		},
		{
			desc:   "nil panic",
			action: func(tf *goyek.TF) { panic(nil) },
		},
		{
			desc:   "runtime.Goexit()",
			action: func(tf *goyek.TF) { runtime.Goexit() },
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			flow := &goyek.Flow{Output: &strings.Builder{}}
			flow.Define(goyek.Task{
				Name:   "task",
				Action: tc.action,
			})

			err := flow.Execute(context.Background(), "task")

			assertFail(t, err, "should return error from first task")
		})
	}
}

func Test_cancelation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	flow := &goyek.Flow{Output: &strings.Builder{}}
	flow.Define(goyek.Task{
		Name: "task",
	})

	err := flow.Execute(ctx, "task")

	assertEqual(t, err, context.Canceled, "should return error canceled")
}

func Test_cancelation_during_last_task(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	flow := &goyek.Flow{Output: &strings.Builder{}}
	flow.Define(goyek.Task{
		Name: "task",
		Action: func(tf *goyek.TF) {
			cancel()
		},
	})

	err := flow.Execute(ctx, "task")

	assertPass(t, err, "should pass as the flow completed")
}

func Test_empty_action(t *testing.T) {
	flow := &goyek.Flow{Output: &strings.Builder{}}
	flow.Define(goyek.Task{
		Name: "task",
	})

	err := flow.Execute(context.Background(), "task")

	assertPass(t, err, "should pass")
}

func Test_invalid_args(t *testing.T) {
	flow := &goyek.Flow{Output: &strings.Builder{}}
	testCases := []struct {
		desc string
		args []string
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
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			err := flow.Execute(context.Background(), tc.args...)

			assertInvalid(t, err, "should return error bad args")
		})
	}
}

func Test_printing(t *testing.T) {
	out := &strings.Builder{}
	flow := &goyek.Flow{
		Output: out,
	}
	skipped := flow.Define(goyek.Task{
		Name: "skipped",
		Action: func(tf *goyek.TF) {
			tf.Skipf("Skipf %d", 0)
		},
	})
	flow.Define(goyek.Task{
		Name: "failing",
		Deps: goyek.Deps{skipped},
		Action: func(tf *goyek.TF) {
			tf.Log("Log", 1)
			tf.Logf("Logf %d", 2)
			tf.Error("Error", 3)
			tf.Errorf("Errorf %d", 4)
			tf.Fatalf("Fatalf %d", 5)
		},
	})

	_ = flow.Execute(context.Background(), "failing")

	assertContains(t, out, "Skipf 0", "should contain proper output from \"skipped\" task")
	assertContains(t, out, "Fatalf 5", "should contain proper output from \"failing\" task")
}

func Test_concurrent_printing(t *testing.T) {
	out := &strings.Builder{}
	flow := goyek.Flow{
		Output: out,
	}
	flow.Define(goyek.Task{
		Name: "task",
		Action: func(tf *goyek.TF) {
			ch := make(chan struct{}, 1)
			go func() {
				defer func() { ch <- struct{}{} }()
				tf.Log("from child goroutine\nwith new line")
			}()
			tf.Error("from main goroutine")
			<-ch
		},
	})

	err := flow.Execute(context.Background(), "task")

	assertFail(t, err, "should fail")
	assertContains(t, out, "from child goroutine", "should contain log from child goroutine")
	assertContains(t, out, "from main goroutine", "should contain log from main goroutine")
}

func Test_concurrent_error(t *testing.T) {
	timeout := time.NewTimer(10 * time.Second)
	defer timeout.Stop()

	out := &strings.Builder{}
	flow := goyek.Flow{
		Output: out,
	}
	flow.Define(goyek.Task{
		Name: "task",
		Action: func(tf *goyek.TF) {
			go func() {
				tf.Fail()
			}()
			for {
				if tf.Failed() {
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

	err := flow.Execute(context.Background(), "task")

	assertFail(t, err, "should fail")
}

func Test_name(t *testing.T) {
	flow := &goyek.Flow{Output: &strings.Builder{}}
	taskName := "my-named-task"
	var got string
	flow.Define(goyek.Task{
		Name: taskName,
		Action: func(tf *goyek.TF) {
			got = tf.Name()
		},
	})

	_ = flow.Execute(context.Background(), taskName)

	assertEqual(t, got, taskName, "should return proper Name value")
}

func Test_output(t *testing.T) {
	out := &strings.Builder{}
	flow := &goyek.Flow{Output: out}
	msg := "hello there"
	flow.Define(goyek.Task{
		Name: "task",
		Action: func(tf *goyek.TF) {
			tf.Output().Write([]byte(msg)) //nolint:errcheck,gosec // not checking errors when writing to output
		},
	})

	_ = flow.Execute(context.Background(), "task")

	assertContains(t, out, msg, "should contain message send via output")
}

func Test_SetDefault(t *testing.T) {
	flow := &goyek.Flow{Output: &strings.Builder{}}
	taskRan := false
	task := flow.Define(goyek.Task{
		Name: "task",
		Action: func(tf *goyek.TF) {
			taskRan = true
		},
	})
	flow.SetDefault(task)

	err := flow.Execute(context.Background())

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

func TestCmd_success(t *testing.T) {
	taskName := "exec"
	out := &strings.Builder{}
	flow := &goyek.Flow{
		Output: out,
	}
	flow.Define(goyek.Task{
		Name: taskName,
		Action: func(tf *goyek.TF) {
			if err := tf.Cmd("go", "version").Run(); err != nil {
				tf.Fatal(err)
			}
		},
	})

	err := flow.Execute(context.Background(), taskName)

	assertPass(t, err, "task should pass")
	assertContains(t, out, "go version go", "output should contain prefix of version report")
}

func TestCmd_error(t *testing.T) {
	taskName := "exec"
	flow := &goyek.Flow{Output: &strings.Builder{}}
	flow.Define(goyek.Task{
		Name: taskName,
		Action: func(tf *goyek.TF) {
			if err := tf.Cmd("go", "wrong").Run(); err != nil {
				tf.Fatal(err)
			}
		},
	})

	err := flow.Execute(nil, taskName) //nolint:staticcheck // present that nil context is handled

	assertFail(t, err, "task should pass")
}

func TestFlow_Tasks(t *testing.T) {
	flow := &goyek.Flow{Output: &strings.Builder{}}
	t1 := flow.Define(goyek.Task{Name: "one"})
	flow.Define(goyek.Task{Name: "two", Usage: "action", Deps: goyek.Deps{t1}})
	flow.Define(goyek.Task{Name: "three"})

	got := flow.Tasks()

	assertEqual(t, len(got), 3, "should return all tasks")
	assertEqual(t, got[0].Name(), "one", "should first return one")
	assertEqual(t, got[1].Name(), "three", "should then return one (sorted)")
	assertEqual(t, got[2].Name(), "two", "should next return two")
	assertEqual(t, got[2].Usage(), "action", "should return usage")
	assertEqual(t, got[2].Deps()[0], "one", "should return dependency")
}

func TestFlow_Default(t *testing.T) {
	flow := &goyek.Flow{Output: &strings.Builder{}}
	task := flow.Define(goyek.Task{Name: "name"})
	flow.SetDefault(task)

	got := flow.Default()

	assertEqual(t, got.Name(), "name", "should return the default task")
}

func TestFlow_Default_empty(t *testing.T) {
	flow := &goyek.Flow{Output: &strings.Builder{}}
	got := flow.Default()

	assertEqual(t, got, nil, "should return nil")
}

func TestFlow_Print(t *testing.T) {
	out := &strings.Builder{}
	flow := &goyek.Flow{Output: out}
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
	flow := &goyek.Flow{Output: out, Logger: goyek.FmtLogger{}}
	flow.Define(goyek.Task{
		Name: "task",
		Action: func(tf *goyek.TF) {
			tf.Log("first")
			tf.Logf("second")
		},
	})

	_ = flow.Execute(context.Background(), "task")

	assertContains(t, out, "first", "should call Log")
	assertContains(t, out, "second", "should call Logf")
}

func TestFlow_Use(t *testing.T) {
	out := &strings.Builder{}
	flow := &goyek.Flow{Output: out}
	flow.Define(goyek.Task{
		Name: "task",
	})
	flow.Use(func(next goyek.Runner) goyek.Runner {
		return func(i goyek.Input) goyek.Result {
			i.Output.Write([]byte("message")) //nolint:errcheck,gosec // not checking errors when writing to output
			return goyek.Result{}
		}
	})

	_ = flow.Execute(context.Background(), "task")

	assertContains(t, out, "message", "should call middleware with proper input")
}

func TestFlow_Use_nil_middleware(t *testing.T) {
	flow := &goyek.Flow{}

	act := func() {
		flow.Use(nil)
	}

	assertPanics(t, act, "should panic on nil middleware")
}

func assertTrue(tb testing.TB, got bool, msg string) {
	tb.Helper()
	if got {
		return
	}
	tb.Errorf("%s\nGOT: %v, WANT: true", msg, got)
}

func assertContains(tb testing.TB, got fmt.Stringer, want string, msg string) {
	tb.Helper()
	gotTxt := got.String()
	if strings.Contains(gotTxt, want) {
		return
	}
	tb.Errorf("%s\nGOT:\n%s\nSHOULD CONTAIN:\n%s", msg, gotTxt, want)
}

func assertNotContains(tb testing.TB, got fmt.Stringer, want string, msg string) {
	tb.Helper()
	gotTxt := got.String()
	if !strings.Contains(gotTxt, want) {
		return
	}
	tb.Errorf("%s\nGOT:\n%s\nSHOULD NOT CONTAIN:\n%s", msg, gotTxt, want)
}

func requireEqual(tb testing.TB, got interface{}, want interface{}, msg string) {
	tb.Helper()
	if reflect.DeepEqual(got, want) {
		return
	}
	tb.Fatalf("%s\nGOT: %v\nWANT: %v", msg, got, want)
}

func assertEqual(tb testing.TB, got interface{}, want interface{}, msg string) {
	tb.Helper()
	if reflect.DeepEqual(got, want) {
		return
	}
	tb.Errorf("%s\nGOT: %v\nWANT: %v", msg, got, want)
}

func assertPass(tb testing.TB, got error, msg string) {
	tb.Helper()
	if got != nil {
		tb.Errorf("%s\nGOT: %v\nWANT: <PASS>", msg, got)
	}
}

func assertFail(tb testing.TB, got error, msg string) {
	tb.Helper()
	if _, ok := got.(*goyek.FailError); !ok {
		tb.Errorf("%s\nGOT: %v\nWANT: <FAIL>", msg, got)
	}
}

func assertInvalid(tb testing.TB, got error, msg string) {
	tb.Helper()
	if _, ok := got.(*goyek.FailError); ok || got == nil {
		tb.Errorf("%s\nGOT: %v\nWANT: <INVALID>", msg, got)
	}
}

func assertPanics(tb testing.TB, fn func(), msg string) {
	tb.Helper()
	tryPanic := func() bool {
		didPanic := false
		func() {
			defer func() {
				if info := recover(); info != nil {
					didPanic = true
				}
			}()
			fn()
		}()
		return didPanic
	}

	if tryPanic() {
		return
	}
	tb.Errorf("%s\ndid not panic, but expected to do so", msg)
}
