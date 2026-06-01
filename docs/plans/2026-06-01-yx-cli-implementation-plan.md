# yx CLI 实施计划

日期：2026-06-01

关联设计文档：[2026-06-01-yx-cli-design.md](./2026-06-01-yx-cli-design.md)

## 实施原则

本计划用于把已确认的 `yx` CLI 设计拆成可执行、可验收、可回滚的原子任务。实现采用 Outside-In TDD：先从用户可见 CLI 行为写测试，再向内实现 application use case、port interface、adapter 和基础设施。

每个任务必须满足：

- 原子性：一次只引入一个可独立理解的行为或基础能力。
- 可测试：任务完成时必须有自动化测试证明行为。
- 可验收：每个任务都有明确命令、测试或文件级验收标准。
- 不越界：不得实现设计文档明确列为非目标的能力。
- 不泄密：token 不得出现在 stdout、stderr、日志、错误、golden snapshot 或测试失败信息中。

## TDD 工作流

每个行为任务遵循红绿重构：

1. Red：先写失败的 CLI contract test 或 application test。
2. Green：写最小实现让测试通过。
3. Refactor：在测试保护下整理结构、命名和重复逻辑。
4. Verify：运行当前包测试；阶段结束运行 `go test ./...`。

如果任务是纯文档、构建脚本或目录整理，允许不执行红绿重构，但必须在任务验收标准中说明原因。

## 分支与提交策略

- 当前仓库处于绿地状态，可在 `main` 上继续推进；如后续接入远端协作，再按功能切分分支。
- 每个里程碑至少一个提交。
- 大任务必须拆成多个小提交，提交信息使用 `type: summary` 格式，例如 `test: add config command contracts`、`feat: implement profile config store`。
- 提交前必须确认 `git status --short` 中没有非预期文件。

## 全局验收标准

- `go test ./...` 在无网络、无云效凭据环境下通过。
- 所有 MVP 命令至少有一个 CLI contract test。
- 所有 Yunxiao adapter 至少有成功和失败路径 mock server test。
- 列表和详情命令支持 `--json`。
- 写命令支持 `--dry-run`。
- 写命令遵守 `--yes` 和 `safety.confirmWrites`。
- 默认输出是终端可读表格。
- 错误输出进入 stderr，并返回非零退出码。
- token 值不出现在任何输出、日志、错误或测试快照中。
- README 覆盖 PAT 登录、profile 配置、代码库、合并请求、工作项和流水线基础命令。

## M1：基础能力

目标：建立可测试的 CLI 骨架、配置系统、输出系统、认证抽象和写操作安全基础。M1 完成后，项目应能离线执行 `go test ./...`，并具备后续命令扩展的稳定边界。

### M1.1 初始化 Go 项目骨架

范围：

- 初始化 `go.mod`。
- 引入 Cobra。
- 创建 `cmd/yx/main.go`。
- 创建 `internal/cli` root command。
- 建立基础测试 helper，用于捕获 stdout、stderr、exit error。

TDD：

- Red：新增 `yx --help` CLI contract test，断言命令名、基础用法、退出码。
- Green：实现最小 root command。
- Refactor：抽出 command 构造函数，避免测试直接调用 `main`。

验收标准：

- `go test ./...` 通过。
- `go run ./cmd/yx --help` 可执行。
- root command 测试不访问网络和用户真实 home 目录。

### M1.2 全局 flag 与执行上下文

范围：

- 支持 `--profile`、`--org`、`--domain`、`--json`、`--verbose`。
- 定义 CLI 运行上下文结构，承载输出 writer、配置路径、profile override 等信息。

TDD：

- Red：测试 root command 能解析全局 flag，并将值传入 fake handler。
- Green：实现全局 flag 绑定。
- Refactor：将 flag 解析结果归一到 `cli.Context` 或等价结构。

验收标准：

- 全局 flag 可被子命令继承。
- 测试可注入临时 config path 和 fake home。
- flag 解析错误进入 stderr。

### M1.3 配置模型与文件存储

范围：

- 实现 `internal/config`。
- 支持默认路径 `~/.config/yx/config.yaml`。
- 支持 `current`、`profiles`、domain、organization、region、output、safety、repoProjectMap。
- 支持临时测试路径覆盖。

