![Edwin logo](cmd/edwin/assets/edwin-icon.png)

# Edwin

App launcher for Linux (Wayland/MangoWM).

## Stack

- **Language:** Go
- **UI:** GTK3 (gotk4 bindings)
- **Wayland:** gtk-layer-shell
- **System tray:** StatusNotifierItem over D-Bus
- **Config:** YAML (~/.config/edwin/config.yaml)

## Build

```bash
go build -mod=mod -buildvcs=false -o edwin ./cmd/edwin/
```

Install directly with:

```bash
curl -fsSL https://raw.githubusercontent.com/tapiaw38/edwin/main/install.sh | sh
```

Or clone the repository and run the installer locally:

```bash
git clone https://github.com/tapiaw38/edwin.git
cd edwin
./install.sh
```

The installer checks for `go`, `pkg-config`, and `git`. Outside an Edwin clone, it clones the latest source into a temporary directory and removes it on exit. On Arch, it installs `gtk3`, `gtk-layer-shell`, `playerctl`, `yt-dlp`, `aspell`, and `hunspell` with `sudo pacman`, then builds Edwin and installs it to `~/.local/bin/edwin`. On other distributions, install equivalent system libraries manually; the installer continues with the build. Add `~/.local/bin` to `PATH` to run `edwin` by name.

## Development

Format:

```bash
gofmt -w $(find . -name '*.go' -not -path './vendor/*')
```

Test:

```bash
go test -mod=mod ./...
go vet ./...
```

## Run

```bash
./edwin
```

## System Tray

Edwin registers a StatusNotifierItem on the session D-Bus while running. Left-click the tray icon to show or hide the launcher; right-click and choose **Quit** to exit. Closing the launcher window hides it while Edwin keeps running in the tray. `Esc` leaves file-action mode first, then hides the launcher. Launching Edwin again activates the existing instance over D-Bus and brings its window forward.

## Setup Hotkey

```bash
./edwin --setup  # Updates ~/.config/mango/bind.conf with hotkey from config
```

## Architecture

Edwin uses a lightweight hexagonal style:

- `internal/modules` owns search logic and returns declarative `Result` values.
- `Result.ActionSpec` describes what should happen when the user presses Enter.
- `cmd/edwin/action_executor.go` is the UI adapter that executes actions against the desktop/session.
- `internal/platform/commands` is the only package that wraps `os/exec`.

Modules should prefer `ActionSpec` helpers such as `OpenAction`, `CopyAction`, `TerminalAction`, `EmailAction`, `FileAction`, `MusicAction`, `SyncAction`, and `SystemAction` instead of running shell commands directly.

## Project Structure

```bash
edwin/
├── LICENSE, README.md, go.mod, go.sum
├── .github/workflows/ci.yml
├── cmd/edwin/
│   ├── main.go, launcher.go, single_instance.go
│   ├── action_executor.go, css.go, keys.go, player.go
│   ├── preview.go, results.go, row.go, search.go, state.go
│   ├── window_email.go, window_fileop.go
│   ├── window_largetype.go, window_stats.go
│   └── assets/ (edwin-icon.png, edwin-tray.png)
├── third_party/gogpu-systray/ # Local SNI implementation with ItemIsMenu disabled
├── internal/
│   ├── apps/apps.go
│   ├── config/ (config.go, paths.go, theme.go)
│   ├── history/history.go
│   ├── platform/commands/commands.go
│   └── modules/
│       ├── modules.go, registry.go, query.go, limits.go, timeouts.go
│       ├── bookmarks.go, calc.go, contacts.go, devtools.go
│       ├── dictionary.go, email.go, emoji.go, files.go
│       ├── file_actions.go, file_ops.go, navigation.go
│       ├── clipboard.go, snippets.go, spell.go, recent.go, help.go
│       ├── large_type.go, stats.go, status.go, sync.go, system.go, text.go
│       ├── kill.go, pass.go, ssh.go, screenshot.go, timer.go, units.go
│       ├── weather.go, web.go, window.go
│       └── music.go, spotify.go, youtube.go, chrome_session.go, firefox_session.go
```

## Features & Prefixes

