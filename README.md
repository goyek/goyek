# taskflow

[![GitHub Release](https://img.shields.io/github/v/release/pellared/taskflow)](https://github.com/pellared/taskflow/releases)
[![go.dev](https://img.shields.io/badge/go.dev-reference-blue.svg)](https://pkg.go.dev/github.com/pellared/taskflow)
[![go.mod](https://img.shields.io/github/go-mod/go-version/pellared/taskflow)](go.mod)
[![Build Status](https://img.shields.io/github/workflow/status/pellared/taskflow/build)](https://github.com/pellared/taskflow/actions?query=workflow%3Abuild+branch%3Amaster)
[![Go Report Card](https://goreportcard.com/badge/github.com/pellared/taskflow)](https://goreportcard.com/report/github.com/pellared/taskflow)

This package ais to simplify creation of task-based workflows in Go.
It can be used for example to create a build workflow in pure Go instead of using scripts or [Make](https://www.gnu.org/software/make/).

`Star` this repository if you find it valuable and worth maintaining.

`Watch` this repository to get notified about new releases, issues, etc.

## FAQ

### Is taskflow stable

No. It is a PoC which is under development.

## Why not to use Make

While [Make](https://www.gnu.org/software/make/) is currently de facto standard it has some pitfalls:

- Requires to learn Make (which is not so easy).
- It is hard to develop a Makefile which is truly cross-platform.

Maybe better expiation can be found [here](https://github.com/magefile/mage#why).

### Why not to use Mage

**taskflow** is intended to be an alternative to [Mage](https://github.com/magefile/mage).

[Mage](https://github.com/magefile/mage) is a framework/tool which magically discovers the [targets](https://magefile.org/targets/) from [magefiles](https://magefile.org/magefiles/).

**taskflow** takes a different approach as it is a regular Go library (package).
This results in following benefits:

- It can be easily used in more places than just for build automation.
- Less magical. Pure Go code.

## Contributing

Simply create an issue or a pull request.
