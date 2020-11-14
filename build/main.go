package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
		Description: "dev build pipeline",
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
		Description: "CI build pipeline",
		Dependencies: taskflow.Deps{
			dev,
			diff,
		},
	})

	tasks.Main()
}

func taskClean(tf *taskflow.TF) {
	files, err := filepath.Glob("coverage.*")
	if err != nil {
		tf.Fatalf("glob failed: %v", err)
	}
	for _, file := range files {
		err := os.Remove(file)
		if err != nil {
			tf.Errorf("failed to remove %s: %v", file, err)
			continue
		}
		tf.Logf("removed %s", file)
	}
}

func taskInstall(tf *taskflow.TF) {
	if err := tf.Exec("tools", nil, "go", "install", "mvdan.cc/gofumpt/gofumports"); err != nil {
		tf.Errorf("go install gofumports: %v", err)
	}
	if err := tf.Exec("tools", nil, "go", "install", "github.com/golangci/golangci-lint/cmd/golangci-lint"); err != nil {
		tf.Errorf("go install golangci-lint: %v", err)
	}
}

func taskBuild(tf *taskflow.TF) {
	if err := tf.Exec("", nil, "go", "build", "./..."); err != nil {
		tf.Errorf("go build: %v", err)
	}
}

func taskFmt(tf *taskflow.TF) {
	tf.Exec("", nil, "gofumports", strings.Split("-l -w -local github.com/pellared/taskflow .", " ")...) //nolint // it is OK if it returns error
}

func taskLint(tf *taskflow.TF) {
	if err := tf.Exec("", nil, "golangci-lint", "run"); err != nil {
		tf.Errorf("golangci-lint: %v", err)
	}
}

func taskTest(tf *taskflow.TF) {
	if err := tf.Exec("", nil, "go", "test", "-race", "-covermode=atomic", "-coverprofile=coverage.out", "./..."); err != nil {
		tf.Errorf("go test: %v", err)
	}
	if err := tf.Exec("", nil, "go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html"); err != nil {
		tf.Errorf("go tool cover: %v", err)
	}
}

func taskModTidy(tf *taskflow.TF) {
	if err := tf.Exec("", nil, "go", "mod", "tidy"); err != nil {
		tf.Errorf("go mod tidy: %v", err)
	}
	if err := tf.Exec("tools", nil, "go", "mod", "tidy"); err != nil {
		tf.Errorf("go mod tidy: %v", err)
	}
}

func taskDiff(tf *taskflow.TF) {
	if err := tf.Exec("", nil, "git", "diff", "--exit-code"); err != nil {
		tf.Errorf("git diff: %v", err)
	}

	sb := &strings.Builder{}
	output := io.MultiWriter(tf.Output(), sb)
	tf.Logf("Exec: git status --porcelain")
	cmd := exec.CommandContext(tf.Context(), "git", "status", "--porcelain")
	cmd.Stdout = output
	cmd.Stderr = output
	if err := cmd.Run(); err != nil {
		tf.Errorf("git status --porcelain: %v", err)
	}
	if sb.Len() > 0 {
		tf.Errorf("git status --porcelain returned output")
	}
}
