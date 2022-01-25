# Releasing

## Pre-release

Create a pull request named `Release v<version>` that does the following:

1. Review all places where the current version is used.
   - Update `go.mod` for submodules to depend on the new release.
1. Remove the changed API warning in [`README.md`](./README.md) if it is present.
1. Add documentation or examples if it they are missing.
1. Update [`CHANGELOG.md`](./CHANGELOG.md).
   - Change the `Unreleased` header to represent the new release.
   - Consider adding a description for the new release.
     Especially if it adds new features or introduces breaking changes.
   - Add a new `Unreleased` header above the new release, with no details.

## Release

Create a GitHib Release named `<version>` with `v<version>` tag.

The release description should include all the release notes
from the [`CHANGELOG.md`](./CHANGELOG.md) for this release.
