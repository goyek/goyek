package main

import "github.com/pellared/taskflow"

func main() {
	flow := taskflow.New()

	taskflow.VerboseParam(flow)
	fmt := flow.MustRegister(taskFmt())
	test := flow.MustRegister(taskTest())

	flow.MustRegister(taskflow.Task{
		Name:        "all",
		Description: "build pipeline",
		Dependencies: taskflow.Deps{
			fmt,
			test,
		},
	})

	flow.Main()
}

func taskFmt() taskflow.Task {
	return taskflow.Task{
		Name:        "fmt",
		Description: "go fmt",
		Command:     taskflow.Exec("go", "fmt", "./..."),
	}
}

func taskTest() taskflow.Task {
	return taskflow.Task{
		Name:        "test",
		Description: "go test with race detector and code covarage",
		Command: func(tf *taskflow.TF) {
			if err := tf.Cmd("go", "test", "-race", "-covermode=atomic", "-coverprofile=coverage.out", "./...").Run(); err != nil {
				tf.Errorf("go test: %v", err)
			}
			if err := tf.Cmd("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html").Run(); err != nil {
				tf.Errorf("go tool cover: %v", err)
			}
		},
	}
}
