package orchestrator

import (
	"context"
	"testing"

	"github.com/insmtx/SingerOS/backend/interaction"
	"github.com/insmtx/SingerOS/backend/llm"
	"github.com/insmtx/SingerOS/backend/skills"
)

// 测试Orchestrator是否可以初始化并注册默认处理器
func TestOrchestratorInit(t *testing.T) {
	// 创建一个简单的subscriber实现
	subscriber := &mockSubscriber{}
	skillManager := skills.NewSimpleSkillManager()
	llmProvider := &mockLLMProvider{}

	orchestrator := NewOrchestrator(subscriber, skillManager, llmProvider)

	// 验证默认事件处理器被正确注册
	if _, exists := orchestrator.handlers["interaction.github.issue_comment"]; !exists {
		t.Errorf("Expected handler for interaction.github.issue_comment to be registered")
	}

	if _, exists := orchestrator.handlers["interaction.github.pull_request"]; !exists {
		t.Errorf("Expected handler for interaction.github.pull_request to be registered")
	}
}

// 辅助方法测试 - 验证注册和获取处理函数
func TestOrchestratorRegisterAndGet(t *testing.T) {
	subscriber := &mockSubscriber{}
	skillManager := skills.NewSimpleSkillManager()
	llmProvider := &mockLLMProvider{}
	orchestrator := NewOrchestrator(subscriber, skillManager, llmProvider)

	// 注册一个自定义处理器
	customTopic := "test.custom.topic"
	called := false
	handler := func(ctx context.Context, event *interaction.Event) error {
		called = true
		return nil
	}

	orchestrator.RegisterHandler(customTopic, handler)

	retrievedHandler, err := orchestrator.GetHandler(customTopic)
	if err != nil {
		t.Errorf("Expected to retrieve handler for %s: %v", customTopic, err)
	}

	if retrievedHandler == nil {
		t.Errorf("Expected non-nil handler for %s", customTopic)
	}

	// 调用处理器以验证它正常工作
	testEvent := &interaction.Event{EventID: "test"}
	err = retrievedHandler(context.Background(), testEvent)
	if err != nil {
		t.Errorf("Unexpected error when calling handler: %v", err)
	}

	if !called {
		t.Error("Expected handler to be called when invoked through GetHandler")
	}
}

// 简单的mock subscriber实现
type mockSubscriber struct{}

func (ms *mockSubscriber) Subscribe(ctx context.Context, topic string, handler func(event any)) error {
	return nil
}

// Mock LLM provider for testing
type mockLLMProvider struct{}

func (m *mockLLMProvider) Name() string { return "mock" }
func (m *mockLLMProvider) Generate(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
	return &llm.GenerateResponse{
		Content: "Mock response for testing",
		Usage: llm.TokenUsage{
			PromptTokens:     1,
			CompletionTokens: 1,
			TotalTokens:      2,
		},
	}, nil
}
func (m *mockLLMProvider) GenerateStream(ctx context.Context, req *llm.GenerateRequest) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	close(ch)
	return ch, nil
}
func (m *mockLLMProvider) CountTokens(text string) int { return len(text) / 4 }
func (m *mockLLMProvider) Models() []string            { return []string{"mock-model"} }
