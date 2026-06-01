# yx CLI 设计文档

日期：2026-06-01

## 目标

`yx` 是一个用 Go 开发的阿里云云效命令行客户端，面向日常研发工作流。它提供接近 `gh` 的使用体验，同时保留云效自身的领域模型：Codeup 代码库与合并请求、Flow 流水线、Projex 项目与工作项。

第一版聚焦常见研发闭环：

- 登录认证与本地 profile 管理。
- 查看、详情、克隆 Codeup 代码库。
- 查看、详情、创建、合并 Codeup 合并请求。
- 查看、详情、创建、更新 Projex 工作项。
- 查看、详情、运行 Flow 流水线，并查看流水线日志。
- 在语义明确的地方提供 GitHub 风格 alias：`pr` 对应合并请求，`issue` 对应工作项。

## 范围

### 第一版范围

- 使用 Go 开发单二进制 CLI。
- 使用 Cobra 作为命令框架。
- 支持多 profile 配置。
- MVP 使用个人访问令牌认证。
- 认证层预留 provider 抽象，后续可扩展 OAuth 或浏览器登录。
- token 优先存入系统 keychain；不可用时降级为权限受限的本地文件。
- 默认输出终端可读表格，同时支持机器可读的 `--json` 输出。
- 开发方式采用 Outside-In TDD。
- 为 MVP 所需云效 OpenAPI 提供 API adapter。
- 仓库克隆通过调用系统 `git` 可执行文件完成。
- 写操作提供安全控制，包括 `--dry-run`、`--yes` 和 profile 级写操作确认开关。

### 第一版不做

- 长期运行或双向代码仓库同步。
- 完整看板视图、列、泳道、布局配置。
- 依赖阿里云 DevOps MCP Server。
- 第一版实现 OAuth。
- 覆盖云效全部 OpenAPI。
- 替代云效管理后台，例如组织级权限管理。

## 功能需求

### FR1：认证与 Profile

- CLI 必须支持 `yx auth login`、`yx auth status`、`yx auth logout`。
- MVP 登录流程必须接受云效个人访问令牌。
- CLI 必须支持多个命名 profile，并支持当前激活 profile。
- CLI 必须支持通过 `--profile` 覆盖单次命令使用的 profile。
- CLI 不得在 stdout、stderr、日志、错误信息或测试快照中打印 token 值。

### FR2：配置

- CLI 默认必须读写 `~/.config/yx/config.yaml`。
- CLI 必须支持 `yx config list`、`yx config get`、`yx config set`、`yx config use`。
- CLI 的 profile 配置必须支持 domain、organization、region、output mode、safety settings、repository-to-project mapping。
- 未知配置 key 必须被拒绝，除非它属于明确声明可扩展的命名空间。

### FR3：代码库命令

- CLI 必须支持列出 Codeup 代码库。
- CLI 必须支持查看单个代码库详情。
- CLI 必须支持通过 Codeup 元数据解析 clone URL，并调用系统 `git` 可执行文件克隆代码库。
- CLI 不得自行实现 Git 传输协议。

### FR4：合并请求命令

- CLI 必须支持列出某个代码库的合并请求。
- CLI 必须支持查看合并请求详情。
- CLI 必须支持基于 source branch 和 target branch 创建合并请求。
- CLI 必须支持合并合并请求。
- CLI 必须提供 `pr` 作为 `mr` 的 alias。

### FR5：项目与工作项命令

- CLI 必须支持列出 Projex 项目。
- CLI 必须支持列出、查看、创建、更新工作项。
- CLI 必须提供 `issue` 作为 `workitem` 的 alias。
- 工作项命令必须以 Projex 项目作为权威上下文。
- 基于仓库的 issue 命令必须要求显式配置 repository-to-project mapping。
- 如果缺少 repository-to-project mapping，命令必须在发送任何云效 API 请求前失败。

### FR6：流水线命令

- CLI 必须支持列出流水线。
- CLI 必须支持查看流水线详情。
- CLI 必须支持针对某个 branch 运行流水线。
- CLI 必须支持查看流水线运行日志。
- 日志跟随可以采用流式或轮询实现，但 `yx pipeline logs <run-id> --follow` 的命令契约必须稳定。

### FR7：输出

- 默认人类可读输出必须采用适合终端阅读的表格样式。
- 列表和详情命令必须支持 `--json`。
- JSON 输出必须使用适合脚本消费的稳定字段名。
- 错误信息必须写入 stderr，并返回非零退出码。

### FR8：写操作安全

- 所有写命令必须支持 `--dry-run`。
- 写命令必须遵守 `profiles.<name>.safety.confirmWrites`。
- 当需要确认时，写命令必须支持 `--yes` 跳过确认。
- 在非交互环境中，如果写命令需要确认但没有传入 `--yes`，命令必须失败。

### FR9：可测试性

