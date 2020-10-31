package taskflow_test

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pellared/taskflow"
)

func Test_successful(t *testing.T) {
	ctx := context.Background()
	tasks := &taskflow.Taskflow{
		Out: ioutil.Discard,
	}
	var executed1 int
	task1 := tasks.MustRegister(taskflow.Task{
		Name: "task-1",
		Command: func(*taskflow.TF) {
			executed1++
		},
	})
	var executed2 int
	tasks.MustRegister(taskflow.Task{
		Name: "task-2",
		Command: func(*taskflow.TF) {
			executed2++
		},
		Dependencies: taskflow.Deps{task1},
	})
	var executed3 int
	tasks.MustRegister(taskflow.Task{
		Name: "task-3",
		Command: func(*taskflow.TF) {
			executed3++
		},
		Dependencies: taskflow.Deps{task1},
	})
	got := func() []int {
		return []int{executed1, executed2, executed3}
	}

	err := tasks.Run(ctx, "task-1")
	require.NoError(t, err, "first execution should pass")
	require.Equal(t, []int{1, 0, 0}, got(), "should execute task 1")

	err = tasks.Run(ctx, "task-2")
	require.NoError(t, err, "second execution should pass")
	require.Equal(t, []int{2, 1, 0}, got(), "should execute task 1 and 2")

	err = tasks.Run(ctx, "task-1", "task-2", "task-3")
	require.NoError(t, err, "third execution should pass")
	require.Equal(t, []int{3, 2, 1}, got(), "should execute task 1 and 2 and 3")
}

func Test_dependency_failure(t *testing.T) {
	ctx := context.Background()
	tasks := &taskflow.Taskflow{
		Out: ioutil.Discard,
	}
	var executed1 int
	task1 := tasks.MustRegister(taskflow.Task{
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
	tasks.MustRegister(taskflow.Task{
		Name: "task-2",
		Command: func(*taskflow.TF) {
			executed2++
		},
		Dependencies: taskflow.Deps{task1},
	})
	var executed3 int
	tasks.MustRegister(taskflow.Task{
		Name: "task-3",
		Command: func(*taskflow.TF) {
			executed3++
		},
		Dependencies: taskflow.Deps{task1},
	})
	got := func() []int {
		return []int{executed1, executed2, executed3}
	}

	err := tasks.Run(ctx, "task-1", "task-2", "task-3")

	assert.Error(t, err, "should return error from first task")
	assert.Equal(t, []int{11, 0, 0}, got(), "should execute task 1")
}

func Test_fail(t *testing.T) {
	ctx := context.Background()
	tasks := &taskflow.Taskflow{
		Out: ioutil.Discard,
	}
	failed := false
	tasks.MustRegister(taskflow.Task{
		Name: "task",
		Command: func(tf *taskflow.TF) {
			defer func() {
				failed = tf.Failed()
			}()
			tf.Fatalf("failing")
		},
	})

	err := tasks.Run(ctx, "task")

	assert.Error(t, err, "should return error")
	assert.True(t, failed, "tf.Failed() should return true")
}

func Test_skip(t *testing.T) {
	ctx := context.Background()
	tasks := &taskflow.Taskflow{
		Out: ioutil.Discard,
	}
	skipped := false
	tasks.MustRegister(taskflow.Task{
		Name: "task",
		Command: func(tf *taskflow.TF) {
			defer func() {
				skipped = tf.Skipped()
			}()
			tf.Skipf("skipping")
		},
	})

	err := tasks.Run(ctx, "task")

	assert.NoError(t, err, "should pass")
	assert.True(t, skipped, "tf.Skipped() should return true")
}

func Test_task_panics(t *testing.T) {
	ctx := context.Background()
	tasks := &taskflow.Taskflow{
		Out: ioutil.Discard,
	}
	tasks.MustRegister(taskflow.Task{
		Name: "task",
		Command: func(tf *taskflow.TF) {
			panic("panicked!")
		},
	})

	err := tasks.Run(ctx, "task")

	assert.Error(t, err, "should return error from first task")
}

func Test_cancelation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	tasks := &taskflow.Taskflow{
		Out: ioutil.Discard,
	}
	tasks.MustRegister(taskflow.Task{
		Name: "task",
	})

	err := tasks.Run(ctx, "task")

	assert.Equal(t, context.Canceled, err, "should return error canceled")
}

func Test_cancelation_during_last_task(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	tasks := &taskflow.Taskflow{
		Out: ioutil.Discard,
	}
	tasks.MustRegister(taskflow.Task{
		Name: "task",
		Command: func(tf *taskflow.TF) {
			cancel()
		},
	})

	err := tasks.Run(ctx, "task")

	assert.Equal(t, context.Canceled, err, "should return error canceled")
}

func Test_empty_command(t *testing.T) {
	ctx := context.Background()
	tasks := &taskflow.Taskflow{
		Out: ioutil.Discard,
	}
	tasks.MustRegister(taskflow.Task{
		Name: "task",
	})

	err := tasks.Run(ctx, "task")

	assert.NoError(t, err, "should pass")
}
