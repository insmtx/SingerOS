// Package runtime defines the unified agent.run boundary for SingerOS.
package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	einomodel "github.com/cloudwego/eino/components/model"
	einoschema "github.com/cloudwego/eino/schema"
	"github.com/insmtx/SingerOS/backend/config"
	runtimeeino "github.com/insmtx/SingerOS/backend/runtime/eino"
	runtimeevents "github.com/insmtx/SingerOS/backend/runtime/events"
	"github.com/insmtx/SingerOS/backend/tools"
	skilltools "github.com/insmtx/SingerOS/backend/tools/skill"
	"github.com/ygpkg/yg-go/logs"
)

const defaultAgentSystemPrompt = "You are the SingerOS agent runtime. Use available skills and tools to analyze incoming events and respond with concrete, evidence-based actions."

// Agent is the SingerOS runtime agent entrypoint.
type Agent struct {
	chatModel     einomodel.ToolCallingChatModel
	toolAdapter   *runtimeeino.ToolAdapter
	skillsCatalog *skilltools.Catalog
	systemPrompt  string
}

// NewAgent creates the SingerOS agent backed by the Eino flow framework.
func NewAgent(ctx context.Context, llmConfig *config.LLMConfig, runtimeConfig Config) (*Agent, error) {
	if llmConfig == nil {
		return nil, fmt.Errorf("llm config is required")
	}
	if runtimeConfig.ToolRegistry == nil {
		return nil, fmt.Errorf("tool registry is required")
	}

	chatModel, err := runtimeeino.NewOpenAIChatModel(ctx, llmConfig)
	if err != nil {
		return nil, err
	}

	return &Agent{
		chatModel:     chatModel,
		toolAdapter:   runtimeeino.NewToolAdapter(runtimeConfig.ToolRegistry),
		skillsCatalog: runtimeConfig.SkillsCatalog,
		systemPrompt:  defaultAgentSystemPrompt,
	}, nil
}

// Run executes one normalized request through the SingerOS agent.
func (a *Agent) Run(ctx context.Context, req *RequestContext) (*RunResult, error) {
	startedAt := time.Now().UTC()
	if a == nil || a.chatModel == nil {
		return nil, fmt.Errorf("eino chat model is not initialized")
	}

	state, err := a.buildRunState(req)
	if err != nil {
		return nil, err
	}
	req = state.req

	if err := emitRunEvent(ctx, state.emitter, req, RunEventStarted, nil); err != nil {
		return nil, err
	}

	flow, err := runtimeeino.NewFlow(ctx, &runtimeeino.FlowConfig{
		Model:        a.chatModel,
		ToolAdapter:  a.toolAdapter,
		Binding:      state.toolBinding,
		SystemPrompt: state.systemPrompt,
		MaxStep:      state.maxStep,
	})
	if err != nil {
		emitRunError(ctx, state.emitter, req, err)
		return nil, err
	}

	var message interface {
		String() string
	}
	var resultMessage string
	var usage *UsagePayload
	if req.EventSink != nil {
		streamedMessage, streamErr := flow.Stream(ctx, state.userInput, state.emitter)
		err = streamErr
		if streamedMessage != nil {
			message = streamedMessage
			resultMessage = strings.TrimSpace(streamedMessage.Content)
			usage = usageFromResponseMeta(streamedMessage.ResponseMeta)
		}
	} else {
		generatedMessage, generateErr := flow.Generate(ctx, state.userInput)
		err = generateErr
		if generatedMessage != nil {
			message = generatedMessage
			resultMessage = strings.TrimSpace(generatedMessage.Content)
			usage = usageFromResponseMeta(generatedMessage.ResponseMeta)
		}
	}
	if err != nil {
		emitRunError(ctx, state.emitter, req, err)
		return nil, err
	}
	if resultMessage == "" && message != nil {
		resultMessage = formatLLMResultForLog(message)
	}

	result := &RunResult{
		RunID:       req.RunID,
		TraceID:     req.TraceID,
		Status:      RunStatusCompleted,
		Message:     resultMessage,
		Usage:       usage,
		StartedAt:   startedAt,
		CompletedAt: time.Now().UTC(),
	}

	if usage != nil {
		_ = state.emitter.Emit(ctx, &RunEvent{
			Type:    RunEventUsage,
			Content: eventContentJSON(usage),
		})
	}
	if err := emitRunEvent(ctx, state.emitter, req, RunEventCompleted, result); err != nil {
		return nil, err
	}

	logs.InfoContextf(ctx, "SingerOS runtime final LLM result: run_id=%s actor=%s result=%s",
		req.RunID, req.Actor.UserID, formatLLMResultForLog(message))

	return result, nil
}

