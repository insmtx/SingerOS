# SingerOS 后端目录结构

> 基于领域驱动设计 (DDD) 的 Go 包结构
>
> **最后更新：2026-04-23**

## 目录概览

```
backend/
├── cmd/                        # 应用入口
│   ├── singer/                # 主服务
│   └── skill-proxy/           # Skill Proxy 服务
│
├── internal/                  # 私有核心代码（强制隔离）
│   ├── eventengine/          # 事件引擎
│   ├── execution/            # 执行引擎
│   ├── agent/                # Agent Runtime
│   ├── connectors/           # 连接器（外部接入）
│   ├── service/              # 服务层（API）
│   ├── skill/                # Skill 系统
│   ├── policy/               # 策略引擎
│   ├── session/              # 会话管理（规划中）
│   ├── workflow/             # 工作流引擎（规划中）
│   └── infra/                # 基础设施
│       ├── mq/               # 消息队列
│       ├── db/               # 数据库
│       └── logger/           # 日志
│
├── pkg/                      # 对外公开接口
│   ├── event/               # Event 定义
│   └── client/              # SDK
│
├── types/                    # 核心类型定义
├── config/                   # 配置管理
├── database/                 # 数据库
├── auth/                     # 认证系统
├── tools/                    # 工具定义
├── toolruntime/              # 工具运行时
├── providers/                # 第三方服务提供者
├── skills/                   # Skill 定义（旧）
├── runtime/                  # Runtime（旧，待删除）
├── interaction/              # Interaction（旧，待删除）
├── gateway/                  # Gateway（旧，待删除）
├── orchestrator/             # Orchestrator（旧，待删除）
├── clientmgr/                # 客户端管理
└── tests/                    # 测试文件
```

## 模块说明

### internal/ - 核心私有代码

| 模块 | 职责 | 状态 |
|------|------|------|
| `eventengine/` | Event Engine - 事件订阅、路由、Handler 调用 | ✅ 已迁移 |
| `execution/` | Execution Engine - 任务执行、重试、超时控制 | ✅ 已完成 |
| `agent/` | Agent Runtime - Agent 生命周期、LLM 调用 | ✅ 已迁移 |
| `connectors/` | Connectors - GitHub/GitLab/WeWork 接入 | ✅ 已迁移 |
| `service/` | Service Layer - API 层、中间件 | 🔄 进行中 |
| `skill/` | Skill System - 技能注册、执行 | 🔄 进行中 |
| `policy/` | Policy Engine - 权限控制、审计 | 📋 规划中 |
| `session/` | Session Management - 会话管理 | 📋 规划中 |
| `workflow/` | Workflow Engine - 流程编排 | 📋 规划中 |
| `infra/` | Infrastructure - MQ、DB、Logger | 🔄 进行中 |

### pkg/ - 对外公开接口

| 模块 | 职责 | 状态 |
|------|------|------|
| `event/` | Event 定义（对外共享） | 📋 规划中 |
| `client/` | SingerOS SDK | 📋 规划中 |

### 遗留模块（待删除）

| 模块 | 新位置 | 状态 |
|------|--------|------|
| `orchestrator/` | `internal/eventengine/` | ⚠️ 待删除 |
| `runtime/` | `internal/agent/` | ⚠️ 待删除 |
| `interaction/` | `internal/connectors/` + `internal/infra/mq/` | ⚠️ 待删除 |
| `gateway/` | `internal/service/` | ⚠️ 待删除 |

## 迁移进度

| 阶段 | 任务 | 状态 |
|------|------|------|
| Phase 1 | 创建目录骨架 | ✅ 完成 |
| Phase 2 | 迁移 Event Engine | ✅ 完成 |
| Phase 3 | 创建 Execution Engine | ✅ 完成 |
| Phase 4 | 迁移 Agent Runtime | ✅ 完成 |
| Phase 5 | 迁移 Connectors/Infra | ✅ 完成 |
| Phase 6 | 实现 Skill API | 📋 待开始 |
| Phase 7 | 清理旧代码 | 📋 待开始 |

## 编译验证

```bash
# 编译所有代码
go build ./...

# 运行所有测试
go test ./...

# 静态检查
go vet ./...

# 代码格式化
gofmt -s -w .
```
