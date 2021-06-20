// Example program for parameters, showcasing the following:
// Sharing of parameters, "private" parameters, and complex parameters encoded in JSON.
// This example also registers the "verbose" parameter, in order to provide output in the task.
// Execute `go run . -v -shared "hello world"` as a first example.
// Execute `go run . -h"` to see all details.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/goyek/goyek"
)

func main() {
	flow := &goyek.Flow{}

	sharedParam := flow.RegisterStringParam(goyek.StringParam{
		Name:    "shared",
		Usage:   "An example parameter shared between tasks",
		Default: "default-value",
	})

	first := flow.Register(taskFirst(sharedParam))
	flow.Register(taskSecond(flow, sharedParam))
	flow.Register(taskComplexParam(flow))
	flow.Register(taskListParam(flow))
	flow.DefaultTask = first

	flow.Main()
}

func taskFirst(sharedParam goyek.RegisteredStringParam) goyek.Task {
	return goyek.Task{
		Name:   "first",
		Usage:  "Showcases a simple parameter",
		Params: goyek.Params{sharedParam},
		Command: func(tf *goyek.TF) {
			tf.Log("Shared parameter named '" + sharedParam.Name() + "', value '" + sharedParam.Get(tf) + "'")
		},
	}
}

func taskSecond(flow *goyek.Flow, sharedParam goyek.RegisteredStringParam) goyek.Task {
	// The following is a "private" parameter, only available to this task.
	privateParam := flow.RegisterStringParam(goyek.StringParam{
		Name:    "private",
		Usage:   "A task-specific parameter",
		Default: "special-default",
	})
	return goyek.Task{
		Name:   "second",
		Usage:  "Showcases shared and task-specific parameters",
		Params: goyek.Params{sharedParam, privateParam},
		Command: func(tf *goyek.TF) {
			tf.Log("Shared parameter named '" + sharedParam.Name() + "', value '" + sharedParam.Get(tf) + "'")
			tf.Log("Private parameter named '" + privateParam.Name() + "', value '" + privateParam.Get(tf) + "'")
		},
	}
}

// complexParam is an example for a serialized complex type.
type complexParam struct {
	StringValue string `json:"stringValue"`
	IntValue    int    `json:"intValue"`
}

// complexParamValue is a wrapper over the complex type, serializing it as JSON.
// While it is possible to implement the goyek.Value interface on the
// complex parameter itself, it is better to separate these concerns.
type complexParamValue complexParam

func (value *complexParamValue) Set(s string) error {
	err := json.Unmarshal([]byte(s), value)
	if err != nil {
		err = errors.New("parse error")
	}
	return err
}

func (value *complexParamValue) Get() interface{} {
	return complexParam(*value)
}

func (value *complexParamValue) String() string {
	bytes, _ := json.Marshal(value)
	return string(bytes)
}

func (value *complexParamValue) IsBool() bool {
	return false
}

// taskComplexParam showcases complex parameters, JSON encoded.
//
// Execute `go run . -v complex -json "{\"stringValue\":\"abc\"}"` as an example.
func taskComplexParam(flow *goyek.Flow) goyek.Task {
	privateParam := flow.RegisterValueParam(goyek.ValueParam{
		Name:  "json",
		Usage: "A complex parameter",
		NewValue: func() goyek.ParamValue {
			return &complexParamValue{
				StringValue: "default",
				IntValue:    123,
			}
		},
	})
	return goyek.Task{
		Name:   "complex",
		Usage:  "Showcases complex parameters",
		Params: goyek.Params{privateParam},
		Command: func(tf *goyek.TF) {
			tf.Log("Private parameter named '" + privateParam.Name() +
				"', value '" + fmt.Sprintf("%v", privateParam.Get(tf).(complexParam)) + "'")
		},
	}
}

type listParamValue []int

func (value *listParamValue) Set(s string) error {
	v, err := strconv.Atoi(s)
	if err != nil {
		err = errors.New("parse error")
	}
	*value = append(*value, v)
	return err
}

func (value *listParamValue) Get() interface{} {
	return []int(*value)
}

func (value *listParamValue) String() string {
	return fmt.Sprintf("%v", *value)
}

func (value *listParamValue) IsBool() bool {
	return false
}

// taskListParam showcases repeatable parameters.
//
// Execute `go run . -v list -port 1 -port 2 -port 3` as an example.
func taskListParam(flow *goyek.Flow) goyek.Task {
	privateParam := flow.RegisterValueParam(goyek.ValueParam{
		Name:  "port",
		Usage: "Integer parameter, can be repeated",
		NewValue: func() goyek.ParamValue {
			var value listParamValue
			return &value
		},
	})
	return goyek.Task{
		Name:   "list",
		Usage:  "Showcases list parameters",
		Params: goyek.Params{privateParam},
		Command: func(tf *goyek.TF) {
			tf.Log("Private parameter named '" + privateParam.Name() +
				"', value " + fmt.Sprintf("%v", privateParam.Get(tf).([]int)))
		},
	}
}
