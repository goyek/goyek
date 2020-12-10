# taskflow

[![go.dev](https://img.shields.io/badge/go.dev-reference-blue.svg)](https://pkg.go.dev/github.com/pellared/taskflow)
[![go.mod](https://img.shields.io/github/go-mod/go-version/pellared/taskflow)](go.mod)
[![Build Status](https://img.shields.io/github/workflow/status/pellared/taskflow/build)](https://github.com/pellared/taskflow/actions?query=workflow%3Abuild+branch%3Amaster)
[![Go Report Card](https://goreportcard.com/badge/github.com/pellared/taskflow)](https://goreportcard.com/report/github.com/pellared/taskflow)
[![codecov](https://codecov.io/gh/pellared/taskflow/branch/master/graph/badge.svg)](https://codecov.io/gh/pellared/taskflow)

This package aims to simplify creation of build pipelines in Go instead of using scripts or [Make](https://www.gnu.org/software/make/).

**taskflow** API is mainly inspired by [testing](https://golang.org/pkg/testing), [http](https://golang.org/pkg/http) and [flag](https://golang.org/pkg/flag) packages.

Check [Go Build Pipeline Demo](https://github.com/pellared/go-build-pipeline-demo) to compare **taskflow** with [Make](https://www.gnu.org/software/make/) and [Mage](https://github.com/magefile/mage).

I am open to any feedback and contribution. Use [Discussions](https://github.com/pellared/taskflow/discussions) or write to me: *Robert Pajak* @ [Gophers Slack](https://invite.slack.golangbridge.org/).

`Star` this repository if you find it valuable and worth maintaining.

## Example

Paste the following code to `build/main.go`:

```go
package main

import "github.com/pellared/taskflow"

func main() {
	tasks := &taskflow.Taskflow{}

	fmt := tasks.MustRegister(taskflow.Task{
		Name:        "fmt",
		Description: "go fmt",
		Command:     taskFmt,
	})

	test := tasks.MustRegister(taskflow.Task{
		Name:        "test",
		Description: "go test with race detector and code covarage",
		Command:     taskTest,
	})

	tasks.MustRegister(taskflow.Task{
		Name:        "all",
		Description: "build pipeline",
		Dependencies: taskflow.Deps{
			fmt,
			test,
		},
	})

	tasks.Main()
}

func taskFmt(tf *taskflow.TF) {
	if err := tf.Exec("", nil, "go", "fmt", "./..."); err != nil {
		tf.Errorf("go fmt: %v", err)
	}
}

func taskTest(tf *taskflow.TF) {
	if err := tf.Exec("", nil, "go", "test", "-race", "-covermode=atomic", "-coverprofile=coverage.out", "./..."); err != nil {
		tf.Errorf("go test: %v", err)
	}
	if err := tf.Exec("", nil, "go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html"); err != nil {
		tf.Errorf("go tool cover: %v", err)
	}
}
```

Sample usage:

```shell
$ go run ./build -h
Usage: [flag(s)] task(s)
Flags:
  -v    verbose
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

Additionally, take a look at the dogfooding [build pipeline](build/main.go).

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

## Contributing

Simply create an issue or a pull request.
