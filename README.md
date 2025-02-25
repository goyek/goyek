# goyek

> Task automation in Go

[![Go Reference](https://pkg.go.dev/badge/github.com/goyek/goyek.svg)](https://pkg.go.dev/github.com/goyek/goyek/v2)
[![Keep a Changelog](https://img.shields.io/badge/changelog-Keep%20a%20Changelog-%23E05735)](CHANGELOG.md)
[![GitHub Release](https://img.shields.io/github/v/release/goyek/goyek)](https://github.com/goyek/goyek/releases)
[![go.mod](https://img.shields.io/github/go-mod/go-version/goyek/goyek)](go.mod)

[![Build Status](https://img.shields.io/github/actions/workflow/status/goyek/goyek/build.yml?branch=main)](https://github.com/goyek/goyek/actions?query=workflow%3Abuild+branch%3Amain)
[![Go Report Card](https://goreportcard.com/badge/github.com/goyek/goyek)](https://goreportcard.com/report/github.com/goyek/goyek)
[![codecov](https://codecov.io/gh/goyek/goyek/branch/main/graph/badge.svg)](https://codecov.io/gh/goyek/goyek)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

Please ⭐ `Star` this repository if you find it valuable and worth maintaining.

[![Introduction](https://img.youtube.com/vi/e-xWEH-fqJ0/hqdefault.jpg)](https://www.youtube.com/watch?v=e-xWEH-fqJ0)

[Slides](https://docs.google.com/presentation/d/1xFAPXeMiOD-92xeIHkUD-SHmJZwc8mSIIgpjuJXEW3U/edit?usp=sharing).

---

Table of Contents:

- [Description](#description)
- [Quick start](#quick-start)
- [Repository template](#repository-template)
- [Examples](#examples)
- [Defining tasks](#defining-tasks)
- [Running programs](#running-programs)
- [Wrapper scripts](#wrapper-scripts)
- [Using middlewares](#using-middlewares)
- [Customizing](#customizing)
- [Alternatives](#alternatives)
  - [Make](#make)
  - [Mage](#mage)
  - [Task](#task)
  - [Bazel](#bazel)
- [Contributing](#contributing)
- [License](#license)

## Description

**goyek** (/ˈɡɔɪæk/ [🔊 listen](http://ipa-reader.xyz/?text=%CB%88%C9%A1%C9%94%C9%AA%C3%A6k))
is a task automation library.

This library is intended to be an alternative to
[Make](https://www.gnu.org/software/make/),
[Mage](https://github.com/magefile/mage),
[Task](https://taskfile.dev/).

The primary properties of goyek are: 

- A library, instead of an application,
  with API inspired by
  [`testing`](https://golang.org/pkg/testing),
  [`cobra`](https://github.com/spf13/cobra),
  [`flag`](https://golang.org/pkg/flag),
  [`http`](https://golang.org/pkg/http)
  packages.
- Cross-platform and shell independent.
- No binary installation needed.
- Easy to debug, like a regular Go code.
- The tasks are defined similarly to
  [`cobra`](https://github.com/spf13/cobra) commands.
- The task action looks like a Go test.
  [`A`](https://pkg.go.dev/github.com/goyek/goyek/v2#A)
  has similar methods to [`testing.T`](https://pkg.go.dev/testing#T).
- Reuse any Go code and library e.g. [`viper`](https://github.com/spf13/viper).
- Highly customizable.
- No third-party dependencies.
- Supplumental features in [`goyek/x`](https://github.com/goyek/x).

## Quick start

> Supplemental packages from [`github.com/goyek/x`](https://pkg.go.dev/github.com/goyek/x)
> are used for convinence.

The convention is to have the build automation
in the `/build` directory (or even Go module).

Put the following content in `/build/hello.go`:

```go
package main

import (
	"flag"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/x/cmd"
)

var msg = flag.String("msg", "greeting message", "Hello world!")

var hello = goyek.Define(goyek.Task{
	Name:  "hello",
	Usage: "demonstration",
	Action: func(a *goyek.A) {
		a.Log(*msg)
		cmd.Exec(a, "go version")
	},
})
```

Put the following content in `/build/main.go`:

```go
package main

import (
	"os"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/x/boot"
)

func main() {
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	goyek.SetDefault(hello)
	boot.Main()
}
```

Run:

```out
$ cd build

$ go mod tidy

$ go run . -h
Usage of build: [flags] [--] [tasks]
Tasks:
  hello  demonstration
Flags:
  -dry-run
        print all tasks without executing actions
  -long-run duration
        print when a task takes longer (default 1m0s)
  -msg string
        Hello world! (default "greeting message")
  -no-color
        disable colorizing output
  -no-deps
        do not process dependencies
  -skip comma-separated tasks
        skip processing the comma-separated tasks
  -v    print all tasks as they are run

$ go run . -v
===== TASK  hello
      hello.go:16: greeting message
      hello.go:17: Exec: go version
go version go1.19.3 windows/amd64
----- PASS: hello (0.12s)
ok      0.123s
```

## Repository template

You can use [goyek/template](https://github.com/goyek/template)
to create a new repository.

For an existing repository you can copy most of its files.

## Examples

- [goyek/template](https://github.com/goyek/template) -
  Go application repository template
- [build](build) -
  dogfooding
- [splunk-otel-go](https://github.com/signalfx/splunk-otel-go/tree/main/build) -
  multi-module repository
- [goyek/demo](https://github.com/goyek/demo)
  and [goyek/workflow](https://github.com/goyek/workflow) -
  demonstratation of the reusability potential

## Defining tasks

Use [`Define`](https://pkg.go.dev/github.com/goyek/goyek/v2#Define)
to register a a task.

You can add dependencies to already defineded tasks using
[`Task.Deps`](https://pkg.go.dev/github.com/goyek/goyek/v2#Task.Deps).
The dependencies are running in sequential order.
Each task runs at most once.

The [`Task.Action`](https://pkg.go.dev/github.com/goyek/goyek/v2#Task.Action)
is a function which executes when a task is running.
A task can have only dependencies and no action to act as a pipeline.

The [`Task.Parallel`](https://pkg.go.dev/github.com/goyek/goyek/v2#Task.Parallel)
can be set to allow a task to be run in parallel with other parallel tasks.

A default task can be assigned using [`SetDefault`](https://pkg.go.dev/github.com/goyek/goyek/v2#SetDefault).

## Running programs

You can use the [`cmd.Exec`](https://pkg.go.dev/github.com/goyek/x/cmd#Exec)
convenient function from [goyek/x](https://github.com/goyek/x)
that should cover most use cases.

Alternatively, you may prefer create your own helpers
like `Exec` in [build/exec.go](build/exec.go).

[#60](https://github.com/goyek/goyek/issues/60) and [#307](https://github.com/goyek/goyek/issues/307)
explain why this feature is not out-of-the-box.

## Wrapper scripts

Instead of executing `go run .` in `build` directory,
you may prefer using the wrapper scripts,
which you can invoke from any location.

Bash: [`goyek.sh`](goyek.sh).
PowerShell: [`goyek.ps1`](goyek.ps1).

## Customizing

Call the [`Use`](https://pkg.go.dev/github.com/goyek/goyek/v2#Use) function
or [`UseExecutor`](https://pkg.go.dev/github.com/goyek/goyek/v2#UseExecutor)
to setup a task runner or flow executor interceptor (middleware).

You can use a middleware, for example to:
generate a task execution report,
add retry logic,
export task execution telemetry, etc.

You can use some reusable middlewares from the
[`middleware`](https://pkg.go.dev/github.com/goyek/goyek/v2/middleware)
package. [`ReportStatus`](https://pkg.go.dev/github.com/goyek/goyek/v2/middleware#ReportStatus)
is the most commonly used middleware.

Notice that the [`boot.Main`](https://pkg.go.dev/github.com/goyek/x/boot#Main)
convenient function from [goyek/x](https://github.com/goyek/x)
sets the most commonly used middlewares and defines flags to configure them.

You can customize the default output by using:

- [`SetOutput`](https://pkg.go.dev/github.com/goyek/goyek/v2#SetOutput)
- [`SetLogger`](https://pkg.go.dev/github.com/goyek/goyek/v2#SetLogger)
- [`SetUsage`](https://pkg.go.dev/github.com/goyek/goyek/v2#SetUsage)
- [`Execute`](https://pkg.go.dev/github.com/goyek/goyek/v2#Execute)
  (instead of [`Main`](https://pkg.go.dev/github.com/goyek/goyek/v2#Main))

You can also study how [github.com/goyek/x](https://github.com/goyek/x)
is customizing the default behavior.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) if you want to help us.

## License

**goyek** is licensed under the terms of the [MIT license](LICENSE).

Note: **goyek** was named **taskflow** before v0.3.0.
