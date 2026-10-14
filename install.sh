#!/usr/bin/env bash
set -euo pipefail

for tool in go pkg-config; do
  if ! command -v "$tool" >/dev/null 2>&1; then
    printf 'Error: required command "%s" is not installed or not in PATH.\n' "$tool" >&2
    exit 1
  fi
done

if command -v pacman >/dev/null 2>&1; then
  if ! command -v sudo >/dev/null 2>&1; then
    printf 'Error: sudo is required to install Arch Linux dependencies.\n' >&2
    exit 1
  fi
  sudo pacman -S --needed gtk3 gtk-layer-shell playerctl yt-dlp aspell hunspell
else
  printf 'Warning: pacman not found. Install GTK3, gtk-layer-shell, playerctl, yt-dlp, aspell, and hunspell with your distribution package manager before building.\n' >&2
fi

go build -mod=mod -buildvcs=false -o edwin ./cmd/edwin/

install_dir="${HOME:?HOME is not set}/.local/bin"
mkdir -p "$install_dir"
install -m 0755 edwin "$install_dir/edwin"

in_path=false
IFS=: read -r -a path_entries <<< "${PATH:-}"
for path_entry in "${path_entries[@]}"; do
  if [[ "${path_entry%/}" == "$install_dir" ]]; then
    in_path=true
    break
  fi
done
if [[ "$in_path" != true ]]; then
  printf 'Warning: %s is not in PATH. Add it to PATH to run edwin by name.\n' "$install_dir" >&2
fi

printf 'Installed Edwin to %s/edwin\n' "$install_dir"
