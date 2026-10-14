# Agent Guide

## Versioning

Edwin follows Semantic Versioning with `MAJOR.MINOR.PATCH` versions:

- Bump `PATCH` for backward-compatible bug fixes.
- Bump `MINOR` for backward-compatible features.
- Bump `MAJOR` for breaking changes.

Do not choose or cut a release without an explicit request from the maintainer.

## Releases

When the maintainer requests a release:

1. Update `CHANGELOG.md`: move entries from `[Unreleased]` into a new `X.Y.Z` section with the release date, then leave a fresh empty `[Unreleased]` section above it.
2. Commit the changelog update.
3. Create and push the version tag and `main`:

   ```sh
   git tag vX.Y.Z
   git push origin main --tags
   ```

   To push only that tag, use `git push origin vX.Y.Z` instead.

Pushing a `v*` tag triggers `.github/workflows/release.yml`, which builds and publishes the `edwin` asset. No separate release action is needed.

## Branches

Small, iterative changes may be made directly on `main`, matching the repository's usual workflow. For large or risky features, use a branch such as `feature/short-name` or `fix/short-name`, then merge it into `main` when ready.

## Commits

Use short English Conventional Commit messages with prefixes already used in this repository, such as `feat:`, `fix:`, `docs:`, `ci:`, `chore:`, and `refactor:`. State what changed briefly; explain why only when the reason is not clear from the diff.

## CI

The CI workflow ignores Markdown files and `docs/**` for both pushes and pull requests. Documentation-only changes do not run the full build and test job. The release workflow runs only when a tag matching `v*` is pushed.

## Go Code

Do not add comments to `.go` files.
