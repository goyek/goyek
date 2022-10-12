// Build is the build pipeline for this repository.
package main

import (
	"io"
	"os"
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
	install := flow.Register(taskInstall())
	build := flow.Register(taskBuild())
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
		install,
		build,
		lint,
		test,
		modTidy,
		diff,
	}))
	flow.DefaultTask = all

	return flow
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
			if err := tf.Cmd("git", "clean", "-fX").Run(); err != nil {
				tf.Fatal(err)
			}
		},
	}
}

func taskInstall() goyek.Task {
	return goyek.Task{
		Name:  "install",
		Usage: " go install tools",
		Action: func(tf *goyek.TF) {
			tools := &strings.Builder{}
			toolsCmd := tf.Cmd("go", "list", `-f={{ join .Imports " " }}`, "-tags=tools")
			toolsCmd.Dir = toolsDir
			toolsCmd.Stdout = tools
			if err := toolsCmd.Run(); err != nil {
				tf.Fatal(err)
			}

			var installArgs []string
			for _, v := range strings.Split(tools.String(), " ") {
				installArgs = append(installArgs, strings.TrimSpace(v))
			}
			installCmd := tf.Cmd("go", append([]string{"install"}, installArgs...)...)
			installCmd.Dir = toolsDir
			if err := installCmd.Run(); err != nil {
				tf.Fatal(err)
			}
		},
	}
}

func taskBuild() goyek.Task {
	return goyek.Task{
		Name:  "build",
		Usage: "go build",
		Action: func(tf *goyek.TF) {
			if err := tf.Cmd("go", "build", "./...").Run(); err != nil {
				tf.Fatal(err)
			}
		},
	}
}

func taskMarkdownLint() goyek.Task {
	return goyek.Task{
		Name:  "markdownlint",
		Usage: "markdownlint-cli (requires docker)",
		Action: func(tf *goyek.TF) {
			curDir, err := os.Getwd()
			if err != nil {
				tf.Fatal(err)
			}

			dockerTag := "markdownlint-cli"
			if err := tf.Cmd("docker", "build", "-t", dockerTag, "-f", toolsDir+"/markdownlint-cli.dockerfile", ".").Run(); err != nil {
				tf.Fatal(err)
			}

			if err := tf.Cmd("docker", "run", "--rm", "-v", curDir+":/workdir", dockerTag, "**/*.md").Run(); err != nil {
				tf.Fatal(err)
			}
		},
	}
}

func taskMisspell() goyek.Task {
	return goyek.Task{
		Name:  "misspell",
		Usage: "misspell",
		Action: func(tf *goyek.TF) {
			misspell := tf.Cmd("misspell", "-error", "-locale=US", "-i=importas", "-w", ".")
			if err := misspell.Run(); err != nil {
				tf.Fatal(err)
			}
		},
	}
}

func taskGolangciLint() goyek.Task {
	return goyek.Task{
		Name:  "golangci-lint",
		Usage: "golangci-lint run --fix",
		Action: func(tf *goyek.TF) {
			exec(tf, rootDir, "golangci-lint run --fix")
			exec(tf, buildDir, "golangci-lint run --fix")
		},
	}
}

func taskTest() goyek.Task {
	return goyek.Task{
		Name:  "test",
		Usage: "go test with race detector and code covarage",
		Action: func(tf *goyek.TF) {
			if err := tf.Cmd("go", "test", "-race", "-covermode=atomic", "-coverprofile=coverage.out", "./...").Run(); err != nil {
				tf.Fatal(err)
			}
		},
	}
}

func taskModTidy() goyek.Task {
	return goyek.Task{
		Name:  "mod-tidy",
		Usage: "go mod tidy",
		Action: func(tf *goyek.TF) {
			exec(tf, rootDir, "go mod tidy")
			exec(tf, buildDir, "go mod tidy")
			exec(tf, toolsDir, "go mod tidy")
		},
	}
}

func taskDiff(ci goyek.RegisteredBoolParam) goyek.Task {
	return goyek.Task{
		Name:   "diff",
		Usage:  "git diff",
		Params: goyek.Params{ci},
		Action: func(tf *goyek.TF) {
			if !ci.Get(tf) {
				tf.Skip("ci param is not set, skipping")
			}

			if err := tf.Cmd("git", "diff", "--exit-code").Run(); err != nil {
				tf.Error(err)
			}

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

// exec runs the command in given directory.
// Calls tf.Error in failure.
func exec(tf *goyek.TF, workDir, cmdLine string) {
	tf.Logf("For %s:", workDir)
	args := strings.Split(cmdLine, " ")
	cmd := tf.Cmd(args[0], args[1:]...)
	cmd.Dir = workDir
	if err := cmd.Run(); err != nil {
		tf.Error(err)
	}
}
