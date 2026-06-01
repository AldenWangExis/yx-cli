# yx CLI Design

Date: 2026-06-01

## Purpose

`yx` is a Go-based command-line client for Alibaba Cloud Yunxiao daily developer workflows. It provides a `gh`-like experience while keeping Yunxiao's domain model explicit: Codeup repositories and merge requests, Flow pipelines, and Projex projects/work items.

The first release targets the common development loop:

- Authenticate and manage local profiles.
- List, inspect, and clone Codeup repositories.
- List, inspect, create, and merge Codeup merge requests.
- List, inspect, create, and update Projex work items.
- List, inspect, run, and inspect logs for Flow pipelines.
- Provide GitHub-style aliases where they are unambiguous: `pr` for merge requests and `issue` for work items.

## Scope

### In Scope

- Go single-binary CLI built with Cobra.
- Multi-profile configuration.
- Personal access token authentication for MVP.
- Authentication provider abstraction that allows OAuth/browser login later.
- Secure token storage with system keychain preferred and a restricted local file fallback.
- Default table output and machine-readable `--json` output.
- Outside-In TDD development flow.
- API client adapters for the Yunxiao OpenAPI surfaces needed by the MVP.
- Git operations for repository clone by invoking the system `git` executable.
- Safety controls for write operations, including `--dry-run`, `--yes`, and profile-level write confirmation.

### Out of Scope

- Long-running or bidirectional repository synchronization.
- Full Kanban board view, column, swimlane, or layout configuration.
- Dependency on Alibaba Cloud DevOps MCP Server.
- OAuth implementation in the first release.
- Complete coverage of every Yunxiao OpenAPI.
- Admin-console replacement features such as organization-wide permission management.

## Functional Requirements

### FR1: Authentication And Profiles

- The CLI must support `yx auth login`, `yx auth status`, and `yx auth logout`.
- The MVP login flow must accept a Yunxiao personal access token.
- The CLI must support multiple named profiles and a current active profile.
- The CLI must allow per-command profile override with `--profile`.
- The CLI must never print token values in stdout, stderr, logs, errors, or snapshots.

### FR2: Configuration

- The CLI must read and write configuration at `~/.config/yx/config.yaml` by default.
- The CLI must support `yx config list`, `yx config get`, `yx config set`, and `yx config use`.
- The CLI must support profile fields for domain, organization, region, output mode, safety settings, and repository-to-project mapping.
- Unknown config keys must be rejected unless they belong to an explicitly extensible namespace.

### FR3: Repository Commands

- The CLI must support listing Codeup repositories.
- The CLI must support viewing a single repository.
- The CLI must support cloning a repository by resolving Codeup metadata and invoking the system `git` executable.
- The CLI must not implement the Git transport protocol itself.

### FR4: Merge Request Commands

- The CLI must support listing merge requests for a repository.
- The CLI must support viewing a merge request.
- The CLI must support creating a merge request from source and target branches.
- The CLI must support merging a merge request.
- The CLI must provide `pr` as an alias for `mr`.

### FR5: Project And Work Item Commands

- The CLI must support listing Projex projects.
- The CLI must support listing, viewing, creating, and updating work items.
- The CLI must provide `issue` as an alias for `workitem`.
- Work item commands must treat Projex project as the authoritative context.
- Repository-based issue commands must require explicit repository-to-project mapping.
- If a repository-to-project mapping is missing, the command must fail before sending any Yunxiao API request.

### FR6: Pipeline Commands

- The CLI must support listing pipelines.
- The CLI must support viewing a pipeline.
- The CLI must support running a pipeline for a branch.
- The CLI must support viewing pipeline run logs.
- Log following may stream or poll, but must keep a stable command contract for `yx pipeline logs <run-id> --follow`.

### FR7: Output

- The default human output must be table-oriented and readable in a terminal.
- List and detail commands must support `--json`.
- JSON output must use stable field names suitable for scripts.
- Error messages must be written to stderr and must produce non-zero exit codes.

### FR8: Write Safety

- All write commands must support `--dry-run`.
- Write commands must honor `profiles.<name>.safety.confirmWrites`.
- Write commands must honor `--yes` when confirmation is required.
- In non-interactive environments, write commands that require confirmation must fail unless `--yes` is present.

### FR9: Testability

- Every MVP command must have at least one CLI contract test before implementation.
- CLI contract tests must use fake application services and must not require Yunxiao credentials or network access.
- Every Yunxiao adapter must have mock server tests for success and failure cases.
- `go test ./...` must pass offline.

## Architecture

The first release is a Go monolith with explicit internal boundaries. Command handlers stay thin; they parse flags, load context, call application services, and render output. Yunxiao HTTP behavior is isolated in adapter packages. Git content operations are delegated to the local `git` executable.

