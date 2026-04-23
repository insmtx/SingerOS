# SingerOS 前端架构文档

本文档是 SingerOS 前端架构的主索引文档，详细文档请查阅子文档。

## 技术栈概览

| 类别 | 技术 | 版本 |
|------|------|------|
| 框架 | React | 19.x |
| 语言 | TypeScript | 5.x (ES2020 target) |
| 构建 | Vite 7 + @vitejs/plugin-react-swc | SWC 编译器 |
| 样式 | TailwindCSS 4 + PostCSS | v4 新架构 |
| 状态管理 | Zustand 5 (traditional + middleware) | subscribeWithSelector + devtools |
| UI 基础 | @base-ui/react (无样式原语) | v1.x |
| 变体系统 | class-variance-authority (CVA) | v0.7 |
| CSS 合并 | clsx + tailwind-merge | cn() 工具函数 |
| 图标 | @tabler/icons-react | v3.x |
| 包管理 | Bun | lockb 格式 |
| 代码检查 | Biome 2 (替代 ESLint + Prettier) | formatter + linter |
| 测试 | Vitest 4 | --coverage 支持 |

## 文档索引

| 文档 | 说明 |
|------|------|
| [状态管理架构](./state-management.md) | Zustand Slice 模式 + ActionImpl 类架构详解 |
| [通信层架构](./communication.md) | HTTP、SSE、WebSocket 通信层实现 |
| [核心机制详解](./core-mechanisms.md) | 路由鉴权、AI交互、CRUD模式、插件化、通知、权限模型 |
| [架构设计模式](./design-patterns.md) | Slice 模式、ActionImpl、Reducer、乐观更新、原语包装、连接层抽象 |
| [组件与布局架构](./components-layout.md) | Shell 三栏架构、BasicLayout、BlankLayout、UI 原语、业务组件库 |
| [工程规范](./engineering-standards.md) | NPM Scripts、路径别名、TypeScript/Biome 配置、样式体系 |
| [待完成事项](./todo.md) | 通信层、布局层、路由层、组件/Hooks、页面模块、API层待完成清单 |

## 项目结构概览

```
frontend/
├── index.html              # 入口 HTML (root 挂载点)
├── vite.config.ts          # Vite 配置 (@ alias, react-swc, 代理)
├── tsconfig.json           # TS 配置 (strict, @/* paths)
├── biome.json              # Biome lint + format 配置
├── package.json            # 依赖与脚本
├── bun.lockb               # Bun 锁文件
├── public/                 # 静态资源 (不经过构建)
│
└── src/
    ├── main.tsx            # React 入口 (StrictMode 渲染 + Chunk 过期刷新)
    ├── App.tsx             # 根组件 (→ Shell/Router)
    ├── index.css           # 全局样式 (@import tailwindcss + @theme)
    │
    ├── api/                # API 层（按业务模块拆分，19 个模块）
    ├── router/             # 路由层（路由定义 + 守卫 + 动态路由 + 菜单配置）
    ├── pages/              # 页面视图（30 个业务模块）
    ├── components/         # UI 组件（layout + ui + business + ai-float）
    ├── store/              # Zustand 状态管理（11 个 Slice）
    ├── hooks/              # 自定义 React Hooks（10 个）
    ├── lib/                # 基础库（request + sse + websocket + utils）
    ├── types/              # 类型定义
    └── assets/             # 静态资源（icons + styles + brand）
```

## 架构分层概览

```
┌──────────────────────────────────────────────────────┐
│                    Pages (页面视图)                    │
│  30 个业务模块页面，按功能领域组织                        │
├──────────────────────────────────────────────────────┤
│               Layouts (布局框架)                       │
│  Shell（三栏布局）/ BasicLayout（侧边栏+头部+内容）       │
│  BlankLayout（独立全屏页）                              │
├──────────────────────────────────────────────────────┤
│            Components (公共组件)                       │
│  business 业务组件库 · ai-float · DockerTerminal        │
│  ui 原语包装组件 (56个)                                 │
├──────────────────────────────────────────────────────┤
│            Hooks (可复用逻辑)                           │
│  useCrudPage · useDialogForm · useInlineAi             │
│  useWebSocket · useNotification · useAiFloat           │
├──────────────────────────────────────────────────────┤
│            Store (Zustand Slice 模式)                  │
│  layout · topic · user · appSettings · plugin          │
│  notification · tabs · config · activeTasks │
├──────────────────────────────────────────────────────┤
│            Router (路由 + 守卫)                        │
│  鉴权守卫 · 角色守卫 · 插件动态路由                      │
├──────────────────────────────────────────────────────┤
│             API (HTTP + SSE 接口)                     │
│  19 个模块 API · HttpClient · SSE EventSource         │
├──────────────────────────────────────────────────────┤
│              Lib (基础库)                              │
│  request.ts · sse.ts · websocket.ts · utils.ts        │
├──────────────────────────────────────────────────────┤
│           @base-ui/react (无样式原语)                   │
│  CVA 变体系统 + cn() 工具函数 + TailwindCSS             │
└──────────────────────────────────────────────────────┘
```

## 快速导航

### 入口层

`main.tsx` → `App.tsx` → Router → Shell / BasicLayout / BlankLayout

- 渲染在 `#root` 挂载点，启用 `React.StrictMode`
- Chunk 过期自动刷新: 监听 `vite:preloadError` → `sessionStorage` 标记 + `reload()`

### 页面模块分类

| 分类 | 模块 |
|------|------|
| 核心工作 | overview, chat, workspace |
| 开发集成 | jenkins |
| AI 能力 | ai-employees, ai-config, skills |
| 知识与报表 | knowledge, reports |
| 运维管理 | schedules, remote-nodes, logs, upgrade |
| 系统配置 | plugins, marketplace, channels, contacts, agents, webhooks |
| 用户 | login, profile, notifications, exception |

## 相关文档

- [后端架构文档](../../backend/ARCHITECTURE.md)
- [布局风格设计](./Orbita_Layout_Arch.md)