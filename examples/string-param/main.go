package main

import (
	"github.com/pellared/taskflow"
)

// Example program for parameter showcase.
// This example also registers the "verbose" parameter, in order to provide output in the task.
// Execute `go run ./main.go -v --example "hello world"` as an example.

func main() {
	flow := taskflow.New()

	taskflow.VerboseParam(flow)
	exampleParam := flow.ConfigureString("default-value", taskflow.ParameterInfo{
		Name:  "example",
		Short: 'e',
		Usage: "An example parameter",
	})

	show := flow.MustRegister(taskShow(exampleParam))
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
