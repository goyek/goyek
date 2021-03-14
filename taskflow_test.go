package taskflow_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pellared/taskflow"
)

func init() {
	taskflow.DefaultOutput = ioutil.Discard
}

func Test_Register(t *testing.T) {
	testCases := []struct {
		desc  string
		task  taskflow.Task
		valid bool
	}{
		{
			desc:  "good task name",
			task:  taskflow.Task{Name: "my-task"},
			valid: true,
		},
		{
			desc:  "missing task name",
			task:  taskflow.Task{},
			valid: false,
		},
		{
			desc:  "invalid dependency",
			task:  taskflow.Task{Name: "my-task", Dependencies: taskflow.Deps{taskflow.RegisteredTask{}}},
			valid: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			flow := taskflow.New()

			_, err := flow.Register(tc.task)

			if tc.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func Test_Register_same_name(t *testing.T) {
	flow := &taskflow.Taskflow{}
	task := taskflow.Task{Name: "task"}
	_, err := flow.Register(task)
	require.NoError(t, err, "should be a valid task")

	_, err = flow.Register(task)

	assert.Error(t, err, "should not be possible to register tasks with same name twice")
}

func Test_MustRegister_panic(t *testing.T) {
	flow := taskflow.New()

	act := func() { flow.MustRegister(taskflow.Task{}) }

	assert.Panics(t, act, "should panic because task name is empty")
}

func Test_successful(t *testing.T) {
	ctx := context.Background()
	flow := &taskflow.Taskflow{}
	var executed1 int
	task1 := flow.MustRegister(taskflow.Task{
		Name: "task-1",
		Command: func(*taskflow.TF) {
			executed1++
		},
	})
	var executed2 int
	flow.MustRegister(taskflow.Task{
		Name: "task-2",
		Command: func(*taskflow.TF) {
			executed2++
		},
		Dependencies: taskflow.Deps{task1},
	})
	var executed3 int
	flow.MustRegister(taskflow.Task{
		Name: "task-3",
		Command: func(*taskflow.TF) {
			executed3++
		},
		Dependencies: taskflow.Deps{task1},
	})
	got := func() []int {
		return []int{executed1, executed2, executed3}
	}

	exitCode := flow.Run(ctx, "task-1")
	require.Equal(t, 0, exitCode, "first execution should pass")
	require.Equal(t, []int{1, 0, 0}, got(), "should execute task 1")

	exitCode = flow.Run(ctx, "task-2")
	require.Equal(t, 0, exitCode, "second execution should pass")
	require.Equal(t, []int{2, 1, 0}, got(), "should execute task 1 and 2")

	exitCode = flow.Run(ctx, "task-1", "task-2", "task-3")
	require.Equal(t, 0, exitCode, "third execution should pass")
	require.Equal(t, []int{3, 2, 1}, got(), "should execute task 1 and 2 and 3")
}

func Test_dependency_failure(t *testing.T) {
	flow := &taskflow.Taskflow{}
	var executed1 int
	task1 := flow.MustRegister(taskflow.Task{
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
	flow.MustRegister(taskflow.Task{
		Name: "task-2",
		Command: func(*taskflow.TF) {
			executed2++
		},
		Dependencies: taskflow.Deps{task1},
	})
	var executed3 int
	flow.MustRegister(taskflow.Task{
		Name: "task-3",
		Command: func(*taskflow.TF) {
			executed3++
		},
		Dependencies: taskflow.Deps{task1},
	})
	got := func() []int {
		return []int{executed1, executed2, executed3}
	}

	exitCode := flow.Run(context.Background(), "task-2", "task-3")

	assert.Equal(t, 1, exitCode, "should return error from first task")
	assert.Equal(t, []int{11, 0, 0}, got(), "should execute task 1")
}

func Test_fail(t *testing.T) {
	flow := &taskflow.Taskflow{}
	failed := false
	flow.MustRegister(taskflow.Task{
		Name: "task",
		Command: func(tf *taskflow.TF) {
			defer func() {
				failed = tf.Failed()
			}()
			tf.Fatal("failing")
		},
	})

	exitCode := flow.Run(context.Background(), "task")

	assert.Equal(t, 1, exitCode, "should return error")
	assert.True(t, failed, "tf.Failed() should return true")
}

func Test_skip(t *testing.T) {
	flow := &taskflow.Taskflow{}
	skipped := false
	flow.MustRegister(taskflow.Task{
		Name: "task",
		Command: func(tf *taskflow.TF) {
			defer func() {
				skipped = tf.Skipped()
			}()
			tf.Skip("skipping")
		},
	})

	exitCode := flow.Run(context.Background(), "task")

	assert.Equal(t, 0, exitCode, "should pass")
	assert.True(t, skipped, "tf.Skipped() should return true")
}

func Test_task_panics(t *testing.T) {
	flow := &taskflow.Taskflow{}
	flow.MustRegister(taskflow.Task{
		Name: "task",
		Command: func(tf *taskflow.TF) {
			panic("panicked!")
		},
	})

	exitCode := flow.Run(context.Background(), "task")

	assert.Equal(t, 1, exitCode, "should return error from first task")
}

func Test_cancelation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	flow := &taskflow.Taskflow{}
	flow.MustRegister(taskflow.Task{
		Name: "task",
	})

	exitCode := flow.Run(ctx, "task")

	assert.Equal(t, 1, exitCode, "should return error canceled")
}

func Test_cancelation_during_last_task(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	flow := &taskflow.Taskflow{}
	flow.MustRegister(taskflow.Task{
		Name: "task",
		Command: func(tf *taskflow.TF) {
			cancel()
		},
	})

	exitCode := flow.Run(ctx, "task")

	assert.Equal(t, 1, exitCode, "should return error canceled")
}

func Test_empty_command(t *testing.T) {
	flow := &taskflow.Taskflow{}
	flow.MustRegister(taskflow.Task{
		Name: "task",
	})

	exitCode := flow.Run(context.Background(), "task")

	assert.Equal(t, 0, exitCode, "should pass")
}

func Test_invalid_args(t *testing.T) {
	flow := taskflow.New()
	flow.MustRegister(taskflow.Task{
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

			assert.Equal(t, 2, exitCode, "should return error bad args")
		})
	}
}

func Test_help(t *testing.T) {
	flow := taskflow.New()
	a := flow.MustRegister(taskflow.Task{
		Name:        "a",
		Description: "some task",
	})
	flow.DefaultTask = a
	flow.Params.SetBool("fast", false)

	exitCode := flow.Run(context.Background(), "-h")

	assert.Equal(t, 2, exitCode, "should return error bad args")
}

func Test_printing(t *testing.T) {
	sb := &strings.Builder{}
	flow := &taskflow.Taskflow{
		Output:  sb,
		Verbose: true,
	}
	skipped := flow.MustRegister(taskflow.Task{
		Name: "skipped",
		Command: func(tf *taskflow.TF) {
			tf.Skipf("Skipf %d", 0)
		},
	})
	flow.MustRegister(taskflow.Task{
		Name:         "failing",
		Dependencies: taskflow.Deps{skipped},
		Command: func(tf *taskflow.TF) {
			tf.Log("Log", 1)
			tf.Logf("Logf %d", 2)
			tf.Error("Error", 3)
			tf.Errorf("Errorf %d", 4)
			tf.Fatalf("Fatalf %d", 5)
		},
	})
	t.Log()

	flow.Run(context.Background(), "failing")

	assert.Contains(t, sb.String(), "Skipf 0", "should contain proper output from \"skipped\" task")
	assert.Contains(t, sb.String(), `Log 1
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
				Verbose: tc.verbose,
				Output:  sb,
			}
			flow.MustRegister(taskflow.Task{
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

			exitCode := flow.Run(context.Background(), "task")

			assert.Equal(t, taskflow.CodeFailure, exitCode, "should fail")
			assert.Contains(t, sb.String(), "from child goroutine")
			assert.Contains(t, sb.String(), "from main goroutine")
		})
	}
}

func Test_name(t *testing.T) {
	flow := &taskflow.Taskflow{}
	taskName := "my-named-task"
	var got string
	flow.MustRegister(taskflow.Task{
		Name: taskName,
		Command: func(tf *taskflow.TF) {
			got = tf.Name()
		},
	})

	exitCode := flow.Run(context.Background(), taskName)

	assert.Equal(t, 0, exitCode, "should pass")
	assert.Equal(t, taskName, got, "should return proper Name value")
}

func Test_verbose(t *testing.T) {
	flow := &taskflow.Taskflow{}
	var got bool
	flow.MustRegister(taskflow.Task{
		Name: "task",
		Command: func(tf *taskflow.TF) {
			got = tf.Verbose()
		},
	})

	exitCode := flow.Run(context.Background(), "-v", "task")

	assert.Equal(t, 0, exitCode, "should pass")
	assert.True(t, got, "should return proper Verbose value")
}

func Test_params(t *testing.T) {
	flow := taskflow.New()
	flow.Params.SetInt("x", 1)
	flow.Params["z"] = "abc"
	var got taskflow.TFParams
	flow.MustRegister(taskflow.Task{
		Name: "task",
		Command: func(tf *taskflow.TF) {
			got = tf.Params()
		},
	})

	exitCode := flow.Run(context.Background(), "y=2", "z=3", "task")

	assert.Equal(t, 0, exitCode, "should pass")
	assert.Equal(t, "1", got.String("x"), "x param")
	assert.Equal(t, 2, got.Int("y"), "y param")
	assert.Equal(t, 3.0, got.Float64("z"), "z param")
}

func Test_defaultTask(t *testing.T) {
	flow := taskflow.New()
	var got taskflow.TFParams
	task := flow.MustRegister(taskflow.Task{
		Name: "task",
		Command: func(tf *taskflow.TF) {
			got = tf.Params()
		},
	})
	flow.DefaultTask = task

	exitCode := flow.Run(context.Background(), "x=a")

	assert.Equal(t, 0, exitCode, "should pass")
	assert.Equal(t, "a", got.String("x"), "x param")
}
