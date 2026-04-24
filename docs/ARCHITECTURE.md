# SingerOS 架构设计文档

> 基于 **Event Engine + Execution Engine + Agent Runtime 三核架构** 构建的企业级 AI 操作系统
>
> **版本：3.1** | **最后更新：2026-04-23**

## 1. 核心愿景

构建一个企业级数字员工平台，让企业可以像管理真实员工一样，创建、配置、授权、调度和审计 AI 数字员工，并实现：

* **多 Agent 协作** - 多个智能体协同工作
* **多运行时执行** - 支持不同 Agent 引擎并存
* **本地 + 云端协同** - Edge 与 Remote Runtime 分工
* **可控、安全、可审计** - 企业级安全控制

数字员工不是单纯的聊天机器人。它需要有独立身份、接收任务的入口、真实执行工作的环境，以及模型、工具、技能、知识库等基础能力。

## 设计原则

* **事件驱动（Event-Driven First）**
  所有行为统一抽象为 Event，通过 Event Bus 传播
* **控制面 / 执行面分离（Control vs Execution）**
  决策与执行彻底解耦
* **三核架构（Three-Core Architecture）**
  Event Engine + Execution Engine + Agent Runtime 职责分离
* **领域驱动设计（Domain-Driven Design）**
  按领域分层（event/execution/agent/skill），而非按技术分层（controller/service/model）
* **接口优先（Interface-Driven）**
  每一层都必须定义 interface，而不是直接依赖实现
* **核心引擎内聚可替换**
  Event Engine、Execution Engine、Agent Runtime 必须可独立替换和部署
* **分层命名（Layered Naming）**
  Engine = 执行能力 | Runtime = 运行时容器 | Service = 对外能力 | Connector = 外部接入
* **边缘优先（Edge-First）**
  本地能力（文件 / GUI）优先由 Edge Runtime 执行
* **安全优先（Security by Design）**
  明确本地与远程执行边界
* **数字助手是最高抽象（Digital Assistant First）**
  代表完整的 AI 数字员工实例
* **强制隔离（Enforced Isolation）**
  使用 `internal/` 目录强制隔离核心实现，`pkg/` 对外公开接口

## 2. 分层架构（四平面模型）

### 2.1 架构总览

```
┌────────────────────────────────────────────┐
│                Client / Edge               │
│  App / CLI / 本地 Agent Runtime (Edge)    │
└────────────────────┬───────────────────────┘
                     │
                     ▼
┌────────────────────────────────────────────┐
│            Interface Layer（接口层）        │
│         Assistant Service / Connector      │
└────────────────────┬───────────────────────┘
                     │
                     ▼
┌────────────────────────────────────────────┐
│          Control Plane（控制面）            │
│  Event Engine / Memory / Policy Engine    │
└────────────────────┬───────────────────────┘
                     │
                     ▼
┌────────────────────────────────────────────┐
│          Execution Plane（执行面）          │
│  Execution Engine / Agent Runtime / Skill  │
└────────────────────────────────────────────┘
```

### 2.2 四平面职责

| 平面 | 组件 | 职责 |
|------|------|------|
| **Edge Plane** | Edge Runtime / Client | 本地文件访问、GUI 自动化、用户环境交互 |
| **Interface Layer** | Assistant Service / Connector | 对外 API / 渠道接入 / 事件标准化 |
| **Control Plane** | Event Engine / Memory / Policy Engine | 决策中心：事件路由、上下文构建、权限控制 |
| **Execution Plane** | Execution Engine / Agent Runtime / Skill | 执行中心：Agent 推理、Skill 调用、Workflow 编排 |

### 2.3 核心数据通道（统一事件流）

```
External Event / User Input
        ↓
Connector（事件标准化）
        ↓
Event Bus（统一事件模型）
        ↓
Event Engine（事件路由）
        ↓
Execution Engine（执行调度）
        ↓
Agent Runtime / Workflow Engine / Skill（执行单元）
        ↓
Event Bus（响应流）
        ↓
Assistant Service → Client / UI
```

> **核心原则**：所有模块之间只能通过 Event Bus 通信

## 3. 核心模块划分

