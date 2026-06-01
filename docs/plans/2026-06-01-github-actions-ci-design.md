# GitHub Actions CI 与 Tag 构建设计

日期：2026-06-01

## 背景

`yx-cli` 是 Go CLI 项目。当前仓库已有本地验证入口：

- `go test ./...`
- `go build -o yx ./cmd/yx`

仓库已发布到 GitHub，当前默认分支为 `main`。目前没有 `.github/workflows`。

## 目标

建立 GitHub Actions workflow，使仓库具备两个自动化路径：

1. 普通 `push` / `pull_request` 只运行测试。
2. 推送版本 tag 时运行测试，并编译上传 CLI 二进制产物。

## 非目标

本设计不包含：

- 自动创建 GitHub Release。
- 自动生成 checksums。
- 发布 Homebrew、Scoop、NPM 或其他包管理器。
- Linux 构建产物。
- 普通 push 时上传构建产物。
- 修改 Go 代码或 CLI 行为。

## Workflow 触发

建议创建 `.github/workflows/ci.yml`。

触发条件：

- `push` 到 `main`：运行测试。
- `pull_request` 到 `main`：运行测试。
- `push` tag `v*`：运行测试，测试通过后构建并上传产物。

## Job 设计

### test

职责：

- 拉取代码。
- 安装 Go `1.26.3`。
- 运行 `go test ./...`。

触发：

- 所有 workflow 触发场景都运行。

失败语义：

- 测试失败时 workflow 失败。
- tag 场景下，测试失败会阻止 build job 运行。

### build

职责：

- 只在 tag push 时运行。
- 依赖 `test` job。
- 使用 matrix 构建两个目标平台：
  - `GOOS=darwin GOARCH=arm64`
  - `GOOS=windows GOARCH=amd64`
- 将产物放入 `dist/`。
- 使用 artifact 上传构建产物。

产物命名：

- `dist/yx-darwin-arm64`
- `dist/yx-windows-amd64.exe`

artifact 命名：

- `yx-darwin-arm64`
- `yx-windows-amd64`

失败语义：

- 任一目标平台构建失败时 workflow 失败。
- 上传 artifact 失败时 workflow 失败。

## 数据流

```text
push / pull_request
  -> test

tag push v*
  -> test
  -> build matrix
       -> dist/yx-darwin-arm64
       -> dist/yx-windows-amd64.exe
       -> upload artifacts
```

## 权限

初版只需要读取仓库和上传 workflow artifact，不创建 Release，不写仓库内容。

建议显式设置：

```yaml
permissions:
  contents: read
```

## 验证方式

本地验证：

```bash
go test ./...
GOOS=darwin GOARCH=arm64 go build -o dist/yx-darwin-arm64 ./cmd/yx
GOOS=windows GOARCH=amd64 go build -o dist/yx-windows-amd64.exe ./cmd/yx
```

GitHub 验证：

1. push 到 `main` 后，Actions 只运行 test job。
2. 推送形如 `v0.1.0` 的 tag 后，Actions 运行 test job 和 build job。
3. tag workflow 完成后，Artifacts 中存在：
   - `yx-darwin-arm64`
   - `yx-windows-amd64`

## 回滚

如 workflow 行为不符合预期，回滚方式是删除或 revert `.github/workflows/ci.yml`。

如果 tag 已经触发构建，删除 workflow 不会删除已有 tag 或历史 artifact；需要单独在 GitHub Actions 页面清理历史运行记录或 artifact。

## 后续可扩展项

后续可以独立增加：

- checksums 生成。
- GitHub Release 自动创建。
- Release assets 上传。
- Linux 构建。
- Homebrew 或 Scoop 发布。
