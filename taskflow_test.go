package taskflow_test

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pellared/taskflow"
)

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
			flow := &taskflow.Taskflow{}

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

func Test_successful(t *testing.T) {
	ctx := context.Background()
	flow := &taskflow.Taskflow{
		Output: ioutil.Discard,
	}
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
	ctx := context.Background()
	flow := &taskflow.Taskflow{
		Output: ioutil.Discard,
	}
	var executed1 int
	task1 := flow.MustRegister(taskflow.Task{
		Name: "task-1",
		Command: func(tf *taskflow.TF) {
			executed1++
			tf.Errorf("it still runs")
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

	exitCode := flow.Run(ctx, "task-1", "task-2", "task-3")

	assert.Equal(t, 1, exitCode, "should return error from first task")
	assert.Equal(t, []int{11, 0, 0}, got(), "should execute task 1")
}

func Test_fail(t *testing.T) {
	ctx := context.Background()
	flow := &taskflow.Taskflow{
		Output: ioutil.Discard,
	}
	failed := false
	flow.MustRegister(taskflow.Task{
		Name: "task",
		Command: func(tf *taskflow.TF) {
			defer func() {
				failed = tf.Failed()
			}()
			tf.Fatalf("failing")
		},
	})

	exitCode := flow.Run(ctx, "task")

	assert.Equal(t, 1, exitCode, "should return error")
	assert.True(t, failed, "tf.Failed() should return true")
}

func Test_skip(t *testing.T) {
	ctx := context.Background()
	flow := &taskflow.Taskflow{
		Output: ioutil.Discard,
	}
	skipped := false
	flow.MustRegister(taskflow.Task{
		Name: "task",
		Command: func(tf *taskflow.TF) {
			defer func() {
				skipped = tf.Skipped()
			}()
			tf.Skipf("skipping")
		},
	})

	exitCode := flow.Run(ctx, "task")

	assert.Equal(t, 0, exitCode, "should pass")
	assert.True(t, skipped, "tf.Skipped() should return true")
}

func Test_task_panics(t *testing.T) {
	ctx := context.Background()
	flow := &taskflow.Taskflow{
		Output: ioutil.Discard,
	}
	flow.MustRegister(taskflow.Task{
		Name: "task",
		Command: func(tf *taskflow.TF) {
			panic("panicked!")
		},
	})

	exitCode := flow.Run(ctx, "task")

	assert.Equal(t, 1, exitCode, "should return error from first task")
}

func Test_cancelation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	flow := &taskflow.Taskflow{
		Output: ioutil.Discard,
	}
	flow.MustRegister(taskflow.Task{
		Name: "task",
	})

	exitCode := flow.Run(ctx, "task")

	assert.Equal(t, 1, exitCode, "should return error canceled")
}

func Test_cancelation_during_last_task(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	flow := &taskflow.Taskflow{
		Output: ioutil.Discard,
	}
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
	ctx := context.Background()
	flow := &taskflow.Taskflow{
		Output: ioutil.Discard,
	}
	flow.MustRegister(taskflow.Task{
		Name: "task",
	})

	exitCode := flow.Run(ctx, "task")

	assert.Equal(t, 0, exitCode, "should pass")
}

func Test_verbose(t *testing.T) {
	ctx := context.Background()
	flow := &taskflow.Taskflow{
		Output: ioutil.Discard,
	}
	var got bool
	flow.MustRegister(taskflow.Task{
		Name: "task",
		Command: func(tf *taskflow.TF) {
			got = tf.Verbose()
		},
	})

	exitCode := flow.Run(ctx, "-v", "task")

	assert.Equal(t, 0, exitCode, "should pass")
	assert.True(t, got, "should return proper Verbose value")
}

func Test_invalid_args(t *testing.T) {
	ctx := context.Background()
	flow := &taskflow.Taskflow{
		Output: ioutil.Discard,
	}
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
			exitCode := flow.Run(ctx, tc.args...)

			assert.Equal(t, 2, exitCode, "should pass")
		})
	}
}

func Test(t *testing.T) {
}
