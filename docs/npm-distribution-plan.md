# npm Distribution Plan

## Decision

Keep `yx` as a Go CLI. npm is an independent distribution channel, not a JavaScript rewrite and not a GitHub Release downloader.

The npm channel uses one main package plus platform-specific binary packages:

```text
@aldenwangexis/yx-cli
@aldenwangexis/yx-cli-darwin-arm64
@aldenwangexis/yx-cli-linux-x64
@aldenwangexis/yx-cli-linux-arm64
@aldenwangexis/yx-cli-win32-x64
```

Users install the main package:

```bash
npm install -g @aldenwangexis/yx-cli
```

npm resolves the matching optional platform package from the npm registry. Installation does not download from GitHub Releases.

## Release Invariant

Git tags are the release transaction boundary. A successful release publishes the same version to every channel:

```text
tag vX.Y.Z
GitHub Release vX.Y.Z
GitHub Release assets built with CLI version vX.Y.Z
@aldenwangexis/yx-cli@X.Y.Z
@aldenwangexis/yx-cli-darwin-arm64@X.Y.Z
@aldenwangexis/yx-cli-linux-x64@X.Y.Z
@aldenwangexis/yx-cli-linux-arm64@X.Y.Z
@aldenwangexis/yx-cli-win32-x64@X.Y.Z
```

The main npm package pins every platform package to the exact same version in `optionalDependencies`. Do not use ranges such as `^X.Y.Z`.

## Package Layout

```text
npm/
  yx-cli/
    package.json
    README.md
    bin/yx.js
    scripts/packages.js
    scripts/prepare-platform-packages.js
    scripts/verify-packages.js
    scripts/install.test.js
  yx-cli-darwin-arm64/
    package.json
    bin/yx              # generated during release
  yx-cli-linux-x64/
    package.json
    bin/yx              # generated during release
  yx-cli-linux-arm64/
    package.json
    bin/yx              # generated during release
  yx-cli-win32-x64/
    package.json
    bin/yx.exe          # generated during release
```

`bin/` directories under platform packages are generated from CI build artifacts and are ignored by git.

## Runtime Behavior

`@aldenwangexis/yx-cli` exposes the `yx` command through `bin/yx.js`.

The shim:

1. detects `process.platform` and `process.arch`;
2. resolves the matching platform package;
3. executes that package's bundled Go binary;
4. forwards args and stdio;
5. sets npm channel metadata:

```text
YX_INSTALL_CHANNEL=npm
YX_NPM_PACKAGE=@aldenwangexis/yx-cli
YX_NPM_PLATFORM_PACKAGE=<resolved-platform-package>
```

The Go binary remains the real CLI implementation.

## Platform Mapping

| Node platform | Node arch | npm platform package | GitHub Release asset |
|---|---|---|---|
| `darwin` | `arm64` | `@aldenwangexis/yx-cli-darwin-arm64` | `yx-darwin-arm64` |
| `linux` | `x64` | `@aldenwangexis/yx-cli-linux-x64` | `yx-linux-amd64` |
| `linux` | `arm64` | `@aldenwangexis/yx-cli-linux-arm64` | `yx-linux-arm64` |
| `win32` | `x64` | `@aldenwangexis/yx-cli-win32-x64` | `yx-windows-amd64.exe` |

Unsupported platforms fail with a clear message.

## CI Release Flow

On tag `vX.Y.Z`:

1. run Go and npm tests;
2. verify `npm/yx-cli/package.json`, all platform package versions, and main package `optionalDependencies` equal `X.Y.Z`;
3. build Go binaries once;
4. upload the binaries as GitHub Actions artifacts;
5. create or update GitHub Release `vX.Y.Z`;
6. if `YX_ENABLE_NPM_PUBLISH=true`, download the same build artifacts;
7. copy the artifacts into npm platform package `bin/` directories;
8. publish platform npm packages first;
9. publish the main npm package last.

Publishing main last matters because it depends on the platform packages.

The npm publish job is disabled by default and runs only when GitHub repository variable `YX_ENABLE_NPM_PUBLISH` is `true`.

## Version Gate

The shared gate is:

```bash
sh scripts/check_release_version.sh vX.Y.Z
```

It fails if:

- main npm package version is not `X.Y.Z`;
- any platform package version differs;
- main package optional dependency pins differ.

Local release check:

```bash
make release-check VERSION=vX.Y.Z
```

## Release Operator Flow

For a normal feature release:

```bash
npm version --prefix npm/yx-cli X.Y.Z --no-git-tag-version
for dir in npm/yx-cli-darwin-arm64 npm/yx-cli-linux-x64 npm/yx-cli-linux-arm64 npm/yx-cli-win32-x64; do
  npm version --prefix "$dir" X.Y.Z --no-git-tag-version
done
node -e '
const fs = require("fs");
const pkgPath = "npm/yx-cli/package.json";
const pkg = require("./" + pkgPath);
for (const name of Object.keys(pkg.optionalDependencies)) pkg.optionalDependencies[name] = pkg.version;
fs.writeFileSync(pkgPath, JSON.stringify(pkg, null, 2) + "\n");
'
make release-check VERSION=vX.Y.Z
git add npm/*/package.json
git commit -m "chore: release vX.Y.Z"
git tag vX.Y.Z
git push origin main --tags
```

Users who do not specify a version receive npm's `latest` dist-tag. npm sets `latest` on normal publish, so default npm installs follow the newest successful npm release. Users who need a specific version can run:

```bash
npm install -g @aldenwangexis/yx-cli@X.Y.Z
```

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
3. all five npm packages use the same version.
4. platform package `bin/` directories have been prepared from the same Go build artifacts.
5. validation gates pass.

Publish platform packages first, then the main package:

```bash
for dir in npm/yx-cli-darwin-arm64 npm/yx-cli-linux-x64 npm/yx-cli-linux-arm64 npm/yx-cli-win32-x64; do
  npm publish --access public "$dir"
done
npm publish --access public npm/yx-cli
```

`--access public` is required for public scoped packages.

## Future Trusted Publishing

For long-term releases, prefer npm Trusted Publishing from GitHub Actions:

- no long-lived npm token in repository secrets;
- OIDC-backed publish identity;
- optional provenance via `npm publish --provenance`.

Before enabling it:

1. publish or configure each npm package in npm;
2. configure npm Trusted Publishing for this GitHub repository and workflow;
3. set GitHub repository variable `YX_ENABLE_NPM_PUBLISH=true`.

## Validation Gates

Run before release:

```bash
make release-check VERSION=vX.Y.Z
```

Run package checks directly:

```bash
cd npm/yx-cli
npm test
npm pack --dry-run
```

To test platform package preparation locally:

```bash
mkdir -p dist
go build -o dist/yx-darwin-arm64 ./cmd/yx
cp dist/yx-darwin-arm64 dist/yx-linux-amd64
cp dist/yx-darwin-arm64 dist/yx-linux-arm64
cp dist/yx-darwin-arm64 dist/yx-windows-amd64.exe
node npm/yx-cli/scripts/prepare-platform-packages.js dist
npm test --prefix npm/yx-cli
```

## Security And Hygiene

- npm install must not download from GitHub Releases.
- Do not commit generated platform package `bin/` directories.
- Do not commit packed `.tgz` artifacts.
- Do not mutate shell profiles from npm install.
- Never commit npm tokens.
- Keep npm package contents constrained with `files` and verify with `npm pack --dry-run`.
