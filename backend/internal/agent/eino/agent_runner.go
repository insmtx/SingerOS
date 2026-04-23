package eino

import (
	"context"
	"fmt"
	"strings"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	einoreact "github.com/cloudwego/eino/flow/agent/react"
	einoschema "github.com/cloudwego/eino/schema"
	runtimeprompt "github.com/insmtx/SingerOS/backend/internal/agent/prompt"
)

// AgentRunner wraps the Eino agent loop behind a stable SingerOS runtime boundary.
type AgentRunner struct {
	agent *einoreact.Agent
}

// AgentRunnerConfig describes the dependencies required to build an Eino agent runner.
type AgentRunnerConfig struct {
	Model        einomodel.ToolCallingChatModel
	ToolAdapter  *ToolAdapter
	Binding      ToolBinding
	SystemPrompt string
	Skills       *runtimeprompt.SkillsContext
	Tools        *runtimeprompt.ToolsContext
	MaxStep      int
}

// NewAgentRunner creates a runnable Eino agent backed by SingerOS tool/runtime layers.
func NewAgentRunner(ctx context.Context, cfg *AgentRunnerConfig) (*AgentRunner, error) {
	if cfg == nil {
		return nil, fmt.Errorf("agent runner config is required")
	}
	if cfg.Model == nil {
		return nil, fmt.Errorf("tool-calling model is required")
	}
	if cfg.ToolAdapter == nil {
		return nil, fmt.Errorf("tool adapter is required")
	}

	toolList, err := cfg.ToolAdapter.EinoTools(cfg.Binding)
	if err != nil {
		return nil, fmt.Errorf("build eino tools: %w", err)
	}

	maxStep := cfg.MaxStep
	if maxStep <= 0 {
		maxStep = 8
	}

	agent, err := einoreact.NewAgent(ctx, &einoreact.AgentConfig{
		ToolCallingModel: cfg.Model,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: toolList,
		},
		MessageModifier: buildMessageModifier(cfg.SystemPrompt, cfg.Skills, cfg.Tools),
		MaxStep:         maxStep,
	})
	if err != nil {
		return nil, fmt.Errorf("create eino react agent: %w", err)
	}

	return &AgentRunner{agent: agent}, nil
}

// Generate runs one user request through the Eino agent loop.
func (r *AgentRunner) Generate(ctx context.Context, userInput string) (*einoschema.Message, error) {
	if r == nil || r.agent == nil {
		return nil, fmt.Errorf("agent runner is not initialized")
	}
	if strings.TrimSpace(userInput) == "" {
		return nil, fmt.Errorf("user input is required")
	}

	return r.agent.Generate(ctx, []*einoschema.Message{
		einoschema.UserMessage(userInput),
	})
}

func buildMessageModifier(systemPrompt string, skills *runtimeprompt.SkillsContext, tools *runtimeprompt.ToolsContext) einoreact.MessageModifier {
	sections := make([]string, 0, 4)
	if trimmed := strings.TrimSpace(systemPrompt); trimmed != "" {
		sections = append(sections, trimmed)
	}
	if skills != nil && strings.TrimSpace(skills.SummarySection) != "" {
		sections = append(sections, skills.SummarySection)
	}
	if tools != nil && strings.TrimSpace(tools.SummarySection) != "" {
		sections = append(sections, tools.SummarySection)
	}
	if skills != nil {
		for _, section := range skills.AlwaysSections {
			if trimmed := strings.TrimSpace(section); trimmed != "" {
				sections = append(sections, trimmed)
			}
		}
	}

	if len(sections) == 0 {
		return nil
	}

	systemMessage := einoschema.SystemMessage(strings.Join(sections, "\n\n"))
	return func(ctx context.Context, input []*einoschema.Message) []*einoschema.Message {
		messages := make([]*einoschema.Message, 0, len(input)+1)
		messages = append(messages, systemMessage)
		messages = append(messages, input...)
		return messages
	}
}
