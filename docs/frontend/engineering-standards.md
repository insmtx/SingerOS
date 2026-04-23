# 工程规范

本文档详细描述 SingerOS 前端的工程规范。

## NPM Scripts

| 命令 | 作用 |
|------|------|
| `bun run dev` | 启动 Vite 开发服务器 (SWC HMR) |
| `bun run build` | TypeScript 检查 + Vite 生产构建 |
| `bun run check` | Biome check (lint + format) |
| `bun run preview` | 预览生产构建 |
| `bun run test` | Vitest 运行测试 |
| `bun run coverage` | Vitest 覆盖率报告 |

## 路径别名

```ts
// tsconfig.json + vite.config.ts
"@/*" → "src/*"
```

全项目统一使用 `@/` 导入路径。

## TypeScript 配置

- `strict: true` — 严格模式
- `noUnusedLocals/noUnusedParameters` — 禁止未使用变量
- `jsx: react-jsx` — 新 JSX 转换 (无需显式 import React)
- `moduleResolution: bundler` — Vite 模式解析

## Biome 规则

- 格式化：空格缩进，单引号
- 检查：recommended + noUnusedVariables + useHookAtTopLevel + noEmptyBlockStatements
- CSS：tailwindDirectives 启用
- VCS：git 集成，默认 main 分支

## 样式体系

### TailwindCSS 4 配置

```css
@import "tailwindcss";

@theme {
  --font-family-sans: "Inter", system-ui, "Avenir", "Helvetica", "Arial", sans-serif;
}
```

使用 TailwindCSS v4 的 `@theme` 指令定义自定义设计令牌。

### 设计系统

| 元素 | Token | 用途 |
|------|-------|------|
| 背景 | `slate-50` | 中性灰页面背景 |
| 表面 | `white` | 内容卡片/面板 |
| 分隔 | `slate-200` | 低对比度分隔线 |
| 主操作 | `blue-500/600` | 发送按钮/选中态 |
| 文字主 | `slate-700` | 正文文字 |
| 文字次 | `slate-500` | 标签/辅助文字 |
| 文字弱 | `slate-400` | 占位符/时间 |

### 视觉规则

- UI 控件：无衬线字体，紧凑布局
- 叙述文本 (AI 回复)：衬线字体 (`font-serif`)，优先阅读体验
- 标签/分类：大写字母 + `tracking-wide` 字间距
- 动效：仅 `transition-colors/opacity`，无干扰性动画