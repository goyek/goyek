# Changelog

All notable changes to this library are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this library adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html)
as well as to [Module version numbering](https://go.dev/doc/modules/version-numbers).

## [Unreleased](https://github.com/goyek/goyek/compare/v2.0.0-rc.11...HEAD)

<!-- markdownlint-disable-next-line line-length -->
## [2.0.0-rc.11](https://github.com/goyek/goyek/compare/v2.0.0-rc.10...v2.0.0-rc.11) - 2022-11-07

### Changed

- `Flow.Main` no longer changes the working directory (undocumented behavior).

<!-- markdownlint-disable-next-line line-length -->
## [2.0.0-rc.10](https://github.com/goyek/goyek/compare/v2.0.0-rc.9...v2.0.0-rc.10) - 2022-11-07

This release adds helpers methods for type `A` that are commonly used
in the [`testing`](https://pkg.go.dev/testing) package.

### Added

- Add `A.Cleanup` method that registers an action's cleanup function.
- Add `A.Setenv` method that sets the environment variable
  and reverts the previous value during cleanup.
- Add `A.TempDir` method that creates a temporary directory
  and removes it during cleanup.

<!-- markdownlint-disable-next-line line-length -->
## [2.0.0-rc.9](https://github.com/goyek/goyek/compare/v2.0.0-rc.8...v2.0.0-rc.9) - 2022-11-06

This release introduces a breaking change as it renames
the heavily used `TF` type to `A`. This follows the convention
used in the [`testing`](https://pkg.go.dev/testing) package
to name the parameter type as the first letter of the "function type".

### Added

- Add `Status.String` method for printing `Status`.

### Changed

- Rename `TF` type to `A` as this is the `Action` parameter.

<!-- markdownlint-disable-next-line line-length -->
## [2.0.0-rc.8](https://github.com/goyek/goyek/compare/v2.0.0-rc.7...v2.0.0-rc.8) - 2022-10-29

### Added

- Make logging more customizable. If `Logger` implements
  `Error`, `Errorf`, `Fatal`, `Fatalf`, `Skip`, `Skipf`
  then they will be used by the `TF` equivalent methods.

<!-- markdownlint-disable-next-line line-length -->
## [2.0.0-rc.7](https://github.com/goyek/goyek/compare/v2.0.0-rc.6...v2.0.0-rc.7) - 2022-10-29

This release has all the features and changes planned for `v2`.
The `v2.0.0` is to be released in two months
to give the users some time for feedback.

### Added

- Add `TF.Helper` and `CodeLineLogger.Helper` methods that work like
  the equivalent method in the `testing` package.

<!-- markdownlint-disable-next-line line-length -->
## [2.0.0-rc.6](https://github.com/goyek/goyek/compare/v2.0.0-rc.5...v2.0.0-rc.6) - 2022-10-27

### Added

- Add `NoDeps` option to skip processing of all dependencies.
- Add `Skip` option to skip processing of given tasks.

### Changed

- Report `NOOP` status if the task's action was nil.

<!-- markdownlint-disable-next-line line-length -->
## [2.0.0-rc.5](https://github.com/goyek/goyek/compare/v2.0.0-rc.4...v2.0.0-rc.5) - 2022-10-25

### Changed

- Change `DefinedTask` to struct.

<!-- markdownlint-disable-next-line line-length -->
## [2.0.0-rc.4](https://github.com/goyek/goyek/compare/v2.0.0-rc.3...v2.0.0-rc.4) - 2022-10-25

This release focuses on improving the API to make creating and customizing
reusable build pipelines easier.

### Added

- Add `DefinedTask.SetName`, `DefinedTask.SetUsage`, `DefinedTask.Action`,
  `DefinedTask.SetAction`, `DefinedTask.SetAction`, `DefinedTask.SetDeps`
  to enable modifying the task after the initial definition.
- Add `Flow.Undefine` to unregister a task.
- Passing `nil` to `Flow.SetDefault` unsets the default task.
- Add `DryRun`, `ReportLongRun` middlewares.

### Changed

- `DefinedTask.Deps` returns `Deps` to facilitate reusing defined task's dependencies
  when creating a new one or redefining existing one.
- Rename `Reporter` middleware to `ReportStatus`.
- Change `Flow.Execute` to accept `[]string` instead of `...string` to make the API
  forward compatible.

<!-- markdownlint-disable-next-line line-length -->
## [2.0.0-rc.3](https://github.com/goyek/goyek/compare/v2.0.0-rc.2...v2.0.0-rc.3) - 2022-10-19

This release focuses on improving usability, extensibility, and customization.

### Added

- Add `Flow.SetLogger` for setting a custom log decorator
  that is used by `TF` logging methods.
- Add `Flow.Logger` for getting the log decorator (`CodeLineLogger` by default).
- Add `CodeLineLogger` which is the default for `Flow.Logger`.
- Add `FmtLogger` which is the default when using `NewRunner`.
- Add `NOOP` status report for tasks that were intentionally not run
  to differentiate from being skipped during execution.
- Add `Flow.Use` method o support task run interception using middlewares.
- Add `middleware` package with `ReportStatus` and `SilentNonFailed` middlewares.
- `TF.Error`, `TF.Errorf`, `TF.Fail` may be called simultaneously from multiple goroutines.
- Add `DefaultFlow` that is the default flow.
- Add the top-level functions such as `Define`, `Main`, and so on which are wrappers
  for the methods of `Flow` called for `DefaultFlow`.

### Changed

- Usually, the task does not call `panic` directly.
  `panic` failure message no longer contains a prefix with file and line information.
  The stack trace is printed instead. The behavior is based on `testing` package.
- `Flow.Main` changes the working directory to parent.
- Rename `Flow.Run` to `Flow.Execute` to reduce possible confusion with `Runner`.
- Report `PASS` for a task without an action.
- Task status reporting is disabled by default.
  It can be enabled by calling `Flow.Use(middleware.ReportStatus)`.
- `Flow.Print` output format is similar to `flag.PrintDefaults`.
  Moreover, it does not print tasks with empty `Task.Usage`.
- Change `Flow.Execute` to return an error instead of returning the exit code
  and printing to output.
- Change `Flow.Output` field to `Flow.SetOutput` setter and `Flow.Output` getter.
- Change `Flow.Usage` field to `Flow.SetUsage` setter and `Flow.Usage` getter.

### Removed

- `Flow.Verbose` is removed.
  To be non-verbose use `Flow.Use(middleware.SilentNonFailed)` instead.

### Fixed

- Fix panic handling so that `panic(nil)` and `runtime.Goexit()` now cause task failure.

<!-- markdownlint-disable-next-line line-length -->
## [2.0.0-rc.2](https://github.com/goyek/goyek/compare/v2.0.0-rc.1...v2.0.0-rc.2) - 2022-10-14

This release focuses on improving usability and encapsulation.

### Added

- Add `Flow.Usage` for setting a custom parsing tasks error handler.

### Changed

- `Flow.Usage` (or `Flow.Print` if it is `nil`)
  is called when `Flow.Run` returns `CodeInvalidArgs`.
- Rename `Flow.Register` to `Flow.Define`.
- Change `Flow.RegisteredTask` to sealed interface `Flow.DefinedTask`.
- Change the `Flow.DefaultTask` field to `Flow.SetDefault` and `Flow.Default` methods.

## [2.0.0-rc.1](https://github.com/goyek/goyek/compare/v1.1.0...v2.0.0-rc.1) - 2022-10-14

This release contains **breaking changes** in the Go API.
It focuses on improving customization mainly by removing the parameters API.
It gives the user the possibility of parsing the input.

### Added

- Add `Flow.Verbose` for controlling the verbose mode.
- Add `Flow.Print` for printing the flow usage.
- `Flow.Main` now exits on receiving the second SIGINT.

### Changed

- `RegisteredTask.Deps` returns `[]string` (dependency names) for easier introspection.

### Removed

- Remove parameters API and out-of-the-box flags (`-v`, `-wd`).

## [1.1.0](https://github.com/goyek/goyek/compare/v1.0.0...v1.1.0) - 2022-10-12

This release focuses on improving output printing.
There are no API changes.

### Added

- The `TF` methods `Log[f]`, `Error[f]`, `Fatal[f]`, `Skip[f]`, `panic`
  prints indented text with a prefix containing file and line information.

### Fixed

- Remove `TF.Cmd` undocumented behavior (printing command name and arguments).

## [1.0.0](https://github.com/goyek/goyek/compare/v0.6.3...v1.0.0) - 2022-09-08

This is the first stable release.

## [0.6.3](https://github.com/goyek/goyek/compare/v0.6.2...v0.6.3) - 2022-02-23

### Added

- `TaskNamePattern` and `ParamNamePattern` has been changed so that
  the plus (`+`) and colon (`:`) characters can be used
  in the task and parameter name (but not as the first character).
  
## [0.6.2](https://github.com/goyek/goyek/compare/v0.6.1...v0.6.2) - 2022-01-25

### Fixed

- The `Taskflow` does not fail when it was requested to cancel,
  but was completed successfully.

## [0.6.1](https://github.com/goyek/goyek/compare/v0.6.0...v0.6.1) - 2021-12-27

This release adds the possibility to change the defaults of the global parameters.

### Added

- Add `RegisterVerboseParam` and `RegisterWorkDirParam` to overwrite
  default values for verbose and work dir parameters.

## [0.6.0](https://github.com/goyek/goyek/compare/v0.5.0...v0.6.0) - 2021-08-02

This release contains multiple **breaking changes** in the Go API.
It is supposed to make it cleaner.

### Added

- `TF.Cmd` also sets `Stdin` to `os.Stdin`.
- Add `Flow.Tasks` and `Flow.Params` to allow introspection
  of the registered tasks and parameters.

### Changed

- Rename `Taskflow` type to `Flow` to simplify the naming.
- Rename `Task.Command` field to `Action` to avoid confusion with
  [`exec.Command`](https://golang.org/pkg/os/exec/#Command) and `TF.Cmd`.

### Removed

- Remove `DefaultOutput` global variable.
- Remove `TF.Exec` method.

## [0.5.0](https://github.com/goyek/goyek/compare/v0.4.0...v0.5.0) - 2021-06-21

### Added

- Add the stack trace when a task panics.

### Changed

- Names of parameters now have to follow `goyek.ParamNamePattern`.
  They were allowed to start with an underscore (`_`), and now no longer are.
- The PowerShell wrapper scripts `goyek.ps1` better handles `stderr` redirection.

## [0.4.0](https://github.com/goyek/goyek/compare/v0.3.0...v0.4.0) - 2021-05-26

### Added

- Add Bash and PowerShell wrapper scripts.
- Add `-wd` global parameter allowing to change the working directory.
  The new `Taskflow.WorkDirParam` method can be used to get its value in
  a task's command.

## [0.3.0](https://github.com/goyek/goyek/compare/v0.2.0...v0.3.0) - 2021-05-03

The repository has been migrated from <https://github.com/pellared/taskflow>
to <https://github.com/goyek/goyek>.

This release contains multiple breaking changes for both the CLI and the Go API.

The biggest change is a redesign of the parameters API,
so they have to be explicitly registered.
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
- Rename `Taskflow.MustRegister` method to `Register`
  and remove previous `Taskflow.Register` implementation.
- Remove `Taskflow.Params` field and `TF.Params` method,
  add `Taskflow.Register*Param` methods and `Task.Params` field instead.
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

<!-- markdownlint-configure-file
{
  "MD024": {
    "siblings_only": true
  }
}
-->
