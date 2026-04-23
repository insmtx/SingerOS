// orchestrator 包提供 SingerOS 的事件编排器功能
//
// 编排器负责从事件总线订阅事件，并根据事件类型分发到相应的处理器进行处理。
// 是 SingerOS 事件驱动架构的核心协调组件。
package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/insmtx/SingerOS/backend/interaction"
	"github.com/insmtx/SingerOS/backend/interaction/eventbus"
	agentruntime "github.com/insmtx/SingerOS/backend/runtime"
	"github.com/ygpkg/yg-go/logs"
)

// EventHandlerFunc 是事件处理函数的类型定义
type EventHandlerFunc func(ctx context.Context, event *interaction.Event) error

// Orchestrator 是事件编排器，负责事件的订阅、分发和处理
type Orchestrator struct {
	subscriber eventbus.Subscriber         // 事件订阅者
	runner     agentruntime.Runner         // 统一任务运行器
	handlers   map[string]EventHandlerFunc // 事件主题到处理器的映射
}

// NewOrchestrator 创建一个新的事件编排器实例
func NewOrchestrator(subscriber eventbus.Subscriber, runner agentruntime.Runner) *Orchestrator {
	orchestrator := &Orchestrator{
		subscriber: subscriber,
		runner:     runner,
		handlers:   make(map[string]EventHandlerFunc),
	}

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

	// 处理GitHub push事件
	o.handlers[interaction.TopicGithubPush] = o.handlePush
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
	logs.InfoContextf(ctx, "Processing GitHub issue comment event with agent runtime: %+v", event)

	return o.runEvent(ctx, event)
}

// handlePullRequest 处理 GitHub Pull Request 事件
func (o *Orchestrator) handlePullRequest(ctx context.Context, event *interaction.Event) error {
	logs.InfoContextf(ctx, "Processing GitHub pull request event with agent runtime: %+v", event)

	return o.runEvent(ctx, event)
}

// handlePush 处理 GitHub Push 提交事件
func (o *Orchestrator) handlePush(ctx context.Context, event *interaction.Event) error {
	logs.InfoContextf(ctx, "Processing GitHub push event with agent runtime: %+v", event)

	return o.runEvent(ctx, event)
}

func (o *Orchestrator) runEvent(ctx context.Context, event *interaction.Event) error {
	if o.runner == nil {
		return fmt.Errorf("agent runtime runner is required")
	}

	result, err := o.runner.Run(ctx, requestFromInteractionEvent(event))
	if err != nil {
		return err
	}
	if result != nil {
		logs.InfoContextf(ctx, "Agent runtime completed event: run_id=%s status=%s", result.RunID, result.Status)
	}
	return nil
}

func requestFromInteractionEvent(event *interaction.Event) *agentruntime.RequestContext {
	if event == nil {
		return &agentruntime.RequestContext{
			Input: agentruntime.InputContext{
				Type: agentruntime.InputTypeEvent,
			},
		}
	}

	return &agentruntime.RequestContext{
		RunID:   event.EventID,
		TraceID: event.TraceID,
		Actor: agentruntime.ActorContext{
			UserID:     event.Actor,
			Channel:    event.Channel,
			ExternalID: event.Actor,
		},
		Input: agentruntime.InputContext{
			Type: agentruntime.InputTypeEvent,
			Text: buildInteractionEventInput(event),
		},
		Metadata: map[string]any{
			"channel":       event.Channel,
			"event_type":    event.EventType,
			"subject":       event.Repository,
			"event_context": mapFromAny(event.Context),
			"event_payload": mapFromAny(event.Payload),
		},
	}
}

func buildInteractionEventInput(event *interaction.Event) string {
	if event == nil {
		return ""
	}

	contextMap := mapFromAny(event.Context)
	payloadMap := mapFromAny(event.Payload)
	sections := []string{
		"You are handling an external event inside SingerOS.",
		buildInteractionEventEnvelope(event),
		buildInteractionEventTask(event.EventType),
	}
	if contextSection := buildJSONSection("Event context", contextMap); contextSection != "" {
		sections = append(sections, contextSection)
	}
	if payloadSection := buildJSONSection("Raw event payload", payloadMap); payloadSection != "" {
		sections = append(sections, payloadSection)
	}

	return strings.Join(filterEmptyStrings(sections), "\n\n")
}

func buildInteractionEventEnvelope(event *interaction.Event) string {
	lines := []string{"Event envelope:"}
	if event.Channel != "" {
		lines = append(lines, "- channel: "+event.Channel)
	}
	if event.EventType != "" {
		lines = append(lines, "- event_type: "+event.EventType)
	}
	if event.Actor != "" {
		lines = append(lines, "- actor: "+event.Actor)
	}
	if event.Repository != "" {
		lines = append(lines, "- subject: "+event.Repository)
	}
	if event.EventID != "" {
		lines = append(lines, "- run_id: "+event.EventID)
	}
	if event.TraceID != "" {
		lines = append(lines, "- trace_id: "+event.TraceID)
	}
	return strings.Join(lines, "\n")
}

func buildInteractionEventTask(eventType string) string {
	base := "Task:\n- Understand what happened from the event payload.\n- Use available skills and tools to gather authoritative details before making claims.\n- If the event requires an external response, decide whether to publish one and keep it evidence-based."

	switch eventType {
	case "pull_request", "github.pull_request", "github.pull_request.opened":
		return base + "\n- This appears to be a GitHub pull request event. Review the change carefully before publishing any GitHub review."
	case "push", "github.push":
		return base + "\n- This appears to be a GitHub push event. Use the commit list and repository context to understand what changed before deciding whether any follow-up is needed."
	case "issue_comment", "github.issue_comment":
		return base + "\n- This appears to be a GitHub issue or pull request comment event. Decide whether a reply is needed."
	default:
		return base
	}
}

func mapFromAny(value any) map[string]any {
	if value == nil {
		return nil
	}
	if typed, ok := value.(map[string]any); ok {
		return typed
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		return nil
	}

	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		return nil
	}
	return decoded
}

func buildJSONSection(title string, value any) string {
	if value == nil {
		return ""
	}

	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return ""
	}

	text := string(encoded)
	if len(text) > 6000 {
		text = text[:6000] + "\n... (truncated)"
	}

	return fmt.Sprintf("%s:\n```json\n%s\n```", title, text)
}

func filterEmptyStrings(values []string) []string {
	filtered := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		filtered = append(filtered, value)
	}
	return filtered
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
