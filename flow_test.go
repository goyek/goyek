package goyek_test

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/goyek/goyek"
)

func Test_Register_empty_name(t *testing.T) {
	flow := &goyek.Flow{Output: &strings.Builder{}}

	act := func() { flow.Register(goyek.Task{}) }

	assertPanics(t, act, "should panic")
}

func Test_Register_same_name(t *testing.T) {
	flow := &goyek.Flow{Output: &strings.Builder{}}
	task := goyek.Task{Name: "task"}
	flow.Register(task)

	act := func() { flow.Register(task) }

	assertPanics(t, act, "should not be possible to register tasks with same name twice")
}

func Test_successful(t *testing.T) {
	ctx := context.Background()
	flow := &goyek.Flow{Output: &strings.Builder{}}
	var executed1 int
	task1 := flow.Register(goyek.Task{
		Name: "task-1",
		Action: func(*goyek.TF) {
			executed1++
		},
	})
	var executed2 int
	flow.Register(goyek.Task{
		Name: "task-2",
		Action: func(*goyek.TF) {
			executed2++
		},
		Deps: goyek.Deps{task1},
	})
	var executed3 int
	flow.Register(goyek.Task{
		Name: "task-3",
		Action: func(*goyek.TF) {
			executed3++
		},
		Deps: goyek.Deps{task1},
	})
	got := func() []int {
		return []int{executed1, executed2, executed3}
	}

	exitCode := flow.Run(ctx, "task-1")
	requireEqual(t, exitCode, 0, "first execution should pass")
	requireEqual(t, got(), []int{1, 0, 0}, "should execute task 1")

	exitCode = flow.Run(ctx, "task-2")
	requireEqual(t, exitCode, 0, "second execution should pass")
	requireEqual(t, got(), []int{2, 1, 0}, "should execute task 1 and 2")

	exitCode = flow.Run(ctx, "task-1", "task-2", "task-3")
	requireEqual(t, exitCode, 0, "third execution should pass")
	requireEqual(t, got(), []int{3, 2, 1}, "should execute task 1 and 2 and 3")
}

func Test_dependency_failure(t *testing.T) {
	flow := &goyek.Flow{Output: &strings.Builder{}}
	var executed1 int
	task1 := flow.Register(goyek.Task{
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
	flow.Register(goyek.Task{
		Name: "task-2",
		Action: func(*goyek.TF) {
			executed2++
		},
		Deps: goyek.Deps{task1},
	})
	var executed3 int
	flow.Register(goyek.Task{
		Name: "task-3",
		Action: func(*goyek.TF) {
			executed3++
		},
		Deps: goyek.Deps{task1},
	})
	got := func() []int {
		return []int{executed1, executed2, executed3}
	}

	exitCode := flow.Run(context.Background(), "task-2", "task-3")

	assertEqual(t, exitCode, 1, "should return error from first task")
	assertEqual(t, got(), []int{11, 0, 0}, "should execute task 1")
}

func Test_fail(t *testing.T) {
	flow := &goyek.Flow{Output: &strings.Builder{}}
	failed := false
	flow.Register(goyek.Task{
		Name: "task",
		Action: func(tf *goyek.TF) {
			defer func() {
				failed = tf.Failed()
			}()
			tf.Fatal("failing")
		},
	})

	exitCode := flow.Run(context.Background(), "task")

	assertEqual(t, exitCode, 1, "should return error")
	assertTrue(t, failed, "tf.Failed() should return true")
}

func Test_skip(t *testing.T) {
	flow := &goyek.Flow{Output: &strings.Builder{}}
	skipped := false
	flow.Register(goyek.Task{
		Name: "task",
		Action: func(tf *goyek.TF) {
			defer func() {
				skipped = tf.Skipped()
			}()
			tf.Skip("skipping")
		},
	})

	exitCode := flow.Run(context.Background(), "task")

	assertEqual(t, exitCode, 0, "should pass")
	assertTrue(t, skipped, "tf.Skipped() should return true")
}

func Test_task_panics(t *testing.T) {
	flow := &goyek.Flow{Output: &strings.Builder{}}
	flow.Register(goyek.Task{
		Name: "task",
		Action: func(tf *goyek.TF) {
			panic("panicked!")
		},
	})

	exitCode := flow.Run(context.Background(), "task")

	assertEqual(t, exitCode, 1, "should return error from first task")
}

func Test_cancelation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	flow := &goyek.Flow{Output: &strings.Builder{}}
	flow.Register(goyek.Task{
		Name: "task",
	})

	exitCode := flow.Run(ctx, "task")

	assertEqual(t, exitCode, 1, "should return error canceled")
}

func Test_cancelation_during_last_task(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	flow := &goyek.Flow{Output: &strings.Builder{}}
	flow.Register(goyek.Task{
		Name: "task",
		Action: func(tf *goyek.TF) {
			cancel()
		},
	})

	exitCode := flow.Run(ctx, "task")

	assertEqual(t, exitCode, 0, "should pass as the flow completed")
}

func Test_empty_action(t *testing.T) {
	flow := &goyek.Flow{Output: &strings.Builder{}}
	flow.Register(goyek.Task{
		Name: "task",
	})

	exitCode := flow.Run(context.Background(), "task")

	assertEqual(t, exitCode, 0, "should pass")
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
			exitCode := flow.Run(context.Background(), tc.args...)

			assertEqual(t, exitCode, 2, "should return error bad args")
		})
	}
}