```text
cmd/yx
  main.go

internal/cli
  root.go
  auth.go
  config.go
  repo.go
  mr.go
  pipeline.go
  project.go
  workitem.go

internal/app
  repo.go
  merge_request.go
  pipeline.go
  project.go
  workitem.go
  ports.go

internal/config
  store.go
  profile.go
  repo_project_map.go

internal/auth
  provider.go
  pat.go
  token_store.go
  keychain_store.go
  file_store.go

internal/yunxiao
  client.go
  request.go
  endpoint.go
  pagination.go
  errors.go

internal/yunxiao/codeup
  repository_adapter.go
  merge_request_adapter.go

internal/yunxiao/flow
  pipeline_adapter.go

internal/yunxiao/projex
  project_adapter.go
  workitem_adapter.go

internal/gitx
  git.go

internal/output
  table.go
  json.go
  golden.go

internal/safety
  confirm.go
```

### Dependency Direction

```text
CLI command
  -> app use case
  -> app port interface
  -> Yunxiao API adapter or Git adapter
  -> external system
```

The CLI and application layers depend on port interfaces, not concrete HTTP clients. Yunxiao adapters implement those interfaces.

This preserves testability and keeps OpenAPI details from leaking into command semantics.

## Command Surface

### Authentication

```bash
yx auth login
yx auth status
yx auth logout
```

MVP login accepts a Yunxiao personal access token. The command stores the token for the active or selected profile.

### Configuration

```bash
yx config list
yx config get <key>
yx config set <key> <value>
yx config use <profile>
```

Global flags available to relevant commands:

```bash
--profile <name>
--org <organization>
--domain <url>
--json
--verbose
```

### Repositories

```bash
yx repo list
yx repo view <repo>
yx repo clone <repo>
```

`repo clone` resolves the repository clone URL through Codeup metadata and calls the system `git` executable. It does not implement the Git protocol in Go.

### Merge Requests

```bash
yx mr list --repo <repo>
yx mr view <mr-id> --repo <repo>
yx mr create --repo <repo> --source <branch> --target <branch> --title <title>
yx mr merge <mr-id> --repo <repo>
```

Alias:

```bash
yx pr ...
```

`pr` is a command alias for `mr`. Output and behavior are identical.

### Projects And Work Items

```bash
yx project list
yx workitem list --project <project>
yx workitem view <workitem-id>
yx workitem create --project <project> --type <type> --title <title>
yx workitem update <workitem-id> --status <status> --assignee <user>
```

Alias:

```bash
yx issue ...
```

`issue` is an alias for `workitem`, but Yunxiao work items remain the source domain model.

The authoritative context for work items is a Projex project. `yx issue list --repo <repo>` is allowed only when the local config maps the repository to a project. If no mapping exists, the command fails with a clear message:

```text
repo "foo" is not mapped to a project; run yx config set repo.foo.project <project-id>
```

The CLI must not infer this mapping by scanning all organization projects.

### Pipelines

```bash
yx pipeline list
yx pipeline view <pipeline-id>
yx pipeline run <pipeline-id> --branch <branch>
yx pipeline logs <run-id> --follow
```

## Configuration Model

The default config path is:

```text
~/.config/yx/config.yaml
```

Config shape:

```yaml
current: default
profiles:
  default:
    domain: https://devops.aliyun.com
    organization: "<organization-id-or-name>"
    region: "center"
    output: table
    safety:
      confirmWrites: false
    repoProjectMap:
      my-repo: "<project-id>"
```

Rules:

- `current` identifies the active profile.
- `--profile` overrides `current` for a single invocation.
- `--domain` and `--org` override profile values for a single invocation.
- `region` is used by the endpoint resolver to handle center and regional endpoint differences.
- Unknown config keys are rejected unless they are under an explicitly extensible namespace.

## Authentication Model

MVP uses PAT authentication:

```bash
yx auth login --profile default
```

The user pastes a Yunxiao personal access token. The token is stored outside the main YAML config.

Token storage priority:

1. System keychain.
2. Local file fallback with `0600` permissions.

`yx auth status` shows the active profile, organization, domain, token presence, and token storage backend. It never prints the token.

The auth package exposes a provider interface so OAuth/browser login can be added later without changing command handlers:

```go
type Provider interface {
    Login(ctx context.Context, profile config.Profile) (Token, error)
    Status(ctx context.Context, profile config.Profile) (Status, error)
    Logout(ctx context.Context, profile config.Profile) error
}
```

The exact interface may evolve during implementation, but the boundary remains: command handlers depend on an auth provider abstraction, not directly on PAT storage.

## Safety Model

Default behavior follows `gh`-style command execution: explicit write commands execute without additional confirmation.

Safety controls:

- `--dry-run` is supported by all write operations and prints the intended operation without sending a write request.
- `--yes` skips confirmation when confirmation is required.
- `profiles.<name>.safety.confirmWrites: true` requires confirmation for write operations such as merge request merge, work item update, and pipeline run.
- Future destructive or force operations must always require confirmation, regardless of `confirmWrites`.
- In non-interactive environments, if confirmation is required and `--yes` is not present, the command fails.

The MVP contains no destructive commands such as repository deletion.

## Yunxiao API Client Design

The base client owns:

