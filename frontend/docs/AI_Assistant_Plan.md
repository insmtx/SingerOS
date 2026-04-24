# SingerOS 前端功能规划 — AI 助手模块

## 变更记录

### Phase 1 + Phase 2 (已完成)

**新增文件**：
| 文件 | 说明 |
|------|------|
| `src/types/chat.ts` | 聊天领域类型（Message, ToolCall, Attachment 等） |
| `src/store/slices/chatSlice.ts` | Zustand chat slice（消息流、输入、Mock 流式生成、resendMessage、tokenUsage） |
| `src/utils/format.ts` | 共享格式化工具（formatTime, formatDate, formatFileSize） |
| `src/mocks/streamSimulator.ts` | Mock 流式数据生成器（逐字输出 + 工具调用模拟） |
| `src/mocks/chatMocks.ts` | 预置会话和消息 Mock 数据（3 个会话场景） |
| `src/hooks/use-mobile.ts` | 修复缺失的 useIsMobile hook |
| `src/components/layout/TopBar.tsx` | 全局状态栏（品牌、AI 状态、用户菜单） |
| `src/components/chat/ChatHeader.tsx` | 会话标题栏（动态标题/模型/token 计数） |
| `src/components/chat/MessageTimeline.tsx` | 消息时间轴（自动滚动、min-h-0 flex-1） |
| `src/components/chat/UserMessageBubble.tsx` | 用户蓝色渐变气泡（hover 复制按钮） |
| `src/components/chat/AIMessageBubble.tsx` | AI 气泡（Markdown + 流式光标 + 复制/重新生成） |
| `src/components/chat/ToolCallBlock.tsx` | 工具调用折叠/展开展示 |
| `src/components/chat/TypingIndicator.tsx` | 脉冲输入指示器 |
| `src/components/chat/DateDivider.tsx` | 日期分割线 |
| `src/components/chat/WelcomeScreen.tsx` | 空状态快捷建议网格 |
| `src/components/input/ChatInput.tsx` | 复合输入框（自适应 textarea + 附件 + 模型选择） |
| `src/components/chat/index.ts` | chat 组件导出 |
| `src/components/input/index.ts` | input 组件导出 |
| `docs/AI_Assistant_Plan.md` | 功能规划文档 |

**修改文件**：
| 文件 | 说明 |
|------|------|
| `package.json` | 新增 react-markdown + remark-gfm |
| `bun.lockb` | 依赖 lockfile 更新 |
| `src/components/layout/Shell.tsx` | 纵向 flex 加入 TopBar |
| `src/components/layout/CenterCanvas.tsx` | 重构为 ChatHeader + MessageTimeline + ChatInput |
| `src/components/layout/LeftRail.tsx` | 5 分组层级导航 + 会话搜索 |
| `src/components/layout/index.ts` | 新增 TopBar 导出 |
| `src/store/appStore.ts` | 合入 chatSlice + useChatStore hook |
| `src/store/slices/layoutSlice.ts` | 新增 navGroups + collapsedNavGroups + conversationSearchQuery |
| `src/components/ui/chart.tsx` | 修复 recharts v3 类型错误 |
| `src/components/ui/resizable.tsx` | 修复 react-resizable-panels v4 导出 |
| `src/components/ui/scroll-area.tsx` | 移除未使用 React import |

---

## 一、规划概览

### 1.1 目标
基于当前前端项目结构，逐步构建 SingerOS 的 **AI 助手（聊天界面）**。当前阶段专注于：
- 聊天交互核心功能（流式对话、工具调用展示）
- 纯前端 Mock 数据驱动
- 保持现有技术栈不变

### 1.2 技术栈确认
| 层级 | 技术 |
|------|------|
| 构建工具 | Vite + SWC |
| UI 框架 | React 19 |
| 样式方案 | Tailwind CSS 4 |
| 组件库 | @base-ui/react (Radix 无样式底层) |
| 图标 | @tabler/icons-react |
| 状态管理 | Zustand 5 (devtools + subscribeWithSelector) |
| 通信 | SSE / WebSocket (已有基础设施) |

