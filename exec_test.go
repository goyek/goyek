package taskflow_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pellared/taskflow"
)

func TestExec_success(t *testing.T) {
	sb := &strings.Builder{}
	var err error
	r := taskflow.Runner{
		Out: sb,
		Command: func(tf *taskflow.TF) {
			err = tf.Exec("", nil, "go", "version")
		},
	}

	result := r.Run()

	assert.NoError(t, err, "should pass")
	assert.Contains(t, sb.String(), "go version go1.")
	assert.True(t, result.Passed(), "task should pass")
}

func TestExec_error(t *testing.T) {
	var err error
	r := taskflow.Runner{
		Command: func(tf *taskflow.TF) {
			err = tf.Exec("", nil, "go", "wrong")
		},
	}

	r.Run()

	assert.Error(t, err, "should error, bad go command")
}
