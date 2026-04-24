# GitHub PR 事件处理流程验证（最新更新 2026-04-18）

## 📊 实现状态摘要

### ✅ 已完成的基础设施

- ✅ Webhook 接收和验证
- ✅ 事件解析和标准化
- ✅ 事件发布到 NATS JetStream
- ✅ Orchestrator 消费和路由
- ✅ Eino Agent Runtime
- ✅ GitHub Tools（PR 读取、文件对比、Review 发布）
- ✅ Skills Catalog（github-pr-review 技能）

### 🔄 进行中

- 🔄 端到端 PR Review 流程验证
- 🔄 OAuth 凭证集成测试

---

## 闭环验证清单

### ✓ Webhook 接收层 (`backend/interaction/connectors/github/`)

- [x] `/github/webhook` 端点已注册 (`github.go:43`)
- [x] 支持多种 GitHub 事件类型 (包括 pull_request)
- [x] HMAC-SHA256 签名验证机制 (`webhook.go`)
- [x] OAuth 回调端点 (`/github/auth`, `/github/callback`)

### ✓ 事件解析层

- [x] 支持 PR 事件 (`pull_request`) 类型解析 (`events.go`)
    - `opened`
    - `synchronize`
    - `reopened`
    - `ready_for_review`
- [x] 自动映射为内部 `interaction.Event` 结构
- [x] 正确填充 Payload 数据 (repo, actor, context)

### ✓ 事件发布层 (`backend/internal/infra/mq/`)

- [x] 发布到 NATS JetStream topic: `interaction.github.pull_request`
- [x] 支持异步消息传输
- [x] Publisher 接口抽象 (`mq/bus.go`)

### ✓ 事件消费层 (`backend/orchestrator/orchestrator.go`)

- [x] Orchestrator 订阅 PR 事件主题
- [x] 路由到 Agent Runtime (`HandleEvent`)
- [x] 支持多事件类型分发

### ✓ Agent Runtime (`backend/runtime/eino_runner.go`)

- [x] Eino LLM 集成
- [x] Skills Catalog 注入
- [x] Tools Registry 集成
- [x] Auth 凭证解析
- [x] 系统提示词定制（根据事件类型）

### ✓ Tools 系统 (`backend/tools/`)

- [x] PR 元数据读取 (`github.pr.get_metadata`)
- [x] PR 文件列表获取 (`github.pr.get_files`)
- [x] Commit 对比工具 (`github.repo.compare_commits`)
- [x] 文件内容读取 (`github.repo.get_file`)
- [x] PR Review 发布 (`github.pr.publish_review`)
- [x] 账户信息工具 (`github.account_info`)

### ✓ Skills 系统 (`backend/skills/`)

- [x] Skill 接口定义
- [x] 文件化技能支持 (SKILL.md)
- [x] Catalog 扫描和加载 (`backend/tools/skill`)
- [x] PR Review 技能定义 (`backend/skills/github-pr-review/SKILL.md`)

---

## 完整事件处理流程