TDD：

- Red：写配置 load/save 测试，覆盖空文件、缺失文件、默认值、profile 切换、未知 key。
- Green：实现 YAML 配置读写。
- Refactor：拆分 store、profile、repo mapping。

验收标准：

- 缺失配置文件时返回可初始化的空配置，而不是 panic。
- 保存配置时创建父目录。
- 配置文件权限不宽于 `0600` 或平台允许的等价权限。
- 未知 key 被拒绝。

### M1.4 config 命令

范围：

- 实现 `yx config list`。
- 实现 `yx config get <key>`。
- 实现 `yx config set <key> <value>`。
- 实现 `yx config use <profile>`。

TDD：

- Red：为每个 config 子命令写 CLI contract test。
- Green：实现最小命令行为。
- Refactor：把 key path 解析和 profile mutation 放入 config 包或 app 层。

验收标准：

- `config set profiles.default.domain https://devops.aliyun.com` 后可读取。
- `config use default` 会更新 `current`。
- 缺少参数返回非零退出码。
- `--json` 输出结构稳定。

### M1.5 输出系统

范围：

- 实现 `internal/output`。
- 支持 table 输出。
- 支持 JSON 输出。
- 支持错误输出约定。

TDD：

- Red：写 table golden test 和 JSON schema/字段测试。
- Green：实现 renderer。
- Refactor：统一 list/detail 输出入口。

验收标准：

- 默认输出为表格。
- `--json` 输出合法 JSON。
- JSON 字段顺序不作为测试依赖。
- golden snapshot 不包含绝对临时路径或 token。

### M1.6 安全确认层

范围：

- 实现 `internal/safety`。
- 支持 `--dry-run`、`--yes`、`confirmWrites`、TTY/非 TTY 判断。

TDD：

- Red：写 safety decision table 测试。
- Green：实现确认决策。
- Refactor：把交互输入和 TTY 检测抽象为接口，方便测试。

验收标准：

- `confirmWrites=false` 时显式写命令默认执行。
- `confirmWrites=true` 且无 `--yes` 时需要确认。
- 非 TTY 且需要确认时失败。
- `--dry-run` 不触发真实 service 调用。

### M1.7 认证抽象与 token store

范围：

- 定义 auth provider interface。
- 实现 PAT provider。
- 实现 token store interface。
- 先实现 file fallback。
- 预留 keychain store 接口与平台实现入口。

TDD：

- Red：写 `auth status/login/logout` CLI contract test，使用 fake provider。
- Red：写 file token store 测试，覆盖权限和 token 不泄漏。
- Green：实现最小 auth 命令和 file store。
- Refactor：拆分 provider、store、status model。

验收标准：

- `auth status` 不打印 token。
- `auth login` 可从测试输入读取 token。
- `auth logout` 删除当前 profile token。
- file fallback 权限为 `0600`。
- 错误信息不包含 token 原文。

## M2：代码库命令

目标：完成 Codeup repository 查询和 clone 的用户路径，并建立第一个 Yunxiao adapter 样板。

### M2.1 repo port 与 app use case

范围：

- 定义 repository service port。
- 实现 `ListRepositories`、`GetRepository`、`CloneRepository` use case。
- Git clone 通过 `gitx` port 抽象，不直接在 app 层执行命令。

TDD：

- Red：写 app use case 测试，使用 fake repository service 和 fake git runner。
- Green：实现 use case。
- Refactor：统一 repo identifier 解析。

验收标准：

- repo use case 不依赖 HTTP。
- clone use case 会先解析 clone URL，再调用 git runner。
- clone 失败返回可读错误。

### M2.2 repo CLI contract

范围：

- 实现 `yx repo list`。
- 实现 `yx repo view <repo>`。
- 实现 `yx repo clone <repo>`。

TDD：

- Red：为三个命令写 CLI contract test。
- Green：接入 app use case fake。
- Refactor：统一 table/JSON 输出模型。

验收标准：

- list/detail 支持 `--json`。
- 缺少 repo 参数时返回非零退出码。
- clone 命令测试不执行真实 git。

### M2.3 Codeup repository adapter

范围：

