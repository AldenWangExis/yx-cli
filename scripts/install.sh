#!/usr/bin/env sh
set -eu

repo="${YX_INSTALL_REPO:-AldenWangExis/yx-cli}"
version="${YX_INSTALL_VERSION:-}"
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
	linux) os="linux" ;;
	*) printf 'yx install: unsupported OS: %s\n' "$os" >&2; exit 1 ;;
	esac

	case "$arch" in
	x86_64|amd64) arch="amd64" ;;
	arm64|aarch64) arch="arm64" ;;
	*) printf 'yx install: unsupported architecture: %s\n' "$arch" >&2; exit 1 ;;
	esac

	printf 'yx-%s-%s' "$os" "$arch"
}

download_url() {
	asset="$1"
	if [ -n "$version" ]; then
		printf 'https://github.com/%s/releases/download/%s/%s' "$repo" "$version" "$asset"
	else
		printf 'https://github.com/%s/releases/latest/download/%s' "$repo" "$asset"
	fi
}

release_label() {
	if [ -n "$version" ]; then
		printf '%s' "$version"
	else
		printf 'latest'
	fi
}

profile_path() {
	case "${SHELL:-}" in
	*/zsh) printf '%s/.zshrc' "$HOME" ;;
	*/bash) printf '%s/.bashrc' "$HOME" ;;
	*) printf '%s/.profile' "$HOME" ;;
	esac
}

path_entry() {
	if [ "$install_dir" = "$HOME/.local/bin" ]; then
		printf 'export PATH="$HOME/.local/bin:$PATH"'
	else
		printf 'export PATH="%s:$PATH"' "$install_dir"
	fi
}

ensure_path() {
	case ":$PATH:" in
	*":$install_dir:"*)
		printf '%s is already on PATH. You can run yx now.\n' "$install_dir"
		return 0
		;;
	esac

	profile="$(profile_path)"
	entry="$(path_entry)"
	if [ -f "$profile" ] && grep -F "$entry" "$profile" >/dev/null 2>&1; then
		printf '%s is already configured in %s.\n' "$install_dir" "$profile"
		printf 'Restart your shell, open a new terminal, or run: . %s\n' "$profile"
		return 0
	fi

	mkdir -p "$(dirname "$profile")"
	{
		printf '\n'
		printf '# yx-cli\n'
		printf '%s\n' "$entry"
	} >>"$profile"
	printf 'Added %s to PATH in %s.\n' "$install_dir" "$profile"
	printf 'Restart your shell, open a new terminal, or run: . %s\n' "$profile"
}

need curl
need uname
need tr
need chmod
need mkdir
need mv
need rm
need grep
need dirname

asset="${YX_INSTALL_ASSET:-$(detect_asset)}"
url="$(download_url "$asset")"
tmp_dir="${TMPDIR:-/tmp}/yx-install.$$"
tmp="$tmp_dir/$asset"
target="$install_dir/yx"

cleanup() {
	rm -rf "$tmp_dir"
}
trap cleanup EXIT INT TERM

mkdir -p "$install_dir"
mkdir -p "$tmp_dir"
printf 'Downloading %s from %s release %s\n' "$asset" "$repo" "$(release_label)"
curl -fL -o "$tmp" "$url"
chmod +x "$tmp"
mv "$tmp" "$target"

printf 'Installed yx to %s\n' "$target"
ensure_path