### 3.1 Connector（连接器）

**职责：**

* 接收外部系统事件（Webhook / API / 用户输入）
* 标准化为内部 Event
* 发布到 Event Bus

**支持渠道：**

* GitHub / GitLab
* 企业微信 / 飞书
* CLI / Web UI

**关键能力：**

* 签名验证
* 多协议适配
* 事件转换

**命名规范：**

```
GitHub Connector
GitLab Connector
Feishu Connector
Slack Connector
Webhook Connector
```

**接口定义：**

```go
type Connector interface {
    ChannelCode() string
    RegisterRoutes(r gin.IRouter)
}
```

### 3.2 Event Bus（事件总线）

**职责：**

系统唯一通信通道

> 所有模块之间只能通过 Event Bus 通信

**实现：**

* NATS JetStream

**接口定义：**

```go
type Publisher interface {
    Publish(ctx context.Context, topic string, event any) error
    Close() error
}

type Subscriber interface {
    Subscribe(ctx context.Context, topic string, handler func(any)) error
    Close() error
}
```

### 标准 Event 模型

Event 是系统内部统一的通信载体，包含以下核心字段：

- **ID** - 事件唯一标识
- **Type** - 事件类型（command.* / response.* / stream.* / state.* / system.*）
- **Source** - 事件来源
- **Target** - 事件目标
- **SessionID** - 会话标识
- **Payload** - 事件载荷
- **Timestamp** - 时间戳

### Event 分类

```
command.*      // 指令事件
response.*     // 响应事件
stream.*       // 流式事件
state.*        // 状态事件
system.*       // 系统事件
```

### 3.3 Assistant Service（助手服务）【原 Gateway】

**职责：**

* 对外统一 API 入口
* 用户请求处理
* 多渠道统一访问
* 调用 Event Engine / Execution Engine

**本质：**

> **"系统的对外接口层"**

### 3.4 Event Engine（事件引擎）【原 Orchestrator】⭐

**职责：**

* 订阅事件总线中的事件
* 事件路由与分发
* 调用 Handler 处理事件
* 触发执行流程

**核心能力：**

* 事件过滤与路由规则
* 事件聚合与防抖
* 事件优先级调度

**接口定义：**

```go
package eventengine

type Engine interface {
    Start(ctx context.Context) error
    RegisterHandler(eventType string, handler Handler)
    GetHandler(eventType string) (Handler, error)
}

type Handler interface {
    Handle(ctx context.Context, event *event.Event) error
}
```

**包结构：**

```
internal/eventengine/
├── engine.go           # Event Engine 核心实现
├── registry.go         # Handler 注册中心（插件化）
├── router.go           # 事件路由（不写死 switch）
└── builtins/           # 内置事件处理器
    ├── pr_handler.go
    ├── issue_handler.go
    └── push_handler.go
```

**本质：**

> **"系统的响应中心"** - 负责响应外部事件并启动执行流程

**⚠️ 常见错误：**

- ❌ 把所有逻辑写进 Event Handler
- ✅ 正确：Handler → 调用 Execution Engine

### 3.5 Execution Engine（执行引擎）【新增关键模块】⭐

**职责：**

* 调用 Skill
* 调用 Workflow
* 调用 Agent
* 控制执行流程（同步 / 异步 / 重试）

**核心能力：**

* 同步/异步执行控制
* 重试与降级机制
* 执行超时控制
* 并发执行管理

**接口定义：**

```go
package execution

type Engine interface {
    Execute(ctx context.Context, task *Task) error
    RegisterExecutor(taskType TaskType, executor Executor)
}

type Executor interface {
    Execute(ctx context.Context, task *Task) error
}

type Task struct {
    Type       TaskType
    Payload    map[string]interface{}
    Timeout    time.Duration
    MaxRetries int
}
```

**包结构：**

```
internal/execution/
├── engine.go           # Execution Engine 核心
├── dispatcher.go       # 调度器（任务分发）
├── executor.go         # 执行器接口
├── sync_executor.go    # 同步执行器
├── async_executor.go   # 异步执行器
├── retry.go            # 重试控制
├── timeout.go          # 超时控制
└── context/            # 执行上下文
    └── execution_context.go
```

