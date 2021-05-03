package goyek_test

import (
	"strings"
	"testing"

	"github.com/goyek/goyek"
)

func TestExec_success(t *testing.T) {
	sb := &strings.Builder{}
	r := goyek.Runner{
		Output: sb,
	}

	got := r.Run(goyek.Exec("go", "version"))

	assertContains(t, sb.String(), "go version go", "output should contain prefix of version report")
	assertTrue(t, got.Passed(), "task should pass")
}

func TestExec_error(t *testing.T) {
	r := goyek.Runner{}

	got := r.Run(goyek.Exec("go", "wrong"))

	assertTrue(t, got.Failed(), "task should fail")
}
