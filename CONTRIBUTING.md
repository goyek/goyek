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

Report suspected vulnerabilities according to the
[Security Policy](SECURITY.md), not in a public issue.

## Developing

Run `./goyek.sh` (Bash) or `.\goyek.ps1` (PowerShell)
to execute the build pipeline.

The repository contains basic configuration for
[Visual Studio Code](https://code.visualstudio.com/).

## Releasing

### Pre-release

Create a pull request named `Release <version>` that prepares the release:

- Keep the `Unreleased` section and add a dated section for the new release to
  [`CHANGELOG.md`](CHANGELOG.md).
- Update the changelog comparison links for both `Unreleased` and the release.
- Consider adding a description for the new release.
  Especially if it adds new features or introduces breaking changes.
- Run the complete release validation from a clean working tree:

  ```sh
  ./goyek.sh release-check -v
  ```

  This checks the normal CI pipeline, public API compatibility with the most
  recent reachable local `v3` release tag, known vulnerabilities, and the final
  Git state. Make sure that release tags are available in the local clone; the
  validation does not download the API baseline.
- Confirm that all required GitHub Actions and security checks pass for the
  release commit.

### Release

1. Add and push a signed tag:

   ```sh
   TAG='v<version>'
   COMMIT='<commit-sha>'
   git tag -s -m "$TAG" "$TAG" "$COMMIT"
   git push origin "$TAG"
   ```

1. Create a GitHub Release named `<version>` with the `v<version>` tag.

   The release description should include all the release notes
   from the [`CHANGELOG.md`](CHANGELOG.md) for this release.
