# SingerOS 后端架构设计文档

> 面向 AI OS 的 Golang 包结构指南
>
> **版本：3.1** | **最后更新：2026-04-23**

## 1. 概述

本文档提供 SingerOS 后端的 **Golang 包结构设计**，与 `ARCHITECTURE.md` 配合使用。

- `ARCHITECTURE.md` - 高层架构设计、模块划分、执行链路
- `ARCHITECTURE_BACKEND.md` - **本文档** - Go 包结构、目录组织

## 2. 设计原则

### 2.1 按"领域分层"，不是按技术分层

> ❌ 旧模式：controller / service / dao / model
> ✅ 新模式：event / execution / agent / skill

**原因：**
- 技术分层导致模块间耦合严重
- 领域分层让每个模块职责清晰、可独立演进

### 2.2 核心引擎必须"内聚 + 可替换"

- Event Engine 可以单独部署
- Execution Engine 可以替换
- Agent Runtime 可扩展

### 2.3 接口优先（interface-driven）

每一层都必须定义 interface，而不是直接依赖实现

### 2.4 强制隔离（Enforced Isolation）

| 目录 | 用途 | 访问控制 |
|------|------|----------|
| `internal/` | 私有核心代码 | Go 编译器强制隔离，只能被本项目内部引用 |
| `pkg/` | 对外公开接口 | 其他项目可安全导入 |

## 3. 推荐的 Golang 包结构

```bash
singeros/backend/
│
├── cmd/                       # 启动入口（多进程）
│   └── singer/                # 主服务（HTTP + Engine）
│
├── internal/                  # 私有核心代码（强制隔离）
│   ├── eventengine/           # ⭐ 事件引擎
│   ├── execution/             # ⭐ 执行引擎
│   ├── agent/                 # ⭐ Agent Runtime
│   ├── skill/                 # ⭐ Skill 体系
│   ├── connectors/            # ⭐ 外部接入
│   ├── service/               # ⭐ 对外 API 层
│   ├── policy/                # 策略引擎
│   ├── session/               # 会话管理（规划中）
│   ├── workflow/              # 工作流引擎（规划中）
│   └── infra/                 # 基础设施
│
├── pkg/                       # 可复用库（可对外）
│   ├── event/                 # Event 定义（对外共享）
│   ├── client/                # SDK（调用 SingerOS）
│   └── providers/             # 第三方服务提供者
│
├── types/                     # 核心类型定义
├── config/                    # 配置管理
├── database/                  # 数据库
├── auth/                      # 认证系统
├── tools/                     # 工具定义
└── toolruntime/               # 工具运行时
```

## 4. 核心模块说明

### 4.1 `internal/eventengine/` - 事件引擎

**职责：** 事件订阅、路由、Handler 调用

**子目录：**
- `engine.go` - Event Engine 核心
- `registry.go` - Handler 注册中心（插件化）
- `router.go` - 事件路由（不写死 switch）
- `builtins/` - 内置事件处理器（PR、Issue、Push 等）

**⚠️ 常见错误：**
- ❌ 把业务逻辑直接写在 Handler 中
- ❌ 使用 `switch` 硬编码路由
- ✅ 正确：Handler → 调用 Execution Engine

### 4.2 `internal/execution/` - 执行引擎

**职责：** 任务调度、执行控制、重试/超时管理

**子目录：**
- `engine.go` - Execution Engine 核心
- `dispatcher.go` - 调度器（任务分发）
- `executor.go` - 执行器接口
- `sync_executor.go` / `async_executor.go` - 同步/异步执行器
- `retry.go` / `timeout.go` - 重试和超时控制
- `context/` - 执行上下文

**关键点：**
- 支持同步/异步执行
- 支持重试和降级
- 支持超时控制

### 4.3 `internal/agent/` - Agent Runtime

**职责：** Agent 生命周期管理、LLM 调用、上下文维护

**子目录：**
- `runtime.go` - Agent Runtime 接口
- `lifecycle.go` - 生命周期管理
- `context.go` - 上下文管理
- `reasoning.go` - 推理循环
- `eino/` - Eino 具体实现

**⚠️ 常见错误：**
- ❌ Agent Runtime 直接调用 MQ / DB
- ✅ 必须通过 Execution Engine / Skill / Infra

### 4.4 `internal/skill/` - Skill 体系

**职责：** 技能注册、执行、管理

**子目录：**
- `registry.go` - Skill 注册中心（必须动态注册）
- `executor.go` - Skill 执行器
- `base_skill.go` - 基础 Skill 实现
- `builtin/` - 内置技能

**⚠️ 常见错误：**
- ❌ Skill 写死在代码中
- ✅ 必须 Registry 化，支持动态注册

