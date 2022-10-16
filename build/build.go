// Build is the build pipeline for this repository.
package main

import (
	"fmt"
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

func main() { //nolint:funlen // build can be long as it is easier to read and maintain
	// tasks
	flow := &goyek.Flow{}

	clean := flow.Define(goyek.Task{
		Name:  "clean",
		Usage: "remove git ignored files",
		Action: func(tf *goyek.TF) {
			Exec(tf, rootDir, "git clean -fXd")
		},
	})

	mod := flow.Define(goyek.Task{
		Name:  "mod",
		Usage: "go mod tidy",
		Action: func(tf *goyek.TF) {
			Exec(tf, rootDir, "go mod tidy")
			Exec(tf, buildDir, "go mod tidy")
			Exec(tf, toolsDir, "go mod tidy")
		},
	})

	install := flow.Define(goyek.Task{
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
	})

	golint := flow.Define(goyek.Task{
		Name:  "golint",
		Usage: "golangci-lint run --fix",
		Action: func(tf *goyek.TF) {
			Exec(tf, rootDir, "golangci-lint run --fix")
			Exec(tf, buildDir, "golangci-lint run --fix")
		},
	})

	spell := flow.Define(goyek.Task{
		Name:  "spell",
		Usage: "misspell",
		Action: func(tf *goyek.TF) {
			Exec(tf, rootDir, "misspell -error -locale=US -i=importas -w .")
		},
	})

	mdlint := flow.Define(goyek.Task{
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
			Exec(tf, rootDir, "docker build -t "+dockerTag+" -f "+toolsDir+"/markdownlint-cli.dockerfile .")
			if tf.Failed() {
				return
			}
			Exec(tf, rootDir, "docker run --rm -v '"+curDir+":/workdir' "+dockerTag+" *.md")
		},
	})

	test := flow.Define(goyek.Task{
		Name:  "test",
		Usage: "go test",
		Action: func(tf *goyek.TF) {
			Exec(tf, rootDir, "go test -race -covermode=atomic -coverprofile=coverage.out ./...")
		},
	})

	diff := flow.Define(goyek.Task{
		Name:  "diff",
		Usage: "git diff",
		Action: func(tf *goyek.TF) {
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
	})

	// pipelines
	lint := flow.Define(goyek.Task{
		Name:  "lint",
		Usage: "all linters",
		Deps: goyek.Deps{
			golint,
			spell,
			mdlint,
		},
	})

	all := flow.Define(goyek.Task{
		Name:  "all",
		Usage: "build pipeline",
		Deps: goyek.Deps{
			clean,
			mod,
			install,
			lint,
			test,
		},
	})

	flow.Define(goyek.Task{
		Name:  "ci",
		Usage: "CI build pipeline",
		Deps: goyek.Deps{
			all,
			diff,
		},
	})

	// set the build pipeline as the default task
	flow.SetDefault(all)

	// change working directory to repo root
	if err := os.Chdir(".."); err != nil {
		fmt.Println(err)
		os.Exit(goyek.CodeInvalidArgs)
	}

	// run the build pipeline
	flow.Main(os.Args[1:])
}
