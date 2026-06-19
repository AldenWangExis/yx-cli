#!/usr/bin/env node
"use strict";

const fs = require("node:fs");
const { spawnSync } = require("node:child_process");
const https = require("node:https");
const path = require("node:path");
const { assetNameForPlatform } = require("./platform");

const packageRoot = path.join(__dirname, "..");
const pkg = require(path.join(packageRoot, "package.json"));
const assetName = assetNameForPlatform(process.platform, process.arch);
const tag = `v${pkg.version}`;
const repo = process.env.YX_NPM_INSTALL_REPO || "AldenWangExis/yx-cli";
const url = `https://github.com/${repo}/releases/download/${tag}/${assetName}`;
const vendorDir = path.join(packageRoot, "vendor");
const binaryName = process.platform === "win32" ? "yx.exe" : "yx";
const target = path.join(vendorDir, binaryName);

async function main() {
  fs.mkdirSync(vendorDir, { recursive: true });
  console.log(`Downloading ${assetName} from ${repo} release ${tag}`);
  await downloadWithRetries(url, target, 3);
  if (process.platform !== "win32") {
    fs.chmodSync(target, 0o755);
  }
  console.log(`Installed yx npm binary to ${target}`);
}

async function downloadWithRetries(sourceUrl, targetPath, attempts) {
  if (downloadWithCurl(sourceUrl, targetPath)) {
    return;
  }

  let lastError;
  for (let attempt = 1; attempt <= attempts; attempt += 1) {
    try {
      await download(sourceUrl, targetPath);
      return;
    } catch (error) {
      lastError = error;
      if (attempt < attempts) {
        console.log(`Download attempt ${attempt} failed: ${error.message}; retrying...`);
      }
    }
  }
  throw lastError;
}

function downloadWithCurl(sourceUrl, targetPath) {
  const result = spawnSync("curl", ["-fL", "--retry", "3", "--connect-timeout", "15", "-o", targetPath, sourceUrl], {
    stdio: "inherit",
  });
  if (result.error && result.error.code === "ENOENT") {
    return false;
  }
  if (result.status === 0) {
    return true;
  }
  fs.rmSync(targetPath, { force: true });
  console.log("curl download failed; falling back to Node.js downloader");
  return false;
}

function download(sourceUrl, targetPath) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(targetPath, { mode: 0o755 });
    const request = https.get(sourceUrl, response => {
      if (response.statusCode >= 300 && response.statusCode < 400 && response.headers.location) {
        file.close();
        fs.rmSync(targetPath, { force: true });
        download(response.headers.location, targetPath).then(resolve, reject);
        return;
      }
      if (response.statusCode !== 200) {
        file.close();
        fs.rmSync(targetPath, { force: true });
        reject(new Error(`Download failed with status ${response.statusCode}: ${sourceUrl}`));
        return;
      }
      response.pipe(file);
      file.on("finish", () => {
        file.close(resolve);
      });
    });
    request.setTimeout(30000, () => {
      request.destroy(new Error(`Download timed out: ${sourceUrl}`));
    });
    request.on("error", error => {
      file.close();
      fs.rmSync(targetPath, { force: true });
      reject(error);
    });
  });
}

main().catch(error => {
  console.error(`yx npm install failed: ${error.message}`);
  process.exit(1);
});
