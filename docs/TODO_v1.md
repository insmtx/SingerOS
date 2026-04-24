# SingerOS 后端开发 TODO 清单（最新更新 2026-04-18）

> 本清单按照可独立提交、可自测、可验收的原则组织,每个条目代表一次完整的开发迭代。

---

## 📊 当前状态摘要（2026-04-18）

### ✅ 已完成的第一阶段任务：

- ✅ LLM Provider 接口设计与实现
- ✅ OpenAI Provider 实现
- ✅ Skill 系统接口和 Manager
- ✅ Event Gateway (GitHub Webhook)
- ✅ Event Bus (NATS JetStream)
- ✅ Orchestrator 基础实现
- ✅ Tools 系统（Registry + Runtime）
- ✅ Auth 系统（OAuth + 账户管理）
- ✅ GitHub 工具集（PR 读写、文件对比等）

### 🔄 进行中：

- 🔄 PR Review 端到端流程验证
- 🔄 Issue Comment 自动回复技能

### ❌ 待开始：

- ❌ DigitalAssistant 管理 API
- ❌ 多渠道扩展（GitLab、企业微信）
- ❌ Memory 系统

---

## 历史：第一阶段: 基础设施完善 (✅ 已完成)

> 以下任务已全部完成，详见各文件实现。



### 1.1 LLM Provider 接口设计与实现

#### Task 1.1.1: 定义LLM统一接口
**文件**: `backend/llm/provider.go`
**提交信息**: `feat(llm): add LLM provider interface definition`

**实现内容**:
```go
// backend/llm/provider.go
package llm

type Provider interface {
    Name() string
    Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
    GenerateStream(ctx context.Context, req *GenerateRequest) (<-chan StreamChunk, error)
    CountTokens(text string) int
    Models() []string
}

type GenerateRequest struct {
    Messages   []Message
    Model      string
    MaxTokens  int
    Temperature float64
    Stop       []string
}

type GenerateResponse struct {
    Content     string
    Usage      TokenUsage
    FinishReason string
}

type Message struct {
    Role    string
    Content string
}

type StreamChunk struct {
    Content   string
    Done     bool
}

type TokenUsage struct {
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
}
```

**验收标准**:
- [x] 接口定义完整,包含同步和流式调用
- [x] 类型定义覆盖所有必要字段
- [x] `go build ./...` 编译通过
- [x] `go test ./backend/llm/...` 可运行(空测试文件即可)

**自测命令**:
```bash
go build ./...
go vet ./backend/llm/...
```

---

#### Task 1.1.2: 实现OpenAI Provider
**文件**: 
- `backend/llm/openai/provider.go`
- `backend/llm/openai/provider_test.go`

**提交信息**: `feat(llm): implement OpenAI provider with streaming support`

**实现内容**:
- 实现Provider接口
- 支持GPT-4, GPT-3.5-turbo模型
- 实现同步Generate方法
- 实现流式GenerateStream方法
- Token计数(使用tiktoken或近似算法)
- 错误处理与重试

**验收标准**:
- [x] 单元测试覆盖率 > 80%
- [x] Mock测试不依赖真实API调用
- [x] 支持context取消
- [x] 错误可识别(区分网络错误、API错误、配额错误)

**自测命令**:
```bash
go test -v -cover ./backend/llm/openai/...
```

---

#### Task 1.1.3: 实现Model Router多模型路由
**文件**: 
- `backend/llm/router.go`
- `backend/llm/router_test.go`

**提交信息**: `feat(llm): add model router with fallback support`

**实现内容**:
- 多Provider注册与管理
- 基于模型名称的路由
- Fallback降级策略
- Token配额检查(预留接口)

**验收标准**:
- [x] 可注册多个Provider
- [x] 根据模型名称正确路由
- [x] 主Provider失败自动降级
- [x] 单元测试覆盖路由逻辑

**自测命令**:
```bash
go test -v -cover ./backend/llm/...
```

---

### 1.2 数据持久化层实现

