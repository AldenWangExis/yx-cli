# yx-cli

`yx` is a Go CLI for Alibaba Cloud Yunxiao workflows. It covers Codeup repositories, merge requests, work items, projects, pipelines, auth, and local profile configuration.

## Install

Install the latest GitHub Release on macOS Apple Silicon:

```bash
curl -fsSL https://raw.githubusercontent.com/AldenWangExis/yx-cli/main/scripts/install.sh | sh
```

Install a specific version:

```bash
YX_INSTALL_VERSION=v0.2.0 \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/AldenWangExis/yx-cli/main/scripts/install.sh)"
```

The installer writes `yx` to `~/.local/bin` by default. If that directory is not in your `PATH`, add it for zsh:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

Verify:

```bash
yx --help
yx auth status
```

Manual macOS install:

```bash
curl -fsSL -o yx https://github.com/AldenWangExis/yx-cli/releases/latest/download/yx-darwin-arm64
chmod +x yx
mkdir -p ~/.local/bin
mv yx ~/.local/bin/yx
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
```

`yx pr ...` is an alias for `yx mr ...`.

Projects and work items:

```bash
yx project list
yx workitem list --project <project-id>
yx workitem view <workitem-id>
yx workitem create --project <project-id> --type task --title "Task title" --dry-run
yx workitem update <workitem-id> --status done --assignee <user-id> --dry-run
```

`yx issue ...` is an alias for `yx workitem ...`. Repository-based issue commands require an explicit repo-to-project mapping:

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
- `yx-windows-amd64.exe`

Release a new version:

```bash
git tag -a v0.2.0 -m "v0.2.0"
git push origin main v0.2.0
git push codeup main v0.2.0
```
