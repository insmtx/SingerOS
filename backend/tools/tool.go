// tools 包定义 SingerOS 的最小 Tool 抽象。
//
// 当前阶段只提供基础接口，便于先把“授权账户使用能力”跑通，
// 后续再在其上补充统一的 Tool Runtime、Policy 与 Approval。
package tools

import (
	"context"

	auth "github.com/insmtx/SingerOS/backend/auth"
)

const (
	// ResourceGitHubResolvedClient stores the provider-specific resolved client inside execution context.
	ResourceGitHubResolvedClient = "github.resolved_client"
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

// ToolInfo 描述一个 Tool 的基本信息。
type ToolInfo struct {
	Name        string
	Description string
	Provider    string
	ReadOnly    bool
	InputSchema *Schema
}

// Tool 是 SingerOS 的最小工具接口。
type Tool interface {
	Info() *ToolInfo
	Validate(input map[string]interface{}) error
	Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)
}

// ExecutionContext carries runtime-resolved identity and provider resources for a tool call.
type ExecutionContext struct {
	UserID          string
	AccountID       string
	Provider        string
	Selector        *auth.AuthSelector
	ResolvedAccount *auth.AuthorizedAccount
	ResolvedBy      string
	Resources       map[string]interface{}
}

// RuntimeTool is implemented by tools that can consume runtime-injected resources.
type RuntimeTool interface {
	ExecuteWithContext(ctx context.Context, execCtx *ExecutionContext, input map[string]interface{}) (map[string]interface{}, error)
}