- 每个 MVP 命令在实现前必须至少有一个 CLI contract test。
- CLI contract test 必须使用 fake application service，不得依赖云效凭据或网络。
- 每个 Yunxiao adapter 必须有 mock server 测试，覆盖成功与失败响应。
- `go test ./...` 必须能离线通过。

## 架构

第一版采用带清晰内部边界的 Go 单体 CLI。Command handler 保持轻量，只负责解析 flag、加载上下文、调用 application service、渲染输出。云效 HTTP 行为隔离在 adapter 包中。Git 内容操作委托给本地 `git` 可执行文件。

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

### 依赖方向

```text
CLI command
  -> app use case
  -> app port interface
  -> Yunxiao API adapter or Git adapter
  -> external system
```

CLI 层和 application 层依赖 port interface，而不是具体 HTTP client。Yunxiao adapter 负责实现这些 interface。

这个方向保证可测试性，也避免 OpenAPI 细节泄漏到命令语义中。

## 命令面

### 认证

```bash
yx auth login
yx auth status
yx auth logout
```

MVP 登录接受云效个人访问令牌，并把 token 保存到当前或指定 profile。

### 配置

```bash
yx config list
yx config get <key>
yx config set <key> <value>
yx config use <profile>
```

相关命令支持的全局 flag：

```bash
--profile <name>
--org <organization>
--domain <url>
--json
--verbose
```

### 代码库

```bash
yx repo list
yx repo view <repo>
yx repo clone <repo>
```

`repo clone` 通过 Codeup 元数据解析仓库 clone URL，并调用系统 `git` 可执行文件。它不在 Go 中实现 Git 协议。

### 合并请求

```bash
yx mr list --repo <repo>
yx mr view <mr-id> --repo <repo>
yx mr create --repo <repo> --source <branch> --target <branch> --title <title>
yx mr merge <mr-id> --repo <repo>
```

Alias：

```bash
yx pr ...
```

`pr` 是 `mr` 的命令 alias，输出和行为保持一致。

### 项目与工作项

```bash
yx project list
yx workitem list --project <project>
yx workitem view <workitem-id>
yx workitem create --project <project> --type <type> --title <title>
yx workitem update <workitem-id> --status <status> --assignee <user>
```

Alias：

```bash
yx issue ...
```

`issue` 是 `workitem` 的 alias，但云效工作项仍然是源领域模型。

工作项的权威上下文是 Projex 项目。只有在本地配置了仓库到项目的映射时，才允许 `yx issue list --repo <repo>`。如果缺少映射，命令必须给出明确错误：

```text
repo "foo" is not mapped to a project; run yx config set repo.foo.project <project-id>
```

CLI 不得通过扫描组织内所有项目来猜测这个映射。

### 流水线

```bash
yx pipeline list
yx pipeline view <pipeline-id>
yx pipeline run <pipeline-id> --branch <branch>
yx pipeline logs <run-id> --follow
```

## 配置模型

默认配置路径：

```text
~/.config/yx/config.yaml
```

配置结构：

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

规则：

- `current` 表示当前激活 profile。
- `--profile` 对单次调用覆盖 `current`。
- `--domain` 和 `--org` 对单次调用覆盖 profile 中的值。
- `region` 由 endpoint resolver 使用，用于处理中心版和 Region 版 endpoint 差异。
- 未知配置 key 必须被拒绝，除非它位于明确声明可扩展的命名空间下。

## 认证模型

MVP 使用 PAT 认证：

```bash
yx auth login --profile default
```

用户粘贴云效个人访问令牌。token 存储在主 YAML 配置之外。

token 存储优先级：

1. 系统 keychain。
2. 权限为 `0600` 的本地文件 fallback。

`yx auth status` 展示当前 profile、organization、domain、token 是否存在、token 存储后端。它绝不打印 token。

auth 包暴露 provider interface，便于后续增加 OAuth 或浏览器登录，而不需要修改 command handler：

```go
type Provider interface {
    Login(ctx context.Context, profile config.Profile) (Token, error)
    Status(ctx context.Context, profile config.Profile) (Status, error)
    Logout(ctx context.Context, profile config.Profile) error
}
```

具体 interface 可以在实现过程中演进，但边界保持不变：command handler 依赖 auth provider 抽象，而不是直接依赖 PAT 存储。

## 安全模型

默认行为接近 `gh`：用户显式执行写命令时，命令直接执行，不额外确认。

安全控制：

- 所有写操作都支持 `--dry-run`，打印将要执行的操作，不发送写请求。
- 当需要确认时，`--yes` 跳过确认。
- `profiles.<name>.safety.confirmWrites: true` 会要求合并 MR、更新工作项、运行流水线等写操作进行确认。
- 未来如果加入删除、强制覆盖等高风险操作，无论 `confirmWrites` 如何配置，都必须确认。
- 在非交互环境中，如果需要确认但没有传入 `--yes`，命令失败。

MVP 不包含删除代码库等破坏性命令。

## Yunxiao API Client 设计

