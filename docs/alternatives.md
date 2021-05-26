# Alternatives

Table of Contents:

- [Alternatives](#alternatives)
  - [Make](#make)
  - [Mage](#mage)
  - [Task](#task)
  - [Bazel](#bazel)

## Make

While [Make](https://www.gnu.org/software/make/) is currently the _de facto_ standard, it has some pitfalls:

- Requires to learn Make, which is not so easy.
- It is hard to develop a Makefile which is truly cross-platform.
- Debugging and testing Make targets is not fun.

However, if you know Make and are happy with it, do not change it.
Make is very powerful and a lot of stuff can be made faster, if you know how to use it.

**goyek** is intended to be simpler and easier to learn, while still being able to handle most use cases.

## Mage

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

## Task

While [Task](https://taskfile.dev/) is simpler and easier to use
than [Make](https://www.gnu.org/software/make/) it still has similar problems:

- Requires to learn Task's YAML structure and
  the [minimalistic, cross-platform interpreter](https://github.com/mvdan/sh#gosh) which it uses.
- Debugging and testing tasks is not fun.
- Hard to make reusable tasks.
- Requires to "install" the tool.

## Bazel

[Bazel](https://bazel.build/) is a very sophisticated tool which is
[created to efficiently handle complex and long-running build pipelines](https://en.wikipedia.org/wiki/Bazel_(software)#Rationale).
It requires the build target inputs and outputs to be fully specified.

**goyek** is just a simple library that is mainly supposed to create a build pipeline
consisting of commands like `go vet`, `go test`, `go build`.
However, take notice that **goyek** is a library. Nothing prevents you from,
for example, using [Mage's target package](https://pkg.go.dev/github.com/magefile/mage/target)
to make your build pipeline more efficient.