### 1.3 现状保留与扩展
当前已实现：
- `Shell` 三栏布局（LeftRail / CenterCanvas / RightRail）
- `layoutSlice`：工作区、会话列表、右侧面板 Tab
- `topicSlice`：Topic CRUD（乐观更新模式）
- 通信基础设施：`http`, `SSEClient`, `WSClient`

**扩展方向**：
- 新增 `chatSlice` 接管聊天核心状态
- 重构 `CenterCanvas` 为消息时间轴 + 输入框组合
- 新增 `TopBar` 组件承载全局状态与导航
- 新增 `ChatHeader` 承载会话级操作

---

## 二、视觉设计 Token（从截图提取）

### 2.1 色彩系统
```
Background:    slate-50  (#f8fafc) — 全局画布背景
Surface:       white     (#ffffff) — 卡片、面板、消息气泡
Border:        slate-200 (#e2e8f0) — 分割线、边框
Text Primary:  slate-900 (#0f172a) — 标题、正文
Text Secondary:slate-500 (#64748b) — 次要文本、标签
Text Muted:    slate-400 (#94a3b8) — 占位符、时间戳
Accent:        blue-500  (#3b82f6) — 主按钮、选中态、链接
Accent Light:  blue-50   (#eff6ff) — 选中背景
User Message:  blue-600  (#2563eb) — 用户消息背景
Success:       green-500 (#22c55e) — AI 在线状态、成功态
```

### 2.2 字体层级
```
UI 控件:   无衬线, text-xs/sm, font-medium, tracking-wide, uppercase
叙事文本:  无衬线, text-sm, font-serif (AI 回复正文)
标签:      text-xs, uppercase, tracking-wider, text-slate-500
```

### 2.3 尺寸规范
```
TopBar 高度:      48px
LeftRail 宽度:    260px (可折叠)
RightRail 宽度:   280px (可折叠)
CenterCanvas 最大内容宽度:  720px (消息区域居中)
ChatInput 最大宽度: 800px
消息气泡圆角:     12px (用户) / 8px (AI)
```

---

## 三、页面架构（AI 助手单页）

```
App
└── Shell (h-screen, flex, overflow-hidden)
    ├── TopBar (48px, flex, border-b)                    ← 新增
    │   ├── Logo "SingerOS" + 版本号
    │   ├── 全局搜索框
    │   └── 右侧：AI 状态指示器 / 通知 / 用户头像菜单
    ├── MainArea (flex-1, flex, overflow-hidden)
    │   ├── LeftRail (260px, 可折叠)
    │   │   ├── 导航分组（AI 助手、工作台、AI 能力...）  ← 重构
    │   │   └── 会话列表（搜索 + 新建 + 历史）
    │   ├── CenterCanvas (flex-1, flex-col)
    │   │   ├── ChatHeader (会话标题栏)                   ← 新增
    │   │   │   ├── AI 头像 + 会话标题 + 下拉
    │   │   │   └── 右侧：Tokens / 搜索 / 设置 / 分享
    │   │   ├── MessageTimeline (flex-1, overflow-y-auto) ← 核心
    │   │   │   ├── WelcomeScreen (无消息时)
    │   │   │   ├── DateDivider (日期分割线)
    │   │   │   ├── UserMessage (右对齐, 蓝色气泡)
    │   │   │   ├── AIMessage (左对齐, 白色, 头像)
    │   │   │   │   ├── ThinkingBlock (思维链, 可折叠)
    │   │   │   │   ├── ToolCallBlock (工具调用, 可展开)
    │   │   │   │   └── ContentBlock (Markdown 渲染)
    │   │   │   └── TypingIndicator (流式输入指示器)
    │   │   └── ChatInput (底部输入区)                    ← 核心
    │   │       ├── 附件预览区
    │   │       ├── textarea (自动高度, 快捷键)
    │   │       └── 底部工具栏 (附件/@/表情/模型/发送)
    │   └── RightRail (280px, 可折叠, Tab切换)
    │       ├── 快捷操作 (上下文建议)
    │       ├── 文件收件箱
    │       └── 工件预览
    └── (可选) CommandPalette / 提及面板 (浮层)
```

