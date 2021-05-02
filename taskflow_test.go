package taskflow_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/pellared/taskflow"
)

func init() {
	taskflow.DefaultOutput = ioutil.Discard
}

func Test_Register_errors(t *testing.T) {
	testCases := []struct {
		desc string
		task taskflow.Task
	}{
		{
			desc: "missing task name",
			task: taskflow.Task{},
		},
		{
			desc: "invalid dependency",
			task: taskflow.Task{Name: "my-task", Deps: taskflow.Deps{taskflow.RegisteredTask{}}},
		},
		{
			desc: "invalid task name",
			task: taskflow.Task{Name: "-flag"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			flow := taskflow.New()

			act := func() { flow.Register(tc.task) }

			assertPanics(t, act, "should panic")
		})
	}
}

func Test_Register_same_name(t *testing.T) {
	flow := &taskflow.Taskflow{}
	task := taskflow.Task{Name: "task"}
	flow.Register(task)

	act := func() { flow.Register(task) }

	assertPanics(t, act, "should not be possible to register tasks with same name twice")
}

func Test_successful(t *testing.T) {
	ctx := context.Background()
	flow := &taskflow.Taskflow{}
	var executed1 int
	task1 := flow.Register(taskflow.Task{
		Name: "task-1",
		Command: func(*taskflow.TF) {
			executed1++
		},
	})
	var executed2 int
	flow.Register(taskflow.Task{
		Name: "task-2",
		Command: func(*taskflow.TF) {
			executed2++
		},
		Deps: taskflow.Deps{task1},
	})
	var executed3 int
	flow.Register(taskflow.Task{
		Name: "task-3",
		Command: func(*taskflow.TF) {
			executed3++
		},
		Deps: taskflow.Deps{task1},
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
	flow := &taskflow.Taskflow{}
	var executed1 int
	task1 := flow.Register(taskflow.Task{
		Name: "task-1",
		Command: func(tf *taskflow.TF) {
			executed1++
			tf.Error("it still runs")
			executed1 += 10
			tf.FailNow()
			executed1 += 100
		},
	})
	var executed2 int
	flow.Register(taskflow.Task{
		Name: "task-2",
		Command: func(*taskflow.TF) {
			executed2++
		},
		Deps: taskflow.Deps{task1},
	})
	var executed3 int
	flow.Register(taskflow.Task{
		Name: "task-3",
		Command: func(*taskflow.TF) {
			executed3++
		},
		Deps: taskflow.Deps{task1},
	})
	got := func() []int {
		return []int{executed1, executed2, executed3}
	}

	exitCode := flow.Run(context.Background(), "task-2", "task-3")

	assertEqual(t, exitCode, 1, "should return error from first task")
	assertEqual(t, got(), []int{11, 0, 0}, "should execute task 1")
}

func Test_fail(t *testing.T) {
	flow := &taskflow.Taskflow{}
	failed := false
	flow.Register(taskflow.Task{
		Name: "task",
		Command: func(tf *taskflow.TF) {
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
	flow := &taskflow.Taskflow{}
	skipped := false
	flow.Register(taskflow.Task{
		Name: "task",
		Command: func(tf *taskflow.TF) {
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
	flow := &taskflow.Taskflow{}
	flow.Register(taskflow.Task{
		Name: "task",
		Command: func(tf *taskflow.TF) {
			panic("panicked!")
		},
	})

	exitCode := flow.Run(context.Background(), "task")

	assertEqual(t, exitCode, 1, "should return error from first task")
}

func Test_cancelation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	flow := &taskflow.Taskflow{}
	flow.Register(taskflow.Task{
		Name: "task",
	})

	exitCode := flow.Run(ctx, "task")

	assertEqual(t, exitCode, 1, "should return error canceled")
}

func Test_cancelation_during_last_task(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	flow := &taskflow.Taskflow{}
	flow.Register(taskflow.Task{
		Name: "task",
		Command: func(tf *taskflow.TF) {
			cancel()
		},
	})

	exitCode := flow.Run(ctx, "task")

	assertEqual(t, exitCode, 1, "should return error canceled")
}

func Test_empty_command(t *testing.T) {
	flow := &taskflow.Taskflow{}
	flow.Register(taskflow.Task{
		Name: "task",
	})

	exitCode := flow.Run(context.Background(), "task")

	assertEqual(t, exitCode, 0, "should pass")
}

func Test_invalid_args(t *testing.T) {
	flow := taskflow.New()
	flow.Register(taskflow.Task{
		Name: "task",
	})

	testCases := []struct {
		desc string
		args []string
	}{
		{
			desc: "missing task name",
		},
		{
			desc: "bad flag",
			args: []string{"-badflag", "task"},
		},
		{
			desc: "bad task name",
			args: []string{"badtask"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			exitCode := flow.Run(context.Background(), tc.args...)

			assertEqual(t, exitCode, 2, "should return error bad args")
		})
	}
}

func Test_help(t *testing.T) {
	flow := taskflow.New()
	fastParam := flow.RegisterBoolParam(false, taskflow.ParamInfo{
		Name:  "fast",
		Usage: "simulates fast-lane processing",
	})
	a := flow.Register(taskflow.Task{
		Name:   "a",
		Params: taskflow.Params{fastParam},
		Usage:  "some task",
	})
	flow.DefaultTask = a

	exitCode := flow.Run(context.Background(), "-h")

	assertEqual(t, exitCode, taskflow.CodePass, "should return OK")
}

func Test_printing(t *testing.T) {
	sb := &strings.Builder{}
	flow := &taskflow.Taskflow{
		Output: sb,
	}
	skipped := flow.Register(taskflow.Task{
		Name: "skipped",
		Command: func(tf *taskflow.TF) {
			tf.Skipf("Skipf %d", 0)
		},
	})
	flow.Register(taskflow.Task{
		Name: "failing",
		Deps: taskflow.Deps{skipped},
		Command: func(tf *taskflow.TF) {
			tf.Log("Log", 1)
			tf.Logf("Logf %d", 2)
			tf.Error("Error", 3)
			tf.Errorf("Errorf %d", 4)
			tf.Fatalf("Fatalf %d", 5)
		},
	})
	t.Log()

	flow.Run(context.Background(), "-v", "failing")

	assertContains(t, sb.String(), "Skipf 0", "should contain proper output from \"skipped\" task")
	assertContains(t, sb.String(), `Log 1
Logf 2
Error 3
Errorf 4
Fatalf 5`, "should contain proper output from \"failing\" task")
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
			flow := taskflow.Taskflow{
				Output: sb,
			}
			flow.Register(taskflow.Task{
				Name: "task",
				Command: func(tf *taskflow.TF) {
					ch := make(chan struct{})
					go func() {
						defer func() { ch <- struct{}{} }()
						tf.Log("from child goroutine")
					}()
					tf.Error("from main goroutine")
					<-ch
				},
			})

			var args []string
			if tc.verbose {
				args = append(args, "-v")
			}
			args = append(args, "task")
			exitCode := flow.Run(context.Background(), args...)

			assertEqual(t, exitCode, taskflow.CodeFailure, "should fail")
			assertContains(t, sb.String(), "from child goroutine", "should contain log from child goroutine")
			assertContains(t, sb.String(), "from main goroutine", "should contain log from main goroutine")
		})
	}
}

func Test_name(t *testing.T) {
	flow := &taskflow.Taskflow{}
	taskName := "my-named-task"
	var got string
	flow.Register(taskflow.Task{
		Name: taskName,
		Command: func(tf *taskflow.TF) {
			got = tf.Name()
		},
	})

	exitCode := flow.Run(context.Background(), taskName)

	assertEqual(t, exitCode, 0, "should pass")
	assertEqual(t, got, taskName, "should return proper Name value")
}

type arrayValue []string

func (value *arrayValue) Set(s string) error {
	return json.Unmarshal([]byte(s), value)
}

func (value *arrayValue) Get() interface{} { return []string(*value) }

func (value *arrayValue) String() string {
	b, _ := json.Marshal(value)
	return string(b)
}

func (value *arrayValue) IsBool() bool { return false }

func Test_params(t *testing.T) {
	flow := taskflow.New()
	boolParam := flow.RegisterBoolParam(true, taskflow.ParamInfo{
		Name: "b",
	})
	intParam := flow.RegisterIntParam(1, taskflow.ParamInfo{
		Name: "i",
	})
	stringParam := flow.RegisterStringParam("abc", taskflow.ParamInfo{
		Name: "s",
	})
	arrayParam := flow.RegisterValueParam(func() taskflow.ParamValue { return &arrayValue{} }, taskflow.ParamInfo{
		Name: "array",
	})
	var gotBool bool
	var gotInt int
	var gotString string
	var gotArray []string
	flow.Register(taskflow.Task{
		Name: "task",
		Params: taskflow.Params{
			boolParam,
			intParam,
			stringParam,
			arrayParam,
		},
		Command: func(tf *taskflow.TF) {
			gotBool = boolParam.Get(tf)
			gotInt = intParam.Get(tf)
			gotString = stringParam.Get(tf)
			gotArray = arrayParam.Get(tf).([]string)
		},
	})

	exitCode := flow.Run(context.Background(), "-b=false", "-i", "9001", "-s", "xyz", "-array", "[\"a\", \"b\"]", "task")

	assertEqual(t, exitCode, 0, "should pass")
	assertEqual(t, gotBool, false, "bool param")
	assertEqual(t, gotInt, 9001, "int param")
	assertEqual(t, gotString, "xyz", "string param")
	assertEqual(t, gotArray, []string{"a", "b"}, "array param")
}

func Test_invalid_params(t *testing.T) {
	flow := taskflow.New()
	flow.Register(taskflow.Task{
		Name:    "task",
		Command: func(tf *taskflow.TF) {},
	})

	exitCode := flow.Run(context.Background(), "-z=3", "task")

	assertEqual(t, exitCode, taskflow.CodeInvalidArgs, "should fail because of unknown parameter")
}

func Test_unused_params(t *testing.T) {
	flow := taskflow.New()
	flow.DefaultTask = flow.Register(taskflow.Task{Name: "task", Command: func(tf *taskflow.TF) {}})
	flow.RegisterBoolParam(false, taskflow.ParamInfo{Name: "unused"})

	assertPanics(t, func() { flow.Run(context.Background()) }, "should fail because of unused parameter")
}

func Test_param_registration_error_empty_name(t *testing.T) {
	flow := taskflow.New()
	assertPanics(t, func() { flow.RegisterBoolParam(false, taskflow.ParamInfo{Name: ""}) }, "empty name")
}

func Test_param_registration_error_double_name(t *testing.T) {
	flow := taskflow.New()
	info := taskflow.ParamInfo{Name: "double"}
	flow.RegisterBoolParam(false, info)
	assertPanics(t, func() { flow.RegisterBoolParam(false, info) }, "double name")
}

func Test_unregistered_params(t *testing.T) {
	foreignParam := taskflow.New().RegisterBoolParam(false, taskflow.ParamInfo{Name: "foreign"})
	flow := taskflow.New()
	flow.Register(taskflow.Task{
		Name: "task",
		Command: func(tf *taskflow.TF) {
			foreignParam.Get(tf)
		},
	})

	exitCode := flow.Run(context.Background(), "task")

	assertEqual(t, taskflow.CodeFailure, exitCode, "should fail because of unregistered parameter")
}

func Test_defaultTask(t *testing.T) {
	flow := taskflow.New()
	taskRan := false
	task := flow.Register(taskflow.Task{
		Name: "task",
		Command: func(tf *taskflow.TF) {
			taskRan = true
		},
	})
	flow.DefaultTask = task

	exitCode := flow.Run(context.Background())

	assertEqual(t, exitCode, 0, "should pass")
	assertTrue(t, taskRan, "task should have run")
}
