# goyek

[![Go Reference](https://pkg.go.dev/badge/github.com/goyek/goyek.svg)](https://pkg.go.dev/github.com/goyek/goyek/v3)
[![Keep a Changelog](https://img.shields.io/badge/changelog-Keep%20a%20Changelog-%23E05735)](CHANGELOG.md)
[![go.mod](https://img.shields.io/github/go-mod/go-version/goyek/goyek)](go.mod)
[![Go Report Card](https://goreportcard.com/badge/github.com/goyek/goyek/v3)](https://goreportcard.com/report/github.com/goyek/goyek/v3)
[![codecov](https://codecov.io/gh/goyek/goyek/branch/main/graph/badge.svg)](https://codecov.io/gh/goyek/goyek)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

**goyek** (/ˈɡɔɪæk/ [🔊 listen](https://ipa-reader.com/?text=%CB%88%C9%A1%C9%94%C9%AA%C3%A6k))
is a task automation library intended to be an alternative to
[Make](https://www.gnu.org/software/make/),
[Mage](https://github.com/magefile/mage),
[Task](https://taskfile.dev/).

The primary properties of goyek are:

- Library, not an application, with API inspired by
  [`testing`](https://golang.org/pkg/testing),
  [`cobra`](https://github.com/spf13/cobra),
  [`flag`](https://golang.org/pkg/flag),
  [`http`](https://golang.org/pkg/http).
- Cross-platform and shell independent.
- No binary installation needed.
- Easy to debug, like regular Go code.
- Tasks are defined similarly to
  [`cobra`](https://github.com/spf13/cobra) commands.
- The task action looks like a Go test.
  [`goyek.A`](https://pkg.go.dev/github.com/goyek/goyek/v3#A)
  has similar methods to [`testing.T`](https://pkg.go.dev/testing#T).
- Reuse any Go code and library e.g. [`viper`](https://github.com/spf13/viper).
- Highly customizable.
- Zero third-party dependencies.
- Additional features in [`goyek/x`](https://github.com/goyek/x).

5-minute video: [Watch here](https://www.youtube.com/watch?v=e-xWEH-fqJ0)
([Slides](https://docs.google.com/presentation/d/1xFAPXeMiOD-92xeIHkUD-SHmJZwc8mSIIgpjuJXEW3U/edit?usp=sharing)).

Please ⭐ `Star` this repository if you find it valuable.

## Usage

For build automation, store your code in the `build` directory.

The following example defines a simple `hello` task that logs a message
and prints the Go version.

Create `build/hello.go`:

```go
package main

import (
	"flag"

	"github.com/goyek/goyek/v3"
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

Create `build/main.go`:

```go
package main

import (
	"os"

	"github.com/goyek/goyek/v3"
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
are used for convenience.

Run help:

```sh
cd build
go mod tidy
go run . -h
```

Expected output:

```out
Usage of build: [tasks] [flags] [--] [args]
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
```

Run with verbose output:

```sh
go run . -v
```

Example output:

```out
===== TASK  hello
      hello.go:16: greeting message
      hello.go:17: Exec: go version
go version go1.24.0 linux/amd64
----- PASS: hello (0.12s)
ok      0.123s
```

Instead of running `go run .` inside `build`, you can use wrapper scripts:

- Bash: [`goyek.sh`](goyek.sh).
- PowerShell: [`goyek.ps1`](goyek.ps1).

Use [`goyek/template`](https://github.com/goyek/template) when creating
a new repository. For existing repositories, simply copy the relevant files.

See the [documentation](https://pkg.go.dev/github.com/goyek/goyek/v3) for more information.

## Examples

The following repositories demonstrate real-world usage of goyek:

- [goyek/template](https://github.com/goyek/template): Go application
  repository template using goyek for build automation.
- [build/](build/): this repository dogfoods goyek for its own build
  pipeline.
- [splunk-otel-go](https://github.com/signalfx/splunk-otel-go/tree/main/build):
  multi-module repository using goyek.
- [curioswitch/go-build](https://github.com/curioswitch/go-build) and projects
  such as [curioswitch/tasuke](https://github.com/curioswitch/tasuke/blob/main/build/main.go):
  an example of extracting reusable build tasks into a shared module consumed by
  multiple repositories.

## Rationale

**goyek** was originally built as a small library for expressing Go build
pipelines in Go itself instead of relying on an external DSL or tool.
Keeping build logic as regular Go code makes it easy to reuse existing
packages, refactor with IDE support, and debug tasks with the standard Go
toolchain.

Although build automation remains the primary use case, goyek is designed as
a general task runner. Tasks are ordinary Go functions wired together through
dependencies, which makes it suitable for other domains as well. It is known
to be used, for example, by SRE and platform teams to automate deployment
pipelines and other operational workflows.

Compared to [Make](https://www.gnu.org/software/make/) and similar tools,
goyek does not require learning a separate DSL or shell tricks and makes it
easier to keep pipelines portable across platforms. Unlike
[Mage](https://github.com/magefile/mage), it avoids build tags, code
generation, and magic discovery of targets – you just write and compose
ordinary Go functions. In contrast to [Task](https://taskfile.dev/), there is
no YAML to learn or additional binary to install, and debugging tasks is
like debugging any other Go program. Finally, while systems such as
[Bazel](https://bazel.build/) focus on large, hermetic build graphs, goyek
deliberately stays small and library‑like, so you can start with a simple
build script and grow into more advanced pipelines only when needed.

## Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

**goyek** is licensed under the terms of the [MIT license](LICENSE).

Note: **goyek** was named **taskflow** before v0.3.0.