---

## 四、功能模块详细规划

### 4.1 TopBar — 全局状态栏（新增）

**功能点**：
- **品牌区**：SingerOS Logo + 版本号 `v0.x.x`
- **全局搜索**：支持 Cmd+K 唤起，搜索会话/消息/知识库
- **AI 状态指示器**：绿色脉冲点 + "AI 在线" 文本，hover 显示模型信息
- **通知中心**：铃铛图标 + 未读红点，下拉面板展示通知列表
- **用户菜单**：头像 + 用户名，下拉包含：个人设置、主题切换、退出

**状态需求**：
- `aiStatus: 'online' | 'busy' | 'offline'`
- `unreadNotifications: number`
- `currentUser: { name, avatar, role }`

---

### 4.2 LeftRail — 侧边导航重构

**当前**：仅工作区 + 会话列表。

**目标**：截图式层级导航（分组折叠 + 子菜单）。

**导航数据结构**：
```typescript
type NavItem = {
  id: string;
  label: string;
  icon: string;        // Tabler icon name
  type: 'route' | 'group' | 'submenu';
  href?: string;
  children?: NavItem[];
  badge?: number;      // 未读数角标
  active?: boolean;
};
```

**一级导航分组**：
1. **核心功能**（无分组标题）
   - AI 助手（当前激活，带会话列表）
   - 工作台
2. **AI 能力**
   - AI 员工
   - 知识库
   - 技能管理
3. **研发协作**
   - InsFlow（可展开：流水线、部署、监控）
   - InsGit（可展开：代码仓库、PR、Issues）
   - Jenkins
   - InsSketch
4. **团队效率**
   - 组织管理
   - 汇报中心
   - 计划任务
5. **系统**
   - 个人设置
   - 权限管理

**会话列表面板**（位于 AI 助手下方或独立浮层）：
- 搜索框：实时过滤会话标题
- 新建会话按钮（`+`）
- 历史会话列表：标题、最后消息预览、时间、hover 显示删除/置顶
- 支持拖拽排序（未来）

---

### 4.3 ChatHeader — 会话标题栏（新增）

**功能点**：
- **左侧**：AI 头像（默认或根据助手类型动态）+ 当前会话标题（可点击下拉切换历史会话）
- **中间**：`+` 按钮快速新建会话
- **右侧工具组**：
  - `239.5K` Token 消耗计数器（hover 显示详细用量）
  - 搜索当前会话消息
  - 会话设置（模型选择、温度、最大长度等）
  - 分享会话（生成链接/导出）
  - 更多菜单（重命名、归档、删除）

**状态需求**：
- `currentConversation: Conversation | null`
- `tokenUsage: { total, currentSession }`
- `modelConfig: { model, temperature, maxTokens }`

---

### 4.4 MessageTimeline — 消息时间轴（核心重构）

#### 4.4.1 消息数据模型
```typescript
type MessageRole = 'user' | 'assistant' | 'system' | 'tool';

type MessageStatus = 'sending' | 'streaming' | 'complete' | 'error';

type ToolCall = {
  id: string;
  name: string;           // e.g., "vortflow_assign"
  arguments: Record<string, unknown>;
  status: 'pending' | 'running' | 'success' | 'error';
  result?: unknown;
  duration?: number;      // ms
};

type Message = {
  id: string;
  conversationId: string;
  role: MessageRole;
  content: string;        // 最终完整内容
  chunks: string[];       // 流式片段（仅 assistant）
  status: MessageStatus;
  timestamp: number;
  toolCalls?: ToolCall[];
  thinking?: string;      // 思维链 / reasoning
  metadata?: {
    model?: string;
    tokens?: number;
    latency?: number;
  };
};
```

#### 4.4.2 消息类型渲染策略

