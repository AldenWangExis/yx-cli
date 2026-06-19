# npm Distribution Plan

## Decision

Add npm as a thin installation channel for `yx`. The npm package does not rebuild or reimplement the CLI; it downloads the matching binary from GitHub Releases during `postinstall` and exposes a `yx` bin shim.

Package name:

```text
@aldenwangexis/yx-cli
```

Why scoped:

- the unscoped `yx-cli` name is already owned by another npm publisher;
- the current npm account is `aldenwangexis`;
- `@aldenwangexis/yx-cli` returned npm 404 during availability checks, which is the expected pre-publish state.

## Release Source Of Truth

Git tags are the release transaction boundary. A successful release must publish the same version to every channel:

1. tag `vX.Y.Z`;
2. GitHub Release `vX.Y.Z`;
3. Release assets built with CLI version `vX.Y.Z`;
4. npm package `@aldenwangexis/yx-cli@X.Y.Z`;
5. npm package install downloads GitHub Release `vX.Y.Z`.

GitHub Releases remain the canonical binary source. npm package versions must map one-to-one to Git tags:

| Git tag | npm version | downloaded release |
|---|---|---|
| `v1.4.0` | `1.4.0` | `v1.4.0` |
| `v1.6.0` | `1.6.0` | `v1.6.0` |

Do not publish an npm version unless the matching GitHub Release assets already exist.

Current branch note: the npm package is set to `1.4.0` because `v1.4.0` already has Release assets and is useful for validating the npm installer end to end. Before publishing a new public npm release for current CLI features, create a new passing GitHub tag/release and bump `npm/yx-cli/package.json` to that version.

The release workflow enforces this with:

```bash
sh scripts/check_release_version.sh vX.Y.Z
```

If `npm/yx-cli/package.json` is not `X.Y.Z`, the tag workflow fails before binary build and npm publish.

## User UX

Install:

```bash
npm install -g @aldenwangexis/yx-cli
```

Verify:

```bash
command -v yx
yx --version
```

Update:

```bash
npm update -g @aldenwangexis/yx-cli
```

When the CLI is launched through the npm shim, the wrapper sets `YX_INSTALL_CHANNEL=npm` and `YX_NPM_PACKAGE=@aldenwangexis/yx-cli`; update hints should therefore show the npm update command instead of the curl installer command.

## Package Layout

```text
npm/yx-cli/
  package.json
  README.md
  bin/yx.js
  scripts/install.js
  scripts/install.test.js
  scripts/platform.js
```

`scripts/install.js`:

1. reads `package.json` version;
2. maps it to GitHub tag `v<version>`;
3. detects platform and architecture;
4. downloads the matching Release asset into `vendor/`;
5. marks the binary executable on POSIX platforms.

Platform mapping:

| Node platform | Node arch | Release asset |
|---|---|---|
| `darwin` | `arm64` | `yx-darwin-arm64` |
| `linux` | `x64` | `yx-linux-amd64` |
| `linux` | `arm64` | `yx-linux-arm64` |
| `win32` | `x64` | `yx-windows-amd64.exe` |

Unsupported platforms fail with a clear error.

`bin/yx.js` resolves `vendor/yx` or `vendor/yx.exe`, forwards all CLI args, inherits stdio, and exits with the child process status.

## GitHub Installer Coexistence

The GitHub installer writes to:

```text
~/.local/bin/yx
```

npm global install writes a Node shim to the npm global prefix, commonly:

```text
~/.nvm/versions/node/<version>/bin/yx
/usr/local/bin/yx
```

They usually do not overwrite each other. The conflict model is PATH precedence: whichever `yx` appears first wins. Users can inspect it with:

```bash
command -v yx
yx --version
```

## First Manual Publish

Prerequisites:

1. npm account exists and is logged in locally.
2. `npm whoami` returns the intended publisher.
3. The package version matches an existing GitHub Release with all required assets.
4. Validation gates pass.

Publish:

```bash
cd npm/yx-cli
npm publish --access public
```

`--access public` is required for a public scoped package.

Do not commit or print npm tokens.

## Future Trusted Publishing

Manual publish is acceptable for the first package. For long-term releases, prefer npm Trusted Publishing from GitHub Actions:

- no long-lived npm token in repository secrets;
- OIDC-backed publish identity;
- optional provenance via `npm publish --provenance`.

The tag workflow includes a disabled-by-default npm publish job. It runs only when repository variable `YX_ENABLE_NPM_PUBLISH` is set to `true`.

Before enabling it:

1. publish the package manually once, or create/configure it in npm;
2. configure npm Trusted Publishing for this GitHub repository and workflow;
3. verify `npm/yx-cli/package.json` equals the tag version;
4. set GitHub repository variable `YX_ENABLE_NPM_PUBLISH=true`.

After that, pushing tag `vX.Y.Z` publishes GitHub Release assets first, then publishes `@aldenwangexis/yx-cli@X.Y.Z` with npm provenance.

## Release Operator Flow

For a normal feature release:

```bash
npm version --prefix npm/yx-cli X.Y.Z --no-git-tag-version
make release-check VERSION=vX.Y.Z
git add npm/yx-cli/package.json
git commit -m "chore: release vX.Y.Z"
git tag vX.Y.Z
git push origin main --tags
```

Users who do not specify a version receive npm's `latest` dist-tag. npm sets `latest` on normal publish, so default npm installs follow the newest successful npm release. Users who need a specific version can run:

```bash
npm install -g @aldenwangexis/yx-cli@X.Y.Z
```

## Validation Gates

Run before publishing:

```bash
go test ./...
sh scripts/test_install.sh
cd npm/yx-cli
npm test
npm pack --dry-run
```

For an end-to-end local install check:

```bash
rm -rf npm/yx-cli/vendor
tmp_prefix="$(mktemp -d)"
HTTPS_PROXY=http://127.0.0.1:7897 HTTP_PROXY=http://127.0.0.1:7897 \
  npm install --prefix "$tmp_prefix" ./npm/yx-cli
"$tmp_prefix/node_modules/.bin/yx" --version
```

Use the proxy environment only when local network requires it.

## Security And Hygiene

- Never download mutable `latest` during npm install; always use the package version's matching GitHub tag.
- Keep the downloaded binary under the package `vendor/` directory.
- Do not mutate shell profiles from the npm installer.
- Do not commit `vendor/` or packed `.tgz` artifacts.
- Keep npm package contents constrained with the `files` field and verify with `npm pack --dry-run`.
