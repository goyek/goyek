package taskflow_test

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pellared/taskflow"
)

func init() {
	taskflow.DefaultOutput = ioutil.Discard
}

func testTF(t *testing.T, flagsAndParams ...string) *taskflow.TF {
	t.Helper()

	flow := &taskflow.Taskflow{}
	var got *taskflow.TF
	flow.MustRegister(taskflow.Task{
		Name: "task",
		Command: func(tf *taskflow.TF) {
			got = tf
		},
	})

	args := append(flagsAndParams, "task")
	exitCode := flow.Run(context.Background(), args...)

	assert.Equal(t, 0, exitCode, "should pass")
	return got
}
