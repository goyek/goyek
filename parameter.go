package goyek

import (
	"errors"
	"strconv"
)

// BoolParam represents a named boolean parameter that can be registered.
type BoolParam struct {
	Name     string
	Usage    string
	Default  bool
	Required bool
}

// IntParam represents a named integer parameter that can be registered.
type IntParam struct {
	Name     string
	Usage    string
	Default  int
	Required bool
}

// StringParam represents a named string parameter that can be registered.
type StringParam struct {
	Name     string
	Usage    string
	Default  string
	Required bool
}

// ValueParam represents a named parameter for a custom type that can be registered.
// NewValue field must be set with a default value factory.
type ValueParam struct {
	Name     string
	Usage    string
	Required bool
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
	// IsSet returns whether the param has a value
	IsSet() bool
}

// RegisteredParam represents a parameter that has been registered to a Flow.
// It can be used as a parameter for a Task.
type RegisteredParam interface {
	Name() string
	Usage() string
	Default() string
	Required() bool
	value(tf *TF) ParamValue
}

// registeredParam is a helper struct encapsulating concrete registered parameter type.
type registeredParam struct {
	name     string
	usage    string
	required bool
	newValue func() ParamValue
}

// Name returns the key of the parameter.
func (p registeredParam) Name() string {
	return p.name
}

// Usage returns the parameter's description.
func (p registeredParam) Usage() string {
	return p.usage
}

// Default returns the parameter's default value formatted as string.
func (p registeredParam) Default() string {
	return p.newValue().String()
}

// Required returns whether the parameter is required
func (p registeredParam) Required() bool {
	return p.required
}

func (p registeredParam) value(tf *TF) ParamValue {
	value, existing := tf.paramValues[p.name]
	if !existing {
		tf.Fatal(&ParamError{Key: p.name, Err: errors.New("parameter not registered")})
	}
	return value
}

// RegisteredValueParam represents a registered parameter based on a generic implementation.
type RegisteredValueParam struct {
	registeredParam
}

// Get returns the concrete instance of the generic value in the given flow.
func (p RegisteredValueParam) Get(tf *TF) interface{} {
	return p.value(tf).Get()
}

type boolValue struct {
	value bool
	set   bool
}

func (value *boolValue) Set(s string) error {
	value.set = true
	if len(s) == 0 {
		value.value = true
		return nil
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		err = errors.New("parse error")
	}
	value.value = v
	return err
}

func (value *boolValue) Get() interface{} { return value.value }

func (value *boolValue) String() string { return strconv.FormatBool(value.value) }

func (value *boolValue) IsBool() bool { return true }

func (value *boolValue) IsSet() bool { return value.set }

// RegisteredBoolParam represents a registered boolean parameter.
type RegisteredBoolParam struct {
	registeredParam
}

// Get returns the boolean value of the parameter in the given flow.
func (p RegisteredBoolParam) Get(tf *TF) bool {
	value := p.value(tf)
	return value.Get().(bool)
}

type intValue struct {
	value int
	set   bool
}

func (value *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err != nil {
		err = errors.New("parse error")
	}
	value.value = int(v)
	value.set = true
	return err
}

func (value *intValue) Get() interface{} { return value.value }

func (value *intValue) String() string { return strconv.Itoa(value.value) }

func (value *intValue) IsBool() bool { return false }

func (value *intValue) IsSet() bool { return value.set }

// RegisteredIntParam represents a registered integer parameter.
type RegisteredIntParam struct {
	registeredParam
}

// Get returns the integer value of the parameter in the given flow.
func (p RegisteredIntParam) Get(tf *TF) int {
	value := p.value(tf)
	return value.Get().(int)
}

type stringValue struct {
	value string
	set   bool
}

func (value *stringValue) Set(val string) error {
	value.value = val
	value.set = true
	return nil
}

func (value *stringValue) Get() interface{} { return value.value }

func (value *stringValue) String() string { return value.value }

func (value *stringValue) IsBool() bool { return false }

func (value *stringValue) IsSet() bool { return value.set }

// RegisteredStringParam represents a registered string parameter.
type RegisteredStringParam struct {
	registeredParam
}

// Get returns the string value of the parameter in the given flow.
func (p RegisteredStringParam) Get(tf *TF) string {
	value := p.value(tf)
	return value.Get().(string)
}
