package taskflow_test

import (
	"errors"
	"testing"

	"github.com/pellared/taskflow"
)

func Test_ParamError(t *testing.T) {
	baseErr := errors.New("some error")
	err := &taskflow.ParamError{Key: "x", Err: baseErr}

	assertEqual(t, baseErr, err.Unwrap(), "should unwrap proper error")
	assertEqual(t, "taskflow: parameter \"x\": some error", err.Error(), "should have proper message")
}