| Prefix            | Feature                       | Example                                                        |
| ----------------- | ----------------------------- | -------------------------------------------------------------- |
| (none)            | App search, calculator        | `firefox`, `2+2`                                                |
| `>`               | Shell command                 | `> htop`                                                       |
| `;`/`snip`        | Snippet                       | `;email`, `snip email`                                          |
| `f`/`file`        | File search                   | `f readme`                                                     |
| `Tab`             | File actions                  | select file, press `Tab`                                       |
| file op window    | Visual file ops               | Tab -> Rename/Copy/Move                                        |
| `nav`/`browse`    | Folder navigation             | `nav ~/Downloads`                                              |
| `pick`            | Destination picker            | `pick copy source \| ~/Downloads`                              |
| `rename`          | Rename file                   | `rename source \| new-name`                                    |
| `copy`            | Copy file                     | `copy source \| target`                                        |
| `move`            | Move file                     | `move source \| target`                                        |
| `undo`            | Undo last file operation      | `undo`                                                         |
| `status`          | Last action/error             | `status`                                                       |
| `buffer`/`buf`    | File buffer                   | `buffer`                                                       |
| `recent`          | Recent documents              | `recent invoice`, `recent app firefox`                         |
| `large`/`type`/`lt` | Large Type                  | `large 555-1234`                                               |
| `large all`       | Large Type all monitors       | `large all 555-1234`                                           |
| `help`/`?`        | Help                          | `help`                                                         |
| `contact(s)`      | Contacts                      | `contact Ada`, `contacts`                                      |
| `email`/`mail`    | Email composer + attachments  | `email contact \| Subject \| Body`                             |
| `stats`/`usage`   | Usage Stats                   | `stats`                                                        |
| `sync`/`settings sync` | Sync Settings             | `sync`, `sync import ~/edwin-settings.zip`                     |
| `clip`/`clipboard`/`cb`/`paste` | Clipboard       | `clip`, `cb query`, `paste query`                              |
| `define`/`def`    | Dictionary                    | `define word`                                                  |
| `spell`/`spelling` | Spelling                     | `spell recieve`                                                |
| `sp`/`spotify`    | Spotify controls/status       | `sp`, `sp status`                                              |
| `yp`/`youtube player` | YouTube player controls    | `yp play`, `yp status`                                         |
| `m`/`music`       | Local music                   | `m song`, `m artists`, `m artist name`, `m albums`, `m genres` |
| `mq`/`music queue`/`queue music` | Music queue      | `mq`, `music queue`                                            |
| web shortcuts     | Web and YouTube search        | `g`, `yt`, `gh`, `wiki`, `ddg`, `r`, `so` (default keys)       |
| system commands   | Session/power actions         | `lock`, `sleep`, `hibernate`, `reboot`, `shutdown`, `logout`, `trash` |
| `emoji`           | Emoji picker                  | `emoji fire`                                                   |
| `b64`/`b64d`      | Base64 encode/decode          | `b64 hello`                                                    |
| `url`/`urld`      | URL encode/decode             | `url a b&c`                                                    |
| `hash`            | SHA-256                       | `hash secret`                                                  |
| `uuid`            | Generate UUID v4              | `uuid`                                                         |
| `epoch`           | Unix time / date              | `epoch`, `epoch 1700000000`                                    |
| (units)           | Unit conversion               | `100 km to mi`, `50f to c`, `5 gb to mb`                       |
| `ssh`             | SSH hosts from ~/.ssh/config  | `ssh`, `ssh prod`                                              |
| `kill`            | Kill process                  | `kill firefox`                                                 |
| `screenshot`/`ss` | Screenshot (grim/slurp)       | `ss`                                                           |
| `w`               | Windows or MangoWM workspaces | `w`, `w firefox`, `w 2`                                        |
| `timer`           | Countdown -> notification     | `timer 5m`                                                     |
| `wt`/`weather`    | Weather from wttr.in          | `wt`, `weather Berlin`                                         |
| `pass`            | Password store (pass)         | `pass github`                                                  |
| `bm`              | Browser bookmarks             | `bm docs`                                                      |

Quick Look: press `Shift`; PDF/doc previews show page/zoom controls, and `PageUp/PageDown` plus `+/-` also work.
File ops: `Tab` on a file opens actions. `Add to Buffer`, `Copy Path`, `Rename...`, `Copy To...`, `Move To...`, and `Move to Trash` keep the launcher visible. `Open`, `Reveal in Files`, and `Email File` hide it after the action. Rename/copy/move pickers include breadcrumbs and folder browsing.

## Config

Location: `~/.config/edwin/config.yaml`

Color themes: `edwin` (default), `dark`, `midnight`, `light`, and `classic`.

The `edwin` preset is inspired by the butler in a tuxedo on the app icon (`cmd/edwin/assets/edwin-icon.png`). Preset color values:

| Theme | Background RGB / alpha | Entry background | Focus border RGB | Text | Selection RGB | Description text |
| ----- | --------------------- | ---------------- | ---------------- | ---- | ------------- | ---------------- |
| `dark` | `30, 30, 40` / `0.95` | `rgba(255, 255, 255, 0.1)` | `100, 150, 255` | `white` | `100, 150, 255` | `rgba(255, 255, 255, 0.6)` |
| `midnight` | `8, 8, 12` / `0.97` | `rgba(255, 255, 255, 0.07)` | `40, 55, 95` | `white` | `40, 55, 95` | `rgba(255, 255, 255, 0.5)` |
| `light` | `245, 245, 248` / `0.97` | `rgba(0, 0, 0, 0.06)` | `90, 140, 255` | `#1c1c1e` | `90, 140, 255` | `rgba(0, 0, 0, 0.55)` |
| `classic` | `238, 232, 213` / `0.97` | `rgba(0, 0, 0, 0.05)` | `38, 139, 210` | `#073642` | `38, 139, 210` | `rgba(7, 54, 66, 0.6)` |
| `edwin` | `24, 26, 34` / `0.97` | `rgba(208, 193, 180, 0.2)` | `245, 215, 140` | `#fdf2d6` | `208, 193, 180` | `rgba(253, 242, 214, 0.62)` |

`theme` selects a preset. Individual color fields in `config.yaml` override that preset when specified. `selection_color` is rendered with 0.3 alpha. Every preset provides `focus_border_color`; it sets the 1px border on `#edwin-entry:focus`.

```yaml
width: 600
max_results: 6
theme: edwin
background_color: "24, 26, 34"
background_alpha: 0.97
border_radius: 12
font_size: 18
text_color: "#fdf2d6"
selection_color: "208, 193, 180"
entry_background: "rgba(208, 193, 180, 0.2)"
focus_border_color: "245, 215, 140"
desc_color: "rgba(253, 242, 214, 0.62)"
show_icons: true
icon_size: 24
margin_top: 100
history_boost: 3
hotkey: "SUPER,s"
spell_language: "en"
file_manager: "ghostty -e spf"
web_shortcuts:
  g:
    name: Google
    url: "https://www.google.com/search?q=%s"
    icon: web-browser
```

`file_manager` is optional. Edwin splits it on spaces and appends the folder path when opening a folder or revealing a file; when unset, it uses `xdg-open`.

## Key Files to Modify

### Adding a new module

1. Create `internal/modules/newmodule.go`
2. Implement `func NewModuleSearch(query string) []Result`
3. Return `Result` values with `ActionSpec` for Enter behavior
4. Register the module in `internal/modules/registry.go`

### Result struct (modules/modules.go)

```go
type Result struct {
    Type            ResultType // Module/result identifier
    Title           string     // Main text
    Desc            string     // Secondary text
    Icon            string     // Icon name or path
    IconText        string     // Text fallback for icon display
    Preview         string     // Preview pane text
    PreviewImage    string     // Preview pane image path
    PreviewImageURL string     // Remote image to cache for preview
    Data            string     // Module-specific data
    KeepOpen        bool       // Keep launcher open after action
    Confirm         bool       // Ask before executing action
    NavigateQuery   string     // Replace query instead of executing
    ActionSpec      ActionSpec // Declarative action executed by cmd/edwin
    Action          func()     // Legacy escape hatch; avoid for new modules
}
```

Common action helpers:

```go
OpenAction(pathOrURL)
CopyAction(text)
TerminalAction(command)
EmailAction(to, subject, body, attachments...)
FileAction(op, args...)
MusicAction(op, args...)
SyncAction(op, args...)
SystemAction(op)
```

### Spotify view

Special UI in `cmd/edwin/player.go`:

- `createSpotifyView()` - Build layout
- `showSpotifyView()` / `hideSpotifyView()` - Toggle
- `refreshSpotifyInfo()` - Update track info

Uses `playerctl` for MPRIS metadata and `ActionSpec` for controls.

## Dependencies

Build and runtime packages (Arch):

```bash
pacman -S go git base-devel gcc pkgconf gtk3 gtk-layer-shell gobject-introspection dbus xdg-utils playerctl yt-dlp aspell hunspell
```

Go 1.26.1 or newer is required by `go.mod`. The tray uses Edwin's Go D-Bus dependency and requires an active session bus; no libappindicator package is needed. Optional integrations use tools such as `playerctl`, `yt-dlp`, `aspell`, `hunspell`, `grim`, `slurp`, `cliphist`, `wl-clipboard`, `wtype` or `ydotool`, `mpv`, `pass`, and `libnotify` (`notify-send`). Go modules are declared in `go.mod` and fetched during build.

On Arch, `install.sh` installs `gtk3`, `gtk-layer-shell`, `playerctl`, `yt-dlp`, `aspell`, and `hunspell`. Go and `pkg-config` must already be available; install other packages from the Arch command above as needed.

## Debug

Run directly to see GTK errors:

```bash
./edwin 2>&1
```

Test playerctl:

```bash
playerctl metadata
playerctl status
```

## License

MIT. See [LICENSE](LICENSE).
