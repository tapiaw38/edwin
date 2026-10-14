#!/usr/bin/env bash
set -eu

if [ -n "${BASH_VERSION:-}" ]; then
  set -o pipefail
fi

for tool in go pkg-config git; do
  if ! command -v "$tool" >/dev/null 2>&1; then
    printf 'Error: required command "%s" is not installed or not in PATH.\n' "$tool" >&2
    exit 1
  fi
done

temp_dir=
cleanup() {
  if [ -n "$temp_dir" ]; then
    rm -rf "$temp_dir"
  fi
}
trap cleanup EXIT

if ! grep -qx 'module github.com/tapiaw38/edwin' go.mod 2>/dev/null; then
  temp_dir=$(mktemp -d)
  git clone --depth 1 https://github.com/tapiaw38/edwin.git "$temp_dir/edwin"
  cd "$temp_dir/edwin"
fi

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

case ":${PATH:-}:" in
  *":$install_dir:"*) ;;
  *) printf 'Warning: %s is not in PATH. Add it to PATH to run edwin by name.\n' "$install_dir" >&2 ;;
esac

printf 'Installed Edwin to %s/edwin\n' "$install_dir"