#### Task 1.2.1: 数据库连接管理
**文件**: 
- `backend/persistence/database.go`
- `backend/persistence/config.go`

**提交信息**: `feat(persistence): add database connection manager`

**实现内容**:
- PostgreSQL连接池配置
- 支持环境变量配置
- 健康检查
- 优雅关闭

**验收标准**:
- [x] 支持连接池参数配置
- [x] Ping检测连接可用性
- [x] 支持日志输出
- [x] 集成测试可启动(使用testcontainers或mock)

**自测命令**:
```bash
go test -v ./backend/persistence/...
```

---

#### Task 1.2.2: Migration自动化
**文件**: 
- `backend/persistence/migrate.go`
- `backend/persistence/migrations/001_init.up.sql`
- `backend/persistence/migrations/001_init.down.sql`

**提交信息**: `feat(persistence): add auto migration support`

**实现内容**:
- 集成gormigrate或goose
- 初始Migration: DigitalAssistant, Event, Skill等表
- 支持回滚
- 版本管理

**验收标准**:
- [x] Migration可正向执行
- [x] Migration可回滚
- [x] 表结构与types定义一致
- [x] 支持幂等执行

**自测命令**:
```bash
# 需要本地PostgreSQL
go test -v ./backend/persistence/... -run TestMigration
```

---

#### Task 1.2.3: Repository层封装
**文件**: 
- `backend/persistence/repository/digital_assistant_repo.go`
- `backend/persistence/repository/event_repo.go`
- `backend/persistence/repository/skill_repo.go`

**提交信息**: `feat(persistence): implement repository layer for core entities`

**实现内容**:
- DigitalAssistantRepository: CRUD
- EventRepository: Create, Query, UpdateStatus
- SkillRepository: Create, GetByID, List
- 事务支持

**验收标准**:
- [x] 每个Repository有完整CRUD测试
- [x] 支持预加载关联
- [x] 支持分页查询
- [x] 错误类型明确

**自测命令**:
```bash
go test -cover ./backend/persistence/repository/...
```

---

### 1.3 Skill Proxy服务完善

#### Task 1.3.1: Skill注册与发现服务
**文件**: 
- `backend/skillproxy/registry.go`
- `backend/skillproxy/registry_test.go`

**提交信息**: `feat(skillproxy): add skill registry service`

**实现内容**:
- Skill注册接口
- 基于内存的Registry实现
- Skill元数据存储
- 按Category/Type查询

**验收标准**:
- [x] 可注册本地Skill
- [x] 可按名称/ID查询
- [x] 支持批量查询
- [x] 注册/注销线程安全

**自测命令**:
```bash
go test -v ./backend/skillproxy/...
```

---

#### Task 1.3.2: Skill执行器
**文件**: 
- `backend/skillproxy/executor.go`
- `backend/skillproxy/executor_test.go`

**提交信息**: `feat(skillproxy): implement skill executor with timeout control`

**实现内容**:
- 执行上下文管理
- 超时控制
- 输入验证
- 执行结果处理
- 日志记录

**验收标准**:
- [x] 执行超时可配置
- [x] 验证输入符合Schema
- [x] 执行错误可捕获
- [x] 使用Example Skill进行集成测试

**自测命令**:
```bash
go test -v ./backend/skillproxy/...
```

---

#### Task 1.3.3: Skill Proxy HTTP API
**文件**: 
- `backend/skillproxy/handler.go`
- `backend/skillproxy/handler_test.go`

**提交信息**: `feat(skillproxy): add HTTP API for skill execution`

**实现内容**:
- POST /api/v1/skills/:id/execute - 执行Skill
- GET /api/v1/skills - 列出所有Skill
- GET /api/v1/skills/:id - 获取Skill详情
- 健康检查端点

**验收标准**:
- [x] HTTP接口符合RESTful规范
- [x] 请求参数验证
- [x] 统一错误响应格式
- [x] httptest单元测试覆盖

**自测命令**:
```bash
go test -v ./backend/skillproxy/...
```

