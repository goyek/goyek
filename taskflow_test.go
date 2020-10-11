package taskflow_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pellared/taskflow"
)

func Test_successful(t *testing.T) {
	ctx := context.Background()
	tasks := &taskflow.Taskflow{}
	var executed1 int
	task1 := tasks.MustRegister(taskflow.Task{
		Name: "task-1",
		Command: func(*taskflow.TF) error {
			executed1++
			return nil
		},
	})
	var executed2 int
	tasks.MustRegister(taskflow.Task{
		Name: "task-2",
		Command: func(*taskflow.TF) error {
			executed2++
			return nil
		},
		Dependencies: []taskflow.Dependency{task1},
	})
	var executed3 int
	tasks.MustRegister(taskflow.Task{
		Name: "task-3",
		Command: func(*taskflow.TF) error {
			executed3++
			return nil
		},
		Dependencies: []taskflow.Dependency{task1},
	})
	got := func() []int {
		return []int{executed1, executed2, executed3}
	}

	tasks.MustExecute(ctx, "task-1")
	require.Equal(t, []int{1, 0, 0}, got(), "should execute task 1")

	tasks.MustExecute(ctx, "task-2")
	require.Equal(t, []int{2, 1, 0}, got(), "should execute task 1 and 2")

	tasks.MustExecute(ctx, "task-1", "task-2", "task-3")
	require.Equal(t, []int{3, 2, 1}, got(), "should execute task 1 and 2 and 3")
}

func Test_dependency_failure(t *testing.T) {
	ctx := context.Background()
	tasks := &taskflow.Taskflow{}
	var executed1 int
	task1 := tasks.MustRegister(taskflow.Task{
		Name: "task-1",
		Command: func(*taskflow.TF) error {
			executed1++
			return errors.New("I failed you") //nolint:goerr113 // test code
		},
	})
	var executed2 int
	tasks.MustRegister(taskflow.Task{
		Name: "task-2",
		Command: func(*taskflow.TF) error {
			executed2++
			return nil
		},
		Dependencies: []taskflow.Dependency{task1},
	})
	var executed3 int
	tasks.MustRegister(taskflow.Task{
		Name: "task-3",
		Command: func(*taskflow.TF) error {
			executed3++
			return nil
		},
		Dependencies: []taskflow.Dependency{task1},
	})
	got := func() []int {
		return []int{executed1, executed2, executed3}
	}

	err := tasks.Execute(ctx, "task-1", "task-2", "task-3")

	assert.Error(t, err, "should return error from first task")
	assert.Equal(t, []int{1, 0, 0}, got(), "should execute task 1")
}
