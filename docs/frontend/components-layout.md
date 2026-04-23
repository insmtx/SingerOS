# 组件与布局架构

本文档详细描述 SingerOS 前端的组件和布局架构。

## 布局层架构

### Shell 三栏架构 (AI 聊天交互)

```
┌─────────────┬─────────────────────┬──────────────┐
│  LeftRail   │   CenterCanvas      │  RightRail   │
│  (260px)    │   (flex-1)          │  (280px)     │
│             │                     │              │
│  会话列表    │  标题栏              │  快捷操作     │
│  工作区分组  │  消息时间轴          │  文件收件箱   │
│             │  输入框 + 模型选择    │  工件文件     │
└─────────────┴─────────────────────┴──────────────┘
```

**Shell** 使用 Flexbox 三栏布局 (`flex h-screen w-screen overflow-hidden bg-slate-50`)，各栏通过 `border-r/border-l border-slate-200` 分隔。

### LeftRail — 会话导航区

- 工作区（远程/本地）分组，可折叠
- 会话列表，支持创建/删除/切换
- 底部：新建工作区按钮

### CenterCanvas — 聊天交互区

- 标题栏 (当前会话名称)
- 消息时间轴 (用户气泡白色 / AI 回复衬线字体)
- 浮动输入框 (textarea + @提及 + 附件 + 模型选择 + 发送/停止)

### RightRail — 信息区

- 三个 Tab：快捷操作 / 文件收件箱 / 工件文件
- Tab 切换通过 Zustand `activeRightTab` 状态控制

### BasicLayout — 业务页面主布局

- Sidebar：侧边栏导航 (menus.ts 菜单配置 + pluginStore 动态菜单)
- Header：顶部栏 (用户信息 · 通知 · 任务指示器)
- PageTabs：多标签页切换 (tabsStore)
- Content：RouterOutlet (各业务页面)
- AiFloat：全局 AI 浮动助手
- 强制改密码弹窗 (userStore.mustChangePassword)

### BlankLayout — 空白布局

- 登录页、原型编辑器等独立全屏页面

## 布局组件目录

```
components/layout/
├── Shell.tsx           # 三栏布局容器
├── BasicLayout.tsx     # 主布局：侧边栏 + 头部 + 标签页 + 内容 + AI浮动
├── BlankLayout.tsx     # 空白布局（登录/原型编辑器）
├── LeftRail.tsx        # 左栏 - 会话导航
├── CenterCanvas.tsx    # 中栏 - 聊天交互区
├── RightRail.tsx       # 右栏 - 快捷/收件/工件
├── Sidebar.tsx         # 侧边栏导航
├── Header.tsx          # 顶部栏
├── PageTabs.tsx        # 多标签页切换
├── NotificationPopover.tsx # 通知弹出面板
└── index.ts            # barrel 导出
```

## UI 原语组件层 (`components/ui/`)

基于 **@base-ui/react** 无样式原语 + **CVA (class-variance-authority)** 变体系统：

```
@base-ui/react (无样式行为原语)
    ↓ 包装
UI 组件 (CVA 变体 + TailwindCSS 样式)
    ↓ 使用
layout 组件 / 业务组件
```

### Button 变体示例

```ts
const buttonVariants = cva('基础样式', {
  variants: {
    variant: ['default', 'outline', 'secondary', 'ghost', 'destructive', 'link'],
    size: ['default', 'xs', 'sm', 'lg', 'icon', 'icon-xs', 'icon-sm', 'icon-lg'],
  },
});
```

### 已实现 UI 组件清单 (56个)

accordion, alert, alert-dialog, aspect-ratio, avatar, badge, breadcrumb, button, button-group, calendar, card, carousel, chart, checkbox, collapsible, combobox, command, context-menu, dialog, drawer, dropdown-menu, empty, field, form, hover-card, input, input-group, input-otp, item, kbd, label, menubar, navigation-menu, pagination, popover, progress, radio-group, resizable, scroll-area, select, separator, sheet, sidebar, skeleton, slider, sonner, spinner, switch, table, tabs, textarea, toggle, toggle-group, tooltip

## 业务组件库 (`components/business/`)

基于 ui 原语组合的业务 UI 库：

```
components/business/
├── index.ts           # 统一导出
├── ProTable.tsx       # 高级数据表格
├── DialogForm.tsx     # 对话框表单
├── PopForm.tsx        # 弹出式表单
├── SearchToolbar.tsx  # 搜索工具栏
├── TableActions.tsx   # 表格操作列
├── DeleteRecord.tsx   # 删除确认
├── DeptTree.tsx       # 部门树
├── MemberPicker.tsx   # 成员选择器
├── MemberWizard.tsx   # 成员向导
├── PopoverSelect.tsx  # 弹出选择器
├── Editor.tsx         # TipTap 富文本编辑器
├── Icon.tsx           # 图标组件
├── AiAssistButton.tsx # AI 辅助按钮
├── BugTracking.tsx    # Bug 追踪
├── StoryTracking.tsx  # Story 追踪
├── WorkItem.tsx       # 工作项
├── ActiveTaskIndicator.tsx # 全局任务执行指示器
├── BrowserPreview.tsx # 浏览器预览
├── DockerTerminal.tsx # Docker 终端 (xterm.js)
└── utils/             # 业务工具函数
```

## AI 浮动助手组件 (`components/ai-float/`)

全局侧边 AI 助手组件：

```
components/ai-float/
├── AiFloat.tsx      # 全局侧边 AI 助手组件
├── AiFloatPrompt.tsx # Prompt 传递
└── index.ts
```

## 依赖关系图

```
App.tsx
  └─ Provider (主题注入 + Store Provider)
  └─ RouterProvider
      ├─ BlankLayout → login
      ├─ Shell (三栏布局) → chat 页面
      └─ BasicLayout
          ├─ Sidebar (menus.ts 菜单配置 + pluginStore 动态菜单)
          ├─ Header (用户信息 · 通知 · 任务指示器)
          ├─ PageTabs (tabsStore)
          ├─ Content → Outlet (各业务页面)
          ├─ AiFloat (全局 AI 浮动助手)
          └─ 强制改密码弹窗 (userStore.mustChangePassword)

每个业务页面:
  └─ business 组件 (ProTable · DialogForm · SearchToolbar)
  └─ hooks (useCrudPage · useDialogForm)
  └─ API 模块
  └─ Slice 模块
```