# Changelog

All notable changes to this library are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this library adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html)
as well as to [Module version numbering](https://go.dev/doc/modules/version-numbers).

## [Unreleased](https://github.com/goyek/goyek/compare/v2.2.0...HEAD)

### Added

- Add `A.WithContext` that creates a derived `A` with a changed context.
  Thanks to it `A` can be reused to pass cancelation and values via context
  to the helper functions.

### Removed

- Drop support for Go 1.13, 1.14, and 1.15.

## [2.2.0](https://github.com/goyek/goyek/compare/v2.1.0...v2.2.0) - 2024-08-19

This release adds flow execution middlewares.

### Added

- Add `Flow.UseExecutor` to support flow execution interception using middlewares.
- Add `middleware.ReportFlow` flow execution middleware which reports the flow
  execution status.

### Change

- Extract the flow result reporting from `Flow.Main` to `middleware.ReportFlow`.
  Add the middleware using `Flow.UseExecutor` to achieve a backward compatible
  behavior.

### Removed

- Drop support for Go 1.11 and 1.12.

## [2.1.0](https://github.com/goyek/goyek/compare/v2.0.0...v2.1.0) - 2024-01-17

This release adds parallel task execution support.

### Added

- Add `Task.Parallel` to allow running tasks in parallel.
- Add `middleware.BufferParallel` that helps in not getting mixed output from
  parallel tasks execution.
- Add `Input.Parallel` to allow middlewares to have special handling for
  parallel tasks.

## [2.0.0](https://github.com/goyek/goyek/compare/v1.1.0...v2.0.0) - 2023-02-08

This release contains many **breaking changes**
that are necessary to improve usability, extensibility, and customization.
You can also find new supplemental features in
[`goyek/x`](https://github.com/goyek/x).

### Added

- Add `DefaultFlow` which is the default flow.
- Add the top-level functions such as `Define`, `Main`, and so on which are wrappers
  for the methods of `Flow` called for `DefaultFlow`.
- `Flow.Main` now exits on receiving the second SIGINT.
- Add `Flow.Print` for printing the flow usage.
- `Flow.Main` and `Flow.Execute` allow passing execution options.
- Add `NoDeps` option to skip processing of all dependencies.
- Add `Skip` option to skip processing of given tasks.
- Add `Flow.Usage`, `Flow.SetUsage` for getting and setting the function
  that is called when an error occurs while parsing the flow.
- Add `NOOP` status report for tasks that were intentionally not run
  to differentiate from being skipped during execution.
- Add `Flow.Use` method o support task run interception using middlewares.
- Add `middleware` package with
  `ReportStatus`, `SilentNonFailed`, `DryRun`, `ReportLongRun` middlewares.
- `TF.Error`, `TF.Errorf`, `TF.Fail` may be called simultaneously from multiple goroutines.
- Add `NewRunner` which can be useful for testing and debugging
  task actions and middlewares.
- Add `Flow.Undefine` to unregister a task.
- Add `DefinedTask.SetName`, `DefinedTask.SetUsage`, `DefinedTask.Action`,
  `DefinedTask.SetAction`, `DefinedTask.SetAction`, `DefinedTask.SetDeps`
  to enable modifying the task after the initial definition.
- Add `Flow.SetLogger` for setting a custom log decorator
  that is used by `A` logging methods.
  If `Logger` implements
  `Error`, `Errorf`, `Fatal`, `Fatalf`, `Skip`, `Skipf`,
  they will be used by the `A` equivalent methods.
- Add `Flow.Logger` for getting the log decorator (`CodeLineLogger` by default).
- Add `CodeLineLogger` which is the default for `Flow.Logger`.
- Add `FmtLogger` which is the default when using `NewRunner`.
- Add `A.Helper` and `CodeLineLogger.Helper` methods that work like
  the equivalent method in the `testing` package.
- Add `A.Cleanup` method that registers an action's cleanup function.
- Add `A.Setenv` method that sets the environment variable
  and reverts the previous value during cleanup.
- Add `A.TempDir` method that creates a temporary directory
  and removes it during cleanup.

### Changed

- Rename `TF` type to `A` as this is the `Action` parameter.
  This is the most impactful breaking change.
  This follows the convention used in the [`testing`](https://pkg.go.dev/testing)
  package to name the parameter type as the first letter of the "function type".
- Task status reporting is disabled by default.
  It can be enabled by calling `Flow.Use(middleware.ReportStatus)`.
  It reports `NOOP` for a task without an action.
- Usually, the task does not call `panic` directly.
  `panic` failure message no longer contains a prefix with file and line information.
  The stack trace is printed instead. The behavior is based on `testing` package.
- `RegisteredTask.Deps` returns `[]string` (dependency names) for easier introspection.
- Rename `Flow.Register` to `Flow.Define`.
- Change `Flow.RegisteredTask` to `Flow.DefinedTask`.
- `DefinedTask.Deps` returns `Deps` to facilitate reusing defined task's dependencies
  when creating a new one or redefining an existing one.
- Change the `Flow.DefaultTask` field to `Flow.SetDefault` and `Flow.Default` methods.
- Change `Flow.Output` field to `Flow.SetOutput` setter and `Flow.Output` getter.
- Change `Flow.Run` to `Flow.Execute` to reduce possible confusion with `Runner`.
- `Flow.Execute` returns an error instead of returning the exit code
  and does not print to output. It also has different arguments.
- `Flow.Execute` accepts `[]string` instead of `...string` to make the API
  forward compatible.

### Removed

- Remove parameters API and out-of-the-box flags (`-v`, `-wd`).
- Remove `A.Cmd`.
  Use [`github.com/goyek/goyek/x/cmd`](https://pkg.go.dev/github.com/goyek/x/cmd)
  or your helper instead.

### Fixed

- Fix panic handling so that `panic(nil)` and `runtime.Goexit()` now cause task failure.

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
