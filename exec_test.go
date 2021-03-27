package taskflow_test

import (
	"strings"
	"testing"

	"github.com/pellared/taskflow"
)

func TestExec_success(t *testing.T) {
	sb := &strings.Builder{}
	r := taskflow.Runner{
		Output: sb,
	}

	got := r.Run(taskflow.Exec("go", "version"))

	assertContains(t, sb.String(), "go version go", "output should contain prefix of version report")
	assertTrue(t, got.Passed(), "task should pass")
}

func TestExec_error(t *testing.T) {
	r := taskflow.Runner{}

	got := r.Run(taskflow.Exec("go", "wrong"))

	assertTrue(t, got.Failed(), "task should fail")
}