### 4.5 `internal/connectors/` - 连接器

**职责：** 外部系统接入（GitHub、GitLab、飞书等）

**子目录：**
- `connector.go` - Connector 接口
- `github/` - GitHub 连接器
- `gitlab/` - GitLab 连接器
- `wework/` - 企业微信连接器

### 4.6 `internal/service/` - 服务层

**职责：** 对外 API 入口

**子目录：**
- `assistant_service.go` - 助手服务
- `session_service.go` - 会话服务（规划中）
- `middleware/` - 中间件（CORS、日志、Recovery）

### 4.7 `internal/policy/` - 策略引擎

**职责：** 权限控制、审计日志

**子目录：**
- `engine.go` - 策略引擎
- `permission.go` - 权限控制
- `audit.go` - 审计日志

### 4.8 `internal/infra/` - 基础设施

**职责：** 统一基础设施访问

**子目录：**
- `mq/` - 消息队列（Publisher / Subscriber）
- `db/` - 数据库
- `logger/` - 日志

### 4.9 `pkg/` - 对外公开接口

**职责：** 对外共享的类型和 SDK

**子目录：**
- `event/` - Event 定义（event.go、topic.go）
- `client/` - SingerOS SDK

## 5. 进程拆分建议

### 5.1 为什么需要进程拆分？

| 优势 | 说明 |
|------|------|
| 水平扩展 | 不同组件独立扩缩容 |
| 解耦 | 故障隔离 |
| 负载分离 | 不同负载类型分开处理 |

### 5.2 推荐的进程拆分方案

#### Phase 1（当前）：单进程

```bash
cmd/singer/               # 主服务（所有功能）
```

#### Phase 2：分离执行节点

```bash
cmd/server/               # API 服务（HTTP + Event Engine）
cmd/worker/               # 执行节点（Execution Engine + Agent Runtime）
```

#### Phase 3：分离连接器

```bash
cmd/connector/            # 连接器进程（Connectors + Event Bus Publisher）
```

### 5.3 进程间通信

所有进程间通过 **Event Bus** 通信：

```
Connector Process → Event Bus → Event Engine Process → Execution Engine Process → Agent Runtime Worker
```

## 6. 常见错误与最佳实践

### 6.1 常见错误

| ❌ 错误做法 | ✅ 正确做法 |
|------------|------------|
| 把所有逻辑写进 Event Handler | Handler → 调用 Execution Engine |
| Event Handler 使用 `switch` 硬编码路由 | Router 独立 + Handler 插件化 |
| Agent Runtime 直接调 MQ / DB | 通过 Execution Engine / Skill / Infra |
| Skill 写死在代码中 | 必须 Registry 化，支持动态注册 |
| 按技术分层（controller/service/model） | 按领域分层（event/execution/agent/skill） |
| 缺少接口定义，直接依赖实现 | 每层定义 interface，支持替换 |

### 6.2 最佳实践

1. **每个包只暴露必要的接口**
2. **使用 `internal/` 强制隔离核心实现**
3. **使用 `pkg/` 对外公开稳定接口**
4. **Handler 必须插件化，不写死 `switch`**
5. **Skill 必须 Registry 化**
6. **每个引擎独立测试**
7. **依赖注入，避免全局变量**

## 7. 与 ARCHITECTURE.md 的对应关系

| ARCHITECTURE.md 概念 | 对应 Go 包 |
|---------------------|-----------|
| Event Engine | `internal/eventengine/` |
| Execution Engine | `internal/execution/` |
| Agent Runtime | `internal/agent/` |
| Skill System | `internal/skill/` |
| Connector | `internal/connectors/` |
| Assistant Service | `internal/service/` |
| Event Bus | `internal/infra/mq/` |
| Policy Engine | `internal/policy/` |

## 8. 下一步行动

### 立即可做（低成本高收益）

1. 创建 `internal/` 目录结构
2. 移动现有代码到对应领域目录
3. 定义各层 interface
4. Event Engine Handler 插件化

### 中期优化

1. 实现 Execution Engine 的重试/超时控制
2. 完善 Skill Registry
3. 添加 Policy Engine 基础框架

### 长期规划

1. 进程拆分（Server / Worker / Connector）
2. 分布式部署
3. 水平扩展

## 9. 总结

SingerOS 后端应该从：

```
MVC / service-based
```

升级为：

```
Event-Driven + Engine-Oriented + Runtime-Based + Domain-Driven
```

**核心原则：**

- 按领域分层，不按技术分层
- 接口优先，支持替换
- 核心引擎内聚可替换
- 强制隔离（internal/）
- 对外公开（pkg/）
