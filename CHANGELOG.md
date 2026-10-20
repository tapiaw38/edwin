# Changelog

All notable changes to Edwin are documented here.

This project follows [Semantic Versioning](https://semver.org/) and the [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format.

## [Unreleased]

## [0.3.0] - 2026-10-20

### Added

- Added Alfred-style numbered shortcuts for running the first nine search results, with visual badges and configurable `shortcut_modifier` (default `ctrl+shift`).

### Fixed

- Prevented modified Shift presses used by numbered shortcuts from toggling Quick Look.

## [0.2.2] - 2026-10-15

### Fixed

- Added `--noconfirm` to Arch dependency installation so `pacman` cannot consume the installer script from stdin when piped to `sh`.

## [0.2.1] - 2026-10-15

### Fixed

- Invalidated pending searches when hiding the launcher to prevent late async results from restoring expanded layout and shifting its next position.

## [0.2.0] - 2026-10-14

### Added

- Added configurable launcher positioning with animated top and bottom slides and a center fade-in.

### Fixed

- Fixed launcher positioning across monitor work areas and changing window states.
- Reset search and collapse result, Spotify, and preview views when hiding the launcher so stale searches cannot affect its next position.

## [0.1.0] - 2026-10-13

### Added

- Renamed the application from Spark to Edwin and added branded launcher and tray icons.
- Added a StatusNotifierItem system tray icon, menu, and D-Bus single-instance activation.
- Added configurable file-manager commands for opening folders and revealing files.
- Added Arch Linux GitHub Actions CI for vetting, building, and testing.
- Added the Edwin color theme as the default, including configurable input focus borders.
- Added tagged GitHub Releases with a precompiled Edwin binary and a standalone installer.

### Changed

- Changed the installer to download the precompiled binary from the latest GitHub Release instead of building from source.
- Kept the launcher visible after `Move to Trash`, `Copy Path`, `Rename`, `Copy To`, and `Move To` actions.
- Skipped CI for documentation-only pushes and pull requests.
- Updated project documentation for current features, configuration, dependencies, and installation.

### Fixed

- Deduplicated installed applications by XDG desktop-file ID.
- Fixed confirmation dialogs that crashed with a nil parent and auxiliary windows that exited the tray process when closed.
- Removed the white halo outside the rounded app icon.
