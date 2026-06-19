"use strict";

function assetNameForPlatform(platform, arch) {
  if (platform === "darwin" && arch === "arm64") {
    return "yx-darwin-arm64";
  }
  if (platform === "linux" && arch === "x64") {
    return "yx-linux-amd64";
  }
  if (platform === "linux" && arch === "arm64") {
    return "yx-linux-arm64";
  }
  if (platform === "win32" && arch === "x64") {
    return "yx-windows-amd64.exe";
  }
  throw new Error(`Unsupported platform for yx: ${platform}/${arch}`);
}

module.exports = { assetNameForPlatform };
