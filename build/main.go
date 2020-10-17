package main

import (
	"os"
	"os/exec"

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
	cmd := exec.CommandContext(tf.Context(), "go", "test", "-v")
	cmd.Stderr = tf.Writer()
	cmd.Stdout = tf.Writer()
	err := cmd.Run()
	require.NoError(tf, err, "go test failed")
}
