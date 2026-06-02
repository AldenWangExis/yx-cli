#!/usr/bin/env sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"

assert_contains() {
	haystack="$1"
	needle="$2"
	if ! printf '%s' "$haystack" | grep -F -- "$needle" >/dev/null 2>&1; then
		printf 'expected to find: %s\nin: %s\n' "$needle" "$haystack" >&2
		exit 1
	fi
}

run_install() {
	version="$1"
	expected_url="$2"

	tmp_dir="$(mktemp -d)"
	fake_bin="$tmp_dir/bin"
	mkdir -p "$fake_bin"

	cat >"$fake_bin/curl" <<'SH'
#!/usr/bin/env sh
set -eu
printf '%s\n' "$*" >"$YX_TEST_CURL_ARGS"
out=""
while [ "$#" -gt 0 ]; do
	if [ "$1" = "-o" ]; then
		shift
		out="$1"
	fi
	shift || true
done
if [ -z "$out" ]; then
	printf 'missing curl -o target\n' >&2
	exit 2
fi
printf '#!/usr/bin/env sh\nprintf "yx test binary\\n"\n' >"$out"
SH
	chmod +x "$fake_bin/curl"

	cat >"$fake_bin/gh" <<'SH'
#!/usr/bin/env sh
set -eu
if [ "${1:-}" = "auth" ] && [ "${2:-}" = "status" ]; then
	exit 0
fi
printf 'install test should use curl download progress, not gh release download\n' >&2
exit 9
SH
	chmod +x "$fake_bin/gh"

	if [ -n "$version" ]; then
		YX_INSTALL_VERSION="$version" \
			YX_INSTALL_ASSET="yx-darwin-arm64" \
			YX_INSTALL_DIR="$tmp_dir/install" \
			YX_TEST_CURL_ARGS="$tmp_dir/curl.args" \
			HOME="$tmp_dir/home" \
			SHELL="/bin/zsh" \
			PATH="$fake_bin:/usr/bin:/bin" \
			sh "$ROOT_DIR/scripts/install.sh" >"$tmp_dir/stdout"
	else
		YX_INSTALL_ASSET="yx-darwin-arm64" \
			YX_INSTALL_DIR="$tmp_dir/install" \
			YX_TEST_CURL_ARGS="$tmp_dir/curl.args" \
			HOME="$tmp_dir/home" \
			SHELL="/bin/zsh" \
			PATH="$fake_bin:/usr/bin:/bin" \
			sh "$ROOT_DIR/scripts/install.sh" >"$tmp_dir/stdout"
	fi

	curl_args="$(cat "$tmp_dir/curl.args")"
	assert_contains "$curl_args" "$expected_url"
	assert_contains "$curl_args" "-fL"
	test -x "$tmp_dir/install/yx"
	rm -rf "$tmp_dir"
}

run_install "" "https://github.com/AldenWangExis/yx-cli/releases/latest/download/yx-darwin-arm64"
run_install "v1.0.0" "https://github.com/AldenWangExis/yx-cli/releases/download/v1.0.0/yx-darwin-arm64"

printf 'install tests passed\n'
