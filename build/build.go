// Build is the build pipeline for this repository.
package main

import (
	"flag"
	"io"
	"os"
	"strings"

	"github.com/goyek/goyek/v2"
)

func configure() {
	// flags
	flag.BoolVar(&flow.Verbose, "v", false, "print all tasks as they are run")
	ci := flag.Bool("ci", false, "whether CI is calling")

	// tasks
	clean := flow.Define(taskClean())
	modTidy := flow.Define(taskModTidy())
	install := flow.Define(taskInstall())
	build := flow.Define(taskBuild())
	markdownlint := flow.Define(taskMarkdownLint())
	misspell := flow.Define(taskMisspell())
	golangciLint := flow.Define(taskGolangciLint())
	test := flow.Define(taskTest())
	diff := flow.Define(taskDiff(ci))

	// pipelines
	lint := flow.Define(taskLint(goyek.Deps{
		misspell,
		markdownlint,
		golangciLint,
	}))
	all := flow.Define(taskAll(goyek.Deps{
		clean,
		modTidy,
		install,
		build,
		lint,
		test,
		diff,
	}))

	// set default task
	flow.SetDefault(all)
}

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
			if err := tf.Cmd("git", "clean", "-fXd").Run(); err != nil {
				tf.Error(err)
			}
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

			if err := Exec(tf, toolsDir, "go install "+strings.TrimSpace(tools.String())); err != nil {
				tf.Error(err)
			}
		},
	}
}

func taskBuild() goyek.Task {
	return goyek.Task{
		Name:  "build",
		Usage: "go build",
		Action: func(tf *goyek.TF) {
			if err := Exec(tf, rootDir, "go build ./..."); err != nil {
				tf.Error(err)
			}
		},
	}
}

func taskMarkdownLint() goyek.Task {
	return goyek.Task{
		Name:  "mdlint",
		Usage: "markdownlint-cli (requires docker)",
		Action: func(tf *goyek.TF) {
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
			if err := Exec(tf, rootDir, "misspell -error -locale=US -i=importas -w ."); err != nil {
				tf.Error(err)
			}
		},
	}
}

func taskGolangciLint() goyek.Task {
	return goyek.Task{
		Name:  "golint",
		Usage: "golangci-lint run --fix",
		Action: func(tf *goyek.TF) {
			if err := Exec(tf, rootDir, "golangci-lint run --fix"); err != nil {
				tf.Error(err)
			}
			if err := Exec(tf, buildDir, "golangci-lint run --fix"); err != nil {
				tf.Error(err)
			}
		},
	}
}

func taskTest() goyek.Task {
	return goyek.Task{
		Name:  "test",
		Usage: "go test",
		Action: func(tf *goyek.TF) {
			if err := Exec(tf, rootDir, "go test -race -covermode=atomic -coverprofile=coverage.out ./..."); err != nil {
				tf.Error(err)
			}
		},
	}
}

func taskModTidy() goyek.Task {
	return goyek.Task{
		Name:  "mod",
		Usage: "go mod tidy",
		Action: func(tf *goyek.TF) {
			if err := Exec(tf, rootDir, "go mod tidy"); err != nil {
				tf.Error(err)
			}
			if err := Exec(tf, buildDir, "go mod tidy"); err != nil {
				tf.Error(err)
			}
			if err := Exec(tf, toolsDir, "go mod tidy"); err != nil {
				tf.Error(err)
			}
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

			if err := Exec(tf, rootDir, "git diff --exit-code"); err != nil {
				tf.Error(err)
			}

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
