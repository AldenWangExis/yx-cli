"use strict";

const scope = "@aldenwangexis";
const mainPackageName = `${scope}/yx-cli`;

const platformPackages = [
  {
    directory: "yx-cli-darwin-arm64",
    name: `${scope}/yx-cli-darwin-arm64`,
    os: "darwin",
    cpu: "arm64",
    asset: "yx-darwin-arm64",
    binary: "yx",
  },
  {
    directory: "yx-cli-linux-x64",
    name: `${scope}/yx-cli-linux-x64`,
    os: "linux",
    cpu: "x64",
    asset: "yx-linux-amd64",
    binary: "yx",
  },
  {
    directory: "yx-cli-linux-arm64",
    name: `${scope}/yx-cli-linux-arm64`,
    os: "linux",
    cpu: "arm64",
    asset: "yx-linux-arm64",
    binary: "yx",
  },
  {
    directory: "yx-cli-win32-x64",
    name: `${scope}/yx-cli-win32-x64`,
    os: "win32",
    cpu: "x64",
    asset: "yx-windows-amd64.exe",
    binary: "yx.exe",
  },
];

function platformPackageFor(platform, arch) {
  const match = platformPackages.find(pkg => pkg.os === platform && pkg.cpu === arch);
  if (!match) {
    throw new Error(`Unsupported platform for yx: ${platform}/${arch}`);
  }
  return match;
}

module.exports = { mainPackageName, platformPackages, platformPackageFor };
