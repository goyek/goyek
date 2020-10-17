package main

import (
	"os"

	"github.com/stretchr/testify/require"

	"github.com/pellared/taskflow"
)

func main() {
	tasks := &taskflow.Taskflow{
		Verbose: true, // move to flags TODO
	}
	tasks.MustRegister(taskflow.Task{
		Name:        "test",
		Description: "go test",
		Command:     taskTest,
	})
	tasks.Main(os.Args...)
}

func taskTest(tf *taskflow.TF) {
	err := taskflow.Exec(tf, "go", "test", "-v")
	require.NoError(tf, err, "go test failed")
}