---

### 1.4 Memory System实现

#### Task 1.4.1: Memory接口定义
**文件**: `backend/memory/store.go`

**提交信息**: `feat(memory): add memory store interface`

**实现内容**:
```go
type MemoryStore interface {
    Store(ctx context.Context, key string, value interface{}, opts ...Option) error
    Retrieve(ctx context.Context, key string) (interface{}, error)
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
    SetTTL(ctx context.Context, key string, ttl time.Duration) error
}

type ShortTermMemory interface {
    MemoryStore
    SetSession(ctx context.Context, sessionID string, data map[string]interface{}) error
    GetSession(ctx context.Context, sessionID string) (map[string]interface{}, error)
}
```

**验收标准**:
- [x] 接口定义清晰
- [x] 支持可选参数(Opt模式)
- [x] 编译通过

---

#### Task 1.4.2: Redis短期记忆实现
**文件**: 
- `backend/memory/redis/store.go`
- `backend/memory/redis/store_test.go`

**提交信息**: `feat(memory): implement redis-based short term memory`

**实现内容**:
- Redis连接管理
- Session级别KV存储
- TTL自动过期
- 可序列化JSON/MsgPack

**验收标准**:
- [x] 支持Session存储与读取
- [x] TTL生效
- [x] 使用miniredis进行单元测试
- [x] 集成测试可选(需本地Redis)

**自测命令**:
```bash
go test -v ./backend/memory/redis/...
```

---

## 第二阶段: 核心引擎实现 (P0)

### 2.1 Orchestrator编排器

#### Task 2.1.1: Orchestrator接口与基础结构
**文件**: 
- `backend/orchestrator/orchestrator.go`
- `backend/orchestrator/types.go`

**提交信息**: `feat(orchestrator): add orchestrator interface and types`

**实现内容**:
```go
type Orchestrator interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    HandleEvent(ctx context.Context, event *interaction.Event) error
}

type ExecutionContext struct {
    ID            string
    Event         *interaction.Event
    Assistant     *types.DigitalAssistant
    Agent         Agent
    State         ExecutionState
    CreatedAt     time.Time
    UpdatedAt     time.Time
    Steps         []ExecutionStep
}

type ExecutionState string

const (
    StatePending    ExecutionState = "pending"
    StateRunning    ExecutionState = "running"
    StateSucceeded  ExecutionState = "succeeded"
    StateFailed     ExecutionState = "failed"
    StateTimeout    ExecutionState = "timeout"
)
```

**验收标准**:
- [x] 类型定义完整
- [x] 状态枚举合理
- [x] 编译通过

---

#### Task 2.1.2: 事件消费者
**文件**: 
- `backend/orchestrator/consumer.go`
- `backend/orchestrator/consumer_test.go`

**提交信息**: `feat(orchestrator): implement event consumer from NATS JetStream`

**实现内容**:
- 订阅NATS JetStream主题
- 消息反序列化为Event
- 错误处理与重试
- 优雅关闭

**验收标准**:
- [x] 消费NATS消息正常
- [x] 连接断开自动重连
- [x] 使用Mock EventBus测试
- [x] 支持并发消费(可配置worker数)

**自测命令**:
```bash
go test -v ./backend/orchestrator/... -run TestConsumer
```

---

#### Task 2.1.3: DigitalAssistant匹配器
**文件**: 
- `backend/orchestrator/matcher.go`
- `backend/orchestrator/matcher_test.go`

**提交信息**: `feat(orchestrator): implement assistant matcher based on event`

**实现内容**:
- 根据事件类型匹配DigitalAssistant
- 根据事件来源(Channel)过滤
- 支持优先级匹配
- 缓存匹配结果

**验收标准**:
- [x] 按事件类型匹配
- [x] 按Channel过滤
- [x] 匹配规则可配置
- [x] 单元测试覆盖多种匹配场景

**自测命令**:
```bash
go test -v ./backend/orchestrator/... -run TestMatcher
```

---

