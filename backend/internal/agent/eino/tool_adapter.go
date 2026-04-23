package eino

import (
	"context"
	"encoding/json"
	"fmt"

	einotool "github.com/cloudwego/eino/components/tool"
	einoschema "github.com/cloudwego/eino/schema"
	auth "github.com/insmtx/SingerOS/backend/auth"
	"github.com/insmtx/SingerOS/backend/toolruntime"
	"github.com/insmtx/SingerOS/backend/tools"
)

// ToolDefinition is the local bridge shape exported to an Eino integration layer.
//
// It intentionally mirrors only the fields we need from SingerOS tools so the
// actual cloudwego/eino binding can be added later without changing registry
// or runtime packages again.
type ToolDefinition struct {
	Name        string
	Description string
	Provider    string
	ReadOnly    bool
	InputSchema *tools.Schema
}

// ToolCallRequest describes one model-initiated tool call.
type ToolCallRequest struct {
	Selector  *auth.AuthSelector
	Name      string
	UserID    string
	AccountID string
	Arguments map[string]interface{}
}

// ToolCallResult contains the execution result returned back to the model loop.
type ToolCallResult struct {
	Name              string
	Output            map[string]interface{}
	ResolvedAccountID string
	ResolvedBy        string
}

// ToolAdapter bridges SingerOS tool registry/runtime to an Eino-facing API.
type ToolAdapter struct {
	registry *tools.Registry
	runtime  *toolruntime.Runtime
}

// ToolBinding carries runtime-bound identity for one Eino agent execution.
type ToolBinding struct {
	Selector  *auth.AuthSelector
	UserID    string
	AccountID string
}

// NewToolAdapter creates a new adapter over the shared tool registry and runtime.
func NewToolAdapter(registry *tools.Registry, runtime *toolruntime.Runtime) *ToolAdapter {
	return &ToolAdapter{
		registry: registry,
		runtime:  runtime,
	}
}

// Definitions returns the registry tools in an Eino-friendly description shape.
func (a *ToolAdapter) Definitions() []ToolDefinition {
	if a == nil || a.registry == nil {
		return nil
	}

	infos := a.registry.ListInfos()
	definitions := make([]ToolDefinition, 0, len(infos))
	for _, info := range infos {
		definitions = append(definitions, ToolDefinition{
			Name:        info.Name,
			Description: info.Description,
			Provider:    info.Provider,
			ReadOnly:    info.ReadOnly,
			InputSchema: info.InputSchema,
		})
	}

	return definitions
}

// EinoTools returns actual Eino tools bound to the current runtime identity.
func (a *ToolAdapter) EinoTools(binding ToolBinding) ([]einotool.BaseTool, error) {
	if a == nil || a.registry == nil {
		return nil, nil
	}

	infos := a.registry.ListInfos()
	result := make([]einotool.BaseTool, 0, len(infos))
	for _, info := range infos {
		toolInfo := info
		result = append(result, &invokableTool{
			adapter: a,
			info:    &toolInfo,
			binding: binding,
		})
	}

	return result, nil
}

// Invoke executes a tool call through SingerOS Tool Runtime.
func (a *ToolAdapter) Invoke(ctx context.Context, req *ToolCallRequest) (*ToolCallResult, error) {
	if req == nil {
		return nil, fmt.Errorf("tool call request is required")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("tool name is required")
	}
	if a == nil || a.runtime == nil {
		return nil, fmt.Errorf("tool runtime is required")
	}

	result, err := a.runtime.Execute(ctx, &toolruntime.ExecuteRequest{
		ToolName:  req.Name,
		Selector:  req.Selector,
		UserID:    req.UserID,
		AccountID: req.AccountID,
		Input:     req.Arguments,
	})
	if err != nil {
		return nil, err
	}

	callResult := &ToolCallResult{
		Name:       result.ToolName,
		Output:     result.Output,
		ResolvedBy: result.ResolvedBy,
	}
	if result.ResolvedAccount != nil {
		callResult.ResolvedAccountID = result.ResolvedAccount.ID
	}

	return callResult, nil
}

type invokableTool struct {
	adapter *ToolAdapter
	info    *tools.ToolInfo
	binding ToolBinding
}

func (t *invokableTool) Info(ctx context.Context) (*einoschema.ToolInfo, error) {
	if t == nil || t.info == nil {
		return nil, fmt.Errorf("tool info is required")
	}

	return toEinoToolInfo(t.info), nil
}

func (t *invokableTool) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...einotool.Option) (string, error) {
	if t == nil || t.adapter == nil {
		return "", fmt.Errorf("tool adapter is required")
	}

	input := make(map[string]interface{})
	if argumentsInJSON != "" {
		if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
			return "", fmt.Errorf("unmarshal tool arguments: %w", err)
		}
	}

	result, err := t.adapter.Invoke(ctx, &ToolCallRequest{
		Selector:  t.binding.Selector,
		Name:      t.info.Name,
		UserID:    t.binding.UserID,
		AccountID: t.binding.AccountID,
		Arguments: input,
	})
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result.Output)
	if err != nil {
		return "", fmt.Errorf("marshal tool output: %w", err)
	}

	return string(output), nil
}

func toEinoToolInfo(info *tools.ToolInfo) *einoschema.ToolInfo {
	if info == nil {
		return nil
	}

	params := make(map[string]*einoschema.ParameterInfo)
	if info.InputSchema != nil {
		for name, property := range info.InputSchema.Properties {
			params[name] = toEinoParameterInfo(property, info.InputSchema.Required, name)
		}
	}

	toolInfo := &einoschema.ToolInfo{
		Name: info.Name,
		Desc: info.Description,
		Extra: map[string]any{
			"provider":  info.Provider,
			"read_only": info.ReadOnly,
		},
	}
	if len(params) > 0 {
		toolInfo.ParamsOneOf = einoschema.NewParamsOneOfByParams(params)
	}

	return toolInfo
}

func toEinoParameterInfo(property *tools.Property, required []string, fieldName string) *einoschema.ParameterInfo {
	if property == nil {
		return nil
	}

	info := &einoschema.ParameterInfo{
		Type:     toEinoDataType(property.Type),
		Desc:     property.Description,
		Enum:     property.Enum,
		Required: isRequired(required, fieldName),
	}
	if property.Items != nil {
		info.ElemInfo = toEinoParameterInfo(property.Items, nil, "")
	}

	return info
}

func toEinoDataType(value string) einoschema.DataType {
	switch value {
	case "object":
		return einoschema.Object
	case "number":
		return einoschema.Number
	case "integer":
		return einoschema.Integer
	case "array":
		return einoschema.Array
	case "boolean":
		return einoschema.Boolean
	case "null":
		return einoschema.Null
	default:
		return einoschema.String
	}
}

func isRequired(required []string, fieldName string) bool {
	for _, candidate := range required {
		if candidate == fieldName {
			return true
		}
	}

	return false
}
