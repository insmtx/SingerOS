package eino

import (
	"context"
	"testing"

	"github.com/insmtx/SingerOS/backend/config"
)

func TestNewOpenAIChatModel(t *testing.T) {
	model, err := NewOpenAIChatModel(context.Background(), &config.LLMConfig{
		Provider: "openai",
		APIKey:   "test-key",
		Model:    "gpt-4o-mini",
		BaseURL:  "https://api.openai.example/v1",
	})
	if err != nil {
		t.Fatalf("create openai chat model: %v", err)
	}
	if model == nil {
		t.Fatalf("expected non-nil chat model")
	}
}
