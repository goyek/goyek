# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased](https://github.com/pellared/taskflow/compare/v0.2.0...HEAD)

This release contains multiple breaking changes for both the CLI and the Go API.

The biggest change is a redesign of the parameters API, so they have to be explicitly registered.
It makes the usage of the parameters more controlled and provides a better help output.
Moreover, the parameters are set via CLI using the flag syntax.

### Added

- Help is printed when `-h`, `--help` or `help` is passed.
- Help contains parameters' information.
- The tasks and parameters can be passed via CLI in any order.
- `Taskflow.Run` panics when a registered parameter is not assigned to any task.
### Changed

- Use the flag syntax for setting parameters via CLI.
- Rename `Task.Description` field to `Usage`.
- Rename `Task.Dependencies` field to `Deps`.
- Rename `Taskflow.MustRegister` method to `Register` and remove previous `Taskflow.Register` implemention.
- Remove previous parameters API, add `Taskflow.Register*Param` methods and `Task.Params` field instead.
- Remove `TF.Verbose`, add `Taskflow.VerboseParam` instead.
- Drop official support for Go 1.10.

## [0.2.0](https://github.com/pellared/taskflow/compare/v0.1.1...v0.2.0) - 2021-03-14

### Added

- Add the possibility to set a default task.

## [0.1.1](https://github.com/pellared/taskflow/compare/v0.1.0...v0.1.1) - 2021-02-28

### Fixed

- Make concurrent printing thread-safe.

## [0.1.0](https://github.com/pellared/taskflow/releases/tag/v0.1.0) - 2021-01-14

### Added

- First release version after the experiential phase.
