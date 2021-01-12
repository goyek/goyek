# taskflow

[![go.dev](https://img.shields.io/badge/go.dev-reference-blue.svg)](https://pkg.go.dev/github.com/pellared/taskflow)
[![go.mod](https://img.shields.io/github/go-mod/go-version/pellared/taskflow)](go.mod)
[![Build Status](https://img.shields.io/github/workflow/status/pellared/taskflow/build)](https://github.com/pellared/taskflow/actions?query=workflow%3Abuild+branch%3Amaster)
[![Go Report Card](https://goreportcard.com/badge/github.com/pellared/taskflow)](https://goreportcard.com/report/github.com/pellared/taskflow)
[![codecov](https://codecov.io/gh/pellared/taskflow/branch/master/graph/badge.svg)](https://codecov.io/gh/pellared/taskflow)

This package aims to simplify the creation of build pipelines in Go instead of using scripts or [Make](https://www.gnu.org/software/make/).

**taskflow** API is mainly inspired by the [testing](https://golang.org/pkg/testing), [http](https://golang.org/pkg/http) and [flag](https://golang.org/pkg/flag) packages.

Check [Go Build Pipeline Demo](https://github.com/pellared/go-build-pipeline-demo) to compare **taskflow** with [Make](https://www.gnu.org/software/make/) and [Mage](https://github.com/magefile/mage).

I am open to any feedback and contribution. Use [Discussions](https://github.com/pellared/taskflow/discussions) or write to me: *Robert Pajak* @ [Gophers Slack](https://invite.slack.golangbridge.org/).

`Star` this repository if you find it valuable and worth maintaining.

## Example

Paste the following code to `build/build.go`:

```go
package main

import "github.com/pellared/taskflow"

func main() {
	flow := taskflow.New()

	fmt := flow.MustRegister(taskFmt())
	test := flow.MustRegister(taskTest())

	flow.MustRegister(taskflow.Task{
		Name:        "all",
		Description: "build pipeline",
		Dependencies: taskflow.Deps{
			fmt,
			test,
		},
	})

	flow.Main()
}

func taskFmt() taskflow.Task {
	return taskflow.Task{
		Name:        "fmt",
		Description: "go fmt",
		Command:     taskflow.Exec("go", "fmt", "./..."),
	}
}

func taskTest() taskflow.Task {
	return taskflow.Task{
		Name:        "test",
		Description: "go test with race detector and code covarage",
		Command: func(tf *taskflow.TF) {
			if err := tf.Cmd("go", "test", "-race", "-covermode=atomic", "-coverprofile=coverage.out", "./...").Run(); err != nil {
				tf.Errorf("go test: %v", err)
			}
			if err := tf.Cmd("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html").Run(); err != nil {
				tf.Errorf("go tool cover: %v", err)
			}	
		},
	}
}
```

Sample usage:

```shell
$ go run ./build -h
Usage: [flag(s)] [key=val] task(s)
Flags:
  -v    Verbose output: log all tasks as they are run. Also print all text from Log and Logf calls even if the task succeeds.
Tasks:
  all     build pipeline
  fmt     go fmt
  test    go test with race detector and code covarage

$ go run ./build all
ok     0.453s

$ go run ./build -v all
===== TASK  fmt
Exec: go fmt ./...
----- PASS: fmt (0.06s)
===== TASK  test
Exec: go test -race -covermode=atomic -coverprofile=coverage.out ./...
?       github.com/pellared/taskflow/example    [no test files]
Exec: go tool cover -html=coverage.out -o coverage.html
----- PASS: test (0.11s)
ok      0.176s
```

Tired of writing `go run ./build` each time? Just add an alias to your shell. For example by adding the line below to `~/.bash_aliases`:

```shell
alias gake='go run ./build'
```

Additionally, take a look at the dogfooding [build pipeline](build/build.go).

## Features

### Task registration

The registered tasks are required to have a non-empty name. For future compatibility, it is strongly suggested to use only the following characters: letters (`a-z` and `A-Z`), digits (`0-9`), underscode (`_`), hyphen (`-`). Do not use equals character (`=`) as it is resvered for assigning parameters. A task which a given name can be only registered once. 

### Task dependencies

During task registration it is possible to add a dependency to an already registered task. When taskflow is processed, it makes sure that the dependency is executed before current task is run. Take notice that each task will be executed at most once.

### Task command

Task command is a function which is executed when a task is executed.  It is not required to to set a command. Not having a command is very handy when registering "pipelines".

### Task runner

You can use [`type Runner`](https://pkg.go.dev/github.com/pellared/taskflow#Runner) for testing the execution of a single command. It may be handy e.g. during development of a new task, when debugging some issue or if you want to have a test suite for your reusable commands.

### Verbose mode

Verbose mode which works like for `go test`. When enabled, test output is streamed when go test -v is used. If disabled, only logs from failed task are send to output.

Verbose mode for the whole task flow can be set via CLI flag by setting `-v` or by setting `Verbose` to `true` in [`type Taskflow`](https://pkg.go.dev/github.com/pellared/taskflow#Taskflow).

It is also possible to set the Verbose mode for the [`type Runner`](https://pkg.go.dev/github.com/pellared/taskflow#Runner) which can be used when testing or debugging a single task's command.

Use [`func (*TF) Verbose`](https://pkg.go.dev/github.com/pellared/taskflow#TF.Verbose) to check if verbose mode was set within the task's command.

### Parameters

The task commands can get the parameters using [`func (*TF) Params`](https://pkg.go.dev/github.com/pellared/taskflow#TF.Params). 

The parameters can be set via CLI using `key=val` syntax after CLI flags. For example `go run ./build -v ci=true all` would run the `all` task with `ci` parameter set to `"true"` in verbose mode.

Default values can be assigned via `Params` field in [`type Taskflow`](https://pkg.go.dev/github.com/pellared/taskflow#Taskflow).

[`type Params`](https://pkg.go.dev/github.com/pellared/taskflow#Params) contains convenient conversion methods like [`func (p Params) Int(key string) (int, error)`](https://pkg.go.dev/github.com/pellared/taskflow#Params.Int).

### Helpers for running programs

Use [`func Exec(name string, args ...string) func(*TF)`](https://pkg.go.dev/github.com/pellared/taskflow#Exec) to create a task's command which only runs a single program.

Use [`func (tf *TF) Cmd(name string, args ...string) *exec.Cmd`](https://pkg.go.dev/github.com/pellared/taskflow#TF.Cmd) if within a task's command function when you want to execute more programs or you need more granular control.

## FAQ

### Is taskflow stable

No, it is in experimental phase.

### Why not to use Make

While [Make](https://www.gnu.org/software/make/) is currently de facto standard, it has some pitfalls:

- Requires to learn Make which is not so easy.
- It is hard to develop a Makefile which is truly cross-platform.
- Debugging and testing Make targets is not fun.

However, if you (and your team) know Make and are happy with it, do not change it.
Make is very powerful and a lot of stuff can be made a lot faster, if you know how to use it.

**taskflow** is intended to be simpler and easier to learn, while still being able to handle most use cases.

### Why not to use Mage

**taskflow** is intended to be an alternative to [Mage](https://github.com/magefile/mage).

[Mage](https://github.com/magefile/mage) is a framework/tool which magically discovers the [targets](https://magefile.org/targets/) from [magefiles](https://magefile.org/magefiles/).

**taskflow** takes a different approach as it is a regular Go library (package).
This results in following benefits:

- It is easy to debug. Like a regular Go application.
- Tasks and helpers are testable. See [exec_test.go](exec_test.go).
- Reusing tasks is easy and readable. Just create a function which registers common tasks. Mage does it in a [hacky way](https://magefile.org/importing/).
- API similar to [testing](https://golang.org/pkg/testing) so it is possible to use e.g. [testify](https://github.com/stretchr/testify) for asserting.

To sum up, **taskflow** is not magical. Write regular Go code. No build tags or special names for functions.

### Why not to use Task

While [Task](https://taskfile.dev/) is simpler and easier to use than [Make](https://www.gnu.org/software/make/) it still has some problems:

- Requires to learn Task's YAML sturcture and the [minimalistic, cross-platform interpreter](https://github.com/mvdan/sh#gosh) which it uses.
- Debugging and testing tasks is not fun.
- Harder to make some reusable tasks.
- Requires to "install" the tool. **taskflow** leverages `go run` and Go Modules so that you can be sure that everyone uses the same version of **taskflow**.

### Why not to use Bazel

[Bazel](https://bazel.build/) is a very sophisticated tool which is [created to efficiently handle complex and long-running build pipelines](https://en.wikipedia.org/wiki/Bazel_(software)#Rationale). It requires the build target inputs and outputs to be fully specified. 

**taskflow** is just a simple library that is mainly supposed to create a build pipeline consisting of commands like `go vet`, `go test`, `go build`. However, take notice that **taskflow** is a library. Nothing prevents you from, for example, using [Mage's target package](https://pkg.go.dev/github.com/magefile/mage/target) to make your build pipeline more efficient.

## Contributing

Simply create an issue or a pull request.
