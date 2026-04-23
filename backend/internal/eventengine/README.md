# Orchestrator

Orchestrator 是 SingerOS 的核心组件之一，负责消费由各种连接器（如 GitHub）发布的事件，并执行相应的业务逻辑。

## 功能特性

- 监听 RabbitMQ 中的各种主题（topics）
- 对 GitHub 事件的实时处理
- 支持多种事件类型的消费者模型
- 可扩展的处理器注册机制

## 当前支持的事件类型

- `interaction.github.issue_comment`: GitHub 问题评论事件
- `interaction.github.pull_request`: GitHub PR 事件

## 如何使用

Orchestrator 在 SingerOS 后台服务启动时同时启动，并自动订阅配置好的事件主题。

## 架构原理

1. 各个连接器（connectors）接收到的外部事件被标准化为 `interaction.Event` 格式
2. 这些事件通过 eventbus 发送到相应的话题（topic）
3. Orchestrator 订阅这些话题并处理事件
4. 后续可以在此基础上增加业务逻辑执行、技能调用等功能