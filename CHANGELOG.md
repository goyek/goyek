# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased](https://github.com/goyek/goyek/compare/v0.3.0...HEAD)

### Added

- Add Bash and PowerShell wrapper scripts.
- Add `-wd` global parameter allowing to change the working directory.
  The new `Taskflow.WorkDirParam` method can be used to get its value in a task's command.

## [0.3.0](https://github.com/goyek/goyek/compare/v0.2.0...v0.3.0) - 2021-05-03

The repository has been migrated from <https://github.com/pellared/taskflow>
to <https://github.com/goyek/goyek>.

This release contains multiple breaking changes for both the CLI and the Go API.

The biggest change is a redesign of the parameters API, so they have to be explicitly registered.
It makes the usage of the parameters more controlled and provides a better help output.
Moreover, the parameters are set via CLI using the flag syntax.

### Added

- Help is printed when `-h`, `--help` or `help` is passed.
- Help contains parameters' information.
- The tasks and parameters can be passed via CLI in any order.
- `Taskflow.Run` handles `nil` passed as `context.Context` argument.
- `Taskflow.Run` panics when a registered parameter is not assigned to any task.

### Changed

- Module path changed from `github.com/pellared/taskflow` to `github.com/goyek/goyek`.
- Rename package `task` to `goyek`.
- Use the flag syntax for setting parameters via CLI.
- Rename `Task.Description` field to `Usage`.
- Rename `Task.Dependencies` field to `Deps`.
- Rename `CodeFailure` constant to `CodeFail`.
- Rename `Taskflow.MustRegister` method to `Register` and remove previous `Taskflow.Register` implementation.
- Remove `Taskflow.Params` field and `TF.Params` method, add `Taskflow.Register*Param` methods and `Task.Params` field instead.
- Remove `TF.Verbose`, add `Taskflow.VerboseParam` instead.
- Unexport `Runner` type, use `Taskflow` in tests instead.
- Enforce patterns for task names (`TaskNamePattern`) and parameter names (`ParamNamePattern`).

### Removed

- Remove `New` function, create instance using `&Taskflow{}` instead.
- Drop official support for Go 1.10.

## [0.2.0](https://github.com/goyek/goyek/compare/v0.1.1...v0.2.0) - 2021-03-14

### Added

- Add the possibility to set a default task.

## [0.1.1](https://github.com/goyek/goyek/compare/v0.1.0...v0.1.1) - 2021-02-28

### Fixed

- Make concurrent printing thread-safe.

## [0.1.0](https://github.com/goyek/goyek/releases/tag/v0.1.0) - 2021-01-14

### Added

- First release version after the experiential phase.