- 实现 Yunxiao repository adapter。
- 覆盖认证 header、分页、成功响应、错误响应。

TDD：

- Red：mock server 测请求 path、query、`x-yunxiao-token`。
- Green：实现 adapter。
- Refactor：抽出基础 request helper。

验收标准：

- adapter 测试离线通过。
- 401/403/404/5xx 转换为结构化错误。
- token 不出现在错误字符串中。

## M3：合并请求命令

目标：完成 `mr` 和 `pr` alias 的核心工作流。

### M3.1 mr app use case

范围：

- 定义 merge request service port。
- 实现 list、view、create、merge use case。
- 写操作接入 safety 和 dry-run。

TDD：

- Red：写 use case 测试，覆盖 create、merge、dry-run、confirmWrites。
- Green：实现 use case。
- Refactor：统一 write operation summary。

验收标准：

- dry-run 不调用 service 写方法。
- merge 在需要确认但未确认时失败。
- create 参数校验在调用 service 前完成。

### M3.2 mr/pr CLI contract

范围：

- 实现 `yx mr list/view/create/merge`。
- 实现 `yx pr` alias。

TDD：

- Red：写 CLI contract test，覆盖 `mr` 和 `pr` 行为一致。
- Green：实现命令。
- Refactor：复用同一 command builder。

验收标准：

- `pr list` 与 `mr list` 调用同一 use case。
- list/view 支持 `--json`。
- create/merge 支持 `--dry-run`。
- 必填参数缺失时不调用 service。

### M3.3 Codeup merge request adapter

范围：

- 实现 merge request adapter。
- 覆盖 list、get、create、merge。

TDD：

- Red：mock server 测成功、分页、校验错误、权限错误。
- Green：实现 adapter。
- Refactor：复用 Codeup endpoint resolver。

验收标准：

- 所有写请求 body 被测试覆盖。
- API 业务错误被转换为结构化错误。
- request ID 在 verbose 可用。

## M4：项目与工作项命令

目标：完成 Projex project/workitem 以及 `issue` alias，明确 repo-to-project mapping 语义。

### M4.1 project/workitem app use case

范围：

- 定义 project service port。
- 定义 workitem service port。
- 实现 project list。
- 实现 workitem list/view/create/update。
- 实现 repo-to-project mapping 解析。

TDD：

- Red：写 use case 测试，覆盖 project-first、repo mapping、mapping 缺失不调用 service。
- Green：实现 use case。
- Refactor：抽出 mapping resolver。

验收标准：

- `issue list --repo foo` 缺少 mapping 时不发 API 请求。
- `workitem list --project p1` 不依赖 repo mapping。
- write use case 支持 dry-run 和确认策略。

### M4.2 workitem/issue CLI contract

范围：

- 实现 `yx project list`。
- 实现 `yx workitem list/view/create/update`。
- 实现 `yx issue` alias。

TDD：

- Red：写 CLI contract test，覆盖 alias、缺参、JSON、dry-run。
- Green：实现命令。
- Refactor：复用 workitem command builder。

验收标准：

- `issue` 与 `workitem` 输出模型一致。
- `issue --repo` 缺少 mapping 时错误文案符合设计文档。
- create/update 支持 `--dry-run`。

### M4.3 Projex adapter

范围：

- 实现 project adapter。
- 实现 workitem adapter。
- 覆盖 search/list/get/create/update。

TDD：

- Red：mock server 测 path、query、body、错误转换。
- Green：实现 adapter。
- Refactor：统一 Projex response normalization。

验收标准：

- search/list 支持分页。
- create/update 请求体被测试覆盖。
- 404 能区分 project/workitem 不存在。

## M5：流水线命令

目标：完成 Flow pipeline 查询、运行和日志查看。

### M5.1 pipeline app use case

范围：

- 定义 pipeline service port。
- 实现 list、view、run、logs。
- `run` 接入 dry-run 和确认策略。
- `logs --follow` 保持命令契约稳定，底层可轮询或流式。

TDD：

- Red：写 use case 测试，覆盖 run dry-run、confirmWrites、logs follow 标记。
- Green：实现 use case。
- Refactor：抽象 log reader。

验收标准：

- dry-run 不触发真实 pipeline run。
- logs 能输出多行日志。
- follow 行为由接口抽象，测试不等待真实时间。

