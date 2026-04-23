package eino

import (
	"context"
	"fmt"

	einoopenai "github.com/cloudwego/eino-ext/components/model/openai"
	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/insmtx/SingerOS/backend/config"
)

// NewOpenAIChatModel creates a real Eino ToolCallingChatModel from SingerOS LLM config.
func NewOpenAIChatModel(ctx context.Context, cfg *config.LLMConfig) (einomodel.ToolCallingChatModel, error) {
	if cfg == nil {
		return nil, fmt.Errorf("llm config is required")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("llm api key is required")
	}

	modelName := cfg.Model
	if modelName == "" {
		modelName = "gpt-4o-mini"
	}

	chatModel, err := einoopenai.NewChatModel(ctx, &einoopenai.ChatModelConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Model:   modelName,
	})
	if err != nil {
		return nil, fmt.Errorf("create eino openai chat model: %w", err)
	}

	return chatModel, nil
}
