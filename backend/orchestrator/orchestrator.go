// orchestrator 包提供 SingerOS 的事件编排器功能
//
// 编排器负责从事件总线订阅事件，并根据事件类型分发到相应的处理器进行处理。
// 是 SingerOS 事件驱动架构的核心协调组件。
package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/insmtx/SingerOS/backend/agent/react"
	"github.com/insmtx/SingerOS/backend/interaction"
	"github.com/insmtx/SingerOS/backend/interaction/eventbus"
	"github.com/insmtx/SingerOS/backend/llm"
	skills "github.com/insmtx/SingerOS/backend/skills"
	"github.com/ygpkg/yg-go/logs"
)

// EventHandlerFunc 是事件处理函数的类型定义
type EventHandlerFunc func(ctx context.Context, event *interaction.Event) error

// Orchestrator 是事件编排器，负责事件的订阅、分发和处理
type Orchestrator struct {
	subscriber        eventbus.Subscriber         // 事件订阅者
	skillManager      skills.SkillManager         // 技能管理器
	agentOrchestrator *react.AgentOrchestrator    // ReAct 代理编排器
	handlers          map[string]EventHandlerFunc // 事件主题到处理器的映射
}

// NewOrchestrator 创建一个新的事件编排器实例
func NewOrchestrator(subscriber eventbus.Subscriber, skillManager skills.SkillManager, llmProvider llm.Provider) *Orchestrator {
	orchestrator := &Orchestrator{
		subscriber:   subscriber,
		skillManager: skillManager,
		handlers:     make(map[string]EventHandlerFunc),
	}

	// 初始化ReAct Agent编排器
	orchestrator.agentOrchestrator = react.NewAgentOrchestrator(llmProvider, skillManager)

	// 注册默认处理器
	orchestrator.registerDefaultHandlers()

	return orchestrator
}

// registerDefaultHandlers 注册默认的事件处理器
func (o *Orchestrator) registerDefaultHandlers() {
	// 处理GitHub issue_comment事件
	o.handlers[interaction.TopicGithubIssueComment] = o.handleIssueComment

	// 处理GitHub pull_request事件
	o.handlers[interaction.TopicGithubPullRequest] = o.handlePullRequest
}

// Start 启动编排器，开始订阅和处理事件
func (o *Orchestrator) Start(ctx context.Context) error {
	for topic, handler := range o.handlers {
		go func(t string, h EventHandlerFunc) {
			logs.InfoContextf(ctx, "Starting subscription for topic: %s", t)
			err := o.subscriber.Subscribe(ctx, t, func(event any) {
				// 将通用interface{}转换为interaction.Event
				jsonBytes, err := json.Marshal(event)
				if err != nil {
					logs.ErrorContextf(ctx, "Failed to marshal event to JSON: %v", err)
					return
				}

				var interactionEvent interaction.Event
				if err := json.Unmarshal(jsonBytes, &interactionEvent); err != nil {
					logs.ErrorContextf(ctx, "Failed to unmarshal event: %v", err)
					return
				}

				logs.DebugContextf(ctx, "Received event on topic %s: %+v", t, interactionEvent)

				if err := h(ctx, &interactionEvent); err != nil {
					logs.ErrorContextf(ctx, "Error handling event on topic %s: %v", t, err)
				}
			})

			if err != nil {
				logs.ErrorContextf(ctx, "Failed to subscribe to topic %s: %v", t, err)
			}
		}(topic, handler)
	}

	return nil
}

// handleIssueComment 处理 GitHub Issue 评论事件
func (o *Orchestrator) handleIssueComment(ctx context.Context, event *interaction.Event) error {
	logs.InfoContextf(ctx, "Processing GitHub issue comment event with ReAct agent: %+v", event)

	// 使用 ReAct Agent 处理事件
	return o.agentOrchestrator.HandleEventAdvanced(ctx, event)
}

// handlePullRequest 处理 GitHub Pull Request 事件
func (o *Orchestrator) handlePullRequest(ctx context.Context, event *interaction.Event) error {
	logs.InfoContextf(ctx, "Processing GitHub pull request event with ReAct agent: %+v", event)

	// 使用 ReAct Agent 处理 PR 事件
	return o.agentOrchestrator.HandleEventAdvanced(ctx, event)
}

// RegisterHandler 允许外部注册新的事件处理器
func (o *Orchestrator) RegisterHandler(topic string, handler EventHandlerFunc) {
	o.handlers[topic] = handler
}

// GetHandler 获取已注册的事件处理器
func (o *Orchestrator) GetHandler(topic string) (EventHandlerFunc, error) {
	handler, exists := o.handlers[topic]
	if !exists {
		return nil, fmt.Errorf("no handler registered for topic: %s", topic)
	}
	return handler, nil
}
