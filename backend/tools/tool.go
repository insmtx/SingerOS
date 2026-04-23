// tools 包定义 SingerOS 的最小 Tool 抽象。
//
// 当前阶段只提供基础接口和 agent 运行时注入的上下文信息。
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Schema describes tool input or output in a provider-agnostic shape.
type Schema struct {
	Type       string               `json:"type"`
	Required   []string             `json:"required,omitempty"`
	Properties map[string]*Property `json:"properties,omitempty"`
}

// Property describes a single field inside a tool schema.
type Property struct {
	Type        string    `json:"type"`
	Description string    `json:"description,omitempty"`
	Enum        []string  `json:"enum,omitempty"`
	Items       *Property `json:"items,omitempty"`
}

// Tool 是 SingerOS 的最小工具接口。
type Tool interface {
	Name() string
	Description() string
	InputSchema() Schema
	Execute(ctx context.Context, input map[string]interface{}) (string, error)
}

// Validator is implemented by tools that perform local input validation before execution.
type Validator interface {
	Validate(input map[string]interface{}) error
}

// BaseTool stores the LLM-facing metadata shared by concrete tools.
type BaseTool struct {
	name        string
	description string
	inputSchema Schema
}

// NewBaseTool creates a reusable metadata base for a concrete tool.
func NewBaseTool(name string, description string, inputSchema Schema) BaseTool {
	return BaseTool{
		name:        strings.TrimSpace(name),
		description: strings.TrimSpace(description),
		inputSchema: inputSchema,
	}
}

// Name returns the stable tool identifier.
func (t BaseTool) Name() string {
	return t.name
}

// Description returns the LLM-facing tool description.
func (t BaseTool) Description() string {
	return t.description
}

// InputSchema returns the tool argument schema.
func (t BaseTool) InputSchema() Schema {
	return t.inputSchema
}

// JSONString encodes structured tool output as the string payload returned to the model.
func JSONString(value interface{}) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("marshal tool output: %w", err)
	}
	return string(encoded), nil
}

// ToolContext carries run-scoped identity and conversation metadata injected by the agent runtime.
type ToolContext struct {
	RunID          string
	TraceID        string
	AssistantID    string
	UserID         string
	AccountID      string
	Channel        string
	ChatID         string
	ConversationID string
	ExternalID     string
	WorkNodeID     string
	Metadata       map[string]any
}

type toolContextKey struct{}

// ContextWithToolContext stores run-scoped tool context on a context.Context.
func ContextWithToolContext(ctx context.Context, toolCtx ToolContext) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, toolContextKey{}, cloneToolContext(toolCtx))
}

// ToolContextFrom returns run-scoped tool context stored on ctx.
func ToolContextFrom(ctx context.Context) (ToolContext, bool) {
	if ctx == nil {
		return ToolContext{}, false
	}
	toolCtx, ok := ctx.Value(toolContextKey{}).(ToolContext)
	if !ok {
		return ToolContext{}, false
	}
	return cloneToolContext(toolCtx), true
}

// RequireToolContext returns run-scoped tool context or an error when it is missing.
func RequireToolContext(ctx context.Context) (ToolContext, error) {
	toolCtx, ok := ToolContextFrom(ctx)
	if !ok {
		return ToolContext{}, fmt.Errorf("tool context is required")
	}
	return toolCtx, nil
}

func cloneToolContext(toolCtx ToolContext) ToolContext {
	cloned := toolCtx
	if toolCtx.Metadata != nil {
		cloned.Metadata = make(map[string]any, len(toolCtx.Metadata))
		for key, value := range toolCtx.Metadata {
			cloned.Metadata[key] = value
		}
	}
	return cloned
}