- Base URL.
- Organization context.
- Region/endpoint resolution.
- Token injection through `x-yunxiao-token`.
- HTTP client.
- Pagination helpers.
- Structured API error conversion.

Domain adapters:

```text
codeup.RepositoryAdapter
codeup.MergeRequestAdapter
flow.PipelineAdapter
projex.ProjectAdapter
projex.WorkitemAdapter
```

Adapters return internal domain models rather than exposing raw OpenAPI response structs to CLI handlers.

### Endpoint Resolution

Center and regional endpoint differences are resolved below the application layer. Command handlers do not build OpenAPI paths.

If endpoint resolution cannot produce a valid request because profile data is incomplete, the command returns an actionable error that names the missing fields.

### Error Handling

Errors are normalized into categories:

- Authentication error: invalid or missing token.
- Authorization error: token lacks required permissions or organization context is wrong.
- Not found: repository, project, work item, merge request, pipeline, or run not found.
- Validation error: invalid command arguments or API-side validation failure.
- Rate limit or transient service error.
- Unknown API error.

Verbose mode may print request path and request ID. It must not print tokens or secrets.

## Data Flow Examples

### `yx mr list --repo demo --json`

```text
CLI parses flags
  -> loads active profile
  -> resolves token
  -> calls app.ListMergeRequests(repo: "demo")
  -> app calls MergeRequestService.List
  -> Codeup adapter sends OpenAPI request
  -> adapter normalizes response
  -> output renders JSON
```

### `yx issue list --repo demo`

```text
CLI parses flags
  -> loads active profile
  -> finds repoProjectMap["demo"]
  -> calls app.ListWorkitems(projectID)
  -> Projex adapter searches work items
  -> output renders table
```

If the mapping is missing, the app layer returns a mapping error and no API request is sent.

### `yx pipeline run pipe1 --branch main --dry-run`

```text
CLI parses flags
  -> app builds intended pipeline run operation
  -> safety layer detects dry run
  -> output renders operation summary
  -> no API request is sent
```

## Outside-In TDD Strategy

Development proceeds from user-visible behavior inward.

### 1. CLI Contract Tests First

Each MVP command starts with a CLI-level test that verifies:

- Command exists.
- Required flags and arguments.
- Exit code.
- stdout and stderr behavior.
- JSON schema when `--json` is set.
- Confirmation and dry-run behavior for writes.

CLI tests use fake application services. They must not require Yunxiao tokens or network access.

### 2. Application Use Case Tests

Application tests cover command semantics that should not live in the CLI parser:

- Repository-to-project mapping.
- Dry-run operation summaries.
- Confirmation decisions.
- Work item alias behavior.
- Merge request alias behavior.
- Normalization of service results before output.

### 3. Adapter Tests With Mock Servers

Yunxiao adapters are tested with mock HTTP servers. These tests verify:

- Request path.
- Query parameters.
- Request body.
- `x-yunxiao-token` header.
- Pagination.
- API error conversion.
- No token leakage in returned errors.

### 4. Optional Real-Service Smoke Tests

Real Yunxiao smoke tests may be added later, gated behind explicit environment variables. They are not part of the default `go test ./...` path.

## Milestones

### M1: Foundation

- Go module and Cobra command skeleton.
- Config load/save.
- Profile selection.
- Output renderer.
- Safety confirmation layer.
- PAT auth provider and token store abstraction.

### M2: Repository Commands

- `repo list`
- `repo view`
- `repo clone`

### M3: Merge Request Commands

- `mr list`
- `mr view`
- `mr create`
- `mr merge`
- `pr` alias

### M4: Project And Work Item Commands

- `project list`
- `workitem list`
- `workitem view`
- `workitem create`
- `workitem update`
- `issue` alias
- repo-to-project mapping support for issue alias commands

### M5: Pipeline Commands

- `pipeline list`
- `pipeline view`
- `pipeline run`
- `pipeline logs`

### M6: Packaging And Documentation

- README with setup and examples.
- Release build instructions.
- Shell completion if low-risk.
- Optional install script or Homebrew tap plan.

## Acceptance Criteria

- `go test ./...` passes without network access or Yunxiao credentials.
- Every MVP command has at least one CLI contract test.
- Every Yunxiao adapter has mock server tests for successful and failing responses.
- List and detail commands support `--json`.
- Default output is readable table output.
- Write commands support `--dry-run`.
- Write commands respect `--yes` and `safety.confirmWrites`.
- Token values never appear in logs, errors, stdout, stderr, or golden snapshots.
- The README explains PAT login, profile configuration, repository commands, merge request commands, work item commands, and pipeline commands.

## Open Semantics Resolved

- The first release is scope B: daily development loop, not full platform administration.
- Authentication is PAT-first with a provider abstraction for future OAuth.
- Work items are project-first; repository-based issue commands require explicit local mapping.
- Write confirmation is configurable; default behavior is direct execution for explicit write commands.
- TDD style is Outside-In: CLI contract tests define behavior before internal adapters are implemented.
- The implementation is a pure Go single-binary CLI. MCP Server integration is a future extension, not a runtime dependency.