基础 client 负责：

- Base URL。
- Organization 上下文。
- Region 与 endpoint 解析。
- 通过 `x-yunxiao-token` 注入 token。
- HTTP client。
- 分页 helper。
- 结构化 API 错误转换。

领域 adapter：

```text
codeup.RepositoryAdapter
codeup.MergeRequestAdapter
flow.PipelineAdapter
projex.ProjectAdapter
projex.WorkitemAdapter
```

adapter 返回内部领域模型，不把原始 OpenAPI response struct 暴露给 CLI handler。

### Endpoint 解析

中心版和 Region 版 endpoint 差异在 application 层以下解决。Command handler 不构造 OpenAPI path。

如果 profile 数据不完整，导致无法解析 endpoint，命令必须返回可操作错误，并指出缺失字段。

### 错误处理

错误归一为以下类别：

- 认证错误：token 无效或缺失。
- 授权错误：token 权限不足或组织上下文不匹配。
- Not found：代码库、项目、工作项、合并请求、流水线或运行实例不存在。
- 校验错误：命令参数无效或 API 侧校验失败。
- 限流或临时服务错误。
- 未知 API 错误。

Verbose 模式可以打印 request path 和 request ID，但不得打印 token 或密钥。

## 数据流示例

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

如果缺少映射，app 层返回 mapping error，并且不发送 API 请求。

### `yx pipeline run pipe1 --branch main --dry-run`

```text
CLI parses flags
  -> app builds intended pipeline run operation
  -> safety layer detects dry run
  -> output renders operation summary
  -> no API request is sent
```

## Outside-In TDD 策略

开发从用户可见行为向内推进。

### 1. 先写 CLI Contract Test

每个 MVP 命令先从 CLI 层测试开始，验证：

- 命令存在。
- 必填 flag 与参数。
- 退出码。
- stdout 与 stderr 行为。
- 设置 `--json` 时的 JSON schema。
- 写命令的确认和 dry-run 行为。

CLI 测试使用 fake application service，不依赖云效 token 或网络。

### 2. Application Use Case 测试

Application 测试覆盖不应该放在 CLI parser 里的命令语义：

- 仓库到项目的映射。
- dry-run 操作摘要。
- 确认策略判断。
- workitem alias 行为。
- merge request alias 行为。
- service result 输出前的规范化。

### 3. Adapter Mock Server 测试

Yunxiao adapter 使用 mock HTTP server 测试。测试必须验证：

- 请求 path。
- query 参数。
- request body。
- `x-yunxiao-token` header。
- 分页。
- API 错误转换。
- 返回错误中没有 token 泄漏。

### 4. 可选真实服务 Smoke Test

后续可以增加真实云效 smoke test，但必须通过显式环境变量开启。它们不属于默认 `go test ./...` 路径。

## 里程碑

### M1：基础能力

- Go module 与 Cobra command skeleton。
- 配置读写。
- Profile 选择。
- 输出 renderer。
- 写操作确认层。
- PAT auth provider 与 token store 抽象。

### M2：代码库命令

- `repo list`
- `repo view`
- `repo clone`

### M3：合并请求命令

- `mr list`
- `mr view`
- `mr create`
- `mr merge`
- `pr` alias

### M4：项目与工作项命令

- `project list`
- `workitem list`
- `workitem view`
- `workitem create`
- `workitem update`
- `issue` alias
- issue alias 命令支持 repo-to-project mapping

### M5：流水线命令

- `pipeline list`
- `pipeline view`
- `pipeline run`
- `pipeline logs`

### M6：打包与文档

- README，包含 setup 和示例。
- Release build 说明。
- 如果风险低，提供 shell completion。
- 可选 install script 或 Homebrew tap 规划。

## 验收标准

- `go test ./...` 在没有网络和云效凭据的情况下通过。
- 每个 MVP 命令至少有一个 CLI contract test。
- 每个 Yunxiao adapter 都有 mock server 测试，覆盖成功和失败响应。
- 列表和详情命令支持 `--json`。
- 默认输出为可读表格。
- 写命令支持 `--dry-run`。
- 写命令遵守 `--yes` 和 `safety.confirmWrites`。
- token 值绝不出现在日志、错误、stdout、stderr 或 golden snapshot 中。
- README 说明 PAT 登录、profile 配置、代码库命令、合并请求命令、工作项命令和流水线命令。

## 已收敛语义

- 第一版范围是 B：日常研发闭环，不做完整平台管理。
- 认证是 PAT-first，并通过 provider 抽象预留未来 OAuth。
- 工作项是项目优先；基于仓库的 issue 命令要求显式本地映射。
- 写操作确认可配置；默认对显式写命令直接执行。
- TDD 风格是 Outside-In：先用 CLI contract test 定义行为，再实现内部 adapter。
- 实现形态是纯 Go 单二进制 CLI。MCP Server 集成是未来扩展，不是运行时依赖。
