# @aldenwangexis/yx-cli

npm package for `yx`, the Alibaba Cloud Yunxiao command line client.

Install:

```bash
npm install -g @aldenwangexis/yx-cli
```

Verify:

```bash
command -v yx
yx --version
```

This package installs a platform-specific npm binary package and copies the bundled `yx-cli` skill to `~/.agents/skills/yx-cli/`. Set `YX_SKIP_SKILL_INSTALL=1` to skip the skill install step. npm installation does not download from GitHub Releases.
