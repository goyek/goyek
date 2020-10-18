package taskflow

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExec_success(t *testing.T) {
	var err error
	fn := func(tf *TF) {
		err = Exec(tf, "git", "help")
	}
	sb := &strings.Builder{}

	Run(fn, RunConfig{Out: sb})

	require.NoError(t, err, "should pass, everyone has git")
	assert.Contains(t, sb.String(), "usage: git")
}

func TestExec_error(t *testing.T) {
	var err error
	fn := func(tf *TF) {
		err = Exec(tf, "git", "wrong")
	}

	Run(fn, RunConfig{})

	assert.Error(t, err, "should error, bad git command")
}
