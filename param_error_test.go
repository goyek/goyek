package taskflow_test

import (
	"errors"
	"testing"

	"github.com/pellared/taskflow"
)

func Test_ParamError(t *testing.T) {
	baseErr := errors.New("some error")
	err := &taskflow.ParamError{Key: "x", Err: baseErr}

	assertEqual(t, err.Unwrap(), baseErr, "should unwrap proper error")
	assertEqual(t, err.Error(), "taskflow: parameter \"x\": some error", "should have proper message")
}
