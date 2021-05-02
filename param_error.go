package goyek

import "strconv"

// ParamError records an error during parameter conversion.
type ParamError struct {
	Key string // the parameter's key
	Err error  // the reason the conversion failure, e.g. *strconv.NumError
}

func (e *ParamError) Error() string {
	return "goyek: parameter " + strconv.Quote(e.Key) + ": " + e.Err.Error()
}

// Unwrap unpacks the wrapped error.
func (e *ParamError) Unwrap() error { return e.Err }
