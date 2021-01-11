package taskflow

import (
	"errors"
	"strconv"
)

// ErrParamNotSet indicates that a parameter is not set.
var ErrParamNotSet = errors.New("parameter is not set")

// Params represents Taskflow parameters used within Taskflow.
// The default values set in the struct are overridden in Run method.
type Params map[string]string

// Int converts the parameter to int using the Go syntax for integer literals.
// ErrParamNotSet error is returned if the parameter was not set.
// *strconv.NumError error is returned if the parameter conversion failed.
func (p Params) Int(key string) (int, error) {
	v := p[key]
	if v == "" {
		return 0, ErrParamNotSet
	}
	i, err := strconv.ParseInt(v, 0, strconv.IntSize)
	return int(i), err
}

// Bool converts the parameter to bool.
// It accepts 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False.
// Any other value returns an error.
// ErrParamNotSet error is returned if the parameter was not set.
// *strconv.NumError error is returned if the parameter conversion failed.
func (p Params) Bool(key string) (bool, error) {
	v := p[key]
	if v == "" {
		return false, ErrParamNotSet
	}
	return strconv.ParseBool(v)
}

// Float64 converts the parameter to float accepting decimal and hexadecimal floating-point number syntax.
// ErrParamNotSet error is returned if the parameter was not set.
// *strconv.NumError error is returned if the parameter conversion failed.
func (p Params) Float64(key string) (float64, error) {
	v := p[key]
	if v == "" {
		return 0, ErrParamNotSet
	}
	return strconv.ParseFloat(v, 64)
}
