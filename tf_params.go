package taskflow

import (
	"encoding"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"time"
)

// TFParams represents Taskflow parameters accessible in task's command.
type TFParams struct {
	params map[string]string
	tf     *TF
}

// String returns the parameter as a string.
func (p TFParams) String(key string) string {
	value, exists := p.params[key]
	if !exists {
		p.tf.Fatal(&ParamError{Key: key, Err: errors.New("parameter not registered")})
	}
	return value
}

// Int converts the parameter to int using the Go syntax for integer literals.
// It fails the task if the conversion failed.
// 0 is returned if the parameter was not set.
func (p TFParams) Int(key string) int {
	v := p.String(key)
	if v == "" {
		return 0
	}
	i, err := strconv.ParseInt(v, 0, strconv.IntSize)
	if err != nil {
		p.tf.Fatal(&ParamError{Key: key, Err: err})
	}
	return int(i)
}

// Bool converts the parameter to bool.
// It fails the task if the conversion failed.
// False is returned if the parameter was not set.
// It accepts 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False.
func (p TFParams) Bool(key string) bool {
	v := p.String(key)
	if v == "" {
		return false
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		p.tf.Fatal(&ParamError{Key: key, Err: err})
	}
	return b
}

// Float64 converts the parameter to float64 accepting decimal and hexadecimal floating-point number syntax.
// It fails the task if the conversion failed.
// 0 is returned if the parameter was not set.
func (p TFParams) Float64(key string) float64 {
	v := p.String(key)
	if v == "" {
		return 0
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		p.tf.Fatal(&ParamError{Key: key, Err: err})
	}
	return f
}

// Duration converts the parameter to time.Duration using time.ParseDuration.
// It fails the task if the conversion failed.
// 0 is returned if the parameter was not set.
// A duration string is a possibly signed sequence of
// decimal numbers, each with optional fraction and a unit suffix,
// such as "300ms", "-1.5h" or "2h45m".
// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
func (p TFParams) Duration(key string) time.Duration {
	v := p.String(key)
	if v == "" {
		return 0
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		p.tf.Fatal(&ParamError{Key: key, Err: err})
	}
	return d
}

// Date converts the parameter to time.Time using time.Parse.
// It fails the task if the conversion failed.
// Zero value is returned if the parameter was not set.
// The layout defines the format by showing how the reference time,
// defined to be
//	Mon Jan 2 15:04:05 -0700 MST 2006
// would be interpreted if it were the value; it serves as an example of
// the input format. The same interpretation will then be made to the
// input string.
func (p TFParams) Date(key string, layout string) time.Time {
	v := p.String(key)
	if v == "" {
		return time.Time{}
	}
	d, err := time.Parse(layout, v)
	if err != nil {
		p.tf.Fatal(&ParamError{Key: key, Err: err})
	}
	return d
}

// ParseText parses the parameter and stores the result
// in the value pointed to by v by using its UnmarshalText method.
// It fails the task if the conversion failed or v is nil or not a pointer.
func (p TFParams) ParseText(key string, v encoding.TextUnmarshaler) {
	if v == nil {
		p.tf.Fatal(&ParamError{Key: key, Err: errors.New("nil passed to ParseText")})
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		p.tf.Fatal(&ParamError{Key: key, Err: errors.New("non-pointer variable passed to ParseText")})
	}

	s := p.String(key)
	if s == "" {
		return
	}
	if err := v.UnmarshalText([]byte(s)); err != nil {
		p.tf.Fatal(&ParamError{Key: key, Err: err})
	}
}

// ParseJSON parses the parameter and stores the result
// in the value pointed to by v by using json.Unmarshal.
// It fails the task if the conversion failed or v is nil or not a pointer.
func (p TFParams) ParseJSON(key string, v interface{}) {
	if v == nil {
		p.tf.Fatal(&ParamError{Key: key, Err: errors.New("nil passed to ParseJSON")})
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		p.tf.Fatal(&ParamError{Key: key, Err: errors.New("non-pointer variable passed to ParseJSON")})
	}

	s := p.String(key)
	if s == "" {
		return
	}
	if err := json.Unmarshal([]byte(s), v); err != nil {
		p.tf.Fatal(&ParamError{Key: key, Err: err})
	}
}
