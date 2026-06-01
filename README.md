# yx-cli

`yx` is a Go CLI for Alibaba Cloud Yunxiao developer workflows.

## Build

```bash
go test ./...
go build -o yx ./cmd/yx
```

## Configure

```bash
yx config set profiles.default.domain https://openapi-rdc.aliyuncs.com
yx config set profiles.default.organization 68086322e3a71588779435e0
yx config set profiles.default.region center
yx config use default
```

## Login

```bash
yx auth login
yx auth status
```

Paste a Yunxiao personal access token when prompted. Tokens are stored outside `config.yaml`.

## Repositories

```bash
yx repo list
yx repo view <repo>
yx repo clone <repo> [destination]
```

## Merge Requests

```bash
yx mr list --repo <repo>
yx mr view <mr-id> --repo <repo>
yx mr create --repo <repo> --source feat/a --target main --title "Add feature" --dry-run
yx mr merge <mr-id> --repo <repo> --yes
```

`yx pr ...` is an alias for `yx mr ...`.

## Work Items

```bash
yx project list
yx workitem list --project <project-id>
yx workitem view <workitem-id>
yx workitem create --project <project-id> --type task --title "Task title" --dry-run
yx workitem update <workitem-id> --status done --assignee <user-id> --dry-run
```

`yx issue ...` is an alias for `yx workitem ...`. Repository-based issue commands require explicit mapping:

```bash
yx config set profiles.default.repoProjectMap.<repo> <project-id>
yx issue list --repo <repo>
```

## Pipelines

```bash
yx pipeline list
yx pipeline view <pipeline-id>
yx pipeline run <pipeline-id> --branch main --dry-run
yx pipeline logs <run-id> --follow
```

## Output And Safety

- Add `--json` to list and detail commands for machine-readable output.
- Add `--dry-run` to write commands to preview without sending write requests.
- Set `profiles.<name>.safety.confirmWrites true` to require confirmation for write commands.