```
1. GitHub Webhook
   ↓
   POST https://your-domain/github/webhook
   Headers: X-Hub-Signature-256, X-GitHub-Event
   Body: { "action": "opened", "pull_request": {...} }
   
2. Webhook Handler (backend/interaction/connectors/github/webhook.go)
   ↓
   a. 验证 HMAC 签名
   b. 解析事件类型 (pull_request)
   c. 提取 actor, repository
   d. 构建 interaction.Event
   
3. Event Publishing (backend/interaction/connectors/github/events.go)
   ↓
   publisher.Publish(ctx, "interaction.github.pull_request", event)
   
4. NATS JetStream Transport (backend/internal/infra/mq/)
   ↓
   Topic: interaction.github.pull_request
   
5. Orchestrator Consumer (backend/orchestrator/orchestrator.go)
   ↓
   subscriber.Subscribe(ctx, "interaction.github.pull_request", handler)
   ↓
   handler: runner.HandleEvent(ctx, event)
   
6. Eino Runner (backend/runtime/eino_runner.go)
   ↓
   a. 构建 AgentRunner
      - System Prompt (PR Review 指令)
      - Skills Context (github-pr-review)
      - Tools Context (GitHub API tools)
   b. 调用 LLM 生成响应
      - LLM 决定调用哪些工具
      - 工具执行并返回结果
   c. LLM 分析工具输出
   d. 生成最终 Review
   
7. Tool Execution (backend/toolruntime/runtime.go)
   ↓
   a. 解析 Auth Account
      - 检查 ExternalRefs (installation_id)
      - 获取 OAuth 凭证
   b. 创建 GitHub Client
   c. 执行具体工具
      例: github.pr.get_metadata
          github.pr.get_files
          github.pr.publish_review
          
8. GitHub API Call
   ↓
   GET /repos/{owner}/{repo}/pulls/{pr_number}
   GET /repos/{owner}/{repo}/pulls/{pr_number}/files
   POST /repos/{owner}/{repo}/pulls/{pr_number}/reviews
   Body: { "body": "Review content", "event": "COMMENT" }
   
9. Result Processing
   ↓
   a. 记录执行日志
   b. 返回结果给 Orchestrator
   c. (Future) 记录到 Event 表
   
10. Developer Sees Review
    ↓
    在 GitHub PR 页面看到 AI Review 评论
```

---

## 主要代码结构

### 1. Orchestrator 组件 (`backend/orchestrator/`)

- `orchestrator.go`: 核心事件消费者
  - 订阅多个 topic
  - 路由到统一的 Agent Runtime
  - 支持自定义 Handler 注册

### 2. GitHub 连接器 (`backend/interaction/connectors/github/`)

- `github.go`: Connector 主结构
- `webhook.go`: Webhook 接收和验证
- `events.go`: 事件解析和发布
- `converter.go`: GitHub 对象转换
- `types.go`: 类型定义

### 3. Agent Runtime (`backend/runtime/`)

- `eino_runner.go`: Eino LLM 执行引擎
- `runner.go`: Runner 接口定义
- `eino/`: Eino 适配器
  - `chatmodel.go`: LLM 模型
  - `agent_runner.go`: Agent 执行器
  - `tool_adapter.go`: 工具适配器
- `prompt/`: 提示词构建
  - `skills.go`: Skills 上下文
  - `tools.go`: Tools 上下文

### 4. Tools 系统 (`backend/tools/`)

- `registry.go`: 工具注册表
- `tool.go`: Tool 接口
- `github/`: GitHub 工具集
  - `pr_read.go`: PR 读取工具
  - `pr_write.go`: PR 写入工具
  - `compare.go`: Commit 对比
  - `common.go`: 通用工具函数

### 5. Skills 系统

- `backend/tools/skill/catalog.go`: Skills Catalog
- `backend/tools/skill/types.go`: Skill 元数据
- `backend/skills/`: 技能目录
  - `github-pr-review/SKILL.md`: PR Review 技能定义

### 6. Auth 系统 (`backend/auth/`)

- `service.go`: Auth 服务
- `store.go`: Store 接口
- `memory_store.go`: 内存存储实现
- `providers/github/`: GitHub OAuth Provider
- `resolver.go`: 账户解析器

### 7. 事件主题定义 (`backend/interaction/topic.go`)

```go
const (
    TopicGithubPullRequest  = "interaction.github.pull_request"
    TopicGithubIssueComment = "interaction.github.issue_comment"
    TopicGithubPush         = "interaction.github.push"
)
```

---

## 完整性测试

### 构建测试

```bash
# 构建主服务
go build -o ./bundles/singer ./backend/cmd/singer/main.go

# 构建 Skill Proxy
go build -o ./bundles/skill-proxy ./backend/cmd/skill-proxy/main.go
```

