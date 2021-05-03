package main

import (
	"io"
	"strings"

	"github.com/goyek/goyek"
)

func main() {
	flow := goyek.New()

	ci := flow.RegisterBoolParam(false, goyek.ParamInfo{
		Name:  "ci",
		Usage: "Whether CI is calling the build script",
	})

	// tasks
	clean := flow.Register(taskClean())
	build := flow.Register(taskBuild())
	fmt := flow.Register(taskFmt())
	lint := flow.Register(taskLint())
	test := flow.Register(taskTest())
	modTidy := flow.Register(taskModTidy())
	diff := flow.Register(taskDiff(ci))

	// pipeline
	all := flow.Register(goyek.Task{
		Name:  "all",
		Usage: "build pipeline",
		Deps: goyek.Deps{
			clean,
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

func taskClean() goyek.Task {
	return goyek.Task{
		Name:    "clean",
		Usage:   "remove git ignored files",
		Command: goyek.Exec("git", "clean", "-fX"),
	}
}

func taskBuild() goyek.Task {
	return goyek.Task{
		Name:    "build",
		Usage:   "go build",
		Command: goyek.Exec("go", "build", "./..."),
	}
}

func taskFmt() goyek.Task {
	return goyek.Task{
		Name:  "fmt",
		Usage: "gofumports",
		Command: func(tf *goyek.TF) {
			installFmt := tf.Cmd("go", "install", "mvdan.cc/gofumpt/gofumports")
			installFmt.Dir = toolsDir
			if err := installFmt.Run(); err != nil {
				tf.Errorf("go install gofumports: %v", err)
			}
			tf.Cmd("gofumports", strings.Split("-l -w -local github.com/goyek/goyek .", " ")...).Run() //nolint // it is OK if it returns error
		},
	}
}

func taskLint() goyek.Task {
	return goyek.Task{
		Name:  "lint",
		Usage: "golangci-lint",
		Command: func(tf *goyek.TF) {
			installLint := tf.Cmd("go", "install", "github.com/golangci/golangci-lint/cmd/golangci-lint")
			installLint.Dir = toolsDir
			if err := installLint.Run(); err != nil {
				tf.Errorf("go install golangci-lint: %v", err)
			}
			lint := tf.Cmd("golangci-lint", "run")
			if err := lint.Run(); err != nil {
				tf.Errorf("golangci-lint run: %v", err)
			}
		},
	}
}

func taskTest() goyek.Task {
	return goyek.Task{
		Name:    "test",
		Usage:   "go test with race detector and code covarage",
		Command: goyek.Exec("go", "test", "-race", "-covermode=atomic", "-coverprofile=coverage.out", "./..."),
	}
}

func taskModTidy() goyek.Task {
	return goyek.Task{
		Name:  "mod-tidy",
		Usage: "go mod tidy",
		Command: func(tf *goyek.TF) {
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

func taskDiff(ci goyek.RegisteredBoolParam) goyek.Task {
	return goyek.Task{
		Name:   "diff",
		Usage:  "git diff",
		Params: goyek.Params{ci},
		Command: func(tf *goyek.TF) {
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
