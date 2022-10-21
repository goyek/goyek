# goyek

> Create build pipelines in Go

[![Go Reference](https://pkg.go.dev/badge/github.com/goyek/goyek.svg)](https://pkg.go.dev/github.com/goyek/goyek/v2)
[![Keep a Changelog](https://img.shields.io/badge/changelog-Keep%20a%20Changelog-%23E05735)](CHANGELOG.md)
[![GitHub Release](https://img.shields.io/github/v/release/goyek/goyek)](https://github.com/goyek/goyek/releases)
[![go.mod](https://img.shields.io/github/go-mod/go-version/goyek/goyek)](go.mod)
[![LICENSE](https://img.shields.io/github/license/goyek/goyek)](LICENSE)

[![Build Status](https://img.shields.io/github/workflow/status/goyek/goyek/build)](https://github.com/goyek/goyek/actions?query=workflow%3Abuild+branch%3Amain)
[![Go Report Card](https://goreportcard.com/badge/github.com/goyek/goyek)](https://goreportcard.com/report/github.com/goyek/goyek)
[![codecov](https://codecov.io/gh/goyek/goyek/branch/main/graph/badge.svg)](https://codecov.io/gh/goyek/goyek)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

Please ⭐ `Star` this repository if you find it valuable and worth maintaining.

---

This README refers to the `v2` RC version.
See the documentation for the latest stable release on [pkg.go.dev](https://pkg.go.dev/github.com/goyek/goyek).

---

Table of Contents:

- [Description](#description)
- [Quick start](#quick-start)
- [Examples](#examples)
- [Repository template](#repository-template)
- [Features](#features)
  - [Wrapper scripts](#wrapper-scripts)
  - [Defining tasks](#defining-tasks)
  - [Running programs](#running-programs)
  - [Parameters](#parameters)
  - [Middlewares](#middlewares)
  - [Custom printing](#custom-printing)
- [Supported Go versions](#supported-go-versions)
- [Alternatives](#alternatives)
  - [Make](#make)
  - [Mage](#mage)
  - [Task](#task)
  - [Bazel](#bazel)
- [Presentations](#presentations)
- [Contributing](#contributing)
- [License](#license)

## Description

**goyek** (/ˈɡɔɪæk/ [🔊 listen](http://ipa-reader.xyz/?text=%CB%88%C9%A1%C9%94%C9%AA%C3%A6k))
is used to create build pipelines in Go.
As opposed to many other tools, it is just a Go library.

The API is mainly inspired by
[`testing`](https://golang.org/pkg/testing),
[`flag`](https://golang.org/pkg/flag),
[`http`](https://golang.org/pkg/http),
[`cobra`](https://github.com/spf13/cobra)
packages.

Here are some good parts:

- It is cross-platform and shell independent.
- No binary installation is needed. Simply add it to `go.mod` like any other Go module.
  - You can be sure that everyone uses the same version of **goyek**.
- It has low learning curve, thanks to the minimal API surface,
  documentation, and examples.
- The task's action look like a unit test.
  It is even possible to use [`testify`](https://github.com/stretchr/testify)
  or [`is`](https://github.com/matryer/is) for asserting.
- It is easy to debug, like a regular Go application.
- You can reuse code like in any Go application.
  It may be helpful to use packages like:
  - [`bitfield/script`](https://github.com/bitfield/script)
  - [`magefile/mage/target`](https://pkg.go.dev/github.com/magefile/mage/target)
  - [`mattn/go-shellwords`](https://pkg.go.dev/github.com/mattn/go-shellwords)
  - [`rjeczalik/notify`](https://github.com/rjeczalik/notify)
  - [`spf13/viper`](https://github.com/spf13/viper)
- **goyek** does not use any third-party dependency other than the Go standard library.

## Quick start

Put the following content in `/build/hello.go`:

```go
package main

import (
	"flag"

	"github.com/goyek/goyek/v2"
)

var msg = flag.String("msg", "greeting message", "Hello world!")

var hello = flow.Define(goyek.Task{
	Name:  "hello",
	Usage: "demonstration",
	Action: func(tf *goyek.TF) {
		tf.Log(*msg)
	},
})
```

Put the following content in `/build/main.go`:

```go
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/goyek/v2/middleware"
)

func main() {
	goyek.SetDefault(hello)

	flag.CommandLine.SetOutput(goyek.Output())
	flag.Usage = usage
	flag.Parse()

	goyek.Use(middleware.Reporter)
	goyek.SetUsage(usage)
	goyek.Main(flag.Args())
}

func usage() {
	fmt.Println("Usage of build: [flags] [--] [tasks]")
	goyek.Print()
	fmt.Println("Flags:")
	flag.PrintDefaults()
}
```

Run:

```shell
$ go mod tidy

$ go run ./build -h
Usage of build: [flags] [--] [tasks]
Tasks:
  hello  demonstration
Flags:
  -msg string
        Hello world! (default "greeting message")

$ go run ./build
===== TASK  hello
      hello.go:15: greeting message
----- PASS: hello (0.00s)
ok      0.001s
```

## Examples

- [example_test.go](example_test.go) -
  demonstrative examples
- [build](build) -
  this repository's own build pipeline (dogfooding)
- [splunk-otel-go](https://github.com/signalfx/splunk-otel-go/tree/main/build) -
  usage in a multi-module monorepo

## Repository template

You can use [goyek/template](https://github.com/goyek/template)
to create a new repository.

For an existing repository you can copy most of its files.

## Features

[example_test.go](example_test.go) demonstrates the key features
of **goyek**.

### Wrapper scripts

The convention is to have the build pipeline
in the `/build` directory (or even module).

Instead of executing `go run ./build`,
you can use the wrapper scripts,
which can be invoked from any location.

- [`goyek.sh`](goyek.sh)
- [`goyek.ps1`](goyek.ps1)

Simply add them to your repository's root directory
and make sure to add the `+x` permission:

```sh
curl -sSfL https://raw.githubusercontent.com/goyek/goyek/v2.0.0-rc.3/goyek.sh -O
curl -sSfL https://raw.githubusercontent.com/goyek/goyek/v2.0.0-rc.3/goyek.ps1 -O
git add --chmod=+x goyek.sh goyek.ps1
```

### Defining tasks

Use [`Flow.Define`](https://pkg.go.dev/github.com/goyek/goyek/v2#Flow.Define)
to register a a task.
A task with a given (non-empty) name can be defined only once.

During task registration it is possible to add a dependency
to another defined task using [`Task.Deps`](https://pkg.go.dev/github.com/goyek/goyek/v2#Task.Deps).
The dependencies are run in sequential order.
Therefore, if possible, you should order them from the fastest
to the slowest.
Each task will be run at most once.

The [`Task.Action`](https://pkg.go.dev/github.com/goyek/goyek/v2#Task.Action)
is a function which is executed when a task is run.
It is not required to set a action.
Not having a action is very handy when registering "pipelines".

Default task can be assigned using [`Flow.SetDefault`](https://pkg.go.dev/github.com/goyek/goyek/v2#Flow.SetDefault).
When the default task is set, then it is run if no task is provided.

### Running programs

Use [`TF.Cmd`](https://pkg.go.dev/github.com/goyek/goyek/v2#TF.Cmd)
to run a program inside a task's action.

You can also use it create your own helpers like `Exec` in [build/exec.go](build/exec.go).

[Here](https://github.com/goyek/goyek/issues/60) is the explanation
why argument splitting is not included out-of-the-box.

### Parameters

As of `v2` the parameters support has been removed
in order to improve customization.

With the new API it is easy to integrate **goyek** with any of these packages:

- [`flag`](https://pkg.go.dev/flag)
- [`pflag`](https://github.com/spf13/pflag)
- [`viper`](https://github.com/spf13/viper)

### Middlewares

**goyek** supports the addition of task run middlewares,
which are executed in the order they are added.

You can use a middleware for example to:
generate a task execution report,
skip some tasks,
add retry logic,
export build execution telemetry, etc.

You can use some reusalbe middlewares from the
[`middleware`](https://pkg.go.dev/github.com/goyek/goyek/v2/middleware)
package. [`Reporter`](https://pkg.go.dev/github.com/goyek/goyek/v2/middleware#Reporter)
is the most commonly used.

### Custom printing

You can customize the default output by using:

- [`Flow.SetOutput`](https://pkg.go.dev/github.com/goyek/goyek/v2#Flow.SetOutput)
- [`Flow.SetLogger`](https://pkg.go.dev/github.com/goyek/goyek/v2#Flow.SetLogger)
- [`Flow.SetUsage`](https://pkg.go.dev/github.com/goyek/goyek/v2#Flow.SetUsage)
- [`Flow.Execute`](https://pkg.go.dev/github.com/goyek/goyek/v2#Flow.Execute)
  instead of [`Flow.Main`](https://pkg.go.dev/github.com/goyek/goyek/v2#Flow.Main)

## Supported Go versions

Minimal supported Go version is 1.11.

## Alternatives

### Make

While [Make (Makefile)](https://www.gnu.org/software/make/) is currently
the _de facto_ standard, it has some pitfalls:

- Requires to learn Make, which is not so easy.
- It is hard to develop a Makefile which is truly cross-platform.
- Debugging and testing Make targets is not fun.

However, if you know Make and are happy with it, do not change it.
Make is very powerful and a lot of stuff can be made faster,
if you know how to use it.

**goyek** is intended to be simpler and easier to learn,
more portable, while still being able to handle most use cases.

### Mage

[Mage](https://github.com/magefile/mage) is a framework/tool which magically discovers
the [targets](https://magefile.org/targets/) from [magefiles](https://magefile.org/magefiles/),
which results in some drawbacks:

- Requires using [build tags](https://magefile.org/magefiles/).
- Reusing tasks is [hacky](https://magefile.org/importing/).
- Requires installation or using [zero install option](https://magefile.org/zeroinstall/)
  which is slow.
- Debugging would be extermly complex.
- Magical by design (of course one may like it).

**goyek** is a non-magical alternative for [Mage](https://github.com/magefile/mage).
Write regular Go code. No build tags, special names for functions, tricky imports.

### Task

While [Task](https://taskfile.dev/) is simpler and easier to use
than [Make](https://www.gnu.org/software/make/),
but it still has similar problems:

- Requires to learn Task's YAML structure and
  the [minimalistic, cross-platform interpreter](https://github.com/mvdan/sh#gosh)
  which it uses.
- Debugging and testing tasks is not fun.
- Hard to make reusable tasks.
- Requires to "install" the tool.

### Bazel

[Bazel](https://bazel.build/) is a very sophisticated tool which is
[created to efficiently handle complex and long-running build pipelines](https://en.wikipedia.org/wiki/Bazel_(software)#Rationale).
It requires the build target inputs and outputs to be fully specified.

**goyek** is just a simple Go library that is mainly supposed to create a build pipeline
consisting of commands like `go vet`, `go test`, `go build`.
However, nothing prevents you from, for example,
using the [Mage's target package](https://pkg.go.dev/github.com/magefile/mage/target)
to make your build pipeline more efficient.

## Presentations

| Date       | Presentation                                                                            | Description                                                                                                                        |
|------------|-----------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------|
| 2021-05-05 | [goyek - Create build pipelines in Go](https://github.com/pellared/goyek-demo)          | **goyek** v0.3.0 demo                                                                                                              |
| 2020-12-14 | [Build pipeline for a Go project](https://github.com/pellared/go-build-pipeline-demo)   | build pipeline using [Make](https://www.gnu.org/software/make/), [Mage](https://github.com/magefile/mage), and **taskflow** v0.1.0 |

Note: **goyek** was named **taskflow** before v0.3.0.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) if you want to help us.

## License

**goyek** is licensed under the terms of the [MIT license](LICENSE).
