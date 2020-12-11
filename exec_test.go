package taskflow_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pellared/taskflow"
)

func TestExec_success(t *testing.T) {
	sb := &strings.Builder{}
	r := taskflow.Runner{
		Output: sb,
	}

	got := r.Run(taskflow.Exec("go", "version"))

	assert.Contains(t, sb.String(), "go version go")
	assert.True(t, got.Passed(), "task should pass")
}

func TestExec_error(t *testing.T) {
	r := taskflow.Runner{}

	got := r.Run(taskflow.Exec("go", "wrong"))

	assert.True(t, got.Failed(), "task should fail")
}
