"use strict";

const assert = require("node:assert/strict");
const pkg = require("../package.json");
const { mainPackageName, platformPackages, platformPackageFor } = require("./packages");

assert.equal(pkg.name, mainPackageName);
assert.equal(pkg.bin.yx, "bin/yx.js");
assert.equal(pkg.publishConfig.access, "public");

for (const platformPackage of platformPackages) {
  assert.equal(
    pkg.optionalDependencies[platformPackage.name],
    pkg.version,
    `${platformPackage.name} must be pinned to the main package version`,
  );
}

assert.equal(platformPackageFor("darwin", "arm64").asset, "yx-darwin-arm64");
assert.equal(platformPackageFor("linux", "x64").asset, "yx-linux-amd64");
assert.equal(platformPackageFor("linux", "arm64").asset, "yx-linux-arm64");
assert.equal(platformPackageFor("win32", "x64").asset, "yx-windows-amd64.exe");
assert.throws(() => platformPackageFor("darwin", "x64"), /Unsupported platform/);

console.log("npm wrapper tests passed");
