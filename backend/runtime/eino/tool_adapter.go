package eino

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	einotool "github.com/cloudwego/eino/components/tool"
	einoschema "github.com/cloudwego/eino/schema"
	runtimeevents "github.com/insmtx/SingerOS/backend/runtime/events"
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
	InputSchema tools.Schema
}

// ToolCallRequest describes one model-initiated tool call.
type ToolCallRequest struct {
	Name        string
	Arguments   map[string]interface{}
	ToolContext tools.ToolContext
}

// ToolCallResult contains the execution result returned back to the model loop.
type ToolCallResult struct {
	Name   string
	Output string
}

// ToolAdapter bridges SingerOS tool registry to an Eino-facing API.
type ToolAdapter struct {
	registry *tools.Registry
}

// ToolBinding carries runtime-bound identity for one Eino agent execution.
type ToolBinding struct {
	ToolContext  tools.ToolContext
	AllowedTools []string
	Emitter      *runtimeevents.Emitter
}

// NewToolAdapter creates a new adapter over the shared tool registry.
func NewToolAdapter(registry *tools.Registry) *ToolAdapter {
	return &ToolAdapter{
		registry: registry,
	}
}

// Definitions returns the registry tools in an Eino-friendly description shape.
func (a *ToolAdapter) Definitions() []ToolDefinition {
	if a == nil || a.registry == nil {
		return nil
	}

	registeredTools := a.registry.List()
	definitions := make([]ToolDefinition, 0, len(registeredTools))
	for _, tool := range registeredTools {
		definitions = append(definitions, ToolDefinition{
			Name:        tool.Name(),
			Description: tool.Description(),
			InputSchema: tool.InputSchema(),
		})
	}

	return definitions
}

// EinoTools returns actual Eino tools bound to the current runtime identity.
func (a *ToolAdapter) EinoTools(binding ToolBinding) ([]einotool.BaseTool, error) {
	if a == nil || a.registry == nil {
		return nil, nil
	}

	boundTools, err := a.boundTools(binding.AllowedTools)
	if err != nil {
		return nil, err
	}

	result := make([]einotool.BaseTool, 0, len(boundTools))
	for _, tool := range boundTools {
		boundTool, err := tools.BindToolContext(tool, binding.ToolContext)
		if err != nil {
			return nil, err
		}
		result = append(result, &invokableTool{
			adapter: a,
			tool:    boundTool,
			binding: binding,
		})
	}

	return result, nil
}

func (a *ToolAdapter) boundTools(allowedTools []string) ([]tools.Tool, error) {
	if len(allowedTools) == 0 {
		return a.registry.List(), nil
	}

	result := make([]tools.Tool, 0, len(allowedTools))
	seen := make(map[string]struct{}, len(allowedTools))
	for _, name := range allowedTools {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}

		tool, err := a.registry.Get(name)
		if err != nil {
			return nil, err
		}
		result = append(result, tool)
	}

	return result, nil
}

// Invoke executes a tool call through the registry-backed adapter.
func (a *ToolAdapter) Invoke(ctx context.Context, req *ToolCallRequest) (*ToolCallResult, error) {
	if req == nil {
		return nil, fmt.Errorf("tool call request is required")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("tool name is required")
	}
	if a == nil || a.registry == nil {
		return nil, fmt.Errorf("tool registry is required")
	}

	tool, err := a.registry.Get(req.Name)
	if err != nil {
		return nil, err
	}

	boundTool, err := tools.BindToolContext(tool, req.ToolContext)
	if err != nil {
		return nil, err
	}
	return invokeTool(ctx, boundTool, req.Arguments, req.ToolContext)
}

func invokeTool(ctx context.Context, tool tools.Tool, arguments map[string]interface{}, toolCtx tools.ToolContext) (*ToolCallResult, error) {
	if tool == nil {
		return nil, fmt.Errorf("tool is required")
	}

	input := cloneToolInput(arguments)
	applyLegacyIdentityInput(input, toolCtx)
	if validator, ok := tool.(tools.Validator); ok {
		if err := validator.Validate(input); err != nil {
			return nil, fmt.Errorf("validate tool %s input: %w", tool.Name(), err)
		}
	}

	output, err := tool.Execute(tools.ContextWithToolContext(ctx, toolCtx), input)
	if err != nil {
		return nil, err
	}

	return &ToolCallResult{
		Name:   tool.Name(),
		Output: output,
	}, nil
}

type invokableTool struct {
	adapter *ToolAdapter
	tool    tools.Tool
	binding ToolBinding
}

func (t *invokableTool) Info(ctx context.Context) (*einoschema.ToolInfo, error) {
	if t == nil || t.tool == nil {
		return nil, fmt.Errorf("tool is required")
	}

	return toEinoToolInfo(t.tool), nil
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

	startedAt := time.Now()
	if err := t.emitToolEvent(ctx, runtimeevents.RunEventToolCallStarted, eventContentJSON(map[string]any{
		"name":      t.tool.Name(),
		"arguments": cloneArguments(input),
	})); err != nil {
		return "", err
	}

	result, err := invokeTool(ctx, t.tool, input, t.binding.ToolContext)
	if err != nil {
		_ = t.emitToolEvent(ctx, runtimeevents.RunEventToolCallFailed, eventContentJSON(map[string]any{
			"name":       t.tool.Name(),
			"elapsed_ms": time.Since(startedAt).Milliseconds(),
		}))
		return "", err
	}

	if err := t.emitToolEvent(ctx, runtimeevents.RunEventToolCallCompleted, eventContentJSON(map[string]any{
		"name":       t.tool.Name(),
		"result":     result.Output,
		"elapsed_ms": time.Since(startedAt).Milliseconds(),
	})); err != nil {
		return "", err
	}

	return result.Output, nil
}

func (t *invokableTool) emitToolEvent(ctx context.Context, eventType runtimeevents.RunEventType, content string) error {
	if t == nil || t.binding.Emitter == nil {
		return nil
	}
	err := t.binding.Emitter.Emit(ctx, &runtimeevents.RunEvent{
		Type:    eventType,
		Content: content,
	})
	_ = err
	return nil
}

func cloneArguments(input map[string]interface{}) map[string]any {
	if len(input) == 0 {
		return nil
	}
	cloned := make(map[string]any, len(input))
	for key, value := range input {
		cloned[key] = value
	}
	return cloned
}

func cloneToolInput(input map[string]interface{}) map[string]interface{} {
	if input == nil {
		return make(map[string]interface{})
	}
	cloned := make(map[string]interface{}, len(input))
	for key, value := range input {
		cloned[key] = value
	}
	return cloned
}

func applyLegacyIdentityInput(input map[string]interface{}, toolCtx tools.ToolContext) {
	if input == nil {
		return
	}
	if toolCtx.UserID != "" {
		if _, exists := input["user_id"]; !exists {
			input["user_id"] = toolCtx.UserID
		}
	}
	if toolCtx.AccountID != "" {
		if _, exists := input["account_id"]; !exists {
			input["account_id"] = toolCtx.AccountID
		}
	}
}

func toEinoToolInfo(tool tools.Tool) *einoschema.ToolInfo {
	if tool == nil {
		return nil
	}

	params := make(map[string]*einoschema.ParameterInfo)
	schema := tool.InputSchema()
	for name, property := range schema.Properties {
		params[name] = toEinoParameterInfo(property, schema.Required, name)
	}

	toolInfo := &einoschema.ToolInfo{
		Name: tool.Name(),
		Desc: tool.Description(),
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
