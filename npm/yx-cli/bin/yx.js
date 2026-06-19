#!/usr/bin/env node
"use strict";

const { spawnSync } = require("node:child_process");
const fs = require("node:fs");
const path = require("node:path");
const { assetNameForPlatform } = require("../scripts/platform");

const pkg = require("../package.json");
const assetName = assetNameForPlatform(process.platform, process.arch);
const binaryName = process.platform === "win32" ? "yx.exe" : "yx";
const binaryPath = path.join(__dirname, "..", "vendor", binaryName);

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
    YX_NPM_ASSET: assetName,
    YX_NPM_PACKAGE: pkg.name,
  },
});

if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}

process.exit(result.status === null ? 1 : result.status);
