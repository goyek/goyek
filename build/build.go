// Build is the build pipeline for this repository.
package main

import (
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/goyek/goyek/v2"
)

const (
	rootDir  = "."
	buildDir = "build"
	toolsDir = "tools"
)

func taskClean() goyek.Task {
	return goyek.Task{
		Name:  "clean",
		Usage: "remove git ignored files",
		Action: func(tf *goyek.TF) {
			Exec(tf, rootDir, "git clean -fXd")
		},
	}
}

func taskInstall() goyek.Task {
	return goyek.Task{
		Name:  "install",
		Usage: "go install tools",
		Action: func(tf *goyek.TF) {
			tools := &strings.Builder{}
			toolsCmd := tf.Cmd("go", "list", `-f={{ join .Imports " " }}`, "-tags=tools")
			toolsCmd.Dir = toolsDir
			toolsCmd.Stdout = tools
			if err := toolsCmd.Run(); err != nil {
				tf.Fatal(err)
			}

			Exec(tf, toolsDir, "go install "+strings.TrimSpace(tools.String()))
		},
	}
}

func taskBuild() goyek.Task {
	return goyek.Task{
		Name:  "build",
		Usage: "go build",
		Action: func(tf *goyek.TF) {
			Exec(tf, rootDir, "go build ./...")
		},
	}
}

func taskMarkdownLint() goyek.Task {
	return goyek.Task{
		Name:  "mdlint",
		Usage: "markdownlint-cli (uses docker)",
		Action: func(tf *goyek.TF) {
			if _, err := exec.LookPath("docker"); err != nil {
				tf.Skip(err)
			}

			curDir, err := os.Getwd()
			if err != nil {
				tf.Fatal(err)
			}

			dockerTag := "markdownlint-cli"
			tf.Log(`Cmd "docker build"`)
			if err := tf.Cmd("docker", "build", "-t", dockerTag, "-f", toolsDir+"/markdownlint-cli.dockerfile", ".").Run(); err != nil {
				tf.Fatal(err)
			}
			tf.Log(`"Cmd "docker run"`)
			if err := tf.Cmd("docker", "run", "--rm", "-v", curDir+":/workdir", dockerTag, "**/*.md").Run(); err != nil {
				tf.Fatal(err)
			}
		},
	}
}

func taskMisspell() goyek.Task {
	return goyek.Task{
		Name:  "spell",
		Usage: "misspell",
		Action: func(tf *goyek.TF) {
			Exec(tf, rootDir, "misspell -error -locale=US -i=importas -w .")
		},
	}
}

func taskGolangciLint() goyek.Task {
	return goyek.Task{
		Name:  "golint",
		Usage: "golangci-lint run --fix",
		Action: func(tf *goyek.TF) {
			Exec(tf, rootDir, "golangci-lint run --fix")
			Exec(tf, buildDir, "golangci-lint run --fix")
		},
	}
}

func taskTest() goyek.Task {
	return goyek.Task{
		Name:  "test",
		Usage: "go test",
		Action: func(tf *goyek.TF) {
			Exec(tf, rootDir, "go test -race -covermode=atomic -coverprofile=coverage.out ./...")
		},
	}
}

func taskModTidy() goyek.Task {
	return goyek.Task{
		Name:  "mod",
		Usage: "go mod tidy",
		Action: func(tf *goyek.TF) {
			Exec(tf, rootDir, "go mod tidy")
			Exec(tf, buildDir, "go mod tidy")
			Exec(tf, toolsDir, "go mod tidy")
		},
	}
}

func taskDiff(ci *bool) goyek.Task {
	return goyek.Task{
		Name:  "diff",
		Usage: "git diff",
		Action: func(tf *goyek.TF) {
			if !*ci {
				tf.Skip("ci param is not set, skipping")
			}

			Exec(tf, rootDir, "git diff --exit-code")

			tf.Log("Cmd: git status --porcelain")
			cmd := tf.Cmd("git", "status", "--porcelain")
			sb := &strings.Builder{}
			cmd.Stdout = io.MultiWriter(tf.Output(), sb)
			if err := cmd.Run(); err != nil {
				tf.Error(err)
			}
			if sb.Len() > 0 {
				tf.Error("git status --porcelain returned output")
			}
		},
	}
}

func taskLint(deps goyek.Deps) goyek.Task {
	return goyek.Task{
		Name:  "lint",
		Usage: "all linters",
		Deps:  deps,
	}
}

func taskAll(deps goyek.Deps) goyek.Task {
	return goyek.Task{
		Name:  "all",
		Usage: "build pipeline",
		Deps:  deps,
	}
}
