package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/goyek/goyek"
)

func main() {
	flow().Main()
}

func flow() *goyek.Flow {
	flow := &goyek.Flow{}

	// parameters
	ci := flow.RegisterBoolParam(goyek.BoolParam{
		Name:  "ci",
		Usage: "Whether CI is calling the build script",
	})

	// tasks
	clean := flow.Register(taskClean())
	build := flow.Register(taskBuild())
	fmt := flow.Register(taskFmt())
	markdownlint := flow.Register(taskMarkdownLint())
	misspell := flow.Register(taskMisspell())
	golangciLint := flow.Register(taskGolangciLint())
	test := flow.Register(taskTest())
	modTidy := flow.Register(taskModTidy())
	diff := flow.Register(taskDiff(ci))

	// pipelines
	lint := flow.Register(taskLint(goyek.Deps{
		misspell,
		markdownlint,
		golangciLint,
	}))
	all := flow.Register(taskAll(goyek.Deps{
		clean,
		build,
		fmt,
		lint,
		test,
		modTidy,
		diff,
	}))
	flow.DefaultTask = all

	return flow
}

const buildDir = "build"

func taskClean() goyek.Task {
	return goyek.Task{
		Name:   "clean",
		Usage:  "remove git ignored files",
		Action: goyek.Exec("git", "clean", "-fX"),
	}
}

func taskBuild() goyek.Task {
	return goyek.Task{
		Name:   "build",
		Usage:  "go build",
		Action: goyek.Exec("go", "build", "./..."),
	}
}

func taskFmt() goyek.Task {
	return goyek.Task{
		Name:  "fmt",
		Usage: "gofumports",
		Action: func(a *goyek.A) {
			installFmt := a.Cmd("go", "install", "mvdan.cc/gofumpt/gofumports")
			installFmt.Dir = buildDir
			if err := installFmt.Run(); err != nil {
				a.Fatalf("go install gofumports: %v", err)
			}
			a.Cmd("gofumports", strings.Split("-l -w -local github.com/goyek/goyek .", " ")...).Run() //nolint // it is OK if it returns error
		},
	}
}

func taskMarkdownLint() goyek.Task {
	return goyek.Task{
		Name:  "markdownlint",
		Usage: "markdownlint-cli (requires docker)",
		Action: func(a *goyek.A) {
			curDir, err := os.Getwd()
			if err != nil {
				a.Fatal(err)
			}

			docsMount := curDir + ":/markdown"
			if err := a.Cmd("docker", "run", "-v", docsMount, "06kellyjac/markdownlint-cli:0.27.1", "**/*.md").Run(); err != nil {
				a.Error(err)
			}

			gitHubTemplatesMount := filepath.Join(curDir, ".github") + ":/markdown"
			if err := a.Cmd("docker", "run", "-v", gitHubTemplatesMount, "06kellyjac/markdownlint-cli:0.27.1", "**/*.md").Run(); err != nil {
				a.Error(err)
			}
		},
	}
}

func taskMisspell() goyek.Task {
	return goyek.Task{
		Name:  "misspell",
		Usage: "misspell",
		Action: func(a *goyek.A) {
			installFmt := a.Cmd("go", "install", "github.com/client9/misspell/cmd/misspell")
			installFmt.Dir = buildDir
			if err := installFmt.Run(); err != nil {
				a.Fatalf("go install misspell: %v", err)
			}
			lint := a.Cmd("misspell", "-error", "-locale=US", "-i=importas", ".")
			if err := lint.Run(); err != nil {
				a.Fatalf("misspell: %v", err)
			}
		},
	}
}

func taskGolangciLint() goyek.Task {
	return goyek.Task{
		Name:  "golangci-lint",
		Usage: "golangci-lint",
		Action: func(a *goyek.A) {
			installLint := a.Cmd("go", "install", "github.com/golangci/golangci-lint/cmd/golangci-lint")
			installLint.Dir = buildDir
			if err := installLint.Run(); err != nil {
				a.Fatalf("go install golangci-lint: %v", err)
			}
			lint := a.Cmd("golangci-lint", "run")
			if err := lint.Run(); err != nil {
				a.Fatalf("golangci-lint run: %v", err)
			}
		},
	}
}

func taskTest() goyek.Task {
	return goyek.Task{
		Name:   "test",
		Usage:  "go test with race detector and code covarage",
		Action: goyek.Exec("go", "test", "-race", "-covermode=atomic", "-coverprofile=coverage.out", "./..."),
	}
}

func taskModTidy() goyek.Task {
	return goyek.Task{
		Name:  "mod-tidy",
		Usage: "go mod tidy",
		Action: func(a *goyek.A) {
			if err := a.Cmd("go", "mod", "tidy").Run(); err != nil {
				a.Errorf("go mod tidy: %v", err)
			}

			toolsModTidy := a.Cmd("go", "mod", "tidy")
			toolsModTidy.Dir = buildDir
			if err := toolsModTidy.Run(); err != nil {
				a.Errorf("go mod tidy: %v", err)
			}
		},
	}
}

func taskDiff(ci goyek.RegisteredBoolParam) goyek.Task {
	return goyek.Task{
		Name:   "diff",
		Usage:  "git diff",
		Params: goyek.Params{ci},
		Action: func(a *goyek.A) {
			if !ci.Get(a) {
				a.Skip("ci param is not set, skipping")
			}

			if err := a.Cmd("git", "diff", "--exit-code").Run(); err != nil {
				a.Errorf("git diff: %v", err)
			}

			cmd := a.Cmd("git", "status", "--porcelain")
			sb := &strings.Builder{}
			cmd.Stdout = io.MultiWriter(a.Output(), sb)
			if err := cmd.Run(); err != nil {
				a.Errorf("git status --porcelain: %v", err)
			}
			if sb.Len() > 0 {
				a.Error("git status --porcelain returned output")
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