**与 Event Engine 的关系：**

```
Event Engine：响应事件 → 决定何时执行
Execution Engine：执行逻辑 → 决定如何执行
```

> **核心原则**：Event Engine 与 Execution Engine 必须解耦

**⚠️ 常见错误：**

- ❌ 直接在 Event Handler 中执行复杂逻辑
- ✅ 正确：Handler → Execution Engine → 具体执行器

### 3.6 Agent Runtime（智能体运行时）

**职责：**

* 管理 Agent 生命周期
* 调用 LLM
* 管理 Memory / Context
* 工具调用（Tool / Skill）

**核心能力：**

* Agent 状态管理
* 上下文维护
* 推理循环（Reasoning Loop）
* 工具调用协调

**接口定义：**

```go
package agent

type Runtime interface {
    Initialize(ctx context.Context, config AgentConfig) error
    Execute(ctx context.Context, task *Task) (*Result, error)
    Shutdown(ctx context.Context) error
}

type Task struct {
    Type      TaskType
    Context   *Context
    Skills    []Skill
    Tools     []Tool
}

type Result struct {
    Output   string
    Metadata map[string]interface{}
}
```

**包结构：**

```
internal/agent/
├── runtime.go           # Agent Runtime 接口
├── lifecycle.go         # 生命周期管理
├── context.go           # 上下文管理
├── reasoning.go         # 推理循环
└── eino/               # Eino 具体实现
    ├── eino_runtime.go
    ├── agent_runner.go
    ├── chatmodel.go
    └── tool_adapter.go
```

**与 Execution Engine 的关系：**

```
Execution Engine 调用 Agent Runtime
Agent Runtime 专注于 Agent 的推理与决策
```

**⚠️ 常见错误：**

- ❌ Agent Runtime 直接调 MQ / DB
- ✅ 必须通过 Execution Engine / Skill / Infra

### 3.7 Workflow Engine（工作流引擎）【规划中】

**职责：**

* 多步骤任务编排
* DAG / 状态机执行
* 长任务执行管理

**包结构：**

```
internal/workflow/
├── engine.go           # 流程引擎
├── definition/         # DAG / YAML 定义
└── runtime/            # 运行时
```

**与 Execution Engine 的关系：**

```
Execution Engine 调用 Workflow Engine
Workflow Engine 专注于复杂流程编排
```

### 3.8 Runtime Manager（运行时调度器）

**职责：**

* 管理所有 Runtime 实例
* 能力注册（Skill / GPU / Browser）
* 负载均衡
* 健康检查

**类比：**

> Kubernetes Scheduler（简化版）

### 3.9 Memory（记忆系统）

**职责：**

* 会话上下文（短期记忆）
* 长期记忆（向量）
* 知识检索（RAG）

### 3.10 Model Router（模型调度）

**职责：**

* 多模型管理
* fallback / 降级
* 成本控制

### 3.11 Policy Engine（策略引擎）【新增关键模块】

**职责：**

* Agent 行为控制
* Skill 调用权限
* 审计日志

**强制规则：**

* Remote Runtime 不得直接访问本地资源
* 所有高权限操作必须经过 Policy Engine

### 3.12 Skills 能力系统

**Skill 定义：** 可复用的 AI 能力单元，是 SingerOS 的核心构建块

**接口定义：**

```go
package skill

type Skill interface {
    Info() *SkillInfo
    Execute(ctx context.Context, input SkillInput) (SkillOutput, error)
    Validate(input SkillInput) error
    GetID() string
    GetName() string
    GetDescription() string
}

type SkillInfo struct {
    ID           string                 `json:"id"`
    Name         string                 `json:"name"`
    Description  string                 `json:"description"`
    Version      string                 `json:"version,omitempty"`
    Category     string                 `json:"category,omitempty"`
    InputSchema  map[string]interface{} `json:"input_schema,omitempty"`
    OutputSchema map[string]interface{} `json:"output_schema,omitempty"`
}
```

**Skill 分类：**