### M5.2 pipeline CLI contract

范围：

- 实现 `yx pipeline list/view/run/logs`。

TDD：

- Red：写 CLI contract test，覆盖 JSON、缺参、dry-run、follow。
- Green：实现命令。
- Refactor：统一 run operation 输出。

验收标准：

- list/view 支持 `--json`。
- run 支持 `--branch` 和 `--dry-run`。
- logs 支持 `--follow`。

### M5.3 Flow adapter

范围：

- 实现 pipeline adapter。
- 覆盖 list、get、run、logs。

TDD：

- Red：mock server 测成功、错误、日志响应。
- Green：实现 adapter。
- Refactor：复用基础 pagination/error helper。

验收标准：

- run 请求体被测试覆盖。
- logs 支持非 follow 和 follow 所需的抽象。
- 5xx/限流错误转换清晰。

## M6：文档、打包与发布准备

目标：让项目具备可安装、可试用、可维护的基础交付面。

### M6.1 README

范围：

- 写 README。
- 覆盖安装、PAT 登录、profile 配置、常用命令、JSON 输出、dry-run、安全确认。

TDD：

- 文档任务不执行红绿重构。
- 使用命令帮助输出或 CLI contract 中的命令面校验 README 示例不漂移。

验收标准：

- README 中所有示例命令在 CLI 中存在。
- README 不包含真实 token、组织 ID 或私有仓库信息。

### M6.2 构建与安装

范围：

- 添加 Makefile 或等价脚本。
- 支持 test、build、lint 可选入口。
- 生成本地二进制。

TDD：

- Red：如果添加构建脚本测试，先写脚本存在性或 smoke test。
- Green：实现构建入口。
- Refactor：保持命令简单，不引入复杂发布系统。

验收标准：

- `go test ./...` 通过。
- `go build ./cmd/yx` 成功。
- 本地二进制可执行 `--help`。

### M6.3 Release 准备

范围：

- 记录 release build 命令。
- 可选增加 shell completion。
- 可选规划 Homebrew tap，不在 MVP 中强制实现。

TDD：

- 文档和 release note 任务不强制红绿重构。

验收标准：

- 发布说明能从干净 checkout 复现 build。
- 可选项不阻塞 MVP 验收。

## 任务执行顺序

严格按以下顺序推进，除非发现设计缺陷需要回到 spec 修订：

1. M1.1 初始化 Go 项目骨架。
2. M1.2 全局 flag 与执行上下文。
3. M1.3 配置模型与文件存储。
4. M1.4 config 命令。
5. M1.5 输出系统。
6. M1.6 安全确认层。
7. M1.7 认证抽象与 token store。
8. M2 代码库命令。
9. M3 合并请求命令。
10. M4 项目与工作项命令。
11. M5 流水线命令。
12. M6 文档、打包与发布准备。

## 每个任务完成时的检查清单

- 是否先写了失败测试。
- 是否只实现了让测试通过的最小代码。
- 是否完成了必要重构。
- 是否运行了相关包测试。
- 是否需要运行 `go test ./...`。
- 是否更新了文档或示例。
- 是否确认没有 token、私有组织信息或临时路径进入快照。
- 是否检查了 `git status --short`。

## 需要及时沟通的语义不收敛点

以下情况出现时，暂停实现并回到设计确认：

- 云效 OpenAPI 的真实字段与设计中的命令契约冲突。
- 某个命令需要在 project、repository、organization 三种上下文之间自动猜测。
- 某个写操作无法可靠支持 `--dry-run`。
- token store 的 keychain 行为在 macOS/Linux/CI 上无法形成一致抽象。
- pipeline logs 的真实 API 无法支持当前 `--follow` 契约。
- adapter 需要引入全局状态或绕过 app port interface。
- 为了通过测试需要放宽 token 不泄漏约束。

## 首个实现切入点

从 M1.1 开始：

1. 写 `yx --help` CLI contract test。
2. 初始化 Go module 和 Cobra root command。
3. 让测试通过。
4. 运行 `go test ./...`。
5. 提交 `test/feat` 组合或单个基础提交。

该切入点最小、可验证，并为后续所有命令提供统一测试入口。
