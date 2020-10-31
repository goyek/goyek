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
	}

	result := r.Run(func(tf *taskflow.TF) {
		err = tf.Exec("", nil, "go", "version")
	})

	assert.NoError(t, err, "should pass")
	assert.Contains(t, sb.String(), "go version go1.")
	assert.True(t, result.Passed(), "task should pass")
}

func TestExec_error(t *testing.T) {
	var err error
	r := taskflow.Runner{}

	r.Run(func(tf *taskflow.TF) {
		err = tf.Exec("", nil, "go", "wrong")
	})

	assert.Error(t, err, "should error, bad go command")
}
