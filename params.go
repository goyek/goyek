package taskflow

import (
	"encoding"
	"encoding/json"
	"errors"
	"strconv"
	"time"
)

// SetInt sets default value for given parameter using strconv.Itoa.
func (p Params) SetInt(key string, v int) {
	p[key] = strconv.Itoa(v)
}

// SetBool sets default value for given parameter using strconv.FormatBool.
func (p Params) SetBool(key string, v bool) {
	p[key] = strconv.FormatBool(v)
}

// SetDuration sets default value for given parameter using time.Duration.String.
func (p Params) SetDuration(key string, v time.Duration) {
	p[key] = v.String()
}

// SetDate sets default value for given parameter using time.Time.Format.
func (p Params) SetDate(key string, v time.Time, layout string) {
	p[key] = v.Format(layout)
}

// SetText sets default value for given parameter using value's MarshalText method.
// If v is nil SetText returns an error.
// An error is also returned if marshaling failed.
func (p Params) SetText(key string, v encoding.TextMarshaler) error {
	if v == nil {
		return &ParamError{Key: key, Err: errors.New("nil passed to SetText")}
	}
	text, err := v.MarshalText()
	if err != nil {
		return &ParamError{Key: key, Err: err}
	}
	p[key] = string(text)
	return nil
}

// SetJSON sets default value for given parameter using json.Marshal.
// If v is nil SetJSON returns an error.
// An error is also returned if marshaling failed.
func (p Params) SetJSON(key string, v interface{}) error {
	if v == nil {
		return &ParamError{Key: key, Err: errors.New("nil passed to SetJSON")}
	}
	text, err := json.Marshal(v)
	if err != nil {
		return &ParamError{Key: key, Err: err}
	}
	p[key] = string(text)
	return nil
}
