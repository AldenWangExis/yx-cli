#!/usr/bin/env sh
set -eu

repo="${YX_INSTALL_REPO:-AldenWangExis/yx-cli}"
version="${YX_INSTALL_VERSION:-latest}"
install_dir="${YX_INSTALL_DIR:-$HOME/.local/bin}"

need() {
	if ! command -v "$1" >/dev/null 2>&1; then
		printf 'yx install: missing required command: %s\n' "$1" >&2
		exit 1
	fi
}

detect_asset() {
	os="$(uname -s | tr '[:upper:]' '[:lower:]')"
	arch="$(uname -m)"

	case "$os" in
	darwin) os="darwin" ;;
	*) printf 'yx install: unsupported OS: %s\n' "$os" >&2; exit 1 ;;
	esac

	case "$arch" in
	arm64|aarch64) arch="arm64" ;;
	*) printf 'yx install: unsupported architecture: %s\n' "$arch" >&2; exit 1 ;;
	esac

	printf 'yx-%s-%s' "$os" "$arch"
}

download_url() {
	asset="$1"
	if [ "$version" = "latest" ]; then
		printf 'https://github.com/%s/releases/latest/download/%s' "$repo" "$asset"
	else
		printf 'https://github.com/%s/releases/download/%s/%s' "$repo" "$version" "$asset"
	fi
}

need curl
need uname
need tr
need chmod
need mkdir
need mv
need rm

asset="${YX_INSTALL_ASSET:-$(detect_asset)}"
url="$(download_url "$asset")"
tmp="${TMPDIR:-/tmp}/yx-install.$$"
target="$install_dir/yx"

cleanup() {
	rm -f "$tmp"
}
trap cleanup EXIT INT TERM

mkdir -p "$install_dir"
printf 'Downloading %s\n' "$url"
curl -fsSL -o "$tmp" "$url"
chmod +x "$tmp"
mv "$tmp" "$target"

printf 'Installed yx to %s\n' "$target"
case ":$PATH:" in
*":$install_dir:"*) ;;
*) printf 'Add %s to PATH before running yx globally.\n' "$install_dir" ;;
esac