#### Task 2.1.4: Orchestrator完整实现
**文件**: 
- `backend/orchestrator/default_orchestrator.go`
- `backend/orchestrator/default_orchestrator_test.go`

**提交信息**: `feat(orchestrator): implement default orchestrator with full workflow`

**实现内容**:
- 整合Consumer、Matcher
- 创建ExecutionContext
- 调用Agent执行
- 状态流转管理
- 执行日志记录

**验收标准**:
- [x] 端到端流程: Event → Match → Execute → Record
- [x] 执行状态正确流转
- [x] 错误能正确捕获与记录
- [x] 集成测试使用真实NATS或容器

**自测命令**:
```bash
go test -v ./backend/orchestrator/... -run TestOrchestrator
```

---

### 2.2 Agent Engine实现

#### Task 2.2.1: Agent接口与基础类型
**文件**: 
- `backend/agent/agent.go`
- `backend/agent/types.go`

**提交信息**: `feat(agent): add agent interface and planning types`

**实现内容**:
```go
type Agent interface {
    ID() string
    Name() string
    Plan(ctx context.Context, task *Task) (*Plan, error)
    Execute(ctx context.Context, plan *Plan) (*Result, error)
    Reflect(ctx context.Context, result *Result) (*Adjustment, error)
}

type Task struct {
    ID          string
    Description string
    Context     map[string]interface{}
    Constraints []string
}

type Plan struct {
    Steps []PlanStep
}

type PlanStep struct {
    ID        string
    SkillID   string
    Input     map[string]interface{}
    DependsOn []string
}

type Result struct {
    Success bool
    Output  map[string]interface{}
    Error   error
}
```

**验收标准**:
- [x] 接口设计支持Planning-Acting-Reflecting循环
- [x] 类型定义完整
- [x] 编译通过

---

#### Task 2.2.2: Code Review Agent实现
**文件**: 
- `backend/agent/codereview/agent.go`
- `backend/agent/codereview/planner.go`
- `backend/agent/codereview/executor.go`

**提交信息**: `feat(agent): implement code review agent`

**实现内容**:
- Plan: 分析PR diff,制定review计划
- Execute: 调用LLM分析代码
- Reflect: 评估结果完整性
- 输出结构化Review结果

**验收标准**:
- [x] 能解析PR diff
- [x] Plan生成合理步骤
- [x] 调用LLM成功
- [x] 输出Review comment格式正确
- [x] 单元测试覆盖planning逻辑

**自测命令**:
```bash
go test -v ./backend/agent/codereview/...
```

---

#### Task 2.2.3: Issue Reply Agent实现
**文件**: 
- `backend/agent/issuereply/agent.go`
- `backend/agent/issuereply/planner.go`

**提交信息**: `feat(agent): implement issue reply agent`

**实现内容**:
- Plan: 分析Issue内容及上下文
- Execute: 生成回复内容
- Reflect: 检查回复质量
- 支持多轮对话

**验收标准**:
- [x] 能解析Issue信息
- [x] 生成合理回复
- [x] 支持引用上下文
- [x] 单元测试覆盖

**自测命令**:
```bash
go test -v ./backend/agent/issuereply/...
```

---

### 2.3 主服务集成

#### Task 2.3.1: 集成Orchestrator到singer服务
**文件**: `backend/cmd/singer/main.go`

**提交信息**: `feat(singer): integrate orchestrator into main service`

**实现内容**:
- 初始化Orchestrator
- 注入EventBus依赖
- 注入DigitalAssistant Repository
- 启动Orchestrator
- 信号处理优雅关闭

**验收标准**:
- [x] 服务启动无报错
- [x] Orchestrator正确初始化
- [x] 收到事件能正确处理
- [x] Ctrl+C优雅关闭

**自测命令**:
```bash
go build -o ./bundles/singer ./backend/cmd/singer/main.go
./bundles/singer --help
```

---

#### Task 2.3.2: 集成数据库到singer服务
**文件**: `backend/cmd/singer/main.go`

