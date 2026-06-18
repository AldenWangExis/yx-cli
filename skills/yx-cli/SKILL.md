---
name: yx-cli
description: >-
  Use this skill whenever the user wants to operate Alibaba Cloud Yunxiao/云效
  with the `yx` CLI: checking auth/config status, installing yx when missing,
  discovering organizations, working with Codeup repos/branches/commits/files/
  merge requests, Projex projects/work items/issues, Flow pipelines/runs/logs,
  or translating a Yunxiao terminal task into an appropriate `yx` workflow.
  Prefer this skill over raw OpenAPI unless yx lacks the needed command.
compatibility: >-
  Requires shell access. Uses local `yx` when installed; the normal macOS
  install path is `~/.local/bin/yx`. Installation and real Yunxiao operations
  require network. Real operations require a Yunxiao personal access token;
  `yx auth login` can discover available organizations and store the selected
  organization ID. Some pipeline workflows need a Codeup service connection ID.
---

# yx-cli

Use `yx` as the primary Yunxiao operational CLI. This skill should guide command choice and safe workflow, not duplicate every flag. For exact syntax, call `yx --help` and `yx <command> --help`.

## First Response Pattern

Start by establishing state:

```bash
if command -v yx >/dev/null 2>&1; then
  yx --version
  yx auth status
else
  echo "yx is not installed"
fi
```

If `yx` is missing, guide installation with the official installer:

```bash
curl -fsSL https://raw.githubusercontent.com/AldenWangExis/yx-cli/main/scripts/install.sh | sh
```

Run the installer only when the user asked to install/setup or approved it. Otherwise show the command and explain that it installs `yx` to `~/.local/bin` and may require a new shell or sourcing the printed profile file.

The installer downloads the latest release by default and shows curl download progress. Pin a version only when the user asks for reproducibility:

```bash
YX_INSTALL_VERSION=v1.2.0 curl -fsSL https://raw.githubusercontent.com/AldenWangExis/yx-cli/main/scripts/install.sh | sh
```

## Core Rules

- Prefer `yx` over direct OpenAPI for normal Yunxiao work.
- Use `--json` when command output will feed another step.
- Use `--dry-run` before writes unless the user explicitly asked for a real mutation.
- Use `--yes` only after the user approved non-interactive writes.
- Treat `repo delete`, `workitem/issue delete`, and `mr/pr close` as destructive or state-changing operations; prefer `--dry-run` first.
- Never print raw PATs, service connection IDs, credentials, or credential-bearing remote URLs.
- If uncertain about flags, run `yx <command> --help` instead of guessing.
- Report exact commands run, with secrets redacted.

## Command Map

| Intent | Start Here | Help Key |
|---|---|---|
| Install/check/login | `yx auth status`, `yx auth login` | `yx auth --help` |
| Profile/domain/org config | `yx config list/get/set/use` | `yx config --help` |
| Organization discovery/selection | `yx org list`, `yx org use <org-id>` | `yx org --help` |
| Codeup repo list/view/create/clone/delete | `yx repo ...` | `yx repo --help` |
| Current repo from git remote | `yx repo current` | `yx repo current --help` |
| Branch/commit/file inspection | `yx repo branch/commit/file ...` | `yx repo branch --help` |
| Merge requests / PRs | `yx mr ...` or `yx pr ...` | `yx mr --help` |
| Projects | `yx project ...` | `yx project --help` |
| Organization members / assignees | `yx member list/search/get` | `yx member --help` |
| Work items / issues | `yx workitem ...` or `yx issue ...` | `yx issue --help` |
| Flow pipelines/runs/logs | `yx pipeline ...` | `yx pipeline --help` |

`pr` is an alias for `mr`; `issue` is an alias for `workitem`.

## Setup Hints

Default profile shape when the organization ID is already known:

```bash
yx config set profiles.default.domain https://devops.aliyun.com
yx config set profiles.default.organization <org-id>
yx config set profiles.default.region center
yx config use default
yx auth login
```

Preferred first-time flow when the organization ID is unknown:

```bash
yx config set profiles.default.domain https://devops.aliyun.com
yx config set profiles.default.region center
yx config use default
yx auth login
yx org list
yx org use <org-id>
```

`yx auth login` shows reference links on separate lines, asks for the PAT, optionally asks for the Codeup service connection ID, and attempts to discover organizations. If exactly one organization is available, it stores it automatically; if several are available, choose one with `yx org use <org-id>`.

Useful URLs shown by `yx auth login/status`:

- PAT: `https://account-devops.aliyun.com/settings/personalAccessToken`
- Service connections: `https://flow.aliyun.com/setting/service-connection`

Optional Codeup service connection:

```bash
yx config set profiles.default.serviceConnections.codeup <service-connection-id>
```

Repo-based issue lookup needs explicit mapping:

```bash
yx config set profiles.default.repoProjectMap.<repo-id> <project-id>
```

## Workflow Guidance

For repo-oriented tasks, first try to resolve context:

```bash
yx repo current --refresh
```

If that fails, inspect remotes and use help:

```bash
git remote -v
yx repo current --help
```

For repository deletion, resolve the target first and dry-run the write:

```bash
yx repo view <repo>
yx repo delete <repo> --dry-run
yx repo delete <repo> --yes
```

For PR/MR work, prefer `yx pr ...` if the user speaks GitHub vocabulary, and `yx mr ...` if they speak Codeup/Yunxiao vocabulary.

Close, not delete, merge requests:

```bash
yx pr close <mr-id> --repo <repo> --dry-run
yx pr close <mr-id> --repo <repo> --yes
```

For issue/work item work, use `yx issue ...` for GitHub-like language and `yx workitem ...` for Yunxiao language. Work items belong to Projex projects, not repos: list-by-repo requires `repoProjectMap`; view/update/delete use the work item ID directly.

Assignees are Yunxiao `USER_ID` values. Use `@me` for the authenticated user; search members before assigning someone else:

```bash
yx member search --name <name>
yx member get --user-id <user-id>
```

Create/update work items with explicit assignees and structured descriptions:

```bash
yx issue create --project <project-id> --type Task --title "Task title" --assignee @me --dry-run
yx issue update <workitem-id> --title "P1 Task title" --description "Updated details" --description-format markdown --dry-run
yx issue update <workitem-id> --assignee @me --yes
```

If creation fails because a project template requires an assignee, retry with `--assignee @me` or search a teammate and pass the returned `USER_ID`.

Delete work items only after confirming the ID:

```bash
yx issue view <workitem-id>
yx issue delete <workitem-id> --dry-run
yx issue delete <workitem-id> --yes
```

For pipeline debugging, proceed in this order:

1. `yx pipeline run view ...`
2. `yx pipeline run steps ...`
3. `yx pipeline run logs ...`

Use `yx pipeline run logs --help` for job, step, build, offset, and limit flags.

## Troubleshooting Cues

| Symptom | Direction |
|---|---|
| `yx: command not found` | Install with the official curl command, then verify PATH/new shell. |
| `unknown flag: --version` | Binary is stale; reinstall or rebuild. |
| Not logged in | Run `yx auth login`; then `yx auth status`. |
| Organization ID unknown | Run `yx org list`; then `yx org use <org-id>`. |
| `unknown command "org"` or missing delete/close commands | Binary is stale; reinstall latest or rebuild from current main. |
| Missing project/repo mapping | Pass `--project` or configure `repoProjectMap`. |
| Current repo cannot resolve | Check `git remote -v`; then `yx repo current --remote <name> --refresh`. |
| Remote URL has credentials | Recommend SSH remote or git credential helper. |
| Pipeline clone/auth issue | Check Codeup service connection configuration. |
| Pipeline tool missing | Inspect logs; fix pipeline runtime/YAML, not historical runs. |

## Reporting Format

Keep replies compact:

1. Status: installed/auth/config summary.
2. Commands: exact commands run or recommended.
3. Result: IDs/status/log highlights.
4. Safety: read-only, dry-run, or real write.
5. Next: one concrete command or decision.
