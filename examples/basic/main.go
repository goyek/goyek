package main

import "github.com/pellared/taskflow"

func main() {
	flow := taskflow.New()

	hello := flow.Register(taskHello())
	fmt := flow.Register(taskFmt())

	flow.Register(taskflow.Task{
		Name:  "all",
		Usage: "build pipeline",
		Deps: taskflow.Deps{
			hello,
			fmt,
		},
	})

	flow.Main()
}

func taskHello() taskflow.Task {
	return taskflow.Task{
		Name:  "hello",
		Usage: "demonstration",
		Command: func(tf *taskflow.TF) {
			tf.Log("Hello world!")
		},
	}
}

func taskFmt() taskflow.Task {
	return taskflow.Task{
		Name:    "fmt",
		Usage:   "go fmt",
		Command: taskflow.Exec("go", "fmt", "./..."),
	}
}
