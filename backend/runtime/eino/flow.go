package eino

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	einoreact "github.com/cloudwego/eino/flow/agent/react"
	einoschema "github.com/cloudwego/eino/schema"
	runtimeevents "github.com/insmtx/SingerOS/backend/runtime/events"
)

// Flow wraps the Eino agent loop used by the SingerOS runtime agent.
type Flow struct {
	agent *einoreact.Agent
}

// FlowConfig describes the dependencies required to build an Eino flow.
type FlowConfig struct {
	Model        einomodel.ToolCallingChatModel
	ToolAdapter  *ToolAdapter
	Binding      ToolBinding
	SystemPrompt string
	MaxStep      int
}

// NewFlow creates a runnable Eino flow backed by SingerOS tool/runtime layers.
func NewFlow(ctx context.Context, cfg *FlowConfig) (*Flow, error) {
	if cfg == nil {
		return nil, fmt.Errorf("flow config is required")
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
		MessageModifier: buildMessageModifier(cfg.SystemPrompt),
		StreamToolCallChecker: func(_ context.Context, sr *einoschema.StreamReader[*einoschema.Message]) (bool, error) {
			defer sr.Close()
			hasTool := false
			for {
				msg, e := sr.Recv()
				if e != nil {
					if e == io.EOF {
						break
					}
					return false, e
				}
				if len(msg.ToolCalls) > 0 {
					hasTool = true
					// 不立刻返回，继续读到EOF以保持一致行为
				}
			}
			return hasTool, nil
		},
		MaxStep: maxStep,
	})
	if err != nil {
		return nil, fmt.Errorf("create eino react agent: %w", err)
	}

	return &Flow{agent: agent}, nil
}

// Generate runs one user request through the Eino agent loop.
func (f *Flow) Generate(ctx context.Context, userInput string) (*einoschema.Message, error) {
	if f == nil || f.agent == nil {
		return nil, fmt.Errorf("flow is not initialized")
	}
	if strings.TrimSpace(userInput) == "" {
		return nil, fmt.Errorf("user input is required")
	}

	return f.agent.Generate(ctx, []*einoschema.Message{
		einoschema.UserMessage(userInput),
	})
}

// Stream runs one user request through the Eino agent loop and emits runtime events.
func (f *Flow) Stream(ctx context.Context, userInput string, emitter *runtimeevents.Emitter) (*einoschema.Message, error) {
	if f == nil || f.agent == nil {
		return nil, fmt.Errorf("flow is not initialized")
	}
	if strings.TrimSpace(userInput) == "" {
		return nil, fmt.Errorf("user input is required")
	}

	stream, err := f.agent.Stream(ctx, []*einoschema.Message{
		einoschema.UserMessage(userInput),
	})
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	chunks := make([]*einoschema.Message, 0)
	var contentSnapshot strings.Builder
	var reasoningSnapshot strings.Builder
	for {
		chunk, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
			break
		}
		if recvErr != nil {
			return nil, recvErr
		}
		if chunk == nil {
			continue
		}
		chunks = append(chunks, chunk)
		if err := emitMessageChunk(ctx, emitter, chunk, &contentSnapshot, &reasoningSnapshot); err != nil {
			return nil, err
		}
	}
	if len(chunks) == 0 {
		return nil, fmt.Errorf("agent stream returned no messages")
	}
	return einoschema.ConcatMessages(chunks)
}

func emitMessageChunk(ctx context.Context, emitter *runtimeevents.Emitter, chunk *einoschema.Message, contentSnapshot *strings.Builder, reasoningSnapshot *strings.Builder) error {
	if emitter == nil || chunk == nil {
		return nil
	}
	if chunk.Content != "" {
		contentSnapshot.WriteString(chunk.Content)
		_ = emitter.Emit(ctx, &runtimeevents.RunEvent{
			Type:    runtimeevents.RunEventMessageDelta,
			Content: chunk.Content,
		})
	}
	if chunk.ReasoningContent != "" {
		reasoningSnapshot.WriteString(chunk.ReasoningContent)
		_ = emitter.Emit(ctx, &runtimeevents.RunEvent{
			Type:    runtimeevents.RunEventReasoningDelta,
			Content: chunk.ReasoningContent,
		})
	}
	for _, toolCall := range chunk.ToolCalls {
		args := map[string]any{}
		if strings.TrimSpace(toolCall.Function.Arguments) != "" {
			args["json"] = toolCall.Function.Arguments
		}
		_ = emitter.Emit(ctx, &runtimeevents.RunEvent{
			Type: runtimeevents.RunEventToolCallArguments,
			Content: eventContentJSON(map[string]any{
				"call_id":   toolCall.ID,
				"name":      toolCall.Function.Name,
				"arguments": args,
			}),
		})
	}
	return nil
}

func eventContentJSON(value interface{}) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(encoded)
}

func buildMessageModifier(systemPrompt string) einoreact.MessageModifier {
	prompt := strings.TrimSpace(systemPrompt)
	if prompt == "" {
		return nil
	}

	systemMessage := einoschema.SystemMessage(prompt)
	return func(ctx context.Context, input []*einoschema.Message) []*einoschema.Message {
		messages := make([]*einoschema.Message, 0, len(input)+1)
		messages = append(messages, systemMessage)
		messages = append(messages, input...)
		return messages
	}
}