### 单元测试

```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./backend/orchestrator/...
go test ./backend/runtime/...
go test ./backend/tools/...
```

### 集成测试（本地）

```bash
# 1. 启动 Docker Compose
docker-compose up -d

# 2. 配置 GitHub App
# - 设置 Webhook URL
# - 配置 Secret
# - 授权 OAuth

# 3. 创建测试 PR
# 在连接的 GitHub 仓库创建 PR

# 4. 查看日志
docker-compose logs -f singer

# 5. 验证 PR Review
# 在 GitHub PR 页面查看 AI Review
```

---

## 验证示例

### 日志输出示例

当 GitHub 发送 `pull_request.opened` 事件时：

```
[INF] Webhook received: event=pull_request, signature=sha256=...
[INF] Processing GitHub pull request event: action=opened, repo=owner/repo, pr=123
[INF] Event published to topic: interaction.github.pull_request
[INF] Orchestrator received event: event_id=xxx, type=pull_request
[INF] Starting Eino agent execution
[INF] Calling tool: github.pr.get_metadata
[INF] Tool executed: github.pr.get_metadata, duration=234ms
[INF] Calling tool: github.pr.get_files
[INF] Tool executed: github.pr.get_files, duration=156ms
[INF] LLM generating review...
[INF] Calling tool: github.pr.publish_review
[INF] Review published: html_url=https://github.com/...
[INF] Agent execution completed successfully
```

### GitHub PR Review 示例

```markdown
## Code Review Summary

### Overall Assessment
This PR introduces a new authentication module with proper token validation and error handling. The implementation follows good security practices.

### Key Findings

1. **Positive: Secure Token Handling**
   - File: `auth/token.go`
   - Token expiration is properly validated
   - Refresh mechanism follows OAuth 2.0 best practices

2. **Minor: Error Message Clarity**
   - File: `auth/handler.go:45`
   - Consider providing more specific error messages for debugging
   - Current: "Authentication failed"
   - Suggested: "Authentication failed: invalid token format"

### No Blocking Issues
The code changes look acceptable. No critical bugs or security concerns detected.
```

---

## 待验证的关键路径

### 🔍 需要手动测试的场景

- [ ] 在测试仓库创建 PR，验证自动 Review
- [ ] 验证 OAuth 凭证正确解析
- [ ] 验证多文件变更的 Review 质量
- [ ] 验证大 PR 的处理性能
- [ ] 验证错误处理和重试机制

### ⚙️ 配置检查清单

- [ ] GitHub App 配置正确（App ID, Private Key）
- [ ] Webhook Secret 配置
- [ ] OAuth Client ID/Secret 配置
- [ ] LLM API Key 配置
- [ ] NATS 连接配置
- [ ] 订阅的 Webhook 事件类型正确

---

## 已知问题和限制

### 当前限制

1. **速率限制**: GitHub API 有速率限制（5000 requests/hour for authenticated）
2. **Token 权限**: 需要正确的 GitHub App 权限（pull-requests, contents）
3. **并发处理**: 单个 Eino Runner 实例，需测试并发性能
4. **错误恢复**: 工具调用失败时的恢复策略待完善

### 改进空间

1. **Review 质量**: 需要优化 LLM 提示词
2. **性能优化**: 添加缓存层减少 API 调用
3. **可配置性**: 支持自定义 Review 规则
4. **多租户**: 当前单租户模式

---

## 下一步行动

1. **立即**: 完成端到端 PR Review 测试
2. **本周**: 
   - 验证 Issue Comment 自动回复
   - 优化 LLM 提示词质量
3. **下周**:
   - 添加更多事件类型支持
   - 实现错误处理和重试机制
4. **后续**:
   - 添加 Performance 监控
   - 实现 DigitalAssistant 管理 API

---

*最后更新: 2026-04-18*
*文档状态: 反映当前实现*
