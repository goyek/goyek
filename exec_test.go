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
	primary := &strings.Builder{}
	message := &strings.Builder{}
	flow := &goyek.Taskflow{
		Output: goyek.Output{
			Primary: primary,
			Message: message,
		},
	}
	flow.Register(goyek.Task{
		Name:    taskName,
		Command: goyek.Exec("go", "version"),
	})

	exitCode := flow.Run(context.Background(), "-v", taskName)

	assertContains(t, primary.String(), "go version go", "output should contain prefix of version report")
	assertEqual(t, exitCode, goyek.CodePass, "task should pass")
}

func TestExec_error(t *testing.T) {
	taskName := "exec"
	flow := &goyek.Taskflow{
		Output: goyek.Output{
			Primary: ioutil.Discard,
			Message: ioutil.Discard,
		},
	}
	flow.Register(goyek.Task{
		Name:    taskName,
		Command: goyek.Exec("go", "wrong"),
	})

	exitCode := flow.Run(nil, taskName) //nolint:staticcheck // present that nil context is handled

	assertEqual(t, exitCode, goyek.CodeFail, "task should pass")
}
