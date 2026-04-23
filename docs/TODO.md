# 🚀 SingerOS 后端开发 TODO（最新更新 2026-04-18）

> 目标：2周内完成 GitHub PR 自动 Review MVP
> 原则：**先闭环，再抽象；先能跑，再优雅**

---

# 📊 当前实现状态摘要

## 已完成的核心组件（Phase 1-2）

| 组件 | 状态 | 说明 |
|------|------|------|
| Event Gateway | ✅ 完成 | GitHub webhook 接收、签名验证、事件解析 |
| Event Bus (RabbitMQ) | ✅ 完成 | Publisher/Subscriber 模式，topic-based 路由 |
| Orchestrator | ✅ 完成 | 事件路由到 Agent Runtime |
| Agent Runtime (Eino) | ✅ 完成 | LLM 代理执行引擎，工具调用 |
| Tools 系统 | ✅ 完成 | 注册表、执行运行时、权限解析 |
| Skills 系统 | ✅ 完成 | Skill 接口、文件化技能目录 |
| Auth 系统 | ✅ 完成 | OAuth 流程、账户管理、凭证解析 |
| GitHub 集成 | ✅ 完成 | Webhook、OAuth、PR 读写工具集 |
| 服务集成 | ✅ 完成 | Singer 主服务、Skill Proxy 框架 |

## 进行中的功能（Phase 3）

| 功能 | 状态 | 说明 |
|------|------|------|
| PR 自动 Review | 🔄 集成中 | 技能已定义，工具已实现，完整流程待验证 |
| Issue 自动回复 | 🔄 基础就绪 | issue_comment 事件已支持，需要 AI 回复技能 |
| DigitalAssistant 管理 | 🔄 类型完成 | 数据库模型完成，API 层待实现 |

# 🧭 总体阶段划分

| 阶段      | 目标          | 时间     |
| ------- | ----------- | ------ |
| Phase 1 | 跑通最小闭环（必须）  | Week 1 |
| Phase 2 | 抽象核心能力（可扩展） | Week 2 |
| Phase 3 | MVP业务完善     | Week 2 |
| Phase 4 | 可扩展能力（延后）   | 可选     |

---

# 🔴 Phase 1：最小闭环（✅ 已完成）

> 核心目标：**PR → 自动评论（端到端跑通）**

## ✅ 完成状态

所有 Phase 1 任务已完成：

- ✅ Docker化项目 - docker-compose 配置完成
- ✅ GitHub Webhook接入 - 签名验证、事件解析
- ✅ 事件发布到 RabbitMQ
- ✅ Orchestrator 事件消费
- ✅ LLM Provider (OpenAI)
- ✅ Skill 基础实现
- ✅ GitHub API 能力（PR 读写工具）
- ✅ 事件 → Skill 路由（Orchestrator）

---

# 🟡 Phase 2：核心抽象（✅ 已完成）

> 这一阶段只做**必要抽象，不做过度设计**

## ✅ 完成状态

- ✅ Skill 接口定义和实现
- ✅ Skill Manager (Catalog 系统)
- ✅ Orchestrator 去硬编码（基于 topic 路由）
- ✅ Agent Runtime (Eino 实现)
- ✅ Tools Registry 和 Runtime
- ✅ Auth 系统和 OAuth 流程

---

# 🟢 Phase 3：MVP业务完善（🔄 进行中）

> 当前重点：完成 PR 自动 Review 完整流程

## 3.1 已完成的基础能力

### ✅ GitHub 工具集

- ✅ 获取 PR 元数据 (`github.pr.get_metadata`)
- ✅ 获取 PR 文件列表 (`github.pr.get_files`)
- ✅ 对比 commits (`github.repo.compare_commits`)
- ✅ 读取文件内容 (`github.repo.get_file`)
- ✅ 发布 PR Review (`github.pr.publish_review`)

### ✅ PR Review Skill

- ✅ 文件化技能定义 (`backend/skills/github-pr-review/SKILL.md`)
- ✅ 审查流程和规则定义
- ✅ 工具依赖声明

### ✅ Agent Runtime

- ✅ Eino LLM 集成
- ✅ 工具调用能力
- ✅ 权限解析和账户管理

