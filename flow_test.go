package goyek_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/goyek/goyek"
)

func init() {
	goyek.DefaultOutput = ioutil.Discard
}

func Test_Register_errors(t *testing.T) {
	testCases := []struct {
		desc string
		task goyek.Task
	}{
		{
			desc: "missing task name",
			task: goyek.Task{},
		},
		{
			desc: "invalid dependency",
			task: goyek.Task{Name: "my-task", Deps: goyek.Deps{goyek.RegisteredTask{}}},
		},
		{
			desc: "invalid task name",
			task: goyek.Task{Name: "-flag"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			flow := &goyek.Flow{}

			act := func() { flow.Register(tc.task) }

			assertPanics(t, act, "should panic")
		})
	}
}

func Test_Register_same_name(t *testing.T) {
	flow := &goyek.Flow{}
	task := goyek.Task{Name: "task"}
	flow.Register(task)

	act := func() { flow.Register(task) }

	assertPanics(t, act, "should not be possible to register tasks with same name twice")
}

func Test_successful(t *testing.T) {
	ctx := context.Background()
	flow := &goyek.Flow{}
	var executed1 int
	task1 := flow.Register(goyek.Task{
		Name: "task-1",
		Action: func(*goyek.Progress) {
			executed1++
		},
	})
	var executed2 int
	flow.Register(goyek.Task{
		Name: "task-2",
		Action: func(*goyek.Progress) {
			executed2++
		},
		Deps: goyek.Deps{task1},
	})
	var executed3 int
	flow.Register(goyek.Task{
		Name: "task-3",
		Action: func(*goyek.Progress) {
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
	flow := &goyek.Flow{}
	var executed1 int
	task1 := flow.Register(goyek.Task{
		Name: "task-1",
		Action: func(p *goyek.Progress) {
			executed1++
			p.Error("it still runs")
			executed1 += 10
			p.FailNow()
			executed1 += 100
		},
	})
	var executed2 int
	flow.Register(goyek.Task{
		Name: "task-2",
		Action: func(*goyek.Progress) {
			executed2++
		},
		Deps: goyek.Deps{task1},
	})
	var executed3 int
	flow.Register(goyek.Task{
		Name: "task-3",
		Action: func(*goyek.Progress) {
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
	flow := &goyek.Flow{}
	failed := false
	flow.Register(goyek.Task{
		Name: "task",
		Action: func(p *goyek.Progress) {
			defer func() {
				failed = p.Failed()
			}()
			p.Fatal("failing")
		},
	})

	exitCode := flow.Run(context.Background(), "task")

	assertEqual(t, exitCode, 1, "should return error")
	assertTrue(t, failed, "p.Failed() should return true")
}

func Test_skip(t *testing.T) {
	flow := &goyek.Flow{}
	skipped := false
	flow.Register(goyek.Task{
		Name: "task",
		Action: func(p *goyek.Progress) {
			defer func() {
				skipped = p.Skipped()
			}()
			p.Skip("skipping")
		},
	})

	exitCode := flow.Run(context.Background(), "task")

	assertEqual(t, exitCode, 0, "should pass")
	assertTrue(t, skipped, "p.Skipped() should return true")
}

func Test_task_panics(t *testing.T) {
	flow := &goyek.Flow{}
	flow.Register(goyek.Task{
		Name: "task",
		Action: func(p *goyek.Progress) {
			panic("panicked!")
		},
	})

	exitCode := flow.Run(context.Background(), "task")

	assertEqual(t, exitCode, 1, "should return error from first task")
}

func Test_cancelation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	flow := &goyek.Flow{}
	flow.Register(goyek.Task{
		Name: "task",
	})

	exitCode := flow.Run(ctx, "task")

	assertEqual(t, exitCode, 1, "should return error canceled")
}

func Test_cancelation_during_last_task(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	flow := &goyek.Flow{}
	flow.Register(goyek.Task{
		Name: "task",
		Action: func(p *goyek.Progress) {
			cancel()
		},
	})

	exitCode := flow.Run(ctx, "task")

	assertEqual(t, exitCode, 1, "should return error canceled")
}

func Test_empty_action(t *testing.T) {
	flow := &goyek.Flow{}
	flow.Register(goyek.Task{
		Name: "task",
	})

	exitCode := flow.Run(context.Background(), "task")

	assertEqual(t, exitCode, 0, "should pass")
}

func Test_invalid_args(t *testing.T) {
	flow := &goyek.Flow{}
	flow.Register(goyek.Task{
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
	flow := &goyek.Flow{}
	fastParam := flow.RegisterBoolParam(goyek.BoolParam{
		Name:  "fast",
		Usage: "simulates fast-lane processing",
	})
	a := flow.Register(goyek.Task{
		Name:   "a",
		Params: goyek.Params{fastParam},
		Usage:  "some task",
	})
	flow.DefaultTask = a

	exitCode := flow.Run(context.Background(), "-h")

	assertEqual(t, exitCode, goyek.CodePass, "should return OK")
}

func Test_printing(t *testing.T) {
	sb := &strings.Builder{}
	flow := &goyek.Flow{
		Output: sb,
	}
	skipped := flow.Register(goyek.Task{
		Name: "skipped",
		Action: func(p *goyek.Progress) {
			p.Skipf("Skipf %d", 0)
		},
	})
	flow.Register(goyek.Task{
		Name: "failing",
		Deps: goyek.Deps{skipped},
		Action: func(p *goyek.Progress) {
			p.Log("Log", 1)
			p.Logf("Logf %d", 2)
			p.Error("Error", 3)
			p.Errorf("Errorf %d", 4)
			p.Fatalf("Fatalf %d", 5)
		},
	})

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
			flow := goyek.Flow{
				Output: sb,
			}
			flow.Register(goyek.Task{
				Name: "task",
				Action: func(p *goyek.Progress) {
					ch := make(chan struct{})
					go func() {
						defer func() { ch <- struct{}{} }()
						p.Log("from child goroutine")
					}()
					p.Error("from main goroutine")
					<-ch
				},
			})

			var args []string
			if tc.verbose {
				args = append(args, "-v")
			}
			args = append(args, "task")
			exitCode := flow.Run(context.Background(), args...)

			assertEqual(t, exitCode, goyek.CodeFail, "should fail")
			assertContains(t, sb.String(), "from child goroutine", "should contain log from child goroutine")
			assertContains(t, sb.String(), "from main goroutine", "should contain log from main goroutine")
		})
	}
}

func Test_name(t *testing.T) {
	flow := &goyek.Flow{}
	taskName := "my-named-task"
	var got string
	flow.Register(goyek.Task{
		Name: taskName,
		Action: func(p *goyek.Progress) {
			got = p.Name()
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
	flow := &goyek.Flow{}
	boolParam := flow.RegisterBoolParam(goyek.BoolParam{
		Name:    "b",
		Default: true,
	})
	intParam := flow.RegisterIntParam(goyek.IntParam{
		Name:    "i",
		Default: 1,
	})
	stringParam := flow.RegisterStringParam(goyek.StringParam{
		Name:    "s",
		Default: "abc",
	})
	arrayParam := flow.RegisterValueParam(goyek.ValueParam{
		Name:     "array",
		NewValue: func() goyek.ParamValue { return &arrayValue{} },
	})
	var gotBool bool
	var gotInt int
	var gotString string
	var gotArray []string
	flow.Register(goyek.Task{
		Name: "task",
		Params: goyek.Params{
			boolParam,
			intParam,
			stringParam,
			arrayParam,
		},
		Action: func(p *goyek.Progress) {
			gotBool = boolParam.Get(p)
			gotInt = intParam.Get(p)
			gotString = stringParam.Get(p)
			gotArray = arrayParam.Get(p).([]string) //nolint:forcetypeassert // test code, it can panic
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
	flow := &goyek.Flow{}
	flow.Register(goyek.Task{
		Name:   "task",
		Action: func(p *goyek.Progress) {},
	})

	exitCode := flow.Run(context.Background(), "-z=3", "task")

	assertEqual(t, exitCode, goyek.CodeInvalidArgs, "should fail because of unknown parameter")
}

func Test_unused_params(t *testing.T) {
	flow := &goyek.Flow{}
	flow.DefaultTask = flow.Register(goyek.Task{Name: "task", Action: func(p *goyek.Progress) {}})
	flow.RegisterBoolParam(goyek.BoolParam{Name: "unused"})

	assertPanics(t, func() { flow.Run(context.Background()) }, "should fail because of unused parameter")
}

func Test_param_registration_error_empty_name(t *testing.T) {
	flow := &goyek.Flow{}
	assertPanics(t, func() { flow.RegisterBoolParam(goyek.BoolParam{Name: ""}) }, "empty name")
}

func Test_param_registration_error_underscore_name_start(t *testing.T) {
	flow := &goyek.Flow{}
	assertPanics(t, func() { flow.RegisterBoolParam(goyek.BoolParam{Name: "_reserved"}) }, "should not start with underscore")
}

func Test_param_registration_error_no_default(t *testing.T) {
	flow := &goyek.Flow{}
	assertPanics(t, func() { flow.RegisterValueParam(goyek.ValueParam{Name: "custom"}) }, "custom parameter must have default value factory")
}

func Test_param_registration_error_double_name(t *testing.T) {
	flow := &goyek.Flow{}
	name := "double"
	flow.RegisterStringParam(goyek.StringParam{Name: name})
	assertPanics(t, func() { flow.RegisterBoolParam(goyek.BoolParam{Name: name}) }, "double name")
}

func Test_unregistered_params(t *testing.T) {
	foreignParam := (&goyek.Flow{}).RegisterBoolParam(goyek.BoolParam{Name: "foreign"})
	flow := &goyek.Flow{}
	flow.Register(goyek.Task{
		Name: "task",
		Action: func(p *goyek.Progress) {
			foreignParam.Get(p)
		},
	})

	exitCode := flow.Run(context.Background(), "task")

	assertEqual(t, goyek.CodeFail, exitCode, "should fail because of unregistered parameter")
}

func Test_defaultTask(t *testing.T) {
	flow := &goyek.Flow{}
	taskRan := false
	task := flow.Register(goyek.Task{
		Name: "task",
		Action: func(p *goyek.Progress) {
			taskRan = true
		},
	})
	flow.DefaultTask = task

	exitCode := flow.Run(context.Background())

	assertEqual(t, exitCode, 0, "should pass")
	assertTrue(t, taskRan, "task should have run")
}

func Test_wd_param(t *testing.T) {
	flow := &goyek.Flow{}
	beforeDir, err := os.Getwd()
	requireEqual(t, err, nil, "should get work dir before the flow")
	dir, cleanup := tempDir(t)
	defer cleanup()
	var got string
	flow.Register(goyek.Task{
		Name: "task",
		Action: func(p *goyek.Progress) {
			var osErr error
			got, osErr = os.Getwd()
			requireEqual(t, osErr, nil, "should get work dir from task")
		},
	})

	exitCode := flow.Run(context.Background(), "task", "-wd", dir)
	afterDir, err := os.Getwd()
	requireEqual(t, err, nil, "should get work dir after the flow")

	assertEqual(t, exitCode, 0, "should pass")
	assertEqual(t, got, dir, "should have changed the working directory in flow")
	assertEqual(t, afterDir, beforeDir, "should change back the working directory after flow")
}

func Test_wd_param_invalid(t *testing.T) {
	flow := &goyek.Flow{}
	beforeDir, err := os.Getwd()
	requireEqual(t, err, nil, "should get work dir before the flow")
	taskRan := false
	flow.Register(goyek.Task{
		Name: "task",
		Action: func(p *goyek.Progress) {
			taskRan = true
		},
	})

	exitCode := flow.Run(context.Background(), "task", "-wd=strange-dir")
	afterDir, err := os.Getwd()
	requireEqual(t, err, nil, "should get work dir after the flow")

	assertEqual(t, exitCode, goyek.CodeInvalidArgs, "should not proceed")
	assertEqual(t, taskRan, false, "should not run the task")
	assertEqual(t, afterDir, beforeDir, "should change back the working directory after flow")
}

func tempDir(t *testing.T) (string, func()) {
	t.Helper()
	dirName := t.Name() + "-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	dir := filepath.Join(os.TempDir(), dirName)
	err := os.Mkdir(dir, 0700)
	requireEqual(t, err, nil, "failed to create a temp directory")
	cleanup := func() {
		err := os.RemoveAll(dir)
		assertEqual(t, err, nil, "should remove temp dir after test")
	}
	return dir, cleanup
}
