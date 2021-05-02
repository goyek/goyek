package main

import (
	"io"
	"strings"

	"github.com/pellared/taskflow"
)

func main() {
	flow := taskflow.New()

	ci := flow.RegisterBoolParam(false, taskflow.ParamInfo{
		Name:  "ci",
		Usage: "Whether CI is calling the build script",
	})

	// tasks
	clean := flow.Register(taskClean())
	install := flow.Register(taskInstall())
	build := flow.Register(taskBuild())
	fmt := flow.Register(taskFmt())
	lint := flow.Register(taskLint())
	test := flow.Register(taskTest())
	modTidy := flow.Register(taskModTidy())
	diff := flow.Register(taskDiff(ci))

	// pipeline
	all := flow.Register(taskflow.Task{
		Name:  "all",
		Usage: "build pipeline",
		Deps: taskflow.Deps{
			clean,
			install,
			build,
			fmt,
			lint,
			test,
			modTidy,
			diff,
		},
	})

	flow.DefaultTask = all
	flow.Main()
}

const toolsDir = "tools"

func taskClean() taskflow.Task {
	return taskflow.Task{
		Name:    "clean",
		Usage:   "remove git ignored files",
		Command: taskflow.Exec("git", "clean", "-fX"),
	}
}

func taskInstall() taskflow.Task {
	return taskflow.Task{
		Name:  "install",
		Usage: "install build tools",
		Command: func(tf *taskflow.TF) {
			installFmt := tf.Cmd("go", "install", "mvdan.cc/gofumpt/gofumports")
			installFmt.Dir = toolsDir
			if err := installFmt.Run(); err != nil {
				tf.Errorf("go install gofumports: %v", err)
			}

			installLint := tf.Cmd("go", "install", "github.com/golangci/golangci-lint/cmd/golangci-lint")
			installLint.Dir = toolsDir
			if err := installLint.Run(); err != nil {
				tf.Errorf("go install golangci-lint: %v", err)
			}
		},
	}
}

func taskBuild() taskflow.Task {
	return taskflow.Task{
		Name:    "build",
		Usage:   "go build",
		Command: taskflow.Exec("go", "build", "./..."),
	}
}

func taskFmt() taskflow.Task {
	return taskflow.Task{
		Name:  "fmt",
		Usage: "gofumports",
		Command: func(tf *taskflow.TF) {
			tf.Cmd("gofumports", strings.Split("-l -w -local github.com/pellared/taskflow .", " ")...).Run() //nolint // it is OK if it returns error
		},
	}
}

func taskLint() taskflow.Task {
	return taskflow.Task{
		Name:    "lint",
		Usage:   "golangci-lint",
		Command: taskflow.Exec("golangci-lint", "run"),
	}
}

func taskTest() taskflow.Task {
	return taskflow.Task{
		Name:    "test",
		Usage:   "go test with race detector and code covarage",
		Command: taskflow.Exec("go", "test", "-race", "-covermode=atomic", "-coverprofile=coverage.out", "./..."),
	}
}

func taskModTidy() taskflow.Task {
	return taskflow.Task{
		Name:  "mod-tidy",
		Usage: "go mod tidy",
		Command: func(tf *taskflow.TF) {
			if err := tf.Cmd("go", "mod", "tidy").Run(); err != nil {
				tf.Errorf("go mod tidy: %v", err)
			}

			toolsModTidy := tf.Cmd("go", "mod", "tidy")
			toolsModTidy.Dir = toolsDir
			if err := toolsModTidy.Run(); err != nil {
				tf.Errorf("go mod tidy: %v", err)
			}
		},
	}
}

func taskDiff(ci taskflow.BoolParam) taskflow.Task {
	return taskflow.Task{
		Name:   "diff",
		Usage:  "git diff",
		Params: taskflow.Params{ci},
		Command: func(tf *taskflow.TF) {
			if !ci.Get(tf) {
				tf.Skip("ci param is not set, skipping")
			}

			if err := tf.Cmd("git", "diff", "--exit-code").Run(); err != nil {
				tf.Errorf("git diff: %v", err)
			}

			cmd := tf.Cmd("git", "status", "--porcelain")
			sb := &strings.Builder{}
			cmd.Stdout = io.MultiWriter(tf.Output(), sb)
			if err := cmd.Run(); err != nil {
				tf.Errorf("git status --porcelain: %v", err)
			}
			if sb.Len() > 0 {
				tf.Error("git status --porcelain returned output")
			}
		},
	}
}
