# Contributing

We seek any feedback and are open to contribution.

Feel free to:

- create an [issue](https://github.com/goyek/goyek/issues),
- propose a [pull request](https://github.com/goyek/goyek/pulls).

It would be very helpful if you:

- tell us what is missing in the documentation and examples,
- share your experience report,
- propose features that you find critical or extremely useful,
- share **goyek** with others by writing a blog post,
  giving a speech at a meetup or conference,
  or even telling your colleagues that you work with.

Make sure to be familiar with our [Code of Conduct](CODE_OF_CONDUCT.md).

## Developing

Go and Docker is required.

Run `./goyek.sh` (Bash) or `.\goyek.ps1` (PowerShell)
[wrapper scripts](README.md#wrapper-scripts) to execute the build pipeline.

The repository contains basic confiugration for
[Visual Studio Code](https://code.visualstudio.com/).

## Releasing

### Pre-release

Create a pull request named `Release <version>` that does the following:

1. Review all places where the current version is used:
   - [`README.md`](README.md)
   - [`build/go.mod`](build/go.mod)
1. Remove the changed API warning in [`README.md`](README.md) if it is present.
1. Add documentation or examples if it they are missing.
1. Update [`CHANGELOG.md`](CHANGELOG.md).
   - Change the `Unreleased` header to represent the new release.
   - Consider adding a description for the new release.
     Especially if it adds new features or introduces breaking changes.
   - Add a new `Unreleased` header above the new release, with no details.

### Release

Create a GitHib Release named `<version>` with `v<version>` tag.

The release description should include all the release notes
from the [`CHANGELOG.md`](CHANGELOG.md) for this release.