| 类型 | 位置 | 样式 | 组件 |
|------|------|------|------|
| UserMessage | 右对齐 | 蓝色渐变背景 `bg-gradient-to-br from-blue-500 to-blue-600`, 白色文字, 圆角 12px | `UserMessageBubble` |
| AIMessage | 左对齐 | 白色背景 `bg-white`, 灰色边框, 深色文字, 左侧头像 | `AIMessageBubble` |
| DateDivider | 居中 | 小字灰色胶囊 | `DateDivider` |
| TypingIndicator | 左对齐 | 三个跳动圆点 | `TypingIndicator` |

#### 4.4.3 工具调用展示（截图重点）

**结构**：
```
▼ 工具调用 (2)
  ├─ ▶ vortflow_assign ×2
  │   └─ [展开后显示参数与结果]
  └─ [其他工具...]
```

**交互**：
- 默认折叠，显示工具名称 + 调用次数
- 点击展开显示每个调用的：参数 JSON、执行状态（spinner → checkmark）、返回结果
- 执行中的工具显示脉冲动画
- 失败工具显示红色错误信息

**组件**：`ToolCallBlock`, `ToolCallItem`

#### 4.4.4 思维链展示（ThinkingBlock）

**结构**：
```
▼ 思考过程
  └─ [灰色斜体文本，展示 AI 推理步骤]
```

**交互**：可折叠，默认折叠以节省空间。

#### 4.4.5 Markdown 内容渲染

- 支持标准 Markdown：标题、列表、代码块、表格、引用
- 代码块支持语法高亮（可集成 `react-markdown` + `shiki` 或 Prism）
- 内联代码：`code` 标签 + 浅色背景
- 链接：蓝色下划线，hover 变色

---

### 4.5 ChatInput — 底部输入区（核心重构）

**截图占位文案**："请描述您问题，支持 Ctrl+V 粘贴图片。输入 @ 提及成员，/ 使用命令，# 引用工作项。"

#### 4.5.1 输入框功能
- **自动高度**：根据内容行数自动扩展（max 8 行），支持 Shift+Enter 换行
- **粘贴图片**：监听 paste 事件，提取图片文件并上传/预览
- **@ 提及**：输入 `@` 弹出成员选择面板（当前 Mock 为固定列表）
- **`/` 命令**：输入 `/` 弹出命令面板（如 `/clear`, `/model`, `/help`）
- **`#` 引用**：输入 `#` 弹出工作项/知识库引用面板
- **快捷键**：
  - `Enter`：发送消息
  - `Shift + Enter`：换行
  - `Escape`：取消输入/关闭面板
  - `↑`（空输入时）：编辑上一条消息

#### 4.5.2 附件预览区
- 位于 textarea 上方
- 支持图片预览（缩略图 + 删除按钮）
- 支持文件列表（图标 + 文件名 + 大小 + 删除）

#### 4.5.3 底部工具栏
- **左区**：
  - 📎 附件按钮（点击上传文件）
  - 🖼️ 图片按钮
  - 😊 表情选择器（可选，首期可简化）
- **右区**：
  - ⚙️ 模型选择下拉（GPT-4 / Claude-3 / DeepSeek）
  - ➤ 发送按钮（蓝色主按钮，空内容时 disabled）
  - 生成中状态：停止按钮（红色边框）

#### 4.5.4 数据模型
```typescript
type ChatInputState = {
  text: string;
  attachments: Attachment[];
  isFocused: boolean;
  isGenerating: boolean;
  selectedModel: string;
  mentionPanelOpen: boolean;
  commandPanelOpen: boolean;
  referencePanelOpen: boolean;
};

type Attachment = {
  id: string;
  type: 'image' | 'file';
  name: string;
  size: number;
  url?: string;      // 本地 blob URL 或上传后 URL
  file?: File;       // 原始文件对象
};
```

---

### 4.6 右侧快捷面板（RightRail 扩展）

当前已实现 Tab 切换（快捷 / 收件箱 / 工件）。

**扩展内容**：

