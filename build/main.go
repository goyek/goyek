package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pellared/taskflow"
)

func main() {
	tasks := &taskflow.Taskflow{}

	// tasks:
	clean := tasks.MustRegister(taskflow.Task{
		Name:        "clean",
		Description: "remove files created during build",
		Command:     taskClean,
	})

	install := tasks.MustRegister(taskflow.Task{
		Name:        "install",
		Description: "install build tools",
		Command:     taskInstall,
	})

	build := tasks.MustRegister(taskflow.Task{
		Name:        "build",
		Description: "go build",
		Command:     taskBuild,
	})

	fmt := tasks.MustRegister(taskflow.Task{
		Name:        "fmt",
		Description: "gofumports",
		Command:     taskFmt,
	})

	lint := tasks.MustRegister(taskflow.Task{
		Name:        "lint",
		Description: "golangci-lint-lintports",
		Command:     taskLint,
	})

	test := tasks.MustRegister(taskflow.Task{
		Name:        "test",
		Description: "go test with race detector and code covarage",
		Command:     taskTest,
	})

	modTidy := tasks.MustRegister(taskflow.Task{
		Name:        "mod-tidy",
		Description: "go mod tidy",
		Command:     taskModTidy,
	})

	diff := tasks.MustRegister(taskflow.Task{
		Name:        "diff",
		Description: "git diff",
		Command:     taskDiff,
	})

	// pipelines:
	dev := tasks.MustRegister(taskflow.Task{
		Name:        "dev",
		Description: "dev build",
		Dependencies: taskflow.Deps{
			clean,
			install,
			build,
			fmt,
			lint,
			test,
			modTidy,
		},
	})

	tasks.MustRegister(taskflow.Task{
		Name:        "ci",
		Description: "CI build",
		Dependencies: taskflow.Deps{
			dev,
			diff,
		},
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

func taskInstall(tf *taskflow.TF) {
	err := tf.Exec("tools", nil, "go", "install", "mvdan.cc/gofumpt/gofumports")
	assert.NoError(tf, err, "install gofumports failed")
	err = tf.Exec("tools", nil, "go", "install", "github.com/golangci/golangci-lint/cmd/golangci-lint")
	assert.NoError(tf, err, "install golangci-lint failed")
}

func taskBuild(tf *taskflow.TF) {
	err := tf.Exec("", nil, "go", "build", "-o", "/dev/null", "./...")
	assert.NoError(tf, err, "go build failed")
}

func taskFmt(tf *taskflow.TF) {
	tf.Exec("", nil, "gofumports", strings.Split("-l -w -local github.com/pellared/taskflow .", " ")...) //nolint // it is OK if it returns error
}

func taskLint(tf *taskflow.TF) {
	err := tf.Exec("", nil, "golangci-lint", "run")
	assert.NoError(tf, err, "linter failed")
}

func taskTest(tf *taskflow.TF) {
	err := tf.Exec("", nil, "go", "test", "-race", "-covermode=atomic", "-coverprofile=coverage.out", "./...")
	assert.NoError(tf, err, "go test failed")
	err = tf.Exec("", nil, "go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html")
	assert.NoError(tf, err, "go tool cover failed")
}

func taskModTidy(tf *taskflow.TF) {
	err := tf.Exec("", nil, "go", "mod", "tidy")
	assert.NoError(tf, err, "go mod tidy failed for root")
	err = tf.Exec("tools", nil, "go", "mod", "tidy")
	assert.NoError(tf, err, "go mod tidy failed for tools")
}

func taskDiff(tf *taskflow.TF) {
	err := tf.Exec("", nil, "git", "diff", "--exit-code")
	assert.NoError(tf, err, "git diff failed")

	tf.Logf("Exec: git status --porcelain")
	output := &strings.Builder{}
	cmd := exec.CommandContext(tf.Context(), "git", "status", "--porcelain")
	cmd.Stdout = output
	cmd.Stderr = output
	err = cmd.Run()
	assert.NoError(tf, err, "git status --porcelain failed")
	res := output.String()
	if res != "" {
		tf.Logf(res)
		tf.Errorf("git status --porcelain returned something")
	}
}
