package simplechat_test

import (
	"context"
	"os"
	"testing"

	"github.com/insmtx/SingerOS/backend/internal/agent"
	"github.com/insmtx/SingerOS/backend/internal/agent/simplechat"
)

func TestLoadFromEnv(t *testing.T) {
	cfg := simplechat.LoadFromEnv()
	if cfg == nil {
		t.Fatal("expected config to be non-nil")
	}
	if cfg.LLMProvider != "openai" {
		t.Errorf("expected provider openai, got %s", cfg.LLMProvider)
	}
}

func TestNewSimpleChat(t *testing.T) {
	cfg := &simplechat.Config{
		LLMProvider: "openai",
		APIKey:      "test-key",
		Model:       "gpt-4",
	}

	ctx := context.Background()
	sc, err := simplechat.New(ctx, cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if sc == nil {
		t.Fatal("expected simplechat to be non-nil")
	}
}

func TestSimpleChat_Run_RequiresAPIKey(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("skipping test: OPENAI_API_KEY not set")
	}

	cfg := &simplechat.Config{
		LLMProvider: "openai",
		APIKey:      os.Getenv("OPENAI_API_KEY"),
		Model:       "gpt-4",
	}

	ctx := context.Background()
	sc, err := simplechat.New(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to create simplechat: %v", err)
	}

	req := &agent.RequestContext{
		Input: agent.InputContext{
			Type: agent.InputTypeMessage,
			Text: "Hello, how are you?",
		},
	}

	result, err := sc.Run(ctx, req)
	if err != nil {
		t.Fatalf("failed to run simplechat: %v", err)
	}

	if result == nil {
		t.Fatal("expected result to be non-nil")
	}

	if result.Message == "" {
		t.Error("expected non-empty message")
	}
}

func TestSimpleChat_Run_EmptyInput(t *testing.T) {
	cfg := &simplechat.Config{
		LLMProvider: "openai",
		APIKey:      "test-key",
		Model:       "gpt-4",
	}

	ctx := context.Background()
	sc, err := simplechat.New(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to create simplechat: %v", err)
	}

	req := &agent.RequestContext{
		Input: agent.InputContext{
			Type: agent.InputTypeMessage,
			Text: "",
		},
	}

	_, err = sc.Run(ctx, req)
	if err == nil {
		t.Error("expected error for empty input")
	}
}