func (a *Agent) buildRunState(req *RequestContext) (*runState, error) {
	if req == nil {
		return nil, errors.New("request context is required")
	}
	ensureRunDefaults(req)

	userInput := buildUserInput(req)
	if userInput == "" {
		userInput = string(req.Input.Type)
	}

	systemPrompt, err := a.buildSystemPrompt(req)
	if err != nil {
		return nil, err
	}

	emitter := runtimeevents.NewEmitter(req.RunID, req.TraceID, sinkForRequest(req))
	toolCtx := tools.ToolContext{
		RunID:          req.RunID,
		TraceID:        req.TraceID,
		UserID:         req.Actor.UserID,
		AccountID:      req.Actor.AccountID,
		Channel:        req.Actor.Channel,
		ChatID:         req.Conversation.ID,
		ConversationID: req.Conversation.ID,
		ExternalID:     req.Actor.ExternalID,
		Metadata:       req.Metadata,
	}
	return &runState{
		req:          req,
		emitter:      emitter,
		userInput:    userInput,
		systemPrompt: systemPrompt,
		toolBinding: runtimeeino.ToolBinding{
			ToolContext:  toolCtx,
			AllowedTools: req.Capability.AllowedTools,
			Emitter:      emitter,
		},
		maxStep: maxStepForRequest(req),
	}, nil
}

func buildUserInput(req *RequestContext) string {
	if req == nil {
		return ""
	}

	switch {
	case strings.TrimSpace(req.Input.Text) != "":
		return strings.TrimSpace(req.Input.Text)
	case len(req.Input.Messages) > 0:
		lines := make([]string, 0, len(req.Input.Messages))
		for _, message := range req.Input.Messages {
			if strings.TrimSpace(message.Content) == "" {
				continue
			}
			role := message.Role
			if role == "" {
				role = "user"
			}
			lines = append(lines, fmt.Sprintf("%s: %s", role, message.Content))
		}
		return strings.Join(lines, "\n")
	default:
		return string(req.Input.Type)
	}
}

func (a *Agent) buildSystemPrompt(req *RequestContext) (string, error) {
	sections := make([]string, 0, 4)
	if a != nil {
		if base := strings.TrimSpace(a.systemPromptForRequest(req)); base != "" {
			sections = append(sections, base)
		}

		skillsContext, err := buildSkillsContext(a.skillsCatalog)
		if err != nil {
			return "", err
		}
		if skillsContext != nil {
			if summary := strings.TrimSpace(skillsContext.SummarySection); summary != "" {
				sections = append(sections, summary)
			}
			for _, section := range skillsContext.AlwaysSections {
				if trimmed := strings.TrimSpace(section); trimmed != "" {
					sections = append(sections, trimmed)
				}
			}
		}
	}
	return strings.Join(sections, "\n\n"), nil
}

func (a *Agent) systemPromptForRequest(req *RequestContext) string {
	prompt := strings.TrimSpace(a.systemPrompt)
	if req != nil && strings.TrimSpace(req.Assistant.SystemPrompt) != "" {
		if prompt == "" {
			prompt = strings.TrimSpace(req.Assistant.SystemPrompt)
		} else {
			prompt += "\n\n" + strings.TrimSpace(req.Assistant.SystemPrompt)
		}
	}
	if req == nil {
		return prompt
	}
	return prompt
}

func ensureRunDefaults(req *RequestContext) {
	if req.RunID == "" {
		req.RunID = fmt.Sprintf("run_%d", time.Now().UTC().UnixNano())
	}
	if req.Input.Type == "" {
		req.Input.Type = InputTypeMessage
	}
}

func maxStepForRequest(req *RequestContext) int {
	if req != nil && req.Runtime.MaxStep > 0 {
		return req.Runtime.MaxStep
	}
	return 12
}

func sinkForRequest(req *RequestContext) runtimeevents.EventSink {
	if req == nil || req.EventSink == nil {
		return runtimeevents.NewNoopSink()
	}
	return req.EventSink
}

func emitRunEvent(ctx context.Context, emitter *runtimeevents.Emitter, req *RequestContext, eventType RunEventType, result *RunResult) error {
	event := &RunEvent{Type: eventType}
	if result != nil {
		event.Content = result.Message
	}
	_ = emitter.Emit(ctx, event)
	return nil
}

func emitRunError(ctx context.Context, emitter *runtimeevents.Emitter, req *RequestContext, err error) {
	if err == nil {
		return
	}
	eventType := RunEventFailed
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		eventType = RunEventCancelled
	}
	_ = emitter.Emit(ctx, &RunEvent{
		Type:    eventType,
		Content: err.Error(),
	})
}

func eventContentJSON(value interface{}) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(encoded)
}

func usageFromResponseMeta(meta *einoschema.ResponseMeta) *UsagePayload {
	if meta == nil || meta.Usage == nil {
		return nil
	}
	return &UsagePayload{
		InputTokens:  meta.Usage.PromptTokens,
		OutputTokens: meta.Usage.CompletionTokens,
		TotalTokens:  meta.Usage.TotalTokens,
	}
}

func formatLLMResultForLog(message interface{ String() string }) string {
	if message == nil {
		return "<nil>"
	}

	formatted := strings.TrimSpace(message.String())
	if formatted == "" {
		return "<empty>"
	}
	if len(formatted) > 2000 {
		return formatted[:2000] + "...(truncated)"
	}
	return formatted
}
