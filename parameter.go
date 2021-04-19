package taskflow

import (
	"errors"
	"flag"
)

type parameter struct {
	info     ParameterInfo
	register func(*flag.FlagSet)
}

// ParameterInfo represents the general information of a parameter for one or more tasks.
type ParameterInfo struct {
	Name  string
	Usage string
}

// Value represents an instance of a generic parameter.
// It deliberately matches the signature of type flag.Value as the flag API is used for the
// underlying implementation.
type Value interface {
	String() string
	Set(string) error
}

// RegisteredParam represents a parameter that has been registered to a Taskflow.
// It can be used as a parameter for a Task.
type RegisteredParam struct {
	name string
}

// Name returns the key of the parameter.
func (p RegisteredParam) Name() string {
	return p.name
}

func (p RegisteredParam) value(tf *TF) Value {
	value, existing := tf.paramValues[p.name]
	if !existing {
		tf.Fatal(&ParamError{Key: p.name, Err: errors.New("parameter not registered")})
	}
	return value
}

// ValueParam represents a registered parameter based on a generic implementation.
type ValueParam struct {
	RegisteredParam
}

// Get returns the concrete instance of the generic parameter in the given flow.
func (p ValueParam) Get(tf *TF) Value {
	return p.value(tf)
}

// BoolParam represents a registered boolean parameter.
type BoolParam struct {
	RegisteredParam
}

// Get returns the boolean value of the parameter in the given flow.
func (p BoolParam) Get(tf *TF) bool {
	value := p.value(tf)
	return value.(flag.Getter).Get().(bool)
}

// IntParam represents a registered integer parameter.
type IntParam struct {
	RegisteredParam
}

// Get returns the integer value of the parameter in the given flow.
func (p IntParam) Get(tf *TF) int {
	value := p.value(tf)
	return value.(flag.Getter).Get().(int)
}

// StringParam represents a registered string parameter.
type StringParam struct {
	RegisteredParam
}

// Get returns the string value of the parameter in the given flow.
func (p StringParam) Get(tf *TF) string {
	value := p.value(tf)
	return value.(flag.Getter).Get().(string)
}

// VerboseParam registers a boolean parameter that controls verbose output.
func VerboseParam(flow *Taskflow) BoolParam {
	param := flow.ConfigureBool(false, ParameterInfo{
		Name:  "v",
		Usage: "Verbose output: log all tasks as they are run. Also print all text from Log and Logf calls even if the task succeeds.",
	})
	flow.Verbose = &param
	return param
}