#### Tab 1: 快捷操作
- 根据当前会话上下文动态生成建议按钮
- 示例："总结当前文档", "生成代码", "解释这段逻辑"
- 点击后自动填充到输入框或直接发送

#### Tab 2: 文件收件箱
- 拖放上传区域
- 当前会话关联的文件列表
- 支持点击插入到消息中

#### Tab 3: 工件预览
- AI 生成的 Markdown 文件、图片、代码文件
- 支持预览模式切换（原始 / 渲染）

---

## 五、状态管理规划（Zustand）

### 5.1 Store 结构
```typescript
type AppStore = LayoutStore & TopicStore & ChatStore;
```

**新增 `chatSlice`**：

```typescript
// State
interface ChatState {
  // 消息
  messagesMap: Record<string, Message>;
  messageIds: string[];           // 当前会话的消息 ID 有序列表
  streamingMessageId: string | null;
  
  // 输入
  inputText: string;
  inputAttachments: Attachment[];
  inputFocused: boolean;
  isGenerating: boolean;
  selectedModel: string;
  
  // 会话
  currentConversationId: string | null;
  conversations: Conversation[];
  
  // 工具/面板
  activeToolCalls: ToolCall[];
  mentionQuery: string | null;
  commandQuery: string | null;
}

// Actions
interface ChatAction {
  // 消息流
  sendMessage: (content: string, attachments?: Attachment[]) => Promise<void>;
  appendChunk: (messageId: string, chunk: string) => void;
  finalizeMessage: (messageId: string) => void;
  cancelGeneration: () => void;
  
  // 输入
  setInputText: (text: string) => void;
  addAttachment: (file: File) => void;
  removeAttachment: (id: string) => void;
  setInputFocused: (focused: boolean) => void;
  
  // 会话
  createConversation: (title?: string) => string;
  switchConversation: (id: string) => void;
  renameConversation: (id: string, title: string) => void;
  deleteConversation: (id: string) => void;
  
  // 工具调用
  registerToolCall: (toolCall: ToolCall) => void;
  updateToolCallStatus: (id: string, status: ToolCall['status'], result?: unknown) => void;
}
```

### 5.2 流式消息处理流程
```
用户点击发送
  → chatSlice.sendMessage(content)
    → 1. 乐观更新：立即添加 UserMessage 到 messagesMap
    → 2. 创建空的 AssistantMessage，设置 status='streaming'
    → 3. 调用 mockStream() / 真实 SSE
    → 4. 收到 chunk → appendChunk(messageId, chunk)
    → 5. 收到 tool_call → registerToolCall()
    → 6. 流结束 → finalizeMessage(messageId) / 或 error → 标记 status='error'
```

### 5.3 实现模式（遵循 Zustand Skill）
- 使用 **Class-based Action Implementation**
- Public Actions：`sendMessage`, `createConversation`
- Internal Actions：`internal_sendMessage`, `internal_streamMessage`
- Dispatch Methods：`#dispatchChat` → `chatReducer`
- Optimistic Update：消息发送后立即渲染，失败时回滚

---

## 六、Mock 数据方案

### 6.1 Mock Service 目录
```
src/mocks/
├── chatMocks.ts       # 消息流、工具调用 Mock
├── conversationMocks.ts # 会话列表 Mock
├── userMocks.ts       # 用户、通知 Mock
└── streamSimulator.ts # 流式数据生成器
```

### 6.2 流式数据模拟器
```typescript
// 模拟 SSE 流，按词/句延迟推送
function mockStreamResponse(
  content: string,
  onChunk: (chunk: string) => void,
  onToolCall?: (tool: ToolCall) => void,
  onComplete?: () => void,
): { cancel: () => void };
```

**示例流**：
1. 延迟 500ms 开始
2. 逐字输出文本内容（每字 20ms）
3. 中途插入 ToolCall 事件
4. ToolCall 延迟 800ms 后返回结果
5. 继续输出剩余文本
6. 输出完成标记

