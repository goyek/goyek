# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased](https://github.com/pellared/taskflow/compare/v0.2.0...HEAD)

### Changed

- Drop official support for Go 1.10. ([#44](https://github.com/pellared/taskflow/pull/44))
- Task parameters need to be registered, both with `Taskflow.Configure*()` and `Task.Parameters`.
- Default behaviour of verbosity changed.

## [0.2.0](https://github.com/pellared/taskflow/compare/v0.1.1...v0.2.0) - 2021-03-14

### Added

- Add the possibility to set a default task.

## [0.1.1](https://github.com/pellared/taskflow/compare/v0.1.0...v0.1.1) - 2021-02-28

### Fixed

- Make concurrent printing thread-safe.

## [0.1.0](https://github.com/pellared/taskflow/releases/tag/v0.1.0) - 2021-01-14

### Added

- First release version after the experiential phase.
