# yx-cli

`yx` is a Go CLI for Alibaba Cloud Yunxiao workflows. It covers Codeup repositories, merge requests, work items, projects, pipelines, auth, and local profile configuration.

## Install

Install `yx`:

```bash
curl -fsSL https://raw.githubusercontent.com/AldenWangExis/yx-cli/main/scripts/install.sh | sh
```

The installer downloads the latest GitHub Release by default, writes `yx` to `~/.local/bin`, adds that directory to your shell profile when needed, and prints the restart or `source` command required for the current terminal.

The installer supports macOS arm64, Linux amd64, and Linux arm64. Ubuntu users can use the same install command; make sure `curl` is installed first.

Node/npm users can install the npm wrapper instead:

```bash
npm install -g @aldenwangexis/yx-cli
```

If both installers are used, `command -v yx` shows which one wins on `PATH`.
The npm channel installs a platform-specific npm binary package and does not download from GitHub Releases during install. The npm package version still matches the GitHub Release tag, so `@aldenwangexis/yx-cli@1.6.0` and GitHub Release `v1.6.0` refer to the same Go CLI version.

Install a specific release when you need a pinned version:

```bash
YX_INSTALL_VERSION=v1.0.0 curl -fsSL https://raw.githubusercontent.com/AldenWangExis/yx-cli/main/scripts/install.sh | sh
```

Verify:

```bash
yx --version
yx --help
yx auth status
```

Windows users can download `yx-windows-amd64.exe` from the GitHub Release assets and place it in a directory on `PATH`.

## Configure

Create a default Yunxiao profile:

```bash
yx config set profiles.default.domain https://devops.aliyun.com
yx config set profiles.default.organization <organization-id>
yx config set profiles.default.region center
yx config use default
```

If you already have a Yunxiao personal access token, you can discover available organizations instead of copying the organization ID manually:

```bash
yx auth login
yx org list
yx org use <organization-id>
```

Optional Codeup service connection configuration:

```bash
yx config set profiles.default.serviceConnections.codeup <service-connection-id>
```

## Login

```bash
yx auth login
yx auth status
```

Paste a Yunxiao personal access token when prompted. Tokens are stored outside `config.yaml`. `yx auth status` masks tokens and service connection IDs in terminal output.

## Common Commands

Repositories:

```bash
yx repo list
yx repo view <repo>
yx repo create --name demo --path demo --visibility private --yes
yx repo clone <repo> [destination]
yx repo delete <repo> --dry-run
yx repo branch list <repo>
yx repo branch sync <repo> --source master --target feat/a --dry-run
yx repo commit list <repo> --ref master
yx repo file view <repo> test.py --ref master
```

Merge requests:

```bash
yx mr list --repo <repo>
yx mr view <mr-id> --repo <repo>
yx mr create --repo <repo> --source feat/a --target main --title "Add feature" --dry-run
yx mr merge <mr-id> --repo <repo> --yes
yx mr close <mr-id> --repo <repo> --dry-run
```

`yx pr ...` is an alias for `yx mr ...`.

Projects, members, and work items:

```bash
yx project list
yx member list                         # discover USER_ID values for assignees
yx member search --name "王子豪"
yx workitem list --project <project-id>
yx workitem view <workitem-id>
yx workitem create --project <project-id> --type task --title "Task title" --assignee @me --dry-run
yx workitem update <workitem-id> --title "P1 Task title" --description "Updated details" --description-format markdown --dry-run
yx workitem update <workitem-id> --status done --assignee <user-id> --dry-run
yx workitem delete <workitem-id> --dry-run
```

`yx issue ...` is an alias for `yx workitem ...`. Use `--assignee @me` for the current user, or `yx member search --name <name>` then pass the returned `USER_ID`. Repository-based issue lists require an explicit repo-to-project mapping:

```bash
yx config set profiles.default.repoProjectMap.<repo> <project-id>
yx issue list --repo <repo>
```

Pipelines:

```bash
yx pipeline list
yx pipeline view <pipeline-id>
yx pipeline create --name yx-cli-ci --file pipeline.yml
yx pipeline run <pipeline-id> --branch main --dry-run
yx pipeline logs <run-id> --follow
```

## Output And Safety

- Add `--json` to list and detail commands for machine-readable output.
- Add `--dry-run` to write commands to preview without sending write requests.
- Add `--yes` to confirmed write commands when you want non-interactive execution.
- `yx` may check GitHub Releases once per day and print update hints to stderr. Disable with `YX_NO_UPDATE_CHECK=1`.
- Set `profiles.<name>.safety.confirmWrites true` to require confirmation for write commands.

## Development

```bash
go test ./...
go build -o yx ./cmd/yx
./yx --help
```

Useful Make targets:

```bash
make test
make build
```

## Release

Pushes to `main` run tests. Tags matching `v*` run tests, build release binaries, and create or update a GitHub Release with these assets:

- `yx-darwin-arm64`
- `yx-linux-amd64`
- `yx-linux-arm64`
- `yx-windows-amd64.exe`

Release a new version:

```bash
git tag -a v0.2.0 -m "v0.2.0"
git push origin main v0.2.0
git push codeup main v0.2.0
```
