package main

import (
	"github.com/pellared/taskflow"
)

// Example program for parameter showcase.
// This example also registers the "verbose" parameter, in order to provide output in the task.
// Execute `go run ./main.go -v --example "hello world"` as an example.
//
// The example also registers two tasks to show that parameters can be shared between tasks.

func main() {
	flow := taskflow.New()

	taskflow.VerboseParam(flow)
	exampleParam := flow.RegisterStringParam("default-value", taskflow.ParameterInfo{
		Name:  "example",
		Short: 'e',
		Usage: "An example parameter",
	})

	show := flow.MustRegister(taskShow(exampleParam))
	flow.MustRegister(taskSecond(flow, exampleParam))
	flow.DefaultTask = show

	flow.Main()
}

func taskShow(exampleParam taskflow.StringParam) taskflow.Task {
	return taskflow.Task{
		Name:        "show",
		Description: "Showcases a registered parameter",
		Parameters:  taskflow.Params{exampleParam},
		Command: func(tf *taskflow.TF) {
			tf.Log("Parameter named '" + exampleParam.Name() + "', value '" + exampleParam.Get(tf) + "'")
		},
	}
}

func taskSecond(flow *taskflow.Taskflow, exampleParam taskflow.StringParam) taskflow.Task {
	// The following is a "private" parameter, only available to this task.
	specialParam := flow.RegisterStringParam("special-default", taskflow.ParameterInfo{
		Name:  "special",
		Short: 's',
		Usage: "Another example parameter",
	})
	return taskflow.Task{
		Name:        "second",
		Description: "Showcases sharing of registered parameter",
		Parameters:  taskflow.Params{exampleParam, specialParam},
		Command: func(tf *taskflow.TF) {
			tf.Log("Parameter named '" + exampleParam.Name() + "', value '" + exampleParam.Get(tf) + "'")
			tf.Log("Parameter named '" + specialParam.Name() + "', value '" + specialParam.Get(tf) + "'")
		},
	}
}
