package goyek

import (
	"errors"
	"strconv"
)

// BoolParam represents a named boolean parameter that can be registered.
type BoolParam struct {
	Name    string
	Usage   string
	Default bool
}

// IntParam represents a named integer parameter that can be registered.
type IntParam struct {
	Name    string
	Usage   string
	Default int
}

// StringParam represents a named string parameter that can be registered.
type StringParam struct {
	Name    string
	Usage   string
	Default string
}

// ValueParam represents a named parameter for a custom type that can be registered.
// NewValue field must be set with a default value factory.
type ValueParam struct {
	Name     string
	Usage    string
	NewValue func() ParamValue
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

// RegisteredParam represents a parameter that has been registered to a Flow.
// It can be used as a parameter for a Task.
type RegisteredParam interface {
	Name() string
	value(p *Progress) ParamValue
}

// registeredParam is a helper struct encapsulating concrete registered parameter type.
type registeredParam struct {
	name     string
	usage    string
	newValue func() ParamValue
}

// Name returns the key of the parameter.
func (p registeredParam) Name() string {
	return p.name
}

func (p registeredParam) value(progress *Progress) ParamValue {
	value, existing := progress.paramValues[p.name]
	if !existing {
		progress.Fatal(&ParamError{Key: p.name, Err: errors.New("parameter not registered")})
	}
	return value
}

// RegisteredValueParam represents a registered parameter based on a generic implementation.
type RegisteredValueParam struct {
	registeredParam
}

// Get returns the concrete instance of the generic value in the given flow.
func (p RegisteredValueParam) Get(progress *Progress) interface{} {
	return p.value(progress).Get()
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

// RegisteredBoolParam represents a registered boolean parameter.
type RegisteredBoolParam struct {
	registeredParam
}

// Get returns the boolean value of the parameter in the given flow.
func (p RegisteredBoolParam) Get(progress *Progress) bool {
	value := p.value(progress)
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

// RegisteredIntParam represents a registered integer parameter.
type RegisteredIntParam struct {
	registeredParam
}

// Get returns the integer value of the parameter in the given flow.
func (p RegisteredIntParam) Get(progress *Progress) int {
	value := p.value(progress)
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

// RegisteredStringParam represents a registered string parameter.
type RegisteredStringParam struct {
	registeredParam
}

// Get returns the string value of the parameter in the given flow.
func (p RegisteredStringParam) Get(progress *Progress) string {
	value := p.value(progress)
	return value.Get().(string)
}
