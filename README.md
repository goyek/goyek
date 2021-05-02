# taskflow

> Create build pipelines in Go

[![Go Reference](https://pkg.go.dev/badge/github.com/pellared/taskflow.svg)](https://pkg.go.dev/github.com/pellared/taskflow)
[![Keep a Changelog](https://img.shields.io/badge/changelog-Keep%20a%20Changelog-%23E05735)](CHANGELOG.md)
[![GitHub Release](https://img.shields.io/github/v/release/pellared/taskflow)](https://github.com/pellared/taskflow/releases)
[![go.mod](https://img.shields.io/github/go-mod/go-version/pellared/taskflow)](go.mod)
[![LICENSE](https://img.shields.io/github/license/pellared/taskflow)](LICENSE)
[![Build Status](https://img.shields.io/github/workflow/status/pellared/taskflow/build)](https://github.com/pellared/taskflow/actions?query=workflow%3Abuild+branch%3Amain)
[![Go Report Card](https://goreportcard.com/badge/github.com/pellared/taskflow)](https://goreportcard.com/report/github.com/pellared/taskflow)
[![codecov](https://codecov.io/gh/pellared/taskflow/branch/main/graph/badge.svg)](https://codecov.io/gh/pellared/taskflow)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)
[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/pellared/taskflow)

> :warning: The `main` branch contains **breaking changes**.
> Here is the [**README for the latest release**](https://github.com/pellared/taskflow/blob/v0.2.0/README.md).

This package aims to simplify the creation of build pipelines in Go instead of using scripts or [Make](https://www.gnu.org/software/make/).

**taskflow** API is mainly inspired by the [testing](https://golang.org/pkg/testing), [http](https://golang.org/pkg/http) and [flag](https://golang.org/pkg/flag) packages.

Check [Go Build Pipeline Demo](https://github.com/pellared/go-build-pipeline-demo) to compare **taskflow** with [Make](https://www.gnu.org/software/make/) and [Mage](https://github.com/magefile/mage).

`Star` this repository if you find it valuable and worth maintaining.

Table of Contents:

- [taskflow](#taskflow)
	- [Example](#example)
	- [Features](#features)
		- [Task registration](#task-registration)
		- [Task command](#task-command)
		- [Task dependencies](#task-dependencies)
		- [Helpers for running programs](#helpers-for-running-programs)
		- [Verbose mode](#verbose-mode)
		- [Default task](#default-task)
		- [Parameters](#parameters)
		- [Task runner](#task-runner)
	- [Supported Go versions](#supported-go-versions)
	- [FAQ](#faq)
		- [Why not use Make?](#why-not-use-make)
		- [Why not use Mage?](#why-not-use-mage)
		- [Why not use Task?](#why-not-use-task)
		- [Why not use Bazel?](#why-not-use-bazel)
	- [Contributing](#contributing)

## Example

Create a file in your project `build/build.go`.

Copy and paste the content from below (this code can also be found in [examples/basic/main.go](examples/basic/main.go)).

```go
package main

import "github.com/pellared/taskflow"

func main() {
	flow := taskflow.New()

	taskflow.VerboseParam(flow)
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
Usage: [flag(s) | task(s)]...
Flags:
  -v    --verbose    Default: false    Verbose output: log all tasks as they are run. Also print all text from Log and Logf calls even if the task succeeds.
Tasks:
  all     build pipeline
  fmt     go fmt
  test    go test with race detector and code covarage
```

```shell
$ go run ./build all
ok     0.453s
```

```shell
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

Tired of writing `go run ./build` each time? Just add an alias to your shell. For example, add the line below to `~/.bash_aliases`:

```shell
alias gake='go run ./build'
```

Additionally, take a look at [taskflow-example](https://github.com/pellared/taskflow-example) and this repository's own build pipeline script - [build/build.go](build/build.go).

## Features

### Task registration

The registered tasks are required to have a non-empty name. For future compatibility, it is strongly suggested to use only the following characters:

- letters (`a-z` and `A-Z`)
- digits (`0-9`)
- underscore (`_`)
- hyphens (`-`)

Do not begin the task name with `-` sign as it is used for assigning parameters.

A task with a given name can be only registered once.

A task without description is not listed in taskflow's CLI usage.

### Task command

Task command is a function which is executed when a task is executed. It is not required to to set a command. Not having a command is very handy when registering "pipelines".

### Task dependencies

During task registration it is possible to add a dependency to an already registered task. When taskflow is processed, it makes sure that the dependency is executed before the current task is run. Take note that each task will be executed at most once.

### Helpers for running programs

Use [`func Exec(name string, args ...string) func(*TF)`](https://pkg.go.dev/github.com/pellared/taskflow#Exec) to create a task's command which only runs a single program.

Use [`func (tf *TF) Cmd(name string, args ...string) *exec.Cmd`](https://pkg.go.dev/github.com/pellared/taskflow#TF.Cmd) if within a task's command function when you want to execute more programs or you need more granular control.

### Verbose mode

Verbose mode streams all logs to the output. If it is disabled, only logs from failed task are send to the output.

By default, the verbose mode is turned on.
It can be controlled by registering a boolean parameter and assigning it to the field [`Taskflow.Verbose`](https://pkg.go.dev/github.com/pellared/taskflow#Taskflow.Verbose).

This is also provided via the convenience function [`taskflow.VerboseParam(flow *taskflow.Taskflow) BoolParam`](https://pkg.go.dev/github.com/pellared/taskflow#VerboseParam), which needs to be explicitly called.
The convenience function registers a boolean parameter `verbose`, short `v`, and assigns it to the [`Taskflow.Verbose`](https://pkg.go.dev/github.com/pellared/taskflow#Taskflow.Verbose) field.
This parameter defaults to `false`, so as soon as this parameter is registered, verbose mode has to be explicitly enabled via the CLI.
This then works similar to `go test -v`.

You can also reuse the registered parameter to check if verbose mode was set within a task's command.

### Default task

Default task can be assigned via [`Taskflow.DefaultTask`](https://pkg.go.dev/github.com/pellared/taskflow#Taskflow.DefaultTask) field.

When the default task is set, then it is run if no task is provided via CLI.

### Parameters

The parameters can be set via CLI using flag syntax. Flags must be registered in a long form, and have an optional short form using only one letter.

On the CLI, flags can be set in the following ways:

- `--longName simple` - for simple single-word values
- `--longName "value with blanks"`
- `--longName="value with blanks"`
- `--booleanParam` - setting boolean parameters implicitly to `true`
- `-l simple` - using short form

For example, `go run ./build -v --ci all` would run the `all` task with `verbose` and `ci` bool parameters both set to `true`.

Parameters must first be registered via [`func (f *Taskflow) RegisterValueParam(newValue func() Value, info ParameterInfo) ValueParam`](https://pkg.go.dev/github.com/pellared/taskflow#Taskflow.RegisterValueParam), or one of the provided methods like [`RegisterStringParam`](https://pkg.go.dev/github.com/pellared/taskflow#Taskflow.RegisterStringParam).

After registration, tasks need to specify which parameters they will read.
Do this by assigning the [`RegisteredParam`](https://pkg.go.dev/github.com/pellared/taskflow#RegisteredParam) instance from the registration result to the [`Task.Parameters`](https://pkg.go.dev/github.com/pellared/taskflow#Task.Parameters) field.
If a task tries to retrieve the value from an unregistered parameter, the task will fail.

When registration is done, the task's command can retrieve the parameter value using the `Get(*TF)` method from the registration result instance during the task's `Command` execution.

See [examples/parameters/main.go](examples/parameters/main.go) for a detailed example.

`Taskflow` will fail execution if there are unused parameters.

### Task runner

You can use [`type Runner`](https://pkg.go.dev/github.com/pellared/taskflow#Runner) to execute a single command.

It may be handy during development of a new task, when debugging some issue or if you want to have a test suite for reusable commands.

## Supported Go versions

Minimal supported Go version is 1.11.

## FAQ

### Why not use Make?

While [Make](https://www.gnu.org/software/make/) is currently the _de facto_ standard, it has some pitfalls:

- Requires to learn Make, which is not so easy.
- It is hard to develop a Makefile which is truly cross-platform.
- Debugging and testing Make targets is not fun.

However, if you (and your team) know Make and are happy with it, do not change it.

Make is very powerful and a lot of stuff can be made a lot faster, if you know how to use it.

**taskflow** is intended to be simpler and easier to learn, while still being able to handle most use cases.

### Why not use Mage?

**taskflow** is intended to be an alternative to [Mage](https://github.com/magefile/mage).

[Mage](https://github.com/magefile/mage) is a framework/tool which magically discovers the [targets](https://magefile.org/targets/) from [magefiles](https://magefile.org/magefiles/).

**taskflow** takes a different approach as it is a regular Go library (package).

This results in following benefits:

- It is easy to debug. Like a regular Go application.
- Tasks and helpers are testable. See [exec_test.go](exec_test.go).
- Reusing tasks is easy and readable. Just create a function which registers common tasks. Mage does it in a [hacky way](https://magefile.org/importing/).
- API similar to [testing](https://golang.org/pkg/testing) so it is possible to use e.g. [testify](https://github.com/stretchr/testify) for asserting.

To sum up, **taskflow** is not magical. Write regular Go code. No build tags or special names for functions.

### Why not use Task?

While [Task](https://taskfile.dev/) is simpler and easier to use than [Make](https://www.gnu.org/software/make/) it still has some problems:

- Requires to learn Task's YAML sturcture and the [minimalistic, cross-platform interpreter](https://github.com/mvdan/sh#gosh) which it uses.
- Debugging and testing tasks is not fun.
- Harder to make some reusable tasks.
- Requires to "install" the tool. **taskflow** leverages `go run` and Go Modules so that you can be sure that everyone uses the same version of **taskflow**.

### Why not use Bazel?

[Bazel](https://bazel.build/) is a very sophisticated tool which is [created to efficiently handle complex and long-running build pipelines](https://en.wikipedia.org/wiki/Bazel_(software)#Rationale). It requires the build target inputs and outputs to be fully specified.

**taskflow** is just a simple library that is mainly supposed to create a build pipeline consisting of commands like `go vet`, `go test`, `go build`. However, take notice that **taskflow** is a library. Nothing prevents you from, for example, using [Mage's target package](https://pkg.go.dev/github.com/magefile/mage/target) to make your build pipeline more efficient.

## Contributing

I am open to any feedback and contribution.

Use [Discussions](https://github.com/pellared/taskflow/discussions) or write to me: *Robert Pajak* @ [Gophers Slack](https://invite.slack.golangbridge.org/).

You can also create an issue, or a pull request.
