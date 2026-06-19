#!/usr/bin/env node
"use strict";

const fs = require("node:fs");
const path = require("node:path");
const { platformPackages } = require("./packages");

const repoRoot = path.join(__dirname, "..", "..", "..");
const distDir = process.argv[2] ? path.resolve(process.argv[2]) : path.join(repoRoot, "dist");

for (const platformPackage of platformPackages) {
  const source = path.join(distDir, platformPackage.asset);
  const targetDir = path.join(repoRoot, "npm", platformPackage.directory, "bin");
  const target = path.join(targetDir, platformPackage.binary);

  if (!fs.existsSync(source)) {
    throw new Error(`Missing release asset for npm package: ${source}`);
  }

  fs.rmSync(targetDir, { recursive: true, force: true });
  fs.mkdirSync(targetDir, { recursive: true });
  fs.copyFileSync(source, target);
  if (platformPackage.os !== "win32") {
    fs.chmodSync(target, 0o755);
  }
  console.log(`Prepared ${platformPackage.name}: ${target}`);
}
