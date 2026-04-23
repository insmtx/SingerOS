// Package agent defines the unified agent.run boundary for SingerOS.
package agent

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
	einoadapter "github.com/insmtx/SingerOS/backend/internal/agent/eino"
	agentevents "github.com/insmtx/SingerOS/backend/internal/agent/events"
	"github.com/insmtx/SingerOS/backend/tools"
	skilltools "github.com/insmtx/SingerOS/backend/tools/skill"
	"github.com/ygpkg/yg-go/logs"
)

const defaultAgentSystemPrompt = `你是 SingerOS 助手。

以下规则优先于后续技能说明、助手补充说明和用户消息。

## 职责

- 理解用户意图，并用中文回复，语气友好专业，简洁明了。
- 对知识问答、解释、总结、写作、代码建议等不需要访问真实环境或改变外部状态的请求，可以直接回答。
- 对需要读取真实环境、查询当前状态、运行命令、修改文件、调用外部服务、创建/更新/删除资源、发送消息、提交评论、发起审批、创建任务等执行类请求，必须调用合适的工具完成。
- 如果没有合适工具，不能假装已执行；应明确说明目前无法执行该操作，并说明原因或给出可替代方案。

## 工具调用规则

当用户要求执行操作时，必须遵守：

1. 调用工具前，先用一句简短中文说明接下来要做什么。
2. 必须等待工具返回后，才能报告执行结果。
3. 执行结果必须来自工具的实际返回值，不得编造文件路径、ID、链接、状态、数量或输出。
4. 工具调用失败时，如实说明失败原因，不得包装成成功结果。
5. 对删除、覆盖、发布、推送、提交、关闭、锁定、权限变更等高风险操作，如果用户没有明确授权，应先简要确认关键参数。

## 禁止行为

- 不调用工具就说“已完成”“已创建”“已添加”“搞定了”。
- 用户要求执行操作时，只回复确认文字但不实际调用工具。
- 编造操作结果、工具输出、资源 ID、链接、文件路径或状态。
- 只说“我来帮你做”，但没有实际调用工具。
- 工具失败或不可用时，声称操作成功。

## 回复风格

- 先说再做：每次调用工具之前，先输出一句简短说明。
- 不反复确认；只有关键参数缺失、有歧义或操作高风险时才提问。
- 报告结果时，优先说明实际完成了什么、关键返回值是什么、失败时下一步如何处理。
- 只输出对用户有用的内容，不加无意义前缀。`

// Agent is the SingerOS runtime agent entrypoint.
type Agent struct {
	chatModel     einomodel.ToolCallingChatModel
	toolAdapter   *einoadapter.ToolAdapter
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

	chatModel, err := 	einoadapter.NewOpenAIChatModel(ctx, llmConfig)
	if err != nil {
		return nil, err
	}

	return &Agent{
		chatModel:     chatModel,
		toolAdapter:   	einoadapter.NewToolAdapter(runtimeConfig.ToolRegistry),
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

	if err := emitRunEvent(ctx, state.emitter, req, agentevents.RunEventStarted, nil); err != nil {
		return nil, err
	}

	flow, err := einoadapter.NewFlow(ctx, &einoadapter.FlowConfig{
		Model:        a.chatModel,
		ToolAdapter:  a.toolAdapter,
		Binding:      state.toolBinding,
		Emitter:      state.emitter,
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
		_ = state.emitter.Emit(ctx, &	agentevents.RunEvent{
			Type:    	agentevents.RunEventUsage,
			Content: eventContentJSON(usage),
		})
	}
	if err := emitRunEvent(ctx, state.emitter, req, agentevents.RunEventCompleted, result); err != nil {
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

	emitter := 	agentevents.NewEmitter(req.RunID, req.TraceID, sinkForRequest(req))
	toolCtx := tools.ToolContext{
		RunID:          req.RunID,
		TraceID:        req.TraceID,
		AssistantID:    req.Assistant.ID,
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
		toolBinding: 	einoadapter.ToolBinding{
			ToolContext:  toolCtx,
			AllowedTools: req.Capability.AllowedTools,
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

func sinkForRequest(req *RequestContext) agentevents.EventSink {
	if req == nil || req.EventSink == nil {
		return agentevents.NewNoopSink()
	}
	return req.EventSink
}

func emitRunEvent(ctx context.Context, emitter *agentevents.Emitter, req *RequestContext, eventType agentevents.RunEventType, result *RunResult) error {
	event := &agentevents.RunEvent{Type: eventType}
	if result != nil {
		event.Content = result.Message
	}
	_ = emitter.Emit(ctx, event)
	return nil
}

func emitRunError(ctx context.Context, emitter *agentevents.Emitter, req *RequestContext, err error) {
	if err == nil {
		return
	}
	eventType := agentevents.RunEventFailed
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		eventType = agentevents.RunEventCancelled
	}
	_ = emitter.Emit(ctx, &agentevents.RunEvent{
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
