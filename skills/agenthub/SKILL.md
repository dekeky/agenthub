---
name: agenthub
description: 浏览 AgentHub 注册表、按 category 筛选 agent 包、安装到本地。当用户需要列出 Hub 上的 agent、查看包详情、安装 picoclaw/openclaw agent 时使用。
metadata:
  requires:
    bins: ["agenthub-cli"]
---

# AgentHub

通过 `agenthub-cli` 连接 AgentHub，按 **category**（`picoclaw`、`openclaw` 等）发现并安装 agent 包。

## 前置条件

- `agenthub-cli` 已安装并在 `PATH` 中（`go build -o agenthub-cli ./cmd/agenthub-cli`）
- 工作目录有 `agenthub-cli.toml`（复制项目根目录 `agenthub-cli.example.toml`）
- 已知本地 runtime category（通常 `picoclaw` 或 `openclaw`）

只读命令（`list`、`categories`、`get`、`file`、`download`）不需要 `upload_token`。`upload`、`update`、`put-file`、`delete` 需要配置 `upload_token`。所有命令加 `--json`，stdout 为 JSON，stderr 为错误。

**Flag 顺序**：所有 flag 必须在位置参数（`<agentName>`、文件路径）之前。

## 命令

### 列出支持的 category

```bash
agenthub-cli --json categories
```

### 按 category 列出 agent

```bash
agenthub-cli --json list --category picoclaw
```

安装前必须先按本地 runtime category 过滤。

### 查看 agent 详情

```bash
agenthub-cli --json get <agentName>
agenthub-cli --json get --version 1.0.0 <agentName>
```

确认返回的 `category` 与本地 runtime 一致，并检查 `files[]`。

### 读取包内文件（可选）

```bash
agenthub-cli --json file <agentName> <path-in-package> [--version VERSION]
```

### 安装 agent（必须校验 category）

```bash
agenthub-cli --json install --expect-category <runtime> [--dest DIR] [--version VERSION] <agentName>
```

- **`--expect-category` 必填**，category 不匹配则中止
- 默认安装路径：`./agents/<agentName>`
- 重复安装会先清空 `--dest`

示例：

```bash
agenthub-cli --json install --expect-category picoclaw --dest ./agents/test test
```

### 上传 agent 包（可选）

```bash
agenthub-cli --json upload --category picoclaw --version 1.0.0 <agentName> <zip-file>
```

| category | ZIP 内必须包含 |
|----------|----------------|
| `picoclaw` | `SKILL.md` 或 `AGENT.md` |
| `openclaw` | `AGENT.md` |

### 更新元信息（可选，需 upload_token）

```bash
agenthub-cli --json update --display-name "Title" --summary "..." [--category picoclaw] <agentName>
```

至少指定 `--display-name`、`--summary`、`--category` 之一。

### 更新包内单文件（可选，需 upload_token）

```bash
agenthub-cli --json put-file <agentName> <path-in-package> --version VERSION [--file LOCAL]
```

未指定 `--file` 时从 stdin 读取内容。

### 删除 agent（可选，需 upload_token）

```bash
agenthub-cli --json delete <agentName>
```

## 标准流程

```bash
# Step 1: 按本地 runtime 筛选
agenthub-cli --json list --category picoclaw

# Step 2: 查看详情，确认 category 和 files
agenthub-cli --json get <agentName>

# Step 3: 安装（不可跳过 --expect-category）
agenthub-cli --json install --expect-category picoclaw --dest ./agents/<agentName> <agentName>

# Step 4: 读取安装目录中的 SKILL.md / AGENT.md 并使用
```

## 注意事项

- **认证**：`upload`、`update`、`put-file`、`delete` 需在 `agenthub-cli.toml` 配置 `upload_token`
- **安全**：`install` 会删除并覆盖 `--dest` 目录
- **失败处理**：`category mismatch` 时不要安装；`config file not found` 时创建 `agenthub-cli.toml`
- 本地使用优先 `install`，`download` 仅保存 ZIP
