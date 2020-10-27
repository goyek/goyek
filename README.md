# taskflow

[![go.dev](https://img.shields.io/badge/go.dev-reference-blue.svg)](https://pkg.go.dev/github.com/pellared/taskflow)
[![go.mod](https://img.shields.io/github/go-mod/go-version/pellared/taskflow)](go.mod)
[![Build Status](https://img.shields.io/github/workflow/status/pellared/taskflow/build)](https://github.com/pellared/taskflow/actions?query=workflow%3Abuild+branch%3Amaster)
[![Go Report Card](https://goreportcard.com/badge/github.com/pellared/taskflow)](https://goreportcard.com/report/github.com/pellared/taskflow)

This package aims to simplify creation of build pipelines in Go instead of using scripts or [Make](https://www.gnu.org/software/make/).

`Star` this repository if you find it valuable and worth maintaining.

`Watch` this repository to get notified about new releases, issues, etc.

## Usage

Take a look at the dogfooding [build pipeline](build/main.go).

Clone this repo and execute:

```shell
$ go run ./build -h
Usage: [flag(s)] task(s)
Flags:
  -v    verbose
Tasks:
  build       go build
  ci          CI build pipeline
  clean       remove files created during build
  dev         dev build pipeline
  diff        git diff
  fmt         gofumports
  install     install build tools
  lint        golangci-lint-lintports
  mod-tidy    go mod tidy
  test        go test with race detector and code covarage

$ go run ./build dev
ok     1.199s

$ go run ./build -v dev
===== TASK  clean
removed coverage.html
removed coverage.out
----- PASS: clean (0.00s)
===== TASK  install
Exec: go install mvdan.cc/gofumpt/gofumports
Exec: go install github.com/golangci/golangci-lint/cmd/golangci-lint
----- PASS: install (0.21s)
===== TASK  build
Exec: go build ./...
----- PASS: build (0.25s)
===== TASK  fmt
Exec: gofumports -l -w -local github.com/pellared/taskflow .
----- PASS: fmt (0.03s)
===== TASK  lint
Exec: golangci-lint run
----- PASS: lint (0.19s)
===== TASK  test
Exec: go test -race -covermode=atomic -coverprofile=coverage.out ./...
ok      github.com/pellared/taskflow    0.029s  coverage: 67.3% of statements
?       github.com/pellared/taskflow/build      [no test files]
Exec: go tool cover -html=coverage.out -o coverage.html
----- PASS: test (0.39s)
===== TASK  mod-tidy
Exec: go mod tidy
Exec: go mod tidy
----- PASS: mod-tidy (0.13s)
ok      1.207s
```

Tired of writing `go run ./build` each time? Just add an alias to your shell. For example by adding the line below to `~/.bash_aliases`:

```shell
alias gake='go run ./build'
```

## FAQ

### Is taskflow stable

No, it is in experimental phase. I am open to any feedback.

### Why not to use Make

While [Make](https://www.gnu.org/software/make/) is currently de facto standard it has some pitfalls:

- Requires to learn Make which is not so easy.
- It is hard to develop a Makefile which is truly cross-platform.
- Debugging and testing Make targets is not fun.

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

## Credits

**taskflow** is mainly inspired by:

- [Mage](https://github.com/magefile/mage),
- [testing](https://golang.org/pkg/testing),
- [flag](https://golang.org/pkg/flag).

## Contributing

Simply create an issue or a pull request.
