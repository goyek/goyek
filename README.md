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

> :warning: The `main` branch contains **new features**.
> Here is the [**README for the latest release**](https://github.com/goyek/goyek/blob/v0.3.0/README.md).

Table of Contents:

- [goyek](#goyek)
  - [Description](#description)
  - [Quick start](#quick-start)
  - [Examples](#examples)
  - [Wrapper scripts](#wrapper-scripts)
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

**goyek** (/ˈɡɔɪæk/ [🔊 listen](http://ipa-reader.xyz/?text=%CB%88%C9%A1%C9%94%C9%AA%C3%A6k))
is used to create build pipelines in Go.
As opposed to many other tools, it is just a Go library.

Here are some good parts:

- No binary installation is needed. Simply add it to `go.mod` like any other Go module.
  - You can be sure that everyone uses the same version of **goyek**.
- It has low learning curve, thanks to the minimal API surface, documentation, and examples.
- One can reuse code like in any Go application. It may be helpful to use packages like:
  - [`github.com/bitfield/script`](https://pkg.go.dev/github.com/bitfield/script)
  - [`github.com/rjeczalik/notify`](https://pkg.go.dev/github.com/rjeczalik/notify)
  - [`github.com/magefile/mage/target`](https://pkg.go.dev/github.com/magefile/mage/target)
- It is easy to debug, like a regular Go application.
- The API is based on [testing](https://golang.org/pkg/testing).
  The task's command look like a unit test.
  It is even possible to use [testify](https://github.com/stretchr/testify) for asserting.
- Tasks and helpers can be easily tested. See [exec_test.go](exec_test.go).

**goyek** API is mainly inspired by the [http](https://golang.org/pkg/http),
[testing](https://golang.org/pkg/testing), and [flag](https://golang.org/pkg/flag) packages.

Please `Star` this repository if you find it valuable and worth maintaining.

## Quick start

Copy and paste the following code into [`build/build.go`](examples/basic/main.go):

```go
package main

import (
	"fmt"
	"os"

	"github.com/goyek/goyek"
)

func main() {
	if err := os.Chdir(".."); err != nil {
		fmt.Println(err)
		os.Exit(goyek.CodeInvalidArgs)
	}

	flow := &goyek.Taskflow{}

	flow.Register(goyek.Task{
		Name:  "hello",
		Usage: "demonstration",
		Command: func(tf *goyek.TF) {
			tf.Log("Hello world!")
		},
	})

	flow.Main()
}
```

Run:

```shell
cd build
go mod tidy
```

Sample usage:

```shell
$ go run . -h
Usage: [flag(s) | task(s)]...
Flags:
  -v     Default: false    Verbose: log all tasks as they are run.
  -wd    Default: .        Working directory: set the working directory.
Tasks:
  hello    demonstration
```

```shell
$ go run . hello
ok     0.000s
```

```shell
$ go run . all -v
===== TASK  hello
Hello world!
----- PASS: hello (0.00s)
ok      0.001s
```

## Examples

- [examples](examples)
- [build/build.go](build/build.go) - this repository's own build pipeline
- [pellared/fluentassert](https://github.com/pellared/fluentassert) - a library using **goyek** without polluting it's root `go.mod`

## Wrapper scripts

Instead of going into `build` directory and executing `go run .`,
we highly recommend using wrapper scripts.

Simply add them to your repository's root directory:

- [`goyek.sh`](goyek.sh) - make sure to add `+x` permission (`git update-index --chmod=+x goyek.sh`):

```bash
#!/bin/bash
set -euo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
cd "$DIR/build"
go run . -wd=".." $@
```

- [`goyek.ps1`](goyek.ps1):

```powershell
$ErrorActionPreference = "Stop"

Push-Location "$PSScriptRoot\build"
& go run . -wd=".." $args
Pop-Location
exit $global:LASTEXITCODE
```

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

Use [`func (tf *TF) Cmd(name string, args ...string) *exec.Cmd`](https://pkg.go.dev/github.com/goyek/goyek#TF.Cmd)
to run a program inside a task's command.

You can use it create your own helpers, for example:

```go
import (
	"fmt"
	"os/exec"

	"github.com/goyek/goyek"
	"github.com/mattn/go-shellwords"
)

func Cmd(tf *goyek.TF, cmdLine string) *exec.Cmd {
	args, err := shellwords.Parse(cmdLine)
	if err != nil {
		tf.Fatalf("parse command line: %v", err)
	}
	return tf.Cmd(args[0], args[1:]...)
}

func Exec(cmdLine string) func(tf *goyek.TF) {
	args, err := shellwords.Parse(cmdLine)
	if err != nil {
		panic(fmt.Sprintf("parse command line: %v", err))
	}
	return func(tf *goyek.TF) {
		if err := tf.Cmd(args[0], args[1:]...).Run(); err != nil {
			tf.Fatal(err)
		}
	}
}
```

[Here](https://github.com/goyek/goyek/issues/60) is the explantion why argument splitting is not included out-of-the-box.

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

For example, `./goyek.sh test -v -pkg ./...` would run the `test` task
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

However, if you know Make and are happy with it, do not change it.
Make is very powerful and a lot of stuff can be made faster, if you know how to use it.

**goyek** is intended to be simpler and easier to learn, while still being able to handle most use cases.

### Mage

[Mage](https://github.com/magefile/mage) is a framework/tool which magically discovers
the [targets](https://magefile.org/targets/) from [magefiles](https://magefile.org/magefiles/),
which results in some drawbacks:

- Requires using [build tags](https://magefile.org/magefiles/).
- Reusing tasks is [hacky](https://magefile.org/importing/).
- Requires installation or using [zero install option](https://magefile.org/zeroinstall/) which is slow.
- Debugging would be extermly complex.
- Magical by design (of course one may like it).

**goyek** is intended to be a non-magical alternative for [Mage](https://github.com/magefile/mage).
Write regular Go code. No build tags, special names for functions, tricky imports.

### Task

While [Task](https://taskfile.dev/) is simpler and easier to use
than [Make](https://www.gnu.org/software/make/) it still has similar problems:

- Requires to learn Task's YAML structure and
  the [minimalistic, cross-platform interpreter](https://github.com/mvdan/sh#gosh) which it uses.
- Debugging and testing tasks is not fun.
- Hard to make reusable tasks.
- Requires to "install" the tool.

### Bazel

[Bazel](https://bazel.build/) is a very sophisticated tool which is
[created to efficiently handle complex and long-running build pipelines](https://en.wikipedia.org/wiki/Bazel_(software)#Rationale).
It requires the build target inputs and outputs to be fully specified.

**goyek** is just a simple library that is mainly supposed to create a build pipeline
consisting of commands like `go vet`, `go test`, `go build`.
However, take notice that **goyek** is a library. Nothing prevents you from,
for example, using [Mage's target package](https://pkg.go.dev/github.com/magefile/mage/target)
to make your build pipeline more efficient.

## Presentations

| Date       | Presentation                                                                            | Description                |
| ---------- | --------------------------------------------------------------------------------------- | -------------------------- |
| 2021-05-05 | [goyek - Create build pipelines in Go](https://github.com/pellared/goyek-demo)          | **goyek** v0.3.0 demo      |
| 2021-03-10 | [taskflow - Create build pipelines in Go](https://github.com/pellared/taskflow-example) | **taskflow** v0.1.1 demo   |
| 2020-12-14 | [Build pipeline for a Go project](https://github.com/pellared/go-build-pipeline-demo)   | build pipeline using [Make](https://www.gnu.org/software/make/), [Mage](https://github.com/magefile/mage), and **taskflow** v0.1.0 |

Note: **goyek** was named **taskflow** before v0.3.0.

## Contributing

We are open to any feedback and contribution.

You can find us on [Gophers Slack](https://invite.slack.golangbridge.org/) in [`#goyek` channel](https://gophers.slack.com/archives/C020UNUK7LL).

Feel free to create an issue, or a pull request.

You may use [Gitpod](https://gitpod.io/#https://github.com/goyek/goyek) to experiment with the codebase.
