package taskflow_test

import (
	"errors"
	"testing"

	"github.com/pellared/taskflow"
	"github.com/stretchr/testify/assert"
)

func Test_ParamError(t *testing.T) {
	baseErr := errors.New("some error")
	err := &taskflow.ParamError{Key: "x", Err: baseErr}

	assert.Equal(t, baseErr, err.Unwrap(), "should unwrap proper error")
	assert.Equal(t, "taskflow: parameter \"x\": some error", err.Error(), "should have proper message")
}
