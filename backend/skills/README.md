# Skills 系统

SingerOS 的 `skills` 目录现在只承载知识型 skills，不再保留可执行 skill 接口、skill manager 或 tool skill 示例。

## 当前目录结构

- `backend/skills/catalog/` - `SKILL.md` 的扫描、解析、读取与索引
- `backend/skills/bundled/` - 跟随服务发布的内置知识型 skills

## 如何添加新 skill

1. 在 `backend/skills/bundled/` 下创建新目录，例如 `github-pr-review/`
2. 在目录中创建 `SKILL.md`
3. 在 `SKILL.md` 顶部使用 YAML frontmatter 描述 `name`、`description`、`version`
4. 如有需要，增加 `references/`、`templates/`、`scripts/`、`assets/`

## 运行时使用方式

运行时分两步使用 skills：

1. 先生成 skill summary/index 注入模型上下文
2. 在需要时按名称读取 skill 正文

## 设计边界

- `skills` 负责知识、流程经验和约束说明
- `tools` 负责最小执行动作
- `toolruntime` 负责账户注入、审批、审计和执行

当前主链：

`Event -> Agent Runtime -> Skills Catalog -> Tool Registry -> Eino Agent -> Tool Runtime`
