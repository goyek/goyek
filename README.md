# goyek

> Create build pipelines in Go

[![Go Reference](https://pkg.go.dev/badge/github.com/goyek/goyek.svg)](https://pkg.go.dev/github.com/goyek/goyek)
[![Keep a Changelog](https://img.shields.io/badge/changelog-Keep%20a%20Changelog-%23E05735)](CHANGELOG.md)
[![GitHub Release](https://img.shields.io/github/v/release/goyek/goyek)](https://github.com/goyek/goyek/releases)
[![go.mod](https://img.shields.io/github/go-mod/go-version/goyek/goyek)](go.mod)
[![LICENSE](https://img.shields.io/github/license/goyek/goyek)](LICENSE)

[![Build Status](https://img.shields.io/github/workflow/status/goyek/goyek/build)](https://github.com/goyek/goyek/actions?query=workflow%3Abuild+branch%3Amain)
[![Go Report Card](https://goreportcard.com/badge/github.com/goyek/goyek)](https://goreportcard.com/report/github.com/goyek/goyek)
[![codecov](https://codecov.io/gh/goyek/goyek/branch/main/graph/badge.svg)](https://codecov.io/gh/goyek/goyek)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

Table of Contents:

- [goyek](#goyek)
  - [Description](#description)
  - [Quick start](#quick-start)
  - [Examples](#examples)
  - [Features](#features)
    - [Task registration](#task-registration)
    - [Task command](#task-command)
    - [Task dependencies](#task-dependencies)
    - [Helpers for running programs](#helpers-for-running-programs)
    - [Verbose mode](#verbose-mode)
    - [Default task](#default-task)
    - [Parameters](#parameters)
    - [Supported Go versions](#supported-go-versions)
  - [Alternatives](#alternatives)
    - [Make](#make)
    - [Mage](#mage)
    - [Task](#task)
    - [Bazel](#bazel)
  - [Presentations](#presentations)
  - [Contributing](#contributing)

## Description

**goyek** is used to create build pipelines in Go.
As opposed to many other tools, it is just a Go library.

Here are some good parts:

- No binary installation is needed. Simply add it to `go.mod` like any other Go module.
- Low learning curve, thanks to the minimal API surface, documentation, and examples.
- Reuse code like in any Go application. It may be helpful to use packages like:
  - [`github.com/bitfield/script`](https://github.com/bitfield/script)
  - [`github.com/magefile/mage/target`](https://pkg.go.dev/github.com/magefile/mage/target)
- Easy to debug. Like a regular Go application.
- The API is based on [testing](https://golang.org/pkg/testing).
  Task implementations look like unit tests. It is even possible to use [testify](https://github.com/stretchr/testify) for asserting.
- Tasks and helpers can be tested. See [exec_test.go](exec_test.go).

**goyek** API is mainly inspired by the [testing](https://golang.org/pkg/testing), [http](https://golang.org/pkg/http) and [flag](https://golang.org/pkg/flag) packages.

Please `Star` this repository if you find it valuable and worth maintaining.

## Quick start

Create a file in your project `build/build.go`. Copy and paste the content from below.

```go
package main

import "github.com/goyek/goyek"

func main() {
	flow := &goyek.Taskflow{}

	hello := flow.Register(taskHello())
	fmt := flow.Register(taskFmt())

	flow.Register(goyek.Task{
		Name:  "all",
		Usage: "build pipeline",
		Deps: goyek.Deps{
			hello,
			fmt,
		},
	})

	flow.Main()
}

func taskHello() goyek.Task {
	return goyek.Task{
		Name:  "hello",
		Usage: "demonstration",
		Command: func(tf *goyek.TF) {
			tf.Log("Hello world!")
		},
	}
}

func taskFmt() goyek.Task {
	return goyek.Task{
		Name:    "fmt",
		Usage:   "go fmt",
		Command: goyek.Exec("go", "fmt", "./..."),
	}
}
```

Sample usage:

```shell
$ go run ./build -h
Usage: [flag(s) | task(s)]...
Flags:
  -v    Default: false    Verbose output: log all tasks as they are run. Also print all text from Log and Logf calls even if the task succeeds.
Tasks:
  all      build pipeline
  fmt      go fmt
  hello    demonstration
```

```shell
$ go run ./build all
ok     0.167s
```

```shell
$ go run ./build all -v
===== TASK  hello
Hello world!
----- PASS: hello (0.00s)
===== TASK  fmt
Cmd: go fmt ./...
----- PASS: fmt (0.18s)
ok      0.183s
```

Tired of writing `go run ./build` each time? Just add an alias to your shell.
For example, add the line below to `~/.bash_aliases`:

```shell
alias goyek='go run ./build'
```

## Examples

- [examples](examples)
- [build/build.go](build/build.go) - this repository's own build pipeline

## Features

### Task registration

The registered tasks are required to have a non-empty name, matching
the regular expression `^[a-zA-Z0-9_][a-zA-Z0-9_-]*$`, available as
[`TaskNamePattern`](https://pkg.go.dev/github.com/goyek/goyek#TaskNamePattern).
This means the following are acceptable:

- letters (`a-z` and `A-Z`)
- digits (`0-9`)
- underscore (`_`)
- hyphens (`-`) - except at the beginning

A task with a given name can be only registered once.

A task without description is not listed in CLI usage.

### Task command

Task command is a function which is executed when a task is executed.
It is not required to to set a command.
Not having a command is very handy when registering "pipelines".

### Task dependencies

During task registration it is possible to add a dependency to an already registered task.
When taskflow is processed, it makes sure that the dependency is executed before the current task is run.
Take note that each task will be executed at most once.

### Helpers for running programs

Use [`func Exec(name string, args ...string) func(*TF)`](https://pkg.go.dev/github.com/goyek/goyek#Exec)
to create a task's command which only runs a single program.

Use [`func (tf *TF) Cmd(name string, args ...string) *exec.Cmd`](https://pkg.go.dev/github.com/goyek/goyek#TF.Cmd)
if within a task's command function when you want to execute more programs or you need more granular control.

### Verbose mode

Enable verbose output using the `-v` CLI flag.
It works similar to `go test -v`. Verbose mode streams all logs to the output.
If it is disabled, only logs from failed task are send to the output.

Use [`func (f *Taskflow) VerboseParam() BoolParam`](https://pkg.go.dev/github.com/goyek/goyek#Taskflow.VerboseParam)
if you need to check if verbose mode was set within a task's command.

### Default task

Default task can be assigned via the [`Taskflow.DefaultTask`](https://pkg.go.dev/github.com/goyek/goyek#Taskflow.DefaultTask) field.

When the default task is set, then it is run if no task is provided via CLI.

### Parameters

The parameters can be set via CLI using the flag syntax.

On the CLI, flags can be set in the following ways:

- `-param simple` - for simple single-word values
- `-param "value with blanks"`
- `-param="value with blanks"`
- `-param` - setting boolean parameters implicitly to `true`

For example, `go run ./build test -v -pkg ./...` would run the `test` task
with `v` bool parameter (verbose mode) set to `true`,
and `pkg` string parameter set to `"./..."`.

Parameters must first be registered via [`func (f *Taskflow) RegisterValueParam(newValue func() ParamValue, info ParamInfo) ValueParam`](https://pkg.go.dev/github.com/goyek/goyek#Taskflow.RegisterValueParam), or one of the provided methods like [`RegisterStringParam`](https://pkg.go.dev/github.com/goyek/goyek#Taskflow.RegisterStringParam).

The registered parameters are required to have a non-empty name, matching
the regular expression `^[a-zA-Z0-9][a-zA-Z0-9_-]*$`, available as
[`ParamNamePattern`](https://pkg.go.dev/github.com/goyek/goyek#ParamNamePattern).
This means the following are acceptable:

- letters (`a-z` and `A-Z`)
- digits (`0-9`)
- underscore (`_`) - except at the beginning
- hyphens (`-`) - except at the beginning

After registration, tasks need to specify which parameters they will read.
Do this by assigning the [`RegisteredParam`](https://pkg.go.dev/github.com/goyek/goyek#RegisteredParam) instance from the registration result to the [`Task.Params`](https://pkg.go.dev/github.com/goyek/goyek#Task.Params) field.
If a task tries to retrieve the value from an unregistered parameter, the task will fail.

When registration is done, the task's command can retrieve the parameter value using the `Get(*TF)` method from the registration result instance during the task's `Command` execution.

See [examples/parameters/main.go](examples/parameters/main.go) for a detailed example.

`Taskflow` will fail execution if there are unused parameters.

### Supported Go versions

Minimal supported Go version is 1.11.

## Alternatives

### Make

While [Make](https://www.gnu.org/software/make/) is currently the _de facto_ standard, it has some pitfalls:

- Requires to learn Make, which is not so easy.
- It is hard to develop a Makefile which is truly cross-platform.
- Debugging and testing Make targets is not fun.

However, if you (and your team) know Make and are happy with it, do not change it.

Make is very powerful and a lot of stuff can be made a lot faster, if you know how to use it.

**goyek** is intended to be simpler and easier to learn, while still being able to handle most use cases.

### Mage

**goyek** is intended to be an alternative implementation of [Mage](https://github.com/magefile/mage).

[Mage](https://github.com/magefile/mage) is a framework/tool which magically discovers the [targets](https://magefile.org/targets/) from [magefiles](https://magefile.org/magefiles/).

**goyek** takes a different approach as it is a regular Go library.

This results in following benefits:

- It is easier to debug.
- Reusing tasks is easier and more readable. Just create a function which registers common tasks.
  Mage does it in a [hacky way](https://magefile.org/importing/).
- Error reporting API is based on [testing](https://golang.org/pkg/testing) package
  so it is possible to use e.g. [testify](https://github.com/stretchr/testify) for asserting.

To sum up, **goyek** is not magical. Write regular Go code. No build tags or special names for functions.

### Task

While [Task](https://taskfile.dev/) is simpler and easier to use than [Make](https://www.gnu.org/software/make/) it still has some problems:

- Requires to learn Task's YAML structure and the [minimalistic, cross-platform interpreter](https://github.com/mvdan/sh#gosh) which it uses.
- Debugging and testing tasks is not fun.
- Harder to make some reusable tasks.
- Requires to "install" the tool. **goyek** leverages `go run` and Go Modules so that you can be sure that everyone uses the same version of **goyek**.

### Bazel

[Bazel](https://bazel.build/) is a very sophisticated tool which is [created to efficiently handle complex and long-running build pipelines](https://en.wikipedia.org/wiki/Bazel_(software)#Rationale). It requires the build target inputs and outputs to be fully specified.

**goyek** is just a simple library that is mainly supposed to create a build pipeline consisting of commands like `go vet`, `go test`, `go build`. However, take notice that **goyek** is a library. Nothing prevents you from, for example, using [Mage's target package](https://pkg.go.dev/github.com/magefile/mage/target) to make your build pipeline more efficient.

## Presentations

- [taskflow - Create build pipelines in Go](https://github.com/pellared/taskflow-example) - **taskflow** (now **goyek**) `v0.1.1` demo
- [Build pipeline for a Go project](https://github.com/pellared/go-build-pipeline-demo) - build pipeline for a naive Go project using [Make](https://www.gnu.org/software/make/), [Mage](https://github.com/magefile/mage), and **taskflow** (now **goyek**) `v0.1.0`

## Contributing

We are open to any feedback and contribution.

You can find us on [Gophers Slack](https://invite.slack.golangbridge.org/) in [`#goyek` channel](https://gophers.slack.com/archives/C020UNUK7LL).

Feel free to create an issue, or a pull request.

You may use [Gitpod](https://gitpod.io/#https://github.com/goyek/goyek) to experiment with the codebase.
