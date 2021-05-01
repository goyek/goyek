// Example program for parameters, showcasing the following:
// Sharing of parameters, "private" parameters, and complex parameters encoded in JSON.
// This example also registers the "verbose" parameter, in order to provide output in the task.
// Execute `go run ./main.go -v --shared "hello world"` as a first example.
// Execute `go run ./main.go -h"` to see all details.

package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/pellared/taskflow"
)

func main() {
	flow := taskflow.New()

	taskflow.VerboseParam(flow)
	sharedParam := flow.RegisterStringParam("default-value", taskflow.ParameterInfo{
		Name:  "shared",
		Short: 's',
		Usage: "An example parameter shared between tasks",
	})

	first := flow.MustRegister(taskFirst(sharedParam))
	flow.MustRegister(taskSecond(flow, sharedParam))
	flow.MustRegister(taskComplexParam(flow))
	flow.DefaultTask = first

	flow.Main()
}

func taskFirst(sharedParam taskflow.StringParam) taskflow.Task {
	return taskflow.Task{
		Name:        "first",
		Description: "Showcases a simple parameter",
		Parameters:  taskflow.Params{sharedParam},
		Command: func(tf *taskflow.TF) {
			tf.Log("Shared parameter named '" + sharedParam.Name() + "', value '" + sharedParam.Get(tf) + "'")
		},
	}
}

func taskSecond(flow *taskflow.Taskflow, sharedParam taskflow.StringParam) taskflow.Task {
	// The following is a "private" parameter, only available to this task.
	privateParam := flow.RegisterStringParam("special-default", taskflow.ParameterInfo{
		Name:  "private",
		Short: 'p',
		Usage: "A task-specific parameter",
	})
	return taskflow.Task{
		Name:        "second",
		Description: "Showcases shared and task-specific parameters",
		Parameters:  taskflow.Params{sharedParam, privateParam},
		Command: func(tf *taskflow.TF) {
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
// While it is possible to implement the taskflow.Value interface on the
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
// Execute `go run ./main.go -v complex --json "{\"stringValue\":\"abc\"}"` as an example.
func taskComplexParam(flow *taskflow.Taskflow) taskflow.Task {
	privateParam := flow.RegisterValueParam(func() taskflow.Value {
		param := complexParamValue{
			StringValue: "default",
			IntValue:    123,
		}
		return &param
	}, taskflow.ParameterInfo{
		Name:  "json",
		Usage: "A complex parameter",
	})
	return taskflow.Task{
		Name:        "complex",
		Description: "Showcases complex parameters",
		Parameters:  taskflow.Params{privateParam},
		Command: func(tf *taskflow.TF) {
			tf.Log("Private parameter named '" + privateParam.Name() +
				"', value '" + fmt.Sprintf("%v", privateParam.Get(tf).(complexParam)) + "'")
		},
	}
}
