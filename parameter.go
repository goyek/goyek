package taskflow

import (
	"errors"
	"strconv"
)

type paramValueFactory struct {
	info     ParamInfo
	newValue func() ParamValue
}

// ParamInfo represents the general information of a parameter for one or more tasks.
type ParamInfo struct {
	Name  string
	Usage string
}

// ParamValue represents an instance of a generic parameter.
type ParamValue interface {
	// String returns the current value formatted as string.
	// The returned format should be in a single line, representing the parameter
	// as it could be provided on the command line.
	String() string
	// IsBool marks parameters that do not explicitly need to be set a value.
	// Set will be called in case the flag is not explicitly parameterized.
	IsBool() bool
	// Get returns the current value, properly typed.
	// Values must return their default value if Set() has not yet been called.
	Get() interface{}
	// Set parses the given string and sets the typed value.
	Set(string) error
}

// RegisteredParam represents a parameter that has been registered to a Taskflow.
// It can be used as a parameter for a Task.
type RegisteredParam interface {
	Name() string
	value(tf *TF) ParamValue
}

// param is a helper struct for implementing concrete parameter types.
type param struct {
	name string
}

// Name returns the key of the parameter.
func (p param) Name() string {
	return p.name
}

func (p param) value(tf *TF) ParamValue {
	value, existing := tf.paramValues[p.name]
	if !existing {
		tf.Fatal(&ParamError{Key: p.name, Err: errors.New("parameter not registered")})
	}
	return value
}

// ValueParam represents a registered parameter based on a generic implementation.
type ValueParam struct {
	param
}

// Get returns the concrete instance of the generic value in the given flow.
func (p ValueParam) Get(tf *TF) interface{} {
	return p.value(tf).Get()
}

type boolValue bool

func (value *boolValue) Set(s string) error {
	if len(s) == 0 {
		*value = true
		return nil
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		err = errors.New("parse error")
	}
	*value = boolValue(v)
	return err
}

func (value *boolValue) Get() interface{} { return bool(*value) }

func (value *boolValue) String() string { return strconv.FormatBool(bool(*value)) }

func (value *boolValue) IsBool() bool { return true }

// BoolParam represents a registered boolean parameter.
type BoolParam struct {
	param
}

// Get returns the boolean value of the parameter in the given flow.
func (p BoolParam) Get(tf *TF) bool {
	value := p.value(tf)
	return value.Get().(bool)
}

type intValue int

func (value *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err != nil {
		err = errors.New("parse error")
	}
	*value = intValue(v)
	return err
}

func (value *intValue) Get() interface{} { return int(*value) }

func (value *intValue) String() string { return strconv.Itoa(int(*value)) }

func (value *intValue) IsBool() bool { return false }

// IntParam represents a registered integer parameter.
type IntParam struct {
	param
}

// Get returns the integer value of the parameter in the given flow.
func (p IntParam) Get(tf *TF) int {
	value := p.value(tf)
	return value.Get().(int)
}

type stringValue string

func (value *stringValue) Set(val string) error {
	*value = stringValue(val)
	return nil
}

func (value *stringValue) Get() interface{} { return string(*value) }

func (value *stringValue) String() string { return string(*value) }

func (value *stringValue) IsBool() bool { return false }

// StringParam represents a registered string parameter.
type StringParam struct {
	param
}

// Get returns the string value of the parameter in the given flow.
func (p StringParam) Get(tf *TF) string {
	value := p.value(tf)
	return value.Get().(string)
}