- **集成类 Skills** - 外部系统集成（GitHub、GitLab、飞书等）
- **AI 类 Skills** - 基于大模型的推理能力（代码审查、摘要生成、分类等）
- **工具类 Skills** - 底层工具能力（Shell 执行、Python 脚本、HTTP 请求等）
- **工作流类 Skills** - 复杂编排能力（PR 审查工作流、Bug 分类工作流等）

**技能加载方式：**

- 文件系统：通过 SKILL.md 文件定义
- 代码嵌入：编译时打包的内置技能
- 远程加载：从技能市场动态下载（规划中）

**包结构：**

```
internal/skill/
├── registry.go         # Skill 注册中心（必须动态注册）
├── executor.go         # Skill 执行器
├── base_skill.go       # 基础 Skill 实现
└── builtin/            # 内置技能
    ├── github_pr_review.go
    └── code_summarize.go
```

**⚠️ 常见错误：**

- ❌ Skill 写死在代码中
- ✅ 必须 Registry 化，支持动态注册

### 3.13 Tools 工具系统

**Tool 定义：** 底层原子能力，提供与外部系统交互的具体实现

**与 Skills 的区别：**

| 维度 | Tools | Skills |
|------|-------|--------|
| 粒度 | 原子操作 | 可组合 |
| 注册 | 系统注册 | 用户可创建 |
| 侧重 | 执行 | 智能决策 |

关系：

```
Agent → Skill → Tool
```

**内置 Tools：**

- HTTP 请求工具
- Shell 命令执行
- Python 脚本执行
- 文件读写操作
- 数据库查询工具

## 4. 数字助手（核心抽象）

数字助手是企业中的"AI 员工"

### 组成：

* 身份信息
* 运行时配置
* 模型配置
* Skills 集合
* 渠道绑定
* Memory
* Policy

### 助手状态：

- **草稿**：配置中，未启用
- **激活**：正常运行，可接收事件
- **停用**：临时禁用
- **归档**：历史版本归档

## 5. 执行面组件

### 5.1 Agent Runtime（远程执行节点）

**职责：**

* 消费任务 Event
* 执行 Agent 推理
* 调用 Skill

**特性（必须满足）：**

* 无状态（或弱状态）
* Worker 模式
* 不暴露 API

### 5.2 Edge Runtime（本地执行节点）

**职责：**

* 本地文件访问
* GUI 自动化（AX / UIA）
* 本地模型
* 用户环境交互

与远程 Runtime 的区别：

| 能力     | Edge | Remote |
| -------- | ---- | ------ |
| 本地文件 | 是   | 否     |
| GUI 操作 | 是   | 否     |
| 云执行   | 否   | 是     |

安全原则：

> Edge Runtime 是唯一可操作用户环境的组件

### 5.3 Skill Proxy（能力代理层）

**职责：**

统一 Skill 调用：

* 本地 Skill
* 远程 Skill
* MCP Skill（未来）

## 6. 关键执行链路（统一模型）

### 6.1 标准执行链路

```
User / Webhook
 ↓
Connector（事件标准化）
 ↓
Event Bus
 ↓
Event Engine（事件路由）
 ↓
Execution Engine（执行调度）
 ↓
┌────────────────────────────────┐
│  Agent Runtime / Workflow      │  ← 执行单元选择
│  Engine / Direct Skill Call    │
└────────────────────────────────┘
 ↓
Skill / Tool 执行
 ↓
Event Bus（流式返回）
 ↓
Assistant Service → Client
```

### 6.2 示例：GitHub PR 自动审查流程

1. **事件触发** - 开发者创建 PR，GitHub 发送 Webhook
2. **事件接收** - GitHub Connector 接收请求
3. **签名验证** - 验证 Webhook 签名确保来源合法
4. **事件标准化** - 转换为内部 Event 格式
5. **事件发布** - 发布到 Event Bus
6. **事件消费** - Event Engine 订阅并处理事件
7. **路由匹配** - Event Engine 根据事件类型选择 Handler
8. **执行触发** - Event Engine 调用 Execution Engine
9. **执行调度** - Execution Engine 决定执行策略（同步/异步/重试）
10. **节点选择** - Runtime Manager 选择合适的 Runtime 节点
11. **配置加载** - Agent Runtime 加载目标数字助手的配置
12. **上下文构建** - 获取 PR 差异内容，构建提示词
13. **能力注入** - 注入代码审查 Skills 和 GitHub Tools
14. **大模型推理** - Agent Runtime 调用 LLM 分析代码并生成审查意见
15. **工具执行** - Execution Engine 调用 GitHub API 发布 Review 评论
16. **结果返回** - 通过 Event Bus 流式返回执行结果
17. **结果记录** - 持久化到事件表

