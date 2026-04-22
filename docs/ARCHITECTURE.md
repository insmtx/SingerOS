# SingerOS 架构设计文档

> 基于 **Agent Execution Kernel + 分布式事件驱动架构** 构建的企业级 AI 操作系统
>
> **版本：2.0** | **最后更新：2026-04-23**

## 1. 核心愿景

构建一个企业级数字员工平台，让企业可以像管理真实员工一样，创建、配置、授权、调度和审计 AI 数字员工，并实现：

* **多 Agent 协作** - 多个智能体协同工作
* **多运行时执行** - 支持不同 Agent 引擎并存
* **本地 + 云端协同** - Edge 与 Remote Runtime 分工
* **可控、安全、可审计** - 企业级安全控制

数字员工不是单纯的聊天机器人。它需要有独立身份、接收任务的入口、真实执行工作的环境，以及模型、工具、技能、知识库等基础能力。

## 设计原则

* **事件驱动（Event-Driven First）**
  所有行为统一抽象为 Event，通过 EventBus 传播
* **控制面 / 执行面分离（Control vs Execution）**
  决策与执行彻底解耦
* **多运行时（Multi-Runtime）**
  支持不同 Agent 引擎并存（Eino / OpenClaw / ClaudeCode）
* **边缘优先（Edge-First）**
  本地能力（文件 / GUI）优先由 Edge Runtime 执行
* **安全优先（Security by Design）**
  明确本地与远程执行边界
* **数字助手是最高抽象（Digital Assistant First）**
  代表完整的 AI 数字员工实例
* **绑定关系为核心**
  IM身份、外部连接、工作节点构成数字员工的工作入口

## 2. 分层架构（三平面模型）

### 2.1 架构总览

```
┌────────────────────────────────────────────┐
│                Client / Edge               │
│  App / CLI / 本地 Agent Runtime (Edge)    │
└────────────────────┬───────────────────────┘
                     │
                     ▼
┌────────────────────────────────────────────┐
│            Control Plane（控制面）         │
│  Gateway / Orchestrator / Memory / Policy │
└────────────────────┬───────────────────────┘
                     │
                     ▼
┌────────────────────────────────────────────┐
│          Execution Plane（执行面）         │
│   Remote Agent Runtime / Skill Workers    │
└────────────────────────────────────────────┘
```

### 2.2 三平面职责

| 平面 | 组件 | 职责 |
|------|------|------|
| **Edge Plane** | Edge Runtime / Client | 本地文件访问、GUI 自动化、用户环境交互 |
| **Control Plane** | Gateway / Orchestrator / Runtime Manager / Memory / Policy | 决策中心：Session 管理、Agent 路由、上下文构建 |
| **Execution Plane** | Remote Agent Runtime / Skill Proxy | 云端执行：Agent 推理、Skill 调用、Tool 执行 |

### 2.3 核心数据通道（统一事件流）

```
External Event / User Input
        ↓
Event Gateway
        ↓
EventBus（统一事件模型）
        ↓
Control Plane（决策）
        ↓
Execution Plane（执行）
        ↓
EventBus（响应流）
        ↓
Client / UI
```

> **核心原则**：所有模块之间只能通过 EventBus 通信

## 3. 核心模块划分

### 3.1 Event Gateway（事件网关）

**职责：**

* 接收外部系统事件（Webhook / API / 用户输入）
* 标准化为内部 Event
* 发布到 EventBus

**支持渠道：**

* GitHub / GitLab
* 企业微信 / 飞书
* CLI / Web UI

**关键能力：**

* 签名验证
* 多协议适配
* 事件转换

### 3.2 EventBus（事件总线）

**职责：**

系统唯一通信通道

> 所有模块之间只能通过 EventBus 通信

**实现：**

* 当前：RabbitMQ
* 推荐演进：NATS（更适合事件驱动架构）

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

### 3.3 Control Plane（控制面 - 核心大脑）

**职责：**

* Session 管理
* Agent 路由
* 上下文构建
* 调用 Orchestrator

本质：

> **"系统的决策中心"**

### 3.4 Orchestrator（编排引擎）

**职责：**

* Agent 调度
* Workflow 执行
* 多 Agent 协作

支持未来：

* DAG / Workflow
* 类 Temporal 执行模型

### 3.5 Runtime Manager（运行时调度器）【新增关键模块】

**职责：**

* 管理所有 Runtime 实例
* 能力注册（Skill / GPU / Browser）
* 负载均衡
* 健康检查

类比：

> Kubernetes Scheduler（简化版）

### 3.6 Memory（记忆系统）

**职责：**

* 会话上下文（短期记忆）
* 长期记忆（向量）
* 知识检索（RAG）

### 3.7 Model Router（模型调度）

**职责：**

* 多模型管理
* fallback / 降级
* 成本控制

### 3.8 Policy（安全与权限）【新增关键模块】

**职责：**

