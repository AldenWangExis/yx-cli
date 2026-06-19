#!/usr/bin/env node
"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const { platformPackages } = require("./packages");

const repoRoot = path.join(__dirname, "..", "..", "..");
const mainPkg = require(path.join(repoRoot, "npm/yx-cli/package.json"));

for (const platformPackage of platformPackages) {
  const pkgPath = path.join(repoRoot, "npm", platformPackage.directory, "package.json");
  const pkg = require(pkgPath);

  assert.equal(pkg.name, platformPackage.name, `${pkgPath}: package name mismatch`);
  assert.equal(pkg.version, mainPkg.version, `${pkg.name}: version must match main package`);
  assert.deepEqual(pkg.os, [platformPackage.os], `${pkg.name}: os mismatch`);
  assert.deepEqual(pkg.cpu, [platformPackage.cpu], `${pkg.name}: cpu mismatch`);
  assert.deepEqual(pkg.files, ["bin"], `${pkg.name}: files must include only bin`);
  assert.equal(pkg.publishConfig.access, "public", `${pkg.name}: package must publish publicly`);

  const binaryPath = path.join(repoRoot, "npm", platformPackage.directory, "bin", platformPackage.binary);
  if (fs.existsSync(binaryPath) && platformPackage.os !== "win32") {
    const mode = fs.statSync(binaryPath).mode & 0o777;
    assert.equal(mode & 0o111, 0o111, `${binaryPath}: binary must be executable`);
  }
}

console.log("npm platform package checks passed");
