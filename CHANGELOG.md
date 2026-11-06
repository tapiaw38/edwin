# Changelog

All notable changes to Edwin are documented here.

This project follows [Semantic Versioning](https://semver.org/) and the [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format.

## [Unreleased]

## [0.8.0] - 2026-11-06

### Added

- Contextual search icons that identify launcher modes including file navigation, music, weather, web, terminal, and system commands.
- `paper` color theme and configurable `font_family` typography setting.

### Changed

- Refined the Edwin theme to a warm charcoal background and redesigned result selection, keyboard shortcut badges, and search focus styling.

## [0.7.0] - 2026-11-06

### Added

- `border_width` and `border_color` config fields for a border around the launcher window. `border_width` defaults to `0`, and `border_color` falls back to `focus_border_color`. Colors accept an `R, G, B` triplet, a `#rrggbb` value, or a CSS `rgb()`/`rgba()` function.
- `input_border_width`, `input_border_color`, `input_border_radius`, and `row_border_radius` config fields for styling the search input and result rows independently.
- **Reload** item in the tray menu, which re-reads `config.yaml` and re-applies styling and window width without restarting Edwin. An invalid file is reported on stderr and leaves the running config untouched.

## [0.6.1] - 2026-11-03

### Fixed

- Reload the installed application list every time the launcher is shown, so apps installed or removed while Edwin runs in the background appear without restarting it. Both the hotkey and the tray icon refresh the list.

## [0.6.0] - 2026-11-02

### Added

- `--background` flag to start Edwin with the launcher window hidden, for autostarting it at login so the first hotkey press activates an already running instance.

### Fixed

- Preload icons on the GTK main loop instead of a goroutine, removing the `GTK_IS_LIST_BOX_ROW` assertion failures logged at startup.

### Changed

- Build release binaries with `-ldflags="-s -w"`, reducing the published `edwin` asset from 72 MB to 47 MB.

## [0.5.0] - 2026-10-26

### Added

- Show up to 12 recently used apps when Down is pressed with an empty search input, using most-used apps as a fallback for older history entries.

## [0.4.3] - 2026-10-26

### Fixed

- Stopped passing an unsupported `-e` flag to `kitty` and `foot` when running shell commands, which made the terminal fail to launch instead of executing the command.
- Added user binary directories to terminal command `PATH` so compositor-launched Edwin can find locally installed executables.
- Kept terminal output visible with the command exit status until Enter is pressed.
- Restored keyboard focus to confirmation dialogs by opening them as overlay layer-shell surfaces on the launcher's monitor.

## [0.4.2] - 2026-10-21

### Fixed

- Trigger Quick Look only when Shift is tapped by itself, preventing shifted letters and symbols from opening the preview while typing.

## [0.4.1] - 2026-10-20

### Fixed

- Let the compositor place the launcher on the focused output, avoiding Wayland's unavailable global pointer coordinates and the primary-monitor fallback.
- Applied estimated horizontal and vertical margins before mapping, then corrected only when the mapped monitor geometry differs; cached the collapsed window height to avoid visible positioning jumps.

## [0.4.0] - 2026-10-20

### Added

- Added a 300px right sidebar shared by file Quick Look and player previews, with a fade-in when shown.
- Kept the search input and results in a fixed-width left column so opening the sidebar expands the window to the right without shifting the input.
- Anchored the launcher to its initial left position for all configured vertical positions.

### Changed

- Moved player controls into the left results column and placed cover art and track metadata in the sidebar.

### Fixed

- Prevented the launcher from sliding in from the left by applying its centered left margin before showing it.

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