## 7. 安全模型

### 三层权限模型

```
Edge Runtime      → 高权限（本地）
Control Plane     → 中权限（调度）
Remote Runtime    → 低权限（执行）
```

### 核心规则

* Remote 不能访问本地
* 所有敏感操作必须经过 Policy Engine
* 全链路审计

### 安全边界

| 组件 | 权限级别 | 可访问资源 |
|------|----------|------------|
| Edge Runtime | 高 | 本地文件、GUI、用户环境 |
| Control Plane | 中 | 调度、路由、配置 |
| Remote Runtime | 低 | 云端资源、API |
| Policy Engine | 最高 | 权限决策、审计 |

## 8. Go 包结构（领域驱动设计）

### 8.1 设计原则

> **按"领域分层"，不是按技术分层**

- ❌ controller / service / dao
- ✅ event / execution / agent / skill

### 8.2 推荐的目录结构

```bash
backend/
│
├── cmd/                        # 启动入口（多进程）
│   ├── singer/                # 主服务（Event Engine + Execution Engine）
│   └── skill-proxy/           # Skill Proxy 服务
│
├── internal/                  # 私有核心代码（强制隔离）
│   ├── eventengine/          # ⭐ 事件引擎
│   │   ├── engine.go         # Event Engine 核心
│   │   ├── registry.go       # Handler 注册中心（插件化）
│   │   └── builtins/         # 内置事件处理器
│   │       ├── pr_handler.go
│   │       └── issue_handler.go
│   │
│   ├── execution/            # ⭐ 执行引擎
│   │   ├── engine.go         # Execution Engine
│   │   ├── dispatcher.go     # 调度器
│   │   ├── executor.go       # 执行器接口
│   │   ├── sync_executor.go  # 同步执行器
│   │   ├── async_executor.go # 异步执行器
│   │   └── context/          # 执行上下文
│   │
│   ├── agent/                # ⭐ Agent Runtime
│   │   ├── runtime.go        # Agent Runtime 接口
│   │   ├── lifecycle.go      # 生命周期管理
│   │   └── eino/             # Eino 实现
│   │       ├── eino_runtime.go
│   │       └── agent_runner.go
│   │
│   ├── connectors/           # 连接器
│   │   ├── connector.go      # Connector 接口
│   │   ├── github/
│   │   ├── gitlab/
│   │   └── wework/
│   │
│   ├── service/              # 服务层
│   │   ├── assistant_service.go
│   │   └── middleware/       # 中间件
│   │
│   ├── skill/                # Skill 系统
│   │   ├── registry.go       # Skill 注册中心
│   │   ├── executor.go       # Skill 执行器
│   │   └── builtin/          # 内置技能
│   │
│   ├── policy/               # 策略引擎
│   └── infra/                # 基础设施
│       ├── mq/               # 消息队列
│       ├── db/               # 数据库
│       └── logger/           # 日志
│
├── pkg/                      # 对外公开接口
│   ├── event/               # Event 定义（对外共享）
│   │   ├── event.go
│   │   └── topic.go
│   └── client/              # SDK（可选）
│
├── types/                    # 核心类型定义
├── config/                   # 配置管理
├── database/                 # 数据库
├── auth/                     # 认证系统
├── tools/                    # 工具定义
└── toolruntime/              # 工具运行时
```

### 8.3 目录说明

**`internal/` 目录：**
- Go 编译器强制保证只能被本项目内部引用
- 明确"内部实现"与"对外接口"的边界
- 为后续拆分多进程/微服务做准备

**`pkg/` 目录：**
- 对外公开的类型和 SDK
- 其他项目可以安全导入

**进程拆分建议：**

