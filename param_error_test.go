package goyek_test

import (
	"errors"
	"testing"

	"github.com/goyek/goyek"
)

func Test_ParamError(t *testing.T) {
	baseErr := errors.New("some error")
	err := &goyek.ParamError{Key: "x", Err: baseErr}

	assertEqual(t, err.Unwrap(), baseErr, "should unwrap proper error")
	assertEqual(t, err.Error(), "goyek: parameter \"x\": some error", "should have proper message")
}