**提交信息**: `feat(singer): integrate database with auto migration`

**实现内容**:
- 加载数据库配置
- 初始化连接池
- 执行Auto Migration
- 注入到Repository

**验收标准**:
- [x] 连接数据库成功
- [x] Migration自动执行
- [x] 服务启动后表存在
- [x] 使用环境变量配置

**自测命令**:
```bash
# 需要本地PostgreSQL
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=singer
go run ./backend/cmd/singer/main.go
```

---

#### Task 2.3.3: 集成LLM Router
**文件**: `backend/cmd/singer/main.go`

**提交信息**: `feat(singer): integrate LLM router with config`

**实现内容**:
- 加载LLM配置(API Key, Model等)
- 初始化OpenAI Provider
- 注册到ModelRouter
- 注入到Agent

**验收标准**:
- [x] API Key从环境变量读取
- [x] Provider初始化成功
- [x] 可配置默认模型
- [x] 日志不输出敏感信息

**自测命令**:
```bash
export OPENAI_API_KEY=sk-xxx
go run ./backend/cmd/singer/main.go
```

---

## 第三阶段: MVP功能实现 (P1)

### 3.1 GitHub集成Skills

#### Task 3.1.1: GitHub API客户端封装
**文件**: 
- `backend/skills/integration/github/client.go`
- `backend/skills/integration/github/client_test.go`

**提交信息**: `feat(skill): add github api client wrapper`

**实现内容**:
- 封装go-github/github库
- 认证(App Token / PAT)
- 常用API封装:
  - GetPR
  - GetPRDiff
  - ListPRFiles
  - CreateReviewComment
  - CreateIssueComment
- Rate Limit处理
- 错误封装

**验收标准**:
- [x] 客户端可正常初始化
- [x] API调用正确
- [x] Rate Limit处理
- [x] 使用mock测试

**自测命令**:
```bash
go test -v ./backend/skills/integration/github/... -run TestClient
```

---

#### Task 3.1.2: GetPRDiff Skill
**文件**: 
- `backend/skills/integration/github/get_pr_diff.go`
- `backend/skills/integration/github/get_pr_diff_test.go`

**提交信息**: `feat(skill): implement get_pr_diff skill`

**实现内容**:
- 输入: owner, repo, pr_number
- 输出: diff_content, files_changed
- 调用GitHub API获取PR diff
- 解析diff结构

**验收标准**:
- [x] Skill注册成功
- [x] 输入Schema验证
- [x] 输出格式正确
- [x] 单元测试覆盖

**自测命令**:
```bash
go test -v ./backend/skills/integration/github/... -run TestGetPRDiff
```

---

#### Task 3.1.3: CreateReviewComment Skill
**文件**: 
- `backend/skills/integration/github/create_review_comment.go`
- `backend/skills/integration/github/create_review_comment_test.go`

**提交信息**: `feat(skill): implement create_review_comment skill`

**实现内容**:
- 输入: owner, repo, pr_number, commit_id, path, position, body
- 输出: comment_id, html_url
- 创建PR Review Comment
- 支持批量创建

**验收标准**:
- [x] 创建Comment成功
- [x] 参数验证
- [x] 返回正确的comment_id
- [x] 单元测试覆盖

**自测命令**:
```bash
go test -v ./backend/skills/integration/github/... -run TestCreateReviewComment
```

---

#### Task 3.1.4: CreateIssueComment Skill
**文件**: 
- `backend/skills/integration/github/create_issue_comment.go`

**提交信息**: `feat(skill): implement create_issue_comment skill`

**实现内容**:
- 输入: owner, repo, issue_number, body
- 输出: comment_id, html_url
- 创建Issue评论

**验收标准**:
- [x] 创建成功
- [x] 参数验证
- [x] 单元测试覆盖

**自测命令**:
```bash
go test -v ./backend/skills/integration/github/... -run TestCreateIssueComment
```

---

### 3.2 AI Skills

