# AgentHub

多运行时 Agent 仓库：上传、下载、浏览 Agent 包，按 **category** 区分运行时（如 `picoclaw`、`openclaw`）。

## 功能

- **Hub API**（`ginx` 统一响应）：列表、详情、文件读取、上传、下载
- **本地存储**：`storage/agents/{agentName}/versions/{version}/` 保存 Agent 包全量文件

## 包格式与类别

**category**（Agent 运行时类别）：

| category | 说明 |
|----------|------|
| `picoclaw` | PicoClaw 格式（默认） |
| `openclaw` | OpenClaw 格式 |
| 其他 | 小写字母/数字/连字符，规则同 agentName |

上传时通过 form 字段 `category` 指定，省略则默认为 `picoclaw`。列表 API 支持 `?category=picoclaw` 过滤。

**包内容校验**（上传时自动检查）：

| category | ZIP 内必须包含 |
|----------|----------------|
| `picoclaw` | `SKILL.md` 或 `AGENT.md` |
| `openclaw` | `AGENT.md` |

查询支持的 category：`agenthub-cli categories` 或 `GET /api/hub/categories`

**包内容示例**（PicoClaw）：

- Skill 包：包含 `SKILL.md` 及附属文件
- Workspace 包：包含 `AGENT.md`（及可选 `skills/` 等目录）

上传时使用 ZIP 压缩包。

## 快速开始

**服务端：**

```bash
cp agenthub-server.example.toml agenthub-server.toml
# 编辑 agenthub-server.toml，设置 upload_token 等
go run ./cmd/agenthub
```

**CLI：**

```bash
cp agenthub-cli.example.toml agenthub-cli.toml
go build -o agenthub-cli ./cmd/agenthub-cli
```

### agenthub-server.toml

```toml
addr = ":8080"
storage_dir = "./storage"
upload_token = "your-secret-token"
```

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `addr` | `:8080` | 监听地址 |
| `storage_dir` | `./storage` | 本地存储目录（相对配置文件目录） |
| `upload_token` | _(空)_ | 上传鉴权 token，未设置则禁止上传 |

服务端默认读取 `./agenthub-server.toml`，可用 `--config` 或 `AGENTHUB_SERVER_CONFIG` 指定路径。

### agenthub-cli.toml

```toml
url = "http://localhost:8080"
upload_token = "your-secret-token"
```

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `url` | `http://localhost:8080` | Hub 地址 |
| `upload_token` | _(空)_ | 上传时使用的 token，需与服务端一致 |

CLI 默认读取 `./agenthub-cli.toml`，可用 `--config` 或 `AGENTHUB_CLI_CONFIG` 指定路径。

## CLI（供 Agent 调用）

```bash
# 浏览
agenthub-cli categories
agenthub-cli list
agenthub-cli list --category openclaw
agenthub-cli get demo-weather
agenthub-cli file demo-weather SKILL.md

# 包管理
agenthub-cli download demo-weather -o demo.zip
agenthub-cli install --expect-category picoclaw --dest ./agents/test test
agenthub-cli upload --category openclaw --version 1.0.0 my-skill ./pkg.zip

# 脚本友好：统一 JSON 输出
agenthub-cli --json list
```

加 `--json` 时 stdout 输出结构化 JSON，stderr 输出错误，便于 Agent 解析。

### Agent Skill

`skills/agenthub/SKILL.md` 供 PicoClaw Agent 使用：教 Agent 直接调用 `agenthub-cli` 列出、查看、安装 Hub 上的 agent 包。打包为 ZIP 后上传到 Hub 即可分发。

## API

### Hub（ginx 响应 `{code, errMsg, body}`）

```bash
# 列表
curl http://localhost:8080/api/hub/agents

# 详情（含文件树）
curl http://localhost:8080/api/hub/agents/my-skill

# 读取文件（用于展示 SKILL.md）
curl http://localhost:8080/api/hub/agents/my-skill/files/SKILL.md

# 上传
curl -X POST http://localhost:8080/api/hub/agents \
  -F "agentName=my-skill" \
  -F "category=openclaw" \
  -F "version=1.0.0" \
  -F "file=@package.zip"

# 下载 ZIP
curl -OJ "http://localhost:8080/api/hub/agents/my-skill/download"
```

## 存储结构

```
storage/
  agents/
    my-skill/
      meta.json
      versions/
        1.0.0/
          SKILL.md
          ...
```
