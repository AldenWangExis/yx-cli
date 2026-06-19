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

node <<'NODE'
const assert = require("node:assert/strict");
const path = require("node:path");
const mainPkg = require("./npm/yx-cli/package.json");
const { platformPackages } = require("./npm/yx-cli/scripts/packages");

for (const platformPackage of platformPackages) {
  const pkg = require(path.join(process.cwd(), "npm", platformPackage.directory, "package.json"));
  assert.equal(pkg.version, mainPkg.version, `${pkg.name}: version must match ${mainPkg.version}`);
  assert.equal(
    mainPkg.optionalDependencies[pkg.name],
    mainPkg.version,
    `${pkg.name}: optionalDependency must be pinned to ${mainPkg.version}`,
  );
}
NODE

echo "release version check passed: $tag matches all npm package versions"
