#!/usr/bin/env bash
set -eu

if [ -n "${BASH_VERSION:-}" ]; then
  set -o pipefail
fi

for tool in curl awk; do
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
  sudo pacman -S --needed --noconfirm gtk3 gtk-layer-shell playerctl yt-dlp aspell hunspell
else
  printf 'Warning: pacman not found. Install GTK3, gtk-layer-shell, playerctl, yt-dlp, aspell, and hunspell with your distribution package manager before running Edwin.\n' >&2
fi

temp_dir=$(mktemp -d)
cleanup() {
  rm -rf "$temp_dir"
}
trap cleanup EXIT

api_url=https://api.github.com/repos/tapiaw38/edwin/releases/latest
curl -fsSL -o "$temp_dir/release.json" "$api_url"
download_url=$(awk '
  /"assets": \[/ { in_assets=1; next }
  in_assets && /"name": "edwin"/ { matched=1; next }
  in_assets && matched && /"browser_download_url":/ {
    sub(/^[[:space:]]*"browser_download_url":[[:space:]]*"/, "")
    sub(/".*$/, "")
    print
    exit
  }
  in_assets && /^[[:space:]]*\]/ { exit }
' "$temp_dir/release.json")

if [ -z "$download_url" ]; then
  printf 'Error: latest GitHub release has no asset named "edwin".\n' >&2
  exit 1
fi

curl -fL -o "$temp_dir/edwin" "$download_url"
chmod 0755 "$temp_dir/edwin"

install_dir="${HOME:?HOME is not set}/.local/bin"
mkdir -p "$install_dir"
mv "$temp_dir/edwin" "$install_dir/edwin"

case ":${PATH:-}:" in
  *":$install_dir:"*) ;;
  *) printf 'Warning: %s is not in PATH. Add it to PATH to run edwin by name.\n' "$install_dir" >&2 ;;
esac

printf 'Installed Edwin to %s/edwin\n' "$install_dir"