```bash
# Phase 1（当前）：单进程
cmd/singer/               # 主服务

# Phase 2：分离执行节点
cmd/server/               # API 服务
cmd/worker/               # 执行节点

# Phase 3：分离连接器
cmd/connector/            # 连接器进程
```

## 9. 技术栈

| 类别     | 技术                                 |
| -------- | ------------------------------------ |
| 语言     | Golang                               |
| 网关     | Gin                                  |
| 事件总线 | NATS JetStream                       |
| 数据库   | PostgreSQL                           |
| 缓存     | Redis                                |
| 向量库   | Qdrant                               |
| LLM      | 多模型（OpenAI / Claude / DeepSeek） |
| 容器化   | Docker + Compose                     |

## 10. 架构演进路径

### Phase 1（当前）

* 单运行时
* GitHub 自动化闭环
* 基础 Event Bus
* Connector 层完成
* Event Engine 与 Execution Engine 分离

### Phase 2

* 多 Runtime（OpenClaw / ClaudeCode）
* Runtime Manager
* 流式事件
* Agent Runtime 独立

### Phase 3

* Workflow Engine
* Memory + RAG
* Policy Engine 完整落地

### Phase 4

* 多租户
* Skill Marketplace
* 企业级治理能力

### Phase 5

* 进程拆分（Server / Worker / Connector）
* 分布式部署
* 水平扩展

## 11. 附录：架构演进历史

### v3.1 (2026-04-23) - Go 包结构优化

引入 **领域驱动设计** 和 **强制隔离** 原则：

- ✅ 使用 `internal/` 实现核心代码隔离
- ✅ 使用 `pkg/` 对外公开接口
- ✅ Event Engine Handler 插件化
- ✅ Skill Registry 化
- ✅ 接口优先设计（每层定义 interface）

### v3.0 (2026-04-23) - 三核架构重构

引入 **Event Engine + Execution Engine + Agent Runtime 三核架构**，解决职责分离问题：

- ✅ Orchestrator → Event Engine（专注事件处理）
- ✅ 新增 Execution Engine（专注执行控制）
- ✅ Agent Runtime 职责明确（专注 Agent 推理）
- ✅ Gateway → Assistant Service（明确对外服务定位）

### v2.0 (2026-04-23) - Agent Execution Kernel 架构

引入 Agent Execution Kernel + 分布式事件驱动架构

### 命名演变

| 版本 | 核心模块命名 |
|------|-------------|
| v1.0 | Gateway / Orchestrator / Agent Runtime |
| v2.0 | Gateway / Orchestrator / Agent Runtime（细化职责） |
| v3.0 | Assistant Service / Event Engine / Execution Engine / Agent Runtime（三核架构） |
| v3.1 | 引入 internal/ 和 pkg/ 强制隔离（领域驱动设计） |

## 12. 总结

### SingerOS 的本质：

> 一个 **事件驱动的分布式 Agent 操作系统**

### 核心能力：

* 多 Agent 编排
* 多 Runtime 执行
* 本地 + 云协同
* 企业级安全控制

### 架构关键词：

```
Event-Driven
Three-Core Architecture
Domain-Driven Design
Interface-First
Control / Execution Separation
Multi-Runtime
Edge + Cloud
Policy-Driven
Enforced Isolation (internal/)
```

### 核心架构公式：

```
Connector → Event → Event Engine → Execution Engine → Capability → Service
                                                ↓
                                    Agent Runtime / Workflow / Skill
```

### 常见错误清单（务必避免）

| ❌ 错误做法 | ✅ 正确做法 |
|------------|------------|
| 把所有逻辑写进 Event Handler | Handler → 调用 Execution Engine |
| Event Handler 使用 `switch` 硬编码路由 | Router 独立 + Handler 插件化 |
| Agent Runtime 直接调 MQ / DB | 通过 Execution Engine / Skill / Infra |
| Skill 写死在代码中 | 必须 Registry 化，支持动态注册 |
| 按技术分层（controller/service/model） | 按领域分层（event/execution/agent/skill） |
| 缺少接口定义，直接依赖实现 | 每层定义 interface，支持替换 |