#### Task 3.2.1: CodeReview Skill
**文件**: 
- `backend/skills/ai/codereview/review.go`
- `backend/skills/ai/codereview/review_test.go`

**提交信息**: `feat(skill): implement ai code_review skill`

**实现内容**:
- 输入: diff_content, language, review_focus
- 输出: review_comments, summary
- 调用LLM分析代码
- 结构化输出Review结果
- Prompt模板管理

**验收标准**:
- [x] Prompt合理
- [x] LLM调用成功
- [x] 输出格式化为Review comments
- [x] 支持多种review focus
- [x] 单元测试(Mock LLM)

**自测命令**:
```bash
go test -v ./backend/skills/ai/codereview/...
```

---

#### Task 3.2.2: Summarize Skill
**文件**: 
- `backend/skills/ai/summarize/summarize.go`

**提交信息**: `feat(skill): implement ai summarize skill`

**实现内容**:
- 输入: content, style (brief/detailed)
- 输出: summary
- 内容摘要生成
- 支持不同风格

**验收标准**:
- [x] 摘要质量合理
- [x] 支持长度控制
- [x] 单元测试覆盖

---

#### Task 3.2.3: GenerateResponse Skill
**文件**: 
- `backend/skills/ai/generate_response/generate.go`

**提交信息**: `feat(skill): implement ai generate_response skill`

**实现内容**:
- 输入: context, question, tone
- 输出: response
- 基于上下文生成回复
- 支持不同语气

**验收标准**:
- [x] 回复相关性好
- [x] 支持语气配置
- [x] 单元测试覆盖

---

### 3.3 Workflow集成

#### Task 3.3.1: PR Review完整流程集成
**文件**: 
- `backend/workflows/pr_review.go`
- `backend/workflows/pr_review_test.go`

**提交信息**: `feat(workflow): implement pr review workflow`

**实现内容**:
- 监听GitHub PR opened事件
- 调用GetPRDiff Skill
- 调用CodeReview Skill
- 调用CreateReviewComment Skill
- 结果通知

**验收标准**:
- [x] GitHub webhook触发流程
- [x] 整体流程无报错
- [x] Review comments创建成功
- [x] 集成测试(需测试仓库)

**自测命令**:
```bash
# 需要配置GitHub App
go test -v ./backend/workflows/... -run TestPRReview -timeout 5m
```

---

#### Task 3.3.2: Issue自动回复流程集成
**文件**: 
- `backend/workflows/issue_reply.go`

**提交信息**: `feat(workflow): implement issue auto-reply workflow`

**实现内容**:
- 监听GitHub issue_comment事件
- 调用GenerateResponse Skill
- 调用CreateIssueComment Skill
- 支持过滤规则

**验收标准**:
- [x] 事件触发正常
- [x] 回复创建成功
- [x] 过滤规则生效
- [x] 集成测试

**自测命令**:
```bash
go test -v ./backend/workflows/... -run TestIssueReply
```

---

### 3.4 DigitalAssistant实际实现

#### Task 3.4.1: CodeAssistantDigitalAssistant
**文件**: 
- `backend/assistant/code_assistant.go`
- `backend/assistant/code_assistant_test.go`

**提交信息**: `feat(assistant): implement code assistant digital assistant`

**实现内容**:
- 配置: Skills列表, Channels列表
- 触发规则: PR事件, Issue事件
- Agent绑定: CodeReviewAgent, IssueReplyAgent
- 配置加载

**验收标准**:
- [x] 可从配置文件加载
- [x] 可匹配对应事件
- [x] 正确调用Agent
- [x] 集成测试

**自测命令**:
```bash
go test -v ./backend/assistant/...
```

---

#### Task 3.4.2: DigitalAssistant配置系统
**文件**: 
- `backend/assistant/config.go`
- `backend/assistant/loader.go`

**提交信息**: `feat(assistant): add assistant config loader from yaml`

**实现内容**:
- YAML配置文件格式
- 配置校验
- 热重载(可选)