func Test_printing(t *testing.T) {
	sb := &strings.Builder{}
	flow := &goyek.Flow{
		Output:  sb,
		Verbose: true,
	}
	skipped := flow.Register(goyek.Task{
		Name: "skipped",
		Action: func(tf *goyek.TF) {
			tf.Skipf("Skipf %d", 0)
		},
	})
	flow.Register(goyek.Task{
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

	flow.Run(context.Background(), "failing")

	assertContains(t, sb.String(), "Skipf 0", "should contain proper output from \"skipped\" task")
	assertContains(t, sb.String(), "Fatalf 5", "should contain proper output from \"failing\" task")
}

func Test_concurrent_printing(t *testing.T) {
	testCases := []struct {
		verbose bool
	}{
		{verbose: false},
		{verbose: true},
	}
	for _, tc := range testCases {
		testName := fmt.Sprintf("Verbose:%v", tc.verbose)
		t.Run(testName, func(t *testing.T) {
			sb := &strings.Builder{}
			flow := goyek.Flow{
				Output:  sb,
				Verbose: tc.verbose,
			}
			flow.Register(goyek.Task{
				Name: "task",
				Action: func(tf *goyek.TF) {
					ch := make(chan struct{})
					go func() {
						defer func() { ch <- struct{}{} }()
						tf.Log("from child goroutine")
					}()
					tf.Error("from main goroutine")
					<-ch
				},
			})

			exitCode := flow.Run(context.Background(), "task")

			assertEqual(t, exitCode, goyek.CodeFail, "should fail")
			assertContains(t, sb.String(), "from child goroutine", "should contain log from child goroutine")
			assertContains(t, sb.String(), "from main goroutine", "should contain log from main goroutine")
		})
	}
}

func Test_name(t *testing.T) {
	flow := &goyek.Flow{Output: &strings.Builder{}}
	taskName := "my-named-task"
	var got string
	flow.Register(goyek.Task{
		Name: taskName,
		Action: func(tf *goyek.TF) {
			got = tf.Name()
		},
	})

	exitCode := flow.Run(context.Background(), taskName)

	assertEqual(t, exitCode, 0, "should pass")
	assertEqual(t, got, taskName, "should return proper Name value")
}

func Test_defaultTask(t *testing.T) {
	flow := &goyek.Flow{Output: &strings.Builder{}}
	taskRan := false
	task := flow.Register(goyek.Task{
		Name: "task",
		Action: func(tf *goyek.TF) {
			taskRan = true
		},
	})
	flow.DefaultTask = task

	exitCode := flow.Run(context.Background())

	assertEqual(t, exitCode, 0, "should pass")
	assertTrue(t, taskRan, "task should have run")
}

func TestCmd_success(t *testing.T) {
	taskName := "exec"
	sb := &strings.Builder{}
	flow := &goyek.Flow{
		Output:  sb,
		Verbose: true,
	}
	flow.Register(goyek.Task{
		Name: taskName,
		Action: func(tf *goyek.TF) {
			if err := tf.Cmd("go", "version").Run(); err != nil {
				tf.Fatal(err)
			}
		},
	})

	exitCode := flow.Run(context.Background(), taskName)

	assertContains(t, sb.String(), "go version go", "output should contain prefix of version report")
	assertEqual(t, exitCode, goyek.CodePass, "task should pass")
}

func TestCmd_error(t *testing.T) {
	taskName := "exec"
	flow := &goyek.Flow{Output: &strings.Builder{}}
	flow.Register(goyek.Task{
		Name: taskName,
		Action: func(tf *goyek.TF) {
			if err := tf.Cmd("go", "wrong").Run(); err != nil {
				tf.Fatal(err)
			}
		},
	})

	exitCode := flow.Run(nil, taskName) //nolint:staticcheck // present that nil context is handled

	assertEqual(t, exitCode, goyek.CodeFail, "task should pass")
}

func TestFlow_Tasks(t *testing.T) {
	flow := &goyek.Flow{Output: &strings.Builder{}}
	t1 := flow.Register(goyek.Task{Name: "one"})
	flow.Register(goyek.Task{Name: "two", Usage: "action", Deps: goyek.Deps{t1}})
	flow.Register(goyek.Task{Name: "three"})

	got := flow.Tasks()

	assertEqual(t, len(got), 3, "should return all tasks")
	assertEqual(t, got[0].Name(), "one", "should first return one")
	assertEqual(t, got[1].Name(), "three", "should then return one (sorted)")
	assertEqual(t, got[2].Name(), "two", "should next return two")
	assertEqual(t, got[2].Usage(), "action", "should return usage")
	assertEqual(t, got[2].Deps()[0], "one", "should return dependency")
}

func assertTrue(tb testing.TB, got bool, msg string) {
	tb.Helper()
	if got {
		return
	}
	tb.Errorf("%s\ngot: [%v], want: [true]", msg, got)
}

func assertContains(tb testing.TB, got string, want string, msg string) {
	tb.Helper()
	if strings.Contains(got, want) {
		return
	}
	tb.Errorf("%s\ngot: [%s], should contain: [%s]", msg, got, want)
}

func requireEqual(tb testing.TB, got interface{}, want interface{}, msg string) {
	tb.Helper()
	if reflect.DeepEqual(got, want) {
		return
	}
	tb.Fatalf("%s\ngot: [%v], want: [%v]", msg, got, want)
}

func assertEqual(tb testing.TB, got interface{}, want interface{}, msg string) {
	tb.Helper()
	if reflect.DeepEqual(got, want) {
		return
	}
	tb.Errorf("%s\ngot: [%v], want: [%v]", msg, got, want)
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
