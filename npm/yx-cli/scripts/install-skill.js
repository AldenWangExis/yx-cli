#!/usr/bin/env node
"use strict";

const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");

const skipValues = new Set(["1", "true", "yes"]);

function shouldSkip() {
  return skipValues.has(String(process.env.YX_SKIP_SKILL_INSTALL || "").toLowerCase());
}

function homeDir() {
  return process.env.HOME || os.homedir();
}

function installSkill() {
  if (shouldSkip()) {
    return { skipped: true };
  }

  const home = homeDir();
  if (!home) {
    return { skipped: true, reason: "home directory is unavailable" };
  }

  const packageRoot = path.join(__dirname, "..");
  const source = path.join(packageRoot, "skills", "yx-cli", "SKILL.md");
  const targetDir = path.join(home, ".agents", "skills", "yx-cli");
  const target = path.join(targetDir, "SKILL.md");

  const content = fs.readFileSync(source);
  if (fs.existsSync(target) && fs.readFileSync(target).equals(content)) {
    return { installed: false, target };
  }

  fs.mkdirSync(targetDir, { recursive: true, mode: 0o700 });
  fs.writeFileSync(target, content, { mode: 0o644 });
  return { installed: true, target };
}

if (require.main === module) {
  try {
    const result = installSkill();
    if (result.reason) {
      console.warn(`yx skill install skipped: ${result.reason}`);
    } else if (result.installed) {
      console.log(`Installed yx skill: ${result.target}`);
    }
  } catch (error) {
    console.warn(`yx skill install skipped: ${error.message}`);
  }
}

module.exports = { installSkill };
