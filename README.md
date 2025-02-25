# goyek

[![Go Reference](https://pkg.go.dev/badge/github.com/goyek/goyek.svg)](https://pkg.go.dev/github.com/goyek/goyek/v2)
[![Keep a Changelog](https://img.shields.io/badge/changelog-Keep%20a%20Changelog-%23E05735)](CHANGELOG.md)
[![GitHub Release](https://img.shields.io/github/v/release/goyek/goyek)](https://github.com/goyek/goyek/releases)
[![go.mod](https://img.shields.io/github/go-mod/go-version/goyek/goyek)](go.mod)
[![Build Status](https://img.shields.io/github/actions/workflow/status/goyek/goyek/build.yml?branch=main)](https://github.com/goyek/goyek/actions?query=workflow%3Abuild+branch%3Amain)
[![Go Report Card](https://goreportcard.com/badge/github.com/goyek/goyek)](https://goreportcard.com/report/github.com/goyek/goyek)
[![codecov](https://codecov.io/gh/goyek/goyek/branch/main/graph/badge.svg)](https://codecov.io/gh/goyek/goyek)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

**goyek** (/ˈɡɔɪæk/ [🔊 listen](http://ipa-reader.xyz/?text=%CB%88%C9%A1%C9%94%C9%AA%C3%A6k))
is a task automation library.

This library is intended to be an alternative to
[Make](https://www.gnu.org/software/make/),
[Mage](https://github.com/magefile/mage),
[Task](https://taskfile.dev/).

Please ⭐ `Star` this repository if you find it valuable and worth maintaining.

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
  [`goyek.A`](https://pkg.go.dev/github.com/goyek/goyek/v2#A)
  has similar methods to [`testing.T`](https://pkg.go.dev/testing#T).
- Reuse any Go code and library e.g. [`viper`](https://github.com/spf13/viper).
- Highly customizable.
- No third-party dependencies.
- Supplumental features in [`goyek/x`](https://github.com/goyek/x).

## Usage

For build automation, the convention is to have the code in the `/build`
directory (or even Go module).

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

The packages from [`github.com/goyek/x`](https://pkg.go.dev/github.com/goyek/x)
are used for convinence.

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

Instead of executing `go run .` in `build` directory,
you may prefer using the wrapper scripts,
which you can invoke from any location.

- Bash: [`goyek.sh`](goyek.sh).
- PowerShell: [`goyek.ps1`](goyek.ps1).

You can use [`goyek/template`](https://github.com/goyek/template)
to create a new repository.

For an existing repository you can copy most of its files.

You can watch a 5 min [video]((https://www.youtube.com/watch?v=e-xWEH-fqJ0))
([slides](https://docs.google.com/presentation/d/1xFAPXeMiOD-92xeIHkUD-SHmJZwc8mSIIgpjuJXEW3U/edit?usp=sharing)).

If you like looking at real usages, check build pipelines of
[`goyek/x`](https://github.com/goyek/x/tree/main/build)
or [`splunk-otel-go`](https://github.com/signalfx/splunk-otel-go/tree/main/build).

See the [documentation](https://pkg.go.dev/github.com/goyek/goyek/v2) for more information.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) if you want to help us.

## License

**goyek** is licensed under the terms of the [MIT license](LICENSE).

Note: **goyek** was named **taskflow** before v0.3.0.