**验收标准**:
- [x] YAML格式正确解析
- [x] 配置校验通过
- [x] 示例配置文件

**配置示例**:
```yaml
# configs/assistants/code-assistant.yaml
code: code-assistant
name: Code Assistant
organization_id: org-001
channels:
  - code: github
    config:
      app_id: "123456"
skills:
  - skill_id: get-pr-diff
    enabled: true
  - skill_id: code-review
    enabled: true
agents:
  - agent_id: pr-review-agent
    trigger: pr_opened
  - agent_id: issue-reply-agent
    trigger: issue_comment
```

---

## 第四阶段: 系统完善 (P2)

### 4.1 更多GitHub事件支持

#### Task 4.1.1: PR同步事件处理
**文件**: `backend/interaction/connectors/github/sync.go`

**提交信息**: `feat(github): add pr synchronize event handler`

**实现内容**:
- 解析synchronize事件
- 触发重新Review

---

#### Task 4.1.2: Issue opened事件处理
**文件**: `backend/interaction/connectors/github/issues.go`

**提交信息**: `feat(github): add issue opened event handler`

**实现内容**:
- 解析issues事件
- 自动分类/分配

---

### 4.2 其他连接器实现

#### Task 4.2.1: GitLab Connector基础实现
**文件**: `backend/interaction/connectors/gitlab/gitlab.go`

**提交信息**: `feat(gitlab): implement basic gitlab connector`

**实现内容**:
- Webhook接收
- MR事件解析
-签名验证

---

#### Task 4.2.2: 企业微信Connector基础实现
**文件**: `backend/interaction/connectors/wework/app.go`

**提交信息**: `feat(wework): implement basic wework connector`

**实现内容**:
- 回调验证
- 消息加解密
- 基础消息接收

---

### 4.3 权限系统基础

#### Task 4.3.1: 权限模型定义
**文件**: `backend/authz/model.go`

**提交信息**: `feat(authz): add permission model definition`

**实现内容**:
- Permission, Role, User模型
- RBAC基础结构
- Policy定义

---

#### Task 4.3.2: 权限检查中间件
**文件**: `backend/authz/middleware.go`

**提交信息**: `feat(authz): add permission check middleware`

**实现内容**:
- HTTP中间件
- Skill调用权限检查
- 资源访问控制

---

### 4.4 可观察性

#### Task 4.4.1: 结构化日志
**文件**: `backend/logger/logger.go`

**提交信息**: `feat(logger): add structured logger with context`

**实现内容**:
- zap或logrus集成
- 上下文传递
- 日志级别配置

---

#### Task 4.4.2: Metrics采集
**文件**: `backend/metrics/metrics.go`

**提交信息**: `feat(metrics): add prometheus metrics collection`

**实现内容**:
- Prometheus集成
- 请求耗时
- Skill调用计数
- 错误率统计

---

#### Task 4.4.3: 分布式追踪
**文件**: `backend/tracing/tracing.go`

**提交信息**: `feat(tracing): add opentelemetry tracing support`

**实现内容**:
- OpenTelemetry集成
- Span传播
- 与Event ID关联

---

## 第五阶段: 测试与文档 (P2)

### 5.1 集成测试

#### Task 5.1.1: 端到端PR Review测试
**提交信息**: `test(e2e): add pr review e2e test`

**实现内容**:
- Mock GitHub Webhook
- 完整流程测试
- 断言各环节

---

#### Task 5.1.2: 端到端Issue Reply测试
**提交信息**: `test(e2e): add issue reply e2e test`

---

### 5.2 性能测试

#### Task 5.2.1: 并发处理测试
**提交信息**: `test(perf): add concurrent event processing test`

---

### 5.3 文档

#### Task 5.3.1: API文档
**提交信息**: `docs: add API documentation with swagger`

**实现内容**:
- Swagger/OpenAPI定义
- 接口文档

---

#### Task 5.3.2: 部署文档
**提交信息**: `docs: add deployment guide`

**实现内容**:
- 环境变量说明
- Docker部署指南
- 配置示例

