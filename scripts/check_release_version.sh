#!/bin/sh
set -eu

tag="${1:-${GITHUB_REF_NAME:-}}"

if [ -z "$tag" ]; then
	echo "release version check skipped: no tag provided"
	exit 0
fi

case "$tag" in
v[0-9]*.[0-9]*.[0-9]*)
	;;
*)
	echo "release version check skipped: $tag is not a vX.Y.Z tag"
	exit 0
	;;
esac

package_version="$(node -p "require('./npm/yx-cli/package.json').version")"
expected_tag="v${package_version}"

if [ "$tag" != "$expected_tag" ]; then
	echo "release version mismatch: tag is $tag but npm/yx-cli/package.json is $package_version"
	echo "Set npm package version to ${tag#v} before publishing this tag."
	exit 1
fi

echo "release version check passed: $tag matches npm package version $package_version"
