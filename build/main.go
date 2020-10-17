package main

import (
	"os"
	"path/filepath"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pellared/taskflow"
)

func main() {
	tasks := &taskflow.Taskflow{}

	clean := tasks.MustRegister(taskflow.Task{
		Name:        "clean",
		Description: "remove files created during build",
		Command:     taskClean,
	})

	test := tasks.MustRegister(taskflow.Task{
		Name:        "test",
		Description: "go test",
		Command:     taskTest,
	})

	tasks.MustRegister(taskflow.Task{
		Name:         "dev",
		Description:  "dev build",
		Dependencies: taskflow.Deps{clean, test},
	})

	tasks.Main()
}

func taskClean(tf *taskflow.TF) {
	files, err := filepath.Glob("coverage.*")
	require.NoError(tf, err, "bad pattern")
	for _, file := range files {
		err := os.Remove(file)
		assert.NoError(tf, err, "failed to remove %s", file)
		tf.Logf("removed %s", file)
	}
}

func taskTest(tf *taskflow.TF) {
	err := taskflow.Exec(tf, "go", "test", "-v")
	require.NoError(tf, err, "go test failed")
}
