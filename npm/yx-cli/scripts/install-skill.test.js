#!/usr/bin/env node
"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { installSkill } = require("./install-skill");

const repoRoot = path.join(__dirname, "..", "..", "..");
const sourceSkill = path.join(repoRoot, "skills", "yx-cli", "SKILL.md");
const packagedSkill = path.join(repoRoot, "npm", "yx-cli", "skills", "yx-cli", "SKILL.md");

assert.equal(
  fs.readFileSync(packagedSkill, "utf8"),
  fs.readFileSync(sourceSkill, "utf8"),
  "packaged skill must match skills/yx-cli/SKILL.md",
);

const originalHome = process.env.HOME;
const originalSkip = process.env.YX_SKIP_SKILL_INSTALL;
const tmpHome = fs.mkdtempSync(path.join(os.tmpdir(), "yx-skill-install-"));

try {
  process.env.HOME = tmpHome;
  delete process.env.YX_SKIP_SKILL_INSTALL;

  const result = installSkill();
  const target = path.join(tmpHome, ".agents", "skills", "yx-cli", "SKILL.md");
  assert.equal(result.installed, true);
  assert.equal(result.target, target);
  assert.equal(fs.readFileSync(target, "utf8"), fs.readFileSync(sourceSkill, "utf8"));

  const again = installSkill();
  assert.equal(again.installed, false, "second install should be idempotent");

  process.env.YX_SKIP_SKILL_INSTALL = "1";
  fs.rmSync(path.join(tmpHome, ".agents"), { recursive: true, force: true });
  const skipped = installSkill();
  assert.equal(skipped.skipped, true);
  assert.equal(fs.existsSync(target), false);
} finally {
  if (originalHome === undefined) {
    delete process.env.HOME;
  } else {
    process.env.HOME = originalHome;
  }
  if (originalSkip === undefined) {
    delete process.env.YX_SKIP_SKILL_INSTALL;
  } else {
    process.env.YX_SKIP_SKILL_INSTALL = originalSkip;
  }
  fs.rmSync(tmpHome, { recursive: true, force: true });
}

console.log("npm skill install tests passed");
