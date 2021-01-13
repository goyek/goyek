package taskflow

import (
	"encoding"
	"encoding/json"
	"errors"
	"strconv"
	"time"
)

func (p Params) SetInt(key string, v int) {
	p[key] = strconv.Itoa(v)
}

func (p Params) SetBool(key string, v bool) {
	p[key] = strconv.FormatBool(v)
}

func (p Params) SetDuration(key string, v time.Duration) {
	p[key] = v.String()
}

func (p Params) SetDate(key string, v time.Time, layout string) {
	p[key] = v.Format(layout)
}

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
