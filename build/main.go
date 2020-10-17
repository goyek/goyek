package main

import (
	"os"

	"github.com/stretchr/testify/require"

	"github.com/pellared/taskflow"
)

func main() {
	tasks := &taskflow.Taskflow{}

	test := tasks.MustRegister(taskflow.Task{
		Name:        "test",
		Description: "go test",
		Command:     taskTest,
	})

	tasks.MustRegister(taskflow.Task{
		Name:         "dev",
		Description:  "dev build",
		Dependencies: taskflow.Deps{test},
	})

	tasks.Main(os.Args...)
}

func taskTest(tf *taskflow.TF) {
	err := taskflow.Exec(tf, "go", "test", "-v")
	require.NoError(tf, err, "go test failed")
}
