# 前端待完成事项

本文档记录 SingerOS 前端的待完成事项，基于当前代码中的 TODO 和桩状态。

## 通信与状态层

| 模块 | 状态 | 待完成 |
|------|------|--------|
| HTTP 客户端 | ✅ 完成 | baseURL 配置、Auth 拦截器接入 |
| SSE 客户端 | ✅ 完成 | 后端 SSE 端点接入 |
| WebSocket 客户端 | ✅ 完成 | 后端 WS 端点接入 |
| 话题 Slice | 🔄 桩 | topicService 实际调用替换 setTimeout 模拟 |
| 布局 Slice | 🔄 桩 | mock 数据替换为后端数据 |
| 其他 Slice | ❌ 未实现 | userSlice, appSettingsSlice, menuSlice, tabsSlice, configSlice, pluginSlice, notificationSlice, activeTasksSlice |

## 布局与交互层

| 模块 | 状态 | 待完成 |
|------|------|--------|
| CenterCanvas | 🔄 桩 | mockMessages 替换为实际 SSE/WS 数据流 |
| RightRail | 🔄 桩 | 文件收件箱/工件数据接入 |
| 模型选择 | 🔄 桩 | 模型路由 API 接入 |
| 输入框 @提及 | 🔄 桩 | DigitalAssistant/Agent 提及功能 |
| 输入框 /命令 | 🔄 桩 | Skill 命令触发功能 |
| 输入框附件 | 🔄 桩 | 文件上传功能 |
| BasicLayout | ❌ 未实现 | Sidebar, Header, PageTabs, NotificationPopover |
| BlankLayout | ❌ 未实现 | 登录页、原型编辑器等独立全屏页面 |

## 路由与认证层

| 模块 | 状态 | 待完成 |
|------|------|--------|
| 路由系统 | ❌ 未实现 | React Router 配置 + 路由守卫 + 动态路由 |
| 路由守卫 | ❌ 未实现 | 鉴权守卫、角色守卫、插件动态路由 |
| 认证系统 | ❌ 未实现 | Login / OAuth / Session / 强制改密码 |
| 菜单系统 | ❌ 未实现 | menus.ts 配置 + pluginStore 动态菜单 |

## 业务组件与 Hooks

| 模块 | 状态 | 待完成 |
|------|------|--------|
| business 组件库 | ❌ 未实现 | ProTable, DialogForm, SearchToolbar, TableActions 等 19 个组件 |
| useCrudPage | ❌ 未实现 | CRUD 分页列表通用逻辑 |
| useDialogForm | ❌ 未实现 | 对话框表单通用逻辑 |
| useDirtyCheck | ❌ 未实现 | 表单脏检查 + 关闭确认 |
| useInlineAi | ❌ 未实现 | 内联 AI 流式交互 |
| useAiFloat | ❌ 未实现 | AI 浮动助手 Prompt 传递 |
| useNotification | ❌ 未实现 | 通知系统（声音/桌面/标签页闪烁） |

## 页面模块 (30 个业务页面)

| 模块 | 状态 | 待完成 |
|------|------|--------|
| overview | ❌ 未实现 | 概览工作台 |
| chat | 🔄 桩 | Shell 三栏布局已实现，需接入 SSE 数据流 |
| login | ❌ 未实现 | 登录页 (BlankLayout) |
| profile | ❌ 未实现 | 个人设置页 |
| notifications | ❌ 未实现 | 通知中心 |
| exception | ❌ 未实现 | 403/404/500 异常页 |

| jenkins | ❌ 未实现 | Jenkins CI/CD + moduleConfig.ts |

| knowledge | ❌ 未实现 | 知识库 + moduleConfig.ts |
| reports | ❌ 未实现 | 报表 + moduleConfig.ts |
| schedules | ❌ 未实现 | 定时任务 + moduleConfig.ts |
| ai-employees | ❌ 未实现 | AI 员工配置 |
| skills | ❌ 未实现 | 技能管理 |
| plugins | ❌ 未实现 | 插件管理 |
| marketplace | ❌ 未实现 | 扩展市场 |
| channels | ❌ 未实现 | 通道管理 |
| contacts | ❌ 未实现 | 组织管理 |
| agents | ❌ 未实现 | Agent 路由 |
| ai-config | ❌ 未实现 | AI 模型配置 |
| webhooks | ❌ 未实现 | Webhook 管理 |
| remote-nodes | ❌ 未实现 | 工作节点 |
| logs | ❌ 未实现 | 运行日志 |
| upgrade | ❌ 未实现 | 系统升级 |

## API 层 (19 个模块 API)

| 模块 | 状态 | 待完成 |
|------|------|--------|
| auth.ts | ❌ 未实现 | 登录/登出/Token刷新 |
| chat.ts | ❌ 未实现 | AI 聊天 SSE 流式接口 |

| knowledge.ts | ❌ 未实现 | 知识库 API |
| ai-employees.ts | ❌ 未实现 | AI 员工 API |
| plugins.ts | ❌ 未实现 | 插件管理 API |
| skills.ts | ❌ 未实现 | 技能管理 API |
| channels.ts | ❌ 未实现 | 通道管理 API |
| contacts.ts | ❌ 未实现 | 组织管理 API |
| agents.ts | ❌ 未实现 | Agent 路由 API |
| webhooks.ts | ❌ 未实现 | Webhook 管理 API |
| 其他模块 API | ❌ 未实现 | reports, schedules, marketplace, remote-nodes, logs, upgrade, notifications, profile |

## AI 交互系统

| 模块 | 状态 | 待完成 |
|------|------|--------|
| AiFloat 组件 | ❌ 未实现 | 全局 AI 浮动助手 |
| AiAssistButton | ❌ 未实现 | AI 辅助按钮 |
| SSE 流式交互 | ❌ 未实现 | EventSource 建立与消息处理 |

## 其他

| 模块 | 状态 | 待完成 |
|------|------|--------|
| 插件化架构 | ❌ 未实现 | moduleConfig.ts 声明 + 动态路由注入 |
| 页面权限模型 | ❌ 未实现 | 路由 meta.requiredRole 控制 |
| Chunk 过期刷新 | ❌ 未实现 | vite:preloadError 监听 + sessionStorage 标记 |