### 6.3 预置场景
- **代码审查**：用户请求审查 PR → AI 分析 → 调用 `github_review` 工具 → 返回审查报告
- **需求指派**：用户请求指派需求 → AI 调用 `vortflow_assign` → 返回指派结果（如截图）
- **知识库问答**：引用 `#知识库条目` → AI 检索 → 返回答案

---

## 七、组件文件组织

```
src/
├── components/
│   ├── layout/
│   │   ├── Shell.tsx              # 布局容器
│   │   ├── TopBar.tsx             # 新增：全局状态栏
│   │   ├── LeftRail.tsx           # 重构：导航 + 会话列表
│   │   ├── CenterCanvas.tsx       # 重构：消息区容器
│   │   └── RightRail.tsx          # 保留：快捷面板
│   ├── chat/
│   │   ├── ChatHeader.tsx         # 新增：会话标题栏
│   │   ├── MessageTimeline.tsx    # 新增：消息列表容器
│   │   ├── UserMessageBubble.tsx  # 新增：用户消息
│   │   ├── AIMessageBubble.tsx    # 新增：AI 消息
│   │   ├── ToolCallBlock.tsx      # 新增：工具调用块
│   │   ├── ThinkingBlock.tsx      # 新增：思维链块
│   │   ├── DateDivider.tsx        # 新增：日期分割
│   │   └── TypingIndicator.tsx    # 新增：输入指示
│   ├── input/
│   │   ├── ChatInput.tsx          # 新增：输入区容器
│   │   ├── AutoResizeTextarea.tsx # 新增：自适应文本域
│   │   ├── AttachmentPreview.tsx  # 新增：附件预览
│   │   ├── MentionPanel.tsx       # 新增：@提及面板
│   │   └── CommandPanel.tsx       # 新增：/命令面板
│   └── ui/                        # 已有：基础 UI 组件库
├── store/
│   ├── appStore.ts                # 合并所有 slices
│   ├── types.ts                   # 已有
│   ├── slices/
│   │   ├── layoutSlice.ts         # 重构扩展
│   │   ├── topicSlice.ts          # 保留
│   │   └── chatSlice.ts           # 新增：聊天核心状态
│   └── utils/
│       └── flattenActions.ts      # 已有
├── hooks/
│   ├── useChat.ts                 # 新增：聊天核心逻辑 Hook
│   ├── useStream.ts               # 新增：流式消息 Hook
│   ├── useMention.ts              # 新增：@提及逻辑
│   ├── useCommand.ts              # 新增：/命令逻辑
│   ├── useWebSocket.ts            # 已有
│   └── useSSE.ts                  # 已有
├── lib/
│   ├── request.ts                 # 已有
│   ├── sse.ts                     # 已有
│   ├── websocket.ts               # 已有
│   └── markdown.ts                # 新增：Markdown 渲染配置
├── mocks/
│   ├── streamSimulator.ts         # 新增
│   ├── chatMocks.ts               # 新增
│   └── conversationMocks.ts       # 新增
├── types/
│   ├── chat.ts                    # 新增：聊天领域类型
│   └── api.ts                     # 已有
└── utils/
    └── format.ts                  # 新增：时间/大小格式化
```

---

## 八、动画与交互规范

### 8.1 允许的动画（遵循 Orbita 架构）
- `transition: opacity, color, background-color, border-color, box-shadow`
- `transform: scale(0.98 → 1)` 按钮按下反馈
- 工具调用执行中：脉冲动画（`animate-pulse`）仅限图标
- 流式文本：无动画，直接追加内容

### 8.2 禁止的动画
- 消息气泡飞入/弹入（干扰阅读）
- 背景粒子/装饰动画
- 页面切换过渡动画（当前单页无需）

### 8.3 流式渲染性能
- 消息内容使用 `dangerouslySetInnerHTML` 或 `react-markdown` 渲染
- 流式更新时仅更新文本节点，避免整组件重渲染
- 长消息列表使用虚拟滚动（如超过 100 条消息，可选）

---

## 九、实施路线图

### Phase 1: 基础骨架 ✅ 已完成
- [x] 编写 `chatSlice`（状态定义 + Mock 数据 + 流式生成器）
- [x] 实现 `TopBar` 组件（品牌区 + AI 状态 + 用户菜单，已移除搜索框）
- [x] 重构 `LeftRail`（5 分组层级导航 + 会话搜索列表）
- [x] 实现 `ChatHeader`（会话标题 + Token 计数器 + 设置按钮）
- [x] 实现 `MessageTimeline` 消息列表 + WelcomeScreen 空状态
- [x] 实现 `UserMessageBubble`（蓝色渐变气泡）
- [x] 实现 `AIMessageBubble`（Markdow 渲染 + 流式光标）
- [x] 实现 `ToolCallBlock`（折叠/展开 + 状态图标）
- [x] 实现 `ChatInput`（自适应 textarea + 附件粘贴 + 模型选择）
- [x] 重构 Shell / CenterCanvas / LeftRail 布局
- [x] 修复已有 UI 组件 TS 错误（chart, resizable, scroll-area, sidebar）
- [x] 安装 react-markdown + remark-gfm

**交付**：可展示完整静态页面 + Mock 流式对话 + 工具调用展示。

### Phase 2: 流式对话体验优化 ✅ 已完成
- [x] 消息自动滚动到底部（新消息/流式追加时）
- [x] 重新生成按钮（消息尾部 hover 操作区）
- [x] Token 计数器随发送动态增加
- [x] 消息操作菜单（复制按钮）
- [x] ChatInput/MessageTimeline flex 布局修复（min-h-0 + flex-1）
- [x] TopBar 移除搜索框
- [x] ChatHeader 动态标题/模型/token
- [-] 模型切换影响 Mock 内容（实际无后端差异，暂不实现）

**交付**：流式对话体验完整可交互。

### Phase 3: 高级交互
- [ ] 实现 `ThinkingBlock`（思维链展示，可折叠）
- [ ] `@` 提及面板（成员列表弹窗）
- [ ] `/` 命令面板（命令列表弹窗）
- [ ] `#` 引用面板（工作项引用弹窗）
- [ ] 会话重命名 / 归档 / 删除下拉菜单
- [ ] 右侧面板快捷操作点击填充到输入框

**交付**：接近截图完整交互体验。

### Phase 4: Polish
- [ ] 键盘快捷键（Esc 关闭面板、↑编辑上一条）
- [ ] 错误状态处理（网络错误、AI 服务异常）
- [ ] 响应式适配（移动端折叠侧边栏）
- [ ] 消息搜索功能

---

## 十、关键技术决策

| 决策点 | 选择 | 理由 |
|--------|------|------|
| Markdown 渲染 | `react-markdown` + `remark-gfm` | 生态成熟，支持插件 |
| 代码高亮 | `shiki` (按需) / 首期 Prism | Shiki 更精美但体积大 |
| 流式模拟 | `setTimeout` 逐字输出 | 简单可控，无需后端 |
| 输入框自动高度 | 原生 `textarea` + scrollHeight | 无需第三方库 |
| 虚拟滚动 | 首期不实现 | 消息量预计 < 100 条 |
| 文件上传预览 | `URL.createObjectURL` | 纯前端预览 |

---

## 十一、与现有代码的兼容性

- **保持 `layoutSlice`**：扩展导航结构，不破坏现有工作区/会话逻辑
- **保持 `topicSlice`**：可独立存在，未来可能与聊天会话合并
- **保持 UI 组件库**：所有新增组件使用 `@base-ui/react` 和 Tailwind
- **保持通信层**：`http` / `SSEClient` / `WSClient` 已就绪，Mock 阶段使用 `streamSimulator` 替代

---

**总结**：本规划以截图中 AI 助手聊天界面为蓝图，分 4 个阶段逐步实施。核心技术栈与现有项目完全兼容，优先使用 Mock 数据实现可交互原型，为后续后端对接预留清晰的接口边界。
