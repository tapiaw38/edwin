# Contributing to Edwin

Thanks for your interest in Edwin, an app launcher for Linux running on Wayland.

## Before you start

Edwin is a GTK3 desktop application. It needs a running Wayland session and a session D-Bus to launch, so some changes can only be verified by hand on a real desktop. Mention in your pull request how you tested.

Open an issue before starting significant work, so we can agree on the approach first. Small fixes can go straight to a pull request.

## Development setup

Install the build and runtime dependencies listed in [Dependencies](README.md#dependencies). On Arch:

```bash
sudo pacman -S --needed go git base-devel gcc pkgconf gtk3 gtk-layer-shell gobject-introspection dbus xdg-utils playerctl yt-dlp aspell hunspell
```

On other distributions, install the equivalent packages. The names differ, but you need at least: Go 1.26.1 or newer, a C toolchain, `pkg-config`, GTK3 development headers, gtk-layer-shell development headers, and GObject introspection. CI builds on Arch, so an Arch container is the closest match to the reference environment.

Build and run:

```bash
go build -mod=mod -buildvcs=false -o edwin ./cmd/edwin/
./edwin
```

See [Debug](README.md#debug) for reading GTK errors and checking integrations.

## Checks

Run these before opening a pull request. CI runs the same ones.

```bash
gofmt -l .          # must print nothing
go vet ./...
go test -mod=mod ./...
```

## Pull requests

1. Branch from `main`: `feature/short-name` or `fix/short-name`.
2. Use [Conventional Commits](https://www.conventionalcommits.org/) with the prefixes already used in this repository: `feat:`, `fix:`, `docs:`, `ci:`, `chore:`, `refactor:`. State what changed; explain why only when the diff does not.
3. Do not add comments to `.go` files. The codebase keeps them out on purpose.
4. Add an entry under `[Unreleased]` in [CHANGELOG.md](CHANGELOG.md) for anything user-facing.
5. Update the README when you change behaviour, config keys, or prefixes.
6. Open the pull request against `main` and fill in the template.

`main` is protected. Every change lands through a pull request with the `build` check passing and one approval, and is merged with squash. The branch is deleted automatically after merge.

## Adding a module

Edwin's features are modules behind prefixes. [Adding a new module](README.md#adding-a-new-module) walks through the registry, the `Result` struct, and where to hook a new prefix in.

## Releases

Releases are cut by the maintainer. Do not bump versions or create tags in a pull request.

## Code of conduct

This project follows the [Code of Conduct](CODE_OF_CONDUCT.md). By taking part, you agree to uphold it.
