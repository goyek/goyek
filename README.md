# goyek

> Build automation in Go

[![Go Reference](https://pkg.go.dev/badge/github.com/goyek/goyek.svg)](https://pkg.go.dev/github.com/goyek/goyek/v2)
[![Keep a Changelog](https://img.shields.io/badge/changelog-Keep%20a%20Changelog-%23E05735)](CHANGELOG.md)
[![GitHub Release](https://img.shields.io/github/v/release/goyek/goyek)](https://github.com/goyek/goyek/releases)
[![go.mod](https://img.shields.io/github/go-mod/go-version/goyek/goyek)](go.mod)
[![LICENSE](https://img.shields.io/github/license/goyek/goyek)](LICENSE)

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
is used to create build automation in Go.
As opposed to many other tools, it is just a Go library
with API inspired by
[`testing`](https://golang.org/pkg/testing),
[`cobra`](https://github.com/spf13/cobra),
[`flag`](https://golang.org/pkg/flag),
[`http`](https://golang.org/pkg/http)
packages.

Here are some good parts:

- It is cross-platform and shell independent.
- No binary installation is needed.
- It is easy to debug, like a regular Go application.
- The tasks are defined similarly to
  [`cobra`](https://github.com/spf13/cobra) commands.
- The task actions look like a Go unit test.
  You may even use [`testify`](https://github.com/stretchr/testify)
  or [`fluentassert`](https://github.com/fluentassert/verify) for asserting.
- You can reuse code like in any Go application.
  It may be helpful to use packages like
  [`fsnotify`](https://github.com/fsnotify/fsnotify) and [`viper`](https://github.com/spf13/viper).
- It is highly customizable.
- It does not use any third-party dependency other than the Go standard library.
  You can find supplumental features in [`goyek/x`](https://github.com/goyek/x).
- Minimal supported Go version is 1.11.

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

var hello = flow.Define(goyek.Task{
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
	"github.com/goyek/goyek/v2"
	"github.com/goyek/x/boot"
)

func main() {
	goyek.SetDefault(hello)
	boot.Main()
}
```

Run:

```out
$ go mod tidy

$ go run ./build -h
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

$ go run ./build -v
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

- [example_test.go](example_test.go) -
  demonstrative examples
- [goyek/template](https://github.com/goyek/template) -
  Go application repository template
- [fluentassert](https://github.com/fluentassert/verify) -
  Go library
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

Instead of executing `go run ./build`,
you may create wrapper scripts,
which you can invoke from any locationn.

Bash - `goyek.sh`:

```bash
#!/bin/bash
set -euo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
cd "$DIR"
go run ./build $@
```

PowerShell - `goyek.ps1`:

```powershell
& go run .\build $args
exit $global:LASTEXITCODE
```

If `/build` is a separate Go module,
check the [goyek.sh](goyek.sh) and [goyek.ps1](goyek.ps1) scripts.

## Using middlewares

Call the [`Use`](https://pkg.go.dev/github.com/goyek/goyek/v2#Use) function
to setup a task runner interceptor (middleware).

You can use a middleware, for example to:
generate a task execution report,
add retry logic,
export build execution telemetry, etc.

You can use some reusable middlewares from the
[`middleware`](https://pkg.go.dev/github.com/goyek/goyek/v2/middleware)
package. [`ReportStatus`](https://pkg.go.dev/github.com/goyek/goyek/v2/middleware#ReportStatus)
is the most commonly used middleware.

Notice that the [`boot.Main`](https://pkg.go.dev/github.com/goyek/x/boot#Main)
convenient function from [goyek/x](https://github.com/goyek/x)
sets the most commonly used middlewares and defines flags to configure them.

## Customizing

You can customize the default output by using:

- [`SetOutput`](https://pkg.go.dev/github.com/goyek/goyek/v2#SetOutput)
- [`SetLogger`](https://pkg.go.dev/github.com/goyek/goyek/v2#SetLogger)
- [`SetUsage`](https://pkg.go.dev/github.com/goyek/goyek/v2#SetUsage)
- [`Execute`](https://pkg.go.dev/github.com/goyek/goyek/v2#Execute)
  (instead of [`Main`](https://pkg.go.dev/github.com/goyek/goyek/v2#Main))

You can also study how [github.com/goyek/x](https://github.com/goyek/x)
is customizing the default behavior.

## Alternatives

### Make

While [Make (Makefile)](https://www.gnu.org/software/make/) is currently
the _de facto_ standard, it has some pitfalls:

- Requires to learn Make (and often Bash).
- It is hard to develop a Makefile which is truly cross-platform.
- Debugging and testing Make targets is not fun.

**goyek** is intended to be simpler, easier to learn,
more portable, while still being able to handle most use cases.

### Mage

[Mage](https://github.com/magefile/mage) is a framework/tool which magically discovers
the [targets](https://magefile.org/targets/) from [magefiles](https://magefile.org/magefiles/),
which results in some drawbacks.

- It requires using [build tags](https://magefile.org/magefiles/).
- Reusing tasks is [hacky](https://magefile.org/importing/).
- It needs installation or use of [zero install option](https://magefile.org/zeroinstall/),
  which is slow.
- Debugging is [complex](https://github.com/magefile/mage/issues/280).
- It is magical by design (of course, one may like it).

**goyek** is a non-magical alternative for [Mage](https://github.com/magefile/mage).
It is easier to customize and extend as it is a library that offers extension points.
Write regular Go code without build tags and tricky imports.

### Task

While [Task](https://taskfile.dev/) is simpler and easier to use
than [Make](https://www.gnu.org/software/make/),
but it still has similar problems:

- Requires to learn Task's YAML structure and
  the [minimalistic, cross-platform interpreter](https://github.com/mvdan/sh#gosh).
- Debugging and testing tasks is not easy.
- Hard to make reusable tasks.
- Requires to install the tool.

### Bazel

[Bazel](https://bazel.build/) is a very sophisticated tool which is
[created to efficiently handle complex and long-running build pipelines](https://en.wikipedia.org/wiki/Bazel_(software)#Rationale).
It requires the build target inputs and outputs to be fully specified.

**goyek** is just a simple Go library.
However, nothing prevents you from, for example,
using the [github.com/magefile/mage/target](https://pkg.go.dev/github.com/magefile/mage/target)
package to make your automation more efficient.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) if you want to help us.

## License

**goyek** is licensed under the terms of the [MIT license](LICENSE).

Note: **goyek** was named **taskflow** before v0.3.0.
