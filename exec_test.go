package goyek_test

import (
	"context"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/goyek/goyek"
)

func TestExec_success(t *testing.T) {
	taskName := "exec"
	sb := &strings.Builder{}
	flow := &goyek.Taskflow{
		Output: sb,
	}
	flow.Register(goyek.Task{
		Name:    taskName,
		Command: goyek.Exec("go", "version"),
	})

	exitCode := flow.Run(context.Background(), "-v", taskName)

	assertContains(t, sb.String(), "go version go", "output should contain prefix of version report")
	assertEqual(t, exitCode, goyek.CodePass, "task should pass")
}

func TestExec_error(t *testing.T) {
	taskName := "exec"
	flow := &goyek.Taskflow{
		Output: ioutil.Discard,
	}
	flow.Register(goyek.Task{
		Name:    taskName,
		Command: goyek.Exec("go", "wrong"),
	})

	exitCode := flow.Run(context.Background(), taskName)

	assertEqual(t, exitCode, goyek.CodeFailure, "task should pass")
}
