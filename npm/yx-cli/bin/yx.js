#!/usr/bin/env node
"use strict";

const { spawnSync } = require("node:child_process");
const fs = require("node:fs");
const path = require("node:path");
const { platformPackageFor } = require("../scripts/packages");

const pkg = require("../package.json");
const platformPackage = platformPackageFor(process.platform, process.arch);
let packageRoot;

try {
  packageRoot = path.dirname(require.resolve(`${platformPackage.name}/package.json`));
} catch (error) {
  const localPackageRoot = path.join(__dirname, "..", "..", platformPackage.directory);
  if (fs.existsSync(path.join(localPackageRoot, "package.json"))) {
    packageRoot = localPackageRoot;
  } else {
    console.error(`yx npm wrapper could not find platform package ${platformPackage.name}.`);
    console.error("Try reinstalling: npm install -g @aldenwangexis/yx-cli");
    process.exit(1);
  }
}

const binaryPath = path.join(packageRoot, "bin", platformPackage.binary);

if (!fs.existsSync(binaryPath)) {
  console.error(`yx npm wrapper could not find installed binary at ${binaryPath}`);
  console.error("Try reinstalling: npm install -g @aldenwangexis/yx-cli");
  process.exit(1);
}

const result = spawnSync(binaryPath, process.argv.slice(2), {
  stdio: "inherit",
  env: {
    ...process.env,
    YX_INSTALL_CHANNEL: "npm",
    YX_NPM_PACKAGE: pkg.name,
    YX_NPM_PLATFORM_PACKAGE: platformPackage.name,
  },
});

if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}

process.exit(result.status === null ? 1 : result.status);
