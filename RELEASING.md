# Releasing

## Pre-release

Create a pull request named `Release v<version>` that does the following:

1. Update `go.mod` for submodules to depend on the new release.
1. Update the [`README.md`](./README.md).
1. Update the [`CHANGELOG.md`](./CHANGELOG.md).

The pull request description should include all the release notes from the [Changelog](./CHANGELOG.md) for this release.

## Release

Create a GitHib Release named `<version>` with `v<version>` tag.

The release description should include all the release notes from the [Changelog](./CHANGELOG.md) for this release.