---

## 第六阶段: 企业级功能 (P3)

### 6.1 多租户

#### Task 6.1.1: 租户隔离
**提交信息**: `feat(multitenant): add tenant isolation support`

---

### 6.2 成本追踪

#### Task 6.2.1: Token使用统计
**提交信息**: `feat(cost): add token usage tracking`

---

#### Task 6.2.2: 成本报表
**提交信息**: `feat(cost): add cost report generation`

---

### 6.3 审计日志

#### Task 6.3.1: 审计日志记录
**提交信息**: `feat(audit): add audit log for sensitive operations`

---

## 验收总则

每项任务提交前必须满足:

1. **代码质量**
   - `go fmt ./...` 无输出
   - `go vet ./...` 无错误
   - `golint ./...` 建议 < 5

2. **测试覆盖**
   - 新增代码测试覆盖率 > 60%
   - 核心逻辑测试覆盖率 > 80%

3. **文档更新**
   - 新增配置需更新README
   - 新增API需更新API文档

4. **CI通过**
   - 所有测试通过
   - 无安全漏洞警告

---

## 快速检查脚本

```bash
#!/bin/bash
# scripts/check.sh

echo "=== 格式检查 ==="
go fmt ./...

echo "=== Vet检查 ==="
go vet ./...

echo "=== 测试 ==="
go test -cover ./...

echo "=== 构建 ==="
go build -o ./bundles/singer ./backend/cmd/singer/main.go
go build -o ./bundles/skill-proxy ./backend/cmd/skill-proxy/main.go

echo "=== 检查完成 ==="
```

---

## 项目里程碑

| 阶段 | 目标 | 预计时间 | 可交付成果 |
|------|------|----------|------------|
| **M1** | 基础设施完善 | Week 1-3 | LLM Provider, 数据库, Skill Proxy完整 |
| **M2** | 核心引擎实现 | Week 4-6 | Orchestrator, Agent Engine, Memory |
| **M3** | MVP功能上线 | Week 7-8 | PR Review, Issue Reply可用 |
| **M4** | 系统完善 | Week 9-10 | 多事件支持, 权限系统, 可观察性 |
| **M5** | 企业级准备 | Week 11-12 | 多租户, 成本追踪, 审计日志 |

---

## 依赖关系图

```
M1 基础设施
├── 1.1 LLM Provider ─────────────┐
├── 1.2 数据持久化 ────┐          │
├── 1.3 Skill Proxy ───┼──────────┼─┐
└── 1.4 Memory ────────┘          │ │
                                  │ │
M2 核心引擎                        │ │
├── 2.1 Orchestrator ─────────────┼─┤
│   └── depends on: Event Bus ✅  │ │
├── 2.2 Agent Engine ─────────────┼─┤
│   └── depends on: LLM Provider  │ │
└── 2.3 主服务集成 ───────────────┘ │
    └── depends on: 以上所有        │
                                    │
M3 MVP功能                          │
├── 3.1 GitHub Skills ─────────────┤
│   └── depends on: Agent Engine ◄─┘
├── 3.2 AI Skills ────────────────┤
│   └── depends on: LLM Provider
├── 3.3 Workflow ─────────────────┤
│   └── depends on: Orchestrator, Skills
└── 3.4 DigitalAssistant ─────────┘
    └── depends on: Workflow, Agent

M4 系统完善
├── 4.1 更多事件 ──────────────────────────┐
├── 4.2 更多连接器 ────────────────────────┤
├── 4.3 权限系统 ─── depends on: 数据库 ◄──┤
└── 4.4 可观察性 ──────────────────────────┘

M5 企业级
├── 6.1 多租户 ──── depends on: 权限系统
├── 6.2 成本追踪 ── depends on: LLM Provider
└── 6.3 审计日志 ── depends on: 数据库
```

---

*本TODO清单最后更新: 2026-03-25*
*请按照阶段顺序逐步实施，每个Task完成后打勾并记录实际完成日期*