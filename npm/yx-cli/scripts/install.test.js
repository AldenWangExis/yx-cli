"use strict";

const assert = require("node:assert/strict");
const pkg = require("../package.json");
const { assetNameForPlatform } = require("./platform");

assert.equal(pkg.name, "@aldenwangexis/yx-cli");
assert.equal(pkg.bin.yx, "bin/yx.js");
assert.equal(pkg.publishConfig.access, "public");
assert.equal(assetNameForPlatform("darwin", "arm64"), "yx-darwin-arm64");
assert.equal(assetNameForPlatform("linux", "x64"), "yx-linux-amd64");
assert.equal(assetNameForPlatform("linux", "arm64"), "yx-linux-arm64");
assert.equal(assetNameForPlatform("win32", "x64"), "yx-windows-amd64.exe");
assert.throws(() => assetNameForPlatform("darwin", "x64"), /Unsupported platform/);

console.log("npm wrapper tests passed");