## 3.2 待完成的集成工作

### Task 13: 完整 PR Review 流程验证 🔄 进行中

**流程：**

```
GitHub PR Opened Webhook
    ↓
Event Gateway 接收并验证
    ↓
发布到 RabbitMQ
    ↓
Orchestrator 消费事件
    ↓
EinoRunner 执行 LLM Agent
    ↓
LLM 调用工具获取 PR 信息
    ↓
LLM 分析代码变更
    ↓
LLM 调用 github.pr.publish_review
    ↓
Review 发布到 GitHub
```

**需要验证：**

- [ ] Webhook 正确接收 PR opened 事件
- [ ] Event 正确发布和消费
- [ ] EinoRunner 正确执行
- [ ] 工具调用成功（需要有效的 GitHub token）
- [ ] Review 正确发布

**验收：**

- [ ] 在测试仓库提交 PR 后自动收到 AI Review
- [ ] Review 包含具体的代码分析
- [ ] 无错误日志

### Task 14: Issue Comment 自动回复 🔄 待实现

**当前状态：**

- ✅ `issue_comment` 事件已支持
- ✅ Event 路由已配置
- ❌ 需要 AI 回复技能

**需要实现：**

1. 创建 `issue-reply` SKILL.md
2. 验证 issue_comment 事件流
3. 测试回复发布

### Task 15: DigitalAssistant 配置管理 ❌ 待实现

**需要实现：**

1. DigitalAssistant CRUD API
2. 配置界面（可选）
3. Runtime 实例化
4. 事件到 Assistant 的绑定

---

## ✅ Phase 3 验收标准

- [ ] PR 自动 Review 端到端跑通
- [ ] Issue 自动回复可用
- [ ] 支持至少 2 种事件类型（PR + Issue）
- [ ] 所有工具调用使用正确的 OAuth 凭证

---

# 🔵 Phase 4：扩展能力（暂不优先）

> 以下功能目前不建议实现，除非有明确需求

## 4.1 Skill Proxy（远程化）

**现状：** ✅ 框架已完成

只有当你需要：

* 多实例部署
* 多语言 Skill
* 资源隔离

才需要进一步完善。

## 4.2 Memory 系统

只有当：

* 需要多轮对话
* 需要长上下文记忆

才需要实现。

**可能的实现：**

- Redis 存储短期记忆
- Vector DB 存储长期记忆
- 对话历史管理

## 4.3 多 Agent 编排

等复杂任务场景出现后再考虑。

## 4.4 多租户

目前单租户模式已足够 MVP 验证。

## 4.5 Workflow Engine

当前通过 Orchestrator + Eino Agent 已能处理基本流程。

---

---

# 📁 当前目录结构

```
backend/
├── cmd/
│   ├── singer/              # 主服务入口
│   └── skill-proxy/         # Skill Proxy 服务
├── config/                  # 配置管理
├── interaction/             # 事件交互层
│   ├── connectors/          # 渠道连接器
│   │   ├── github/          # GitHub 集成（完整）
│   │   ├── gitlab/          # GitLab 集成（stub）
│   │   └── wework/          # 企业微信（stub）
│   ├── eventbus/            # 事件总线
│   │   └── rabbitmq/        # RabbitMQ 实现
│   └── gateway/             # 事件网关
├── orchestrator/            # 事件编排器
├── runtime/                 # Agent Runtime
│   ├── eino/                # Eino 适配器
│   ├── prompt/              # 提示词构建
│   └── eino_runner.go       # Eino Runner 实现
├── tools/                   # Tools 系统
│   ├── registry.go          # 工具注册表
│   └── github/              # GitHub 工具集
├── toolruntime/             # Tool 运行时
├── skills/                  # 技能目录
│   └── github-pr-review/    # PR Review 技能
├── auth/                    # 认证授权系统
│   ├── providers/github/    # GitHub OAuth
│   └── service.go           # Auth 服务
├── types/                   # 领域类型定义
├── database/                # 数据库连接
├── gateway/                 # HTTP Gateway
│   └── trace/               # 请求追踪
└── providers/               # 外部服务提供者
    └── github/              # GitHub Client 工厂
```

---

# 🧪 强制开发规范（必须执行）

每个Task必须：

* ✅ `go build ./...` - 已验证可以通过
* ✅ `go test ./...` - 测试框架完整 
* ✅ 有最小测试 - 测试框架存在
* ✅ 日志可观测 - 已集成yg-go/logs

---

# 🧨 风险控制（重点）

### ❗ 不允许做的事情

* ❌ 不要一开始搞 Skill Proxy
* ❌ 不要设计复杂 Agent（Plan/Reflect）
* ❌ 不要做 Memory
* ❌ 不要做多租户
* ❌ 不要做抽象过度的 Workflow Engine

---

# 🎯 最终MVP定义（非常关键）

当满足：

```
1. PR opened ✅ *事件已捕获*
2. 自动分析diff ❌ *待完成 - 需要PR Diff获取和AI分析技能*
3. 自动评论 ❌ *待完成 - 需要AI分析和评论发布技能*
4. Issue自动回复 ❌ *待完成 - 需要AI回复和发布技能*
```

👉 **项目就算成功**

当前状态：基础事件管道已完成，AI业务处理逻辑(MVP核心部分)待完成
---

# 💡 最后一句（给你团队用的）

> 这个阶段的目标不是“做一个AI平台”，
> 而是**证明这个系统真的能帮人写代码评论**

---

# 🔍 当前项目状态摘要 (截至 2026-04-07)

## 已完成的基础设施组件:

### 1. LLM系统 ✅ 
- 接口定义完整 (`backend/llm/provider.go`)
- OpenAI Provider 实现 (`backend/llm/openai/provider.go`) 
- LLM Router 实现，支持多提供商和降级 (`backend/llm/router.go`)

### 2. Skill系统 ✅  
- 完整的Skill接口和BaseSkill抽象 (`backend/skills/skill.go`)
- SkillManager实现，支持注册、获取和执行 (`backend/skills/manager.go`)
- 示例Skills实现 (`backend/skills/examples/`, `backend/skills/tool_skills/`)

### 3. GitHub连接器 ✅
- Webhook接收和验证 (`backend/interaction/connectors/github/webhook.go`)
- 事件解析和转换为统一Event格式 (`backend/interaction/connectors/github/events.go`)
- 支持PR和Issue Comment事件

### 4. 事件总线和编排器 ✅
- 事件发布到RabbitMQ (`backend/interaction/eventbus/`)
- 事件消费和处理 (`backend/orchestrator/orchestrator.go`)
- 预设处理流程，包括PR和Issue事件

### 5. 数据库系统 ✅
- 数据库连接和初始化 (`backend/database/database.go`)
- 自动迁移机制
- 核心模型定义：DigitalAssistant, Event, User等 (`backend/types/`)

### 6. 主服务集成 ✅
- singer主服务 (`backend/cmd/singer/`)
- skill-proxy服务 (`backend/cmd/skill-proxy/`)

## 待完成的核心业务功能:

### 1. PR自动Review功能 ❌
- 获取PR Diff的具体实现
- 代码分析和审查逻辑  
- PR评论发布功能

### 2. Agent引擎 ❌ 
- Agent接口的实现
- 决策和执行逻辑
- 与Skills的编排机制

### 3. 数字助手(DigitalAssistant)配置和管理 ❌
- 数字助手实例的配置和运行
- 绑定PR/Issue事件到具体代理(agent)
- 持久化和管理接口

### 4. 更完整的Skill集成 ❌
- GitHub API调用的Skills（如PR评论、Issue回复）
- AI相关的Skills（代码分析、摘要生成等）

## 总体状态评估:

- Phase 1 基础架构: ✅ **大部分完成** (除Docker化)
- Phase 2 核心抽象: ✅ **Skill系统和LLM完成，去硬编码和Agent待完成**
- Phase 3 MVP业务: ❌ **核心AI业务功能需要补充**
- 风险: 连续性和数据一致性机制需要加强测试

---

如果你下一步要推进，我可以帮你再做一版：

👉 **“团队分工版本（2-3人怎么拆任务）”**
👉 **“每个Task对应PR粒度（直接能用Git管理）”**
