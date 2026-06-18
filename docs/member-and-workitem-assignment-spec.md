# Member Discovery And Workitem Assignment Spec

## Purpose

`yx-cli` should let users complete the normal Yunxiao work item loop without hand-written OpenAPI calls:

1. Discover organization members and their user IDs.
2. Create work items with an assignee when project templates require it.
3. Update work item title, description, status, and assignee through the CLI.

This document defines the behavior and constraints before implementation.

## OpenAPI Boundary

Use the current `yx-cli` architecture: Yunxiao Projex and Platform OpenAPI adapters under `internal/yunxiao`.

Do not mix in legacy work item APIs unless the current Projex API cannot support a required behavior.

### Current User

Endpoint:

```text
GET /oapi/v1/platform/user
```

Observed/expected fields include:

- `id`: current Yunxiao user ID. Use this for `@me`.
- `name`
- `email`

### Organization Members

Endpoint:

```text
POST /oapi/v1/platform/organizations/{organizationId}/members:search
GET  /oapi/v1/platform/organizations/{organizationId}/members
```

Returned member records include:

- `id`: organization member record ID.
- `organizationId`
- `userId`: the value accepted by work item `assignedTo`.
- `name`
- `email`
- `roleIds`
- `status`

Permission required by official docs: organization member read access.

In the current environment, the active token can read organization members for organization `68086322e3a71588779435e0`.

### Project Members

Endpoint observed as readable:

```text
GET /oapi/v1/projex/organizations/{organizationId}/projects/{projectId}/members
```

Returned project member records include:

- `userId`
- `userName`
- `roleId`
- `roleName`

This endpoint is useful for project-scoped assignment validation and role display, but the first implementation can focus on organization members.

### Work Item Create

Endpoint:

```text
POST /oapi/v1/projex/organizations/{organizationId}/workitems
```

Relevant request fields:

- `spaceId`: project ID.
- `workitemTypeId`: work item type ID/name as currently accepted by the CLI.
- `subject`: title.
- `description`
- `formatType`: `MARKDOWN` or `RICHTEXT`.
- `assignedTo`: user ID.

Some Yunxiao project templates require `assignedTo` at creation time. If it is missing, the server can reject the request before a later `update --assignee` can run.

### Work Item Update

Endpoint:

```text
PUT /oapi/v1/projex/organizations/{organizationId}/workitems/{id}
```

Relevant request fields:

- `subject`: title.
- `description`
- `formatType`: `MARKDOWN` or `RICHTEXT`.
- `status`
- `assignedTo`: user ID.

## Command Design

### Member Commands

Add a top-level `member` command.

#### `yx member list`

Lists organization members.

Expected flags:

```text
--status <status>
--json
```

Behavior:

- Defaults to the current profile and organization.
- Returns all accessible organization members, paging internally when needed.
- Displays a stable table with:
  - `USER_ID`
  - `NAME`
  - `EMAIL`
  - `STATUS`
  - `ROLES`
- `ROLES` may initially display comma-joined `roleIds` when role names are not available.

#### `yx member search`

Searches organization members by explicit criteria.

First-version flags:

```text
--name <name>
--email <email>
--status <status>
--json
```

Behavior:

- At least one search criterion is required.
- `--name` searches by member name.
- `--email` searches by email.
- Search can return zero, one, or many records.
- Do not use positional `search <name>` in the first version. Explicit flags avoid ambiguity and leave room for future search fields.

#### `yx member get`

Reads exactly one member by exact ID.

First-version flag:

```text
--user-id <user-id>
--json
```

Behavior:

- `--user-id` is required.
- Expected result cardinality is exactly one.
- If no member matches, return a not-found error.
- If multiple records somehow match, return an error rather than guessing.

Reasoning:

- `search` returns candidate lists.
- `get` is an exact lookup for a known object.
- The flag name must be `--user-id`, not just `--id`, because Yunxiao member payloads include both an organization member record `id` and a `userId`.

### Work Item Create

Extend:

```text
yx issue create
yx workitem create
```

New flag:

```text
--assignee <user-id|@me>
```

Behavior:

- `--assignee <user-id>` sends that value as `assignedTo`.
- `--assignee @me` resolves the current user via `GET /oapi/v1/platform/user` and sends its `id` as `assignedTo`.
- Do not support `--assignee-name` in the first version. Writing by name is unsafe because names can be duplicated.
- Do not silently default to the current user when `--assignee` is omitted.

Rationale:

- Explicit `@me` gives a fast path for solo/agent workflows.
- Direct user ID preserves automation and scripting support.
- Avoiding implicit defaults prevents surprising assignment behavior.

### Work Item Update

Extend:

```text
yx issue update <workitem-id>
yx workitem update <workitem-id>
```

New flags:

```text
--title <title>
--description <description>
--description-format markdown|richtext
--assignee <user-id|@me>
```

Existing flags remain:

```text
--status <status>
--dry-run
--yes
```

Behavior:

- `--title` maps to Projex `subject`.
- `--description` maps to `description`.
- `--description-format markdown` maps to `formatType: MARKDOWN`.
- `--description-format richtext` maps to `formatType: RICHTEXT`.
- `--assignee @me` resolves current user ID.
- The command requires at least one mutation field: status, assignee, title, or description.
- `--description-format` without `--description` is allowed only if the API supports changing format independently. If not verified during implementation, reject it locally.

## Error Handling

### Missing Assignee On Create

When the server rejects create because the project template requires an assignee, surface an actionable hint.

Example message shape:

```text
create work item failed: assignee is required by the project template.
Try assigning yourself:
  yx issue create ... --assignee @me
Or find a teammate first:
  yx member search --name <name>
  yx issue create ... --assignee <user-id>
```

The implementation should detect common server phrases in both Chinese and English, such as:

- `指派人不能为空`
- `负责人不能为空`
- `assignee`
- `assignedTo`

Do not hide the original server error; append the hint.

### Ambiguous Member Search

Search commands may return multiple rows. This is normal.

Create/update commands must not accept names in the first version, so they do not need ambiguity handling for writes.

If a future `--assignee-name` is added, multiple matches must fail and print candidates.

## Safety Constraints

- Member list/search/get are read-only.
- Work item create/update remain write operations and must honor existing `--dry-run`, `--yes`, and profile safety behavior.
- Never print tokens or service connection IDs.
- Do not log raw request headers.
- Do not use member names as stable identifiers in write operations.
- Do not silently assign to the current user.

## Output Contracts

### Member JSON

Stable JSON fields should be:

```json
{
  "userId": "string",
  "memberId": "string",
  "name": "string",
  "email": "string",
  "status": "string",
  "roleIds": ["string"]
}
```

Project member extensions can later add:

```json
{
  "projectRoleId": "string",
  "projectRoleName": "string"
}
```

### Member Table

Default table columns:

```text
USER_ID  NAME  EMAIL  STATUS  ROLES
```

Use `USER_ID` first because it is the value users need for `--assignee`.

## Implementation Plan

1. Add app-level member models and service/use case interfaces.
2. Add Platform adapter methods:
   - `ListMembers`
   - `SearchMembers`
   - `GetMemberByUserID`
3. Add `yx member` command with `list`, `search`, and `get`.
4. Add assignee resolution helper:
   - passthrough for normal user IDs.
   - `@me` -> current user `id`.
5. Extend `CreateWorkitemInput` with `Assignee`.
6. Extend `workitemCreateRequest` with `assignedTo`.
7. Extend `UpdateWorkitemInput` with `Title`, `Description`, `DescriptionFormat`.
8. Extend `workitemUpdateRequest` with `subject`, `description`, `formatType`.
9. Add actionable error hint for missing assignee server errors.
10. Update README and `skills/yx-cli/SKILL.md`.

## Test Plan

Unit tests:

- `member list/search/get` command parsing and rendering.
- Member adapter decodes both array and envelope response shapes if observed.
- `--assignee @me` resolves current user ID.
- `--assignee <user-id>` passes through without lookup.
- Work item create request includes `assignedTo`.
- Work item update request includes `subject`, `description`, and `formatType`.
- Invalid description format fails locally.
- Missing assignee server error appends the hint.

Integration-style tests with `httptest`:

- Platform member endpoints.
- Projex create/update request body assertions.
- Error body redaction still protects tokens.

Manual smoke tests:

```bash
yx member list
yx member search --name 王
yx member get --user-id <user-id>
yx issue create --project <project-id> --type Task --title "Smoke" --assignee @me --dry-run
yx issue update <workitem-id> --title "P1 Smoke" --dry-run
```

## Non-Goals

- No interactive member picker in the first version.
- No write-by-name assignment in the first version.
- No automatic default assignee.
- No role management.
- No project-member-only enforcement unless required by server behavior.
- No support for legacy work item APIs unless the Projex API proves insufficient.

## Acceptance Criteria

- Users can list all accessible organization members and see `USER_ID`, name, status, and role IDs.
- Users can search members by name and identify the user ID to assign.
- Users can get a member by `--user-id` and see the name.
- Users can create a work item with `--assignee @me`.
- Users can create a work item with `--assignee <user-id>`.
- Missing-assignee server errors guide users to `--assignee @me` and `yx member search --name`.
- Users can update a work item title through `yx issue update --title`.
- Users can update a work item description through `yx issue update --description --description-format`.
- All new behavior is covered by tests and `go test ./...` passes.
