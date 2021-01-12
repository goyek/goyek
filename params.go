package taskflow

import (
	"strconv"
	"time"
)

// ParamError records an error during parameter conversion.
type ParamError struct {
	Key string // the parameter's key
	Err error  // the reason the conversion failed (e.g. ErrParamNotSet, *strconv.NumError, etc.)
}

func (e *ParamError) Error() string {
	return "taskflow: parameter " + strconv.Quote(e.Key) + ": " + e.Err.Error()
}

// Unwrap unpacks the wrapped error.
func (e *ParamError) Unwrap() error { return e.Err }

// Params represents Taskflow parameters used within Taskflow.
// The default values set in the struct are overridden in Run method.
type Params map[string]string

// Int converts the parameter to int using the Go syntax for integer literals.
// 0 is returned if the parameter was not set.
// *strconv.NumError error is returned if the parameter conversion failed.
func (p Params) Int(key string) (int, error) {
	v := p[key]
	if v == "" {
		return 0, nil
	}
	i, err := strconv.ParseInt(v, 0, strconv.IntSize)
	if err != nil {
		return 0, &ParamError{Key: key, Err: err}
	}
	return int(i), nil
}

// Bool converts the parameter to bool.
// It accepts 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False.
// False is returned if the parameter was not set.
// Any other value returns an error.
// *strconv.NumError error is returned if the parameter conversion failed.
func (p Params) Bool(key string) (bool, error) {
	v := p[key]
	if v == "" {
		return false, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, &ParamError{Key: key, Err: err}
	}
	return b, nil
}

// Float64 converts the parameter to float64 accepting decimal and hexadecimal floating-point number syntax.
// 0 is returned if the parameter was not set.
// *strconv.NumError error is returned if the parameter conversion failed.
func (p Params) Float64(key string) (float64, error) {
	v := p[key]
	if v == "" {
		return 0, nil
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, &ParamError{Key: key, Err: err}
	}
	return f, nil
}

// Duration converts the parameter to time.Duration using time.ParseDuration.
// A duration string is a possibly signed sequence of
// decimal numbers, each with optional fraction and a unit suffix,
// such as "300ms", "-1.5h" or "2h45m".
// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
// 0 is returned if the parameter was not set.
// An error is also returned if the parameter conversion failed.
func (p Params) Duration(key string) (time.Duration, error) {
	v := p[key]
	if v == "" {
		return 0, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, &ParamError{Key: key, Err: err}
	}
	return d, nil
}
