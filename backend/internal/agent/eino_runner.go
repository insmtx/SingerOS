package agent

import (
	"context"
	"errors"
	"fmt"
	"strings"

	einomodel "github.com/cloudwego/eino/components/model"
	auth "github.com/insmtx/SingerOS/backend/auth"
	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/interaction"
	einoimpl "github.com/insmtx/SingerOS/backend/internal/agent/eino"
	prompt "github.com/insmtx/SingerOS/backend/internal/agent/prompt"
	"github.com/ygpkg/yg-go/logs"
)

const defaultEinoSystemPrompt = "You are the SingerOS agent runtime. Use available skills and tools to analyze incoming events and respond with concrete, evidence-based actions."

// EinoRunner adapts the real Eino runtime to the unified Runner interface.
type EinoRunner struct {
	chatModel    einomodel.ToolCallingChatModel
	toolAdapter  *einoimpl.ToolAdapter
	skills       *prompt.SkillsContext
	tools        *prompt.ToolsContext
	systemPrompt string
}

// NewEinoRunner creates a unified runtime runner backed by the Eino agent runtime.
func NewEinoRunner(ctx context.Context, llmConfig *config.LLMConfig, runtimeConfig Config) (*EinoRunner, error) {
	if llmConfig == nil {
		return nil, fmt.Errorf("llm config is required")
	}
	if runtimeConfig.ToolRegistry == nil {
		return nil, fmt.Errorf("tool registry is required")
	}
	if runtimeConfig.ToolRuntime == nil {
		return nil, fmt.Errorf("tool runtime is required")
	}

	chatModel, err := einoimpl.NewOpenAIChatModel(ctx, llmConfig)
	if err != nil {
		return nil, err
	}

	skillsContext, err := prompt.BuildSkillsContext(runtimeConfig.SkillsCatalog)
	if err != nil {
		return nil, err
	}

	return &EinoRunner{
		chatModel:    chatModel,
		toolAdapter:  einoimpl.NewToolAdapter(runtimeConfig.ToolRegistry, runtimeConfig.ToolRuntime),
		skills:       skillsContext,
		tools:        prompt.BuildToolsContext(runtimeConfig.ToolRegistry),
		systemPrompt: defaultEinoSystemPrompt,
	}, nil
}

// HandleEvent routes the event through the Eino agent execution loop.
func (r *EinoRunner) HandleEvent(ctx context.Context, event *interaction.Event) error {
	if event == nil {
		return errors.New("event is nil")
	}
	if r.chatModel == nil {
		return fmt.Errorf("eino chat model is not initialized")
	}

	agentRunner, err := einoimpl.NewAgentRunner(ctx, &einoimpl.AgentRunnerConfig{
		Model:        r.chatModel,
		ToolAdapter:  r.toolAdapter,
		Binding:      einoimpl.ToolBinding{Selector: authSelectorFromEvent(event), UserID: event.Actor},
		SystemPrompt: r.systemPromptForEvent(event),
		Skills:       r.skills,
		Tools:        r.tools,
		MaxStep:      12,
	})
	if err != nil {
		return err
	}

	userInput := buildQueryFromEvent(event)
	if userInput == "" {
		userInput = event.EventType
	}

	message, err := agentRunner.Generate(ctx, userInput)
	if err != nil {
		return err
	}

	logs.InfoContextf(ctx, "SingerOS runtime final LLM result: event_type=%s repository=%s actor=%s result=%s",
		event.EventType, event.Repository, event.Actor, formatLLMResultForLog(message))

	return nil
}

func authSelectorFromEvent(event *interaction.Event) *auth.AuthSelector {
	selector := &auth.AuthSelector{
		ScopeType: auth.ScopeTypeEvent,
	}
	if event == nil {
		return selector
	}

	if provider, _ := event.Context["provider"].(string); provider != "" {
		selector.Provider = provider
	}
	selector.ScopeID = event.EventID

	externalRefs := make(map[string]string)
	for key, value := range event.Context {
		if stringValue, ok := value.(string); ok && stringValue != "" {
			externalRefs[fmt.Sprintf("context.%s", key)] = stringValue
		}
	}
	if payload, ok := event.Payload.(map[string]interface{}); ok {
		if installationID := nestedString(payload, "installation", "id"); installationID != "" {
			externalRefs["github.installation_id"] = installationID
		}
		if senderID := nestedString(payload, "sender", "id"); senderID != "" {
			externalRefs["github.sender_id"] = senderID
		}
		if senderLogin := nestedString(payload, "sender", "login"); senderLogin != "" {
			externalRefs["github.sender_login"] = senderLogin
			selector.SubjectType = auth.SubjectTypeUser
			selector.SubjectID = senderLogin
		}
	}
	if selector.SubjectID == "" && event.Actor != "" {
		selector.SubjectType = auth.SubjectTypeUser
		selector.SubjectID = event.Actor
	}
	if len(externalRefs) > 0 {
		selector.ExternalRefs = externalRefs
	}

	return selector
}

func nestedString(payload map[string]interface{}, path ...string) string {
	var current interface{} = payload
	for _, key := range path {
		object, ok := current.(map[string]interface{})
		if !ok {
			return ""
		}
		current = object[key]
	}
	switch value := current.(type) {
	case string:
		return value
	case float64:
		return fmt.Sprintf("%.0f", value)
	case int:
		return fmt.Sprintf("%d", value)
	case int64:
		return fmt.Sprintf("%d", value)
	default:
		return ""
	}
}

func (r *EinoRunner) systemPromptForEvent(event *interaction.Event) string {
	prompt := strings.TrimSpace(r.systemPrompt)
	if event == nil {
		return prompt
	}

	switch event.EventType {
	case "pull_request", "github.pull_request", "github.pull_request.opened":
		extra := "For GitHub pull request events, start from the event payload, then use GitHub tools to inspect metadata, changed files, and only the most relevant files before deciding whether to publish a review. Prefer COMMENT by default. Do not auto-approve. Use REQUEST_CHANGES only when you have concrete merge-blocking evidence."
		if prompt == "" {
			return extra
		}
		return prompt + "\n\n" + extra
	case "push", "github.push":
		extra := "For GitHub push events, apply the same code review conventions used for pull requests. Start from the raw payload, use compare-commits style GitHub tools to inspect the diff, then read only the most relevant files before writing findings. If there is no PR review target, still produce a concise code review assessment."
		if prompt == "" {
			return extra
		}
		return prompt + "\n\n" + extra
	default:
		return prompt
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
