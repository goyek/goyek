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
			err = Exec(tf, "git", "help")
		},
	}

	r.Run()

	require.NoError(t, err, "should pass, everyone has git")
	assert.Contains(t, sb.String(), "usage: git")
}

func TestExec_error(t *testing.T) {
	var err error
	r := Runner{
		Command: func(tf *TF) {
			err = Exec(tf, "git", "wrong")
		},
	}

	r.Run()

	assert.Error(t, err, "should error, bad git command")
}
