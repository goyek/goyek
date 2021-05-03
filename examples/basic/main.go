package main

import "github.com/goyek/goyek"

func main() {
	flow := &goyek.Taskflow{}

	hello := flow.Register(taskHello())
	fmt := flow.Register(taskFmt())

	flow.Register(goyek.Task{
		Name:  "all",
		Usage: "build pipeline",
		Deps: goyek.Deps{
			hello,
			fmt,
		},
	})

	flow.Main()
}

func taskHello() goyek.Task {
	return goyek.Task{
		Name:  "hello",
		Usage: "demonstration",
		Command: func(tf *goyek.TF) {
			tf.Log("Hello world!")
		},
	}
}

func taskFmt() goyek.Task {
	return goyek.Task{
		Name:    "fmt",
		Usage:   "go fmt",
		Command: goyek.Exec("go", "fmt", "./..."),
	}
}
