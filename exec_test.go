package taskflow

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExec_success(t *testing.T) {
	sb := &strings.Builder{}
	var err error
	r := Runner{
		Out: sb,
		Command: func(tf *TF) {
			err = tf.Exec("", nil, "go", "version")
		},
	}

	r.Run()

	require.NoError(t, err, "should pass")
	assert.Contains(t, sb.String(), "go version go1.")
}

func TestExec_error(t *testing.T) {
	var err error
	r := Runner{
		Command: func(tf *TF) {
			err = tf.Exec("", nil, "go", "wrong")
		},
	}

	r.Run()

	assert.Error(t, err, "should error, bad go command")
}