* Agent 行为控制
* Skill 调用权限
* 审计日志

强制规则：

* Remote Runtime 不得直接访问本地资源
* 所有高权限操作必须经过 Policy

### 3.9 Skills 能力系统

**Skill 定义：** 可复用的 AI 能力单元，是 SingerOS 的核心构建块

**Skill 元数据：**
- 唯一标识符
- 名称和描述
- 版本号
- 分类（集成类、AI 类、工具类、工作流类）
- 输入输出模式定义
- 权限声明

**Skill 分类：**
- **集成类 Skills** - 外部系统集成（GitHub、GitLab、飞书等）
- **AI 类 Skills** - 基于大模型的推理能力（代码审查、摘要生成、分类等）
- **工具类 Skills** - 底层工具能力（Shell 执行、Python 脚本、HTTP 请求等）
- **工作流类 Skills** - 复杂编排能力（PR 审查工作流、Bug 分类工作流等）

**技能加载方式：**
- 文件系统：通过 SKILL.md 文件定义
- 代码嵌入：编译时打包的内置技能
- 远程加载：从技能市场动态下载（规划中）

### 3.10 Tools 工具系统

**Tool 定义：** 底层原子能力，提供与外部系统交互的具体实现

**与 Skills 的区别：**
- Tools 是原子操作，Skills 可以组合多个 Tools
- Tools 由系统注册，Skills 可以由用户创建
- Tools 侧重执行，Skills 侧重智能决策

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

## 5. Execution Plane（执行面）

### 5.1 Agent Runtime（远程执行节点）

**职责：**

* 消费任务 Event
* 执行 Agent 推理
* 调用 Skill

**特性（必须满足）：**

* 无状态（或弱状态）
* Worker 模式
* 不暴露 API

### 5.2 Edge Runtime（本地执行节点）【新增关键模块】

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

```
User / Webhook
 ↓
Event Gateway
 ↓
EventBus
 ↓
Control Plane
 ↓
Orchestrator
 ↓
Runtime Manager（选节点）
 ↓
Agent Runtime / Edge Runtime
 ↓
Skill / Tool 执行
 ↓
EventBus（流式返回）
 ↓
Client
```

### 示例：GitHub PR 自动审查流程

1. **事件触发** - 开发者创建 PR，GitHub 发送 Webhook
2. **事件接收** - Event Gateway 的 GitHub Connector 接收请求
3. **签名验证** - 验证 Webhook 签名确保来源合法
4. **事件标准化** - 转换为内部 Event 格式
5. **事件发布** - 发布到 EventBus
6. **事件消费** - Control Plane 订阅并处理事件
7. **路由匹配** - Orchestrator 根据事件类型选择处理器
8. **节点选择** - Runtime Manager 选择合适的 Runtime 节点
9. **配置加载** - Runtime 加载目标数字助手的配置
10. **上下文构建** - 获取 PR 差异内容，构建提示词
11. **能力注入** - 注入代码审查 Skills 和 GitHub Tools
12. **大模型推理** - 调用 LLM 分析代码并生成审查意见
13. **工具执行** - 调用 GitHub API 发布 Review 评论
14. **结果返回** - 通过 EventBus 流式返回执行结果
15. **结果记录** - 持久化到事件表

## 7. 安全模型

### 三层权限模型

```
Edge Runtime      → 高权限（本地）
Control Plane     → 中权限（调度）
Remote Runtime    → 低权限（执行）
```

### 核心规则

* Remote 不能访问本地
* 所有敏感操作必须经过 Policy
* 全链路审计

### 安全边界

| 组件 | 权限级别 | 可访问资源 |
|------|----------|------------|
| Edge Runtime | 高 | 本地文件、GUI、用户环境 |
| Control Plane | 中 | 调度、路由、配置 |
| Remote Runtime | 低 | 云端资源、API |
| Policy | 最高 | 权限决策、审计 |

## 8. 技术栈

| 类别     | 技术                                 |
| -------- | ------------------------------------ |
| 语言     | Golang                               |
| 网关     | Gin                                  |
| 事件总线 | NATS（推荐）/ RabbitMQ（当前）       |
| 数据库   | PostgreSQL                           |
| 缓存     | Redis                                |
| 向量库   | Qdrant                               |
| LLM      | 多模型（OpenAI / Claude / DeepSeek） |
| 容器化   | Docker + Compose                     |

## 9. 架构演进路径

### Phase 1（当前）

* 单运行时（Eino）
* GitHub 自动化闭环
* 基础 EventBus

### Phase 2

* 多 Runtime（OpenClaw / ClaudeCode）
* Runtime Manager
* 流式事件

### Phase 3

* Workflow Engine
* Memory + RAG
* Policy 完整落地

### Phase 4

* 多租户
* Skill Marketplace
* 企业级治理能力

## 10. 总结

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
Control / Execution Separation
Multi-Runtime
Edge + Cloud
Policy-Driven
```