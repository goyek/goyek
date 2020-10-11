# Go Repository Template

[![GitHub Release](https://img.shields.io/github/v/release/golang-templates/seed)](https://github.com/golang-templates/seed/releases)
[![go.dev](https://img.shields.io/badge/go.dev-reference-blue.svg)](https://pkg.go.dev/github.com/golang-templates/seed)
[![go.mod](https://img.shields.io/github/go-mod/go-version/golang-templates/seed)](go.mod)
[![Build Status](https://img.shields.io/github/workflow/status/golang-templates/seed/build)](https://github.com/golang-templates/seed/actions?query=workflow%3Abuild+branch%3Amaster)
[![Go Report Card](https://goreportcard.com/badge/github.com/golang-templates/seed)](https://goreportcard.com/report/github.com/golang-templates/seed)

This is a GitHub repository template for Go. It has been created for ease-of-use for anyone who wants to:

- quickly get into Go without losing too much time on environment setup,
- create a new repoisitory with basic Continous Integration.

It includes:

- continous integration via [GitHub Actions](https://github.com/features/actions),
- build automation via [Make](https://www.gnu.org/software/make),
- dependency management using [Go Modules](https://github.com/golang/go/wiki/Modules),
- code formatting using [gofumpt](https://github.com/mvdan/gofumpt),
- linting with [golangci-lint](https://github.com/golangci/golangci-lint),
- unit testing with [testify](https://github.com/stretchr/testify), [race detector](https://blog.golang.org/race-detector) and [code covarage HTML report](https://blog.golang.org/cover),
- releasing using [GoReleaser](https://github.com/goreleaser/goreleaser),
- dependencies scanning and updating thanks to [Dependabot](https://dependabot.com),
- [Visual Studio Code](https://code.visualstudio.com) configuration with [Go](https://code.visualstudio.com/docs/languages/go) and [Remote Container](https://code.visualstudio.com/docs/remote/containers) support.

`Star` this repository if you find it valuable and worth maintaining.

`Watch` this repository to get notified about new releases, issues, etc.

## Usage

1. Click the `Use this template` button (alt. clone or download this repository).
1. Replace all occurences of `golang-templates/seed` to `your_org/repo_name` in all files.
1. Rename folder `cmd/seed` to `cmd/repo_name` and update [.goreleaser.yml](.goreleaser.yml).
1. Update [LICENSE](LICENSE) and [README.md](README.md).

## Build

- Terminal: `make` to get help for make targets.
- Visual Studio Code: `Terminal` → `Run Build Task... (CTRL+ALT+B)` to execute a fast build.

## Release

The release workflow is triggered each time a tag with `v` prefix is pushed.

_CAUTION_: Make sure to understand the consequences before you bump the major version. More info: [Go Wiki](https://github.com/golang/go/wiki/Modules#releasing-modules-v2-or-higher), [Go Blog](https://blog.golang.org/v2-go-modules).

## Maintainance

Remember to update Go version in [.github/workflows](.github/workflows), [Makefile](Makefile) and [devcontainer.json](.devcontainer/devcontainer.json).

Notable files:
- [devcontainer.json](.devcontainer/devcontainer.json) - Visual Studio Code Remote Container configuration,
- [.github/workflows](.github/workflows) - GitHub Actions workflows,
- [.github/dependabot.yml](.github/dependabot.yml) - Dependabot configuration,
- [.vscode](.vscode) - Visual Studio Code configuration files,
- [.golangci.yml](.golangci.yml) - golangci-lint configuration,
- [.goreleaser.yml](.goreleaser.yml) - GoReleaser configuration,
- [Makefile](Makefile) - Make targets used for development, [CI build](.github/workflows) and [.vscode/tasks.json](.vscode/tasks.json),
- [go.mod](go.mod) - [Go module definition](https://github.com/golang/go/wiki/Modules#gomod),
- [tools.go](tools.go) - [build tools](https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module).

## FAQ

### Why Visual Studio Code editor configuration

Developers that use Visual Studio Code can take advantage of the editor configuration. While others do not have to care about it. Setting configs for each repo is unnecessary time consuming. VS Code is the most popular Go editor ([survey](https://blog.golang.org/survey2019-results)) and it is officially [supported by the Go team](https://blog.golang.org/vscode-go). 

You can always remove the [.devcontainer](.devcontainer) and [.vscode](.vscode) directories if it really does not help you.

### Why GitHub Actions, not any other CI server

GitHub Actions is out-of-the-box if you are already using GitHub.
[Here](https://github.com/mvdan/github-actions-golang) you can learn how to use it for Go.

However, changing to any other CI server should be very simple, because this repository has build logic and tooling installation in Makefile. 

You can also use the `docker` make target to run the build using a docker container.

### How can I use Make on Windows

Use [WSL (Windows Subsystem for Linux)](https://docs.microsoft.com/en-us/windows/wsl/install-win10).
Alternatively install [tdm-gcc](https://jmeubank.github.io/tdm-gcc/) and [Make Windows port](https://gist.github.com/evanwill/0207876c3243bbb6863e65ec5dc3f058).

### How can I create a Docker image, deb/rpm/snap package, Homebrew Tap, Scoop App Manifest etc.

Take a look at GoReleaser [docs](https://goreleaser.com/customization/) as well as [its repo](https://github.com/goreleaser/goreleaser/) how it is dogfooding its functionality.


### How can I create an application installation script

1. Install [GoDownloader](https://github.com/goreleaser/godownloader).
1. Execute: 

```bash
godownloader --repo=your_org/repo_name > ./install.sh
```

3. Push `install.sh` to your repository.
1. Add installation instructions to your `README.md` e.g.:

```bash
curl -sSfL https://raw.githubusercontent.com/your_org/repo_name/master/install.sh | sh -s -- -b /usr/local/bin
```

### How can I create a library instead of an application

You can change the [.goreleaser.yml](.goreleaser.yml) to contain:

```yaml
build:
  skip: true
release:
  github:
  prerelease: auto
```

Alternatively, you can completly remove the usage of GoReleaser if you prefer handcrafted release notes.

### Why the code coverage results are not accurate

By default `go test` records code coverage for the package that is currently tested. If you want to get more accurate (cross-package) coverage, then consider using [go-acc](https://github.com/ory/go-acc). [Read more](https://www.ory.sh/golang-go-code-coverage-accurate/).

### How to integrate with Codecov

1. Sing up on [Codecov](https://codecov.io/) and setup your repository.
1. Additionally consider configuring [Codecov GitHub Application](https://github.com/apps/codecov) for better integration.
1. Add [codecov-action](https://github.com/codecov/codecov-action) to [.github/workflows/build.yml](.github/workflows/build.yml):

```yaml
      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v1.0.10
        with:
          file: ./coverage.out
```

4. Add a badge in [README.md](README.md):

```md
[![codecov](https://codecov.io/gh/your_org/repo_name/branch/master/graph/badge.svg)](https://codecov.io/gh/your_org/repo_name)
```

Codecov integration example:
- [repository](https://github.com/pellared/codecov-integration-example),
- [coverage report](https://codecov.io/gh/pellared/codecov-integration-example/src/master/cmd/seed/main.go),
- [pull request comment](https://github.com/pellared/codecov-integration-example/pull/3#issuecomment-653814033).

Read [Codecov docs](https://docs.codecov.io/docs/quick-start) for further information and troubleshooting.

### How to automate generating git tags for next release version

Auto-tagging can be done in many ways e.g. by using GitHub Actions like:

- [Github Tag Bump](https://github.com/marketplace/actions/github-tag-bump),
- [bumpr](https://github.com/marketplace/actions/bumpr-bump-version-when-merging-pull-request-with-specific-labels),
- [Increment Semantic Version](https://github.com/marketplace/actions/increment-semantic-version),
- [Github Tag](https://github.com/marketplace/actions/github-tag).

However, creating a release tag manually is often the optimal approach. Take notice that this template executes a release workflow each time a git tag with `v` prefix is pushed.

## Contributing

Simply create an issue or a pull request.
