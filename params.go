package taskflow

import (
	"errors"
	"strconv"
)

var ErrParamNotSet = errors.New("parameter is not set")

// Params represents Taskflow parameters used within Taskflow.
// The default values set in the struct are overridden in Run method.
type Params map[string]string

func (p Params) Int(key string) (int, error) {
	v := p[key]
	if v == "" {
		return 0, ErrParamNotSet
	}
	return strconv.Atoi(v)
}
