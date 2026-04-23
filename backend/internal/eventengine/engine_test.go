package eventengine

import (
	"context"
	"testing"

	"github.com/insmtx/SingerOS/backend/interaction"
	"github.com/insmtx/SingerOS/backend/internal/execution"
)

type mockSubscriber struct{}

func (ms *mockSubscriber) Subscribe(ctx context.Context, topic string, handler func(event any)) error {
	return nil
}

type mockExecutor struct{}

func (me *mockExecutor) Execute(ctx context.Context, task *execution.Task) *execution.Result {
	return &execution.Result{Success: true}
}

func TestEventEngineInit(t *testing.T) {
	subscriber := &mockSubscriber{}
	execEngine := execution.NewExecutionEngine()
	execEngine.RegisterExecutor(execution.TaskTypeAgent, &mockExecutor{})
	engine := NewEventEngine(subscriber, execEngine)

	if _, exists := engine.handlers["interaction.github.issue_comment"]; !exists {
		t.Errorf("Expected handler for interaction.github.issue_comment to be registered")
	}

	if _, exists := engine.handlers["interaction.github.pull_request"]; !exists {
		t.Errorf("Expected handler for interaction.github.pull_request to be registered")
	}

	if _, exists := engine.handlers["interaction.github.push"]; !exists {
		t.Errorf("Expected handler for interaction.github.push to be registered")
	}
}

func TestEventEngineRegisterAndGet(t *testing.T) {
	subscriber := &mockSubscriber{}
	execEngine := execution.NewExecutionEngine()
	execEngine.RegisterExecutor(execution.TaskTypeAgent, &mockExecutor{})
	engine := NewEventEngine(subscriber, execEngine)

	customTopic := "test.custom.topic"
	called := false
	handler := func(ctx context.Context, event *interaction.Event) error {
		called = true
		return nil
	}

	engine.RegisterHandler(customTopic, handler)

	retrievedHandler, err := engine.GetHandler(customTopic)
	if err != nil {
		t.Errorf("Expected to retrieve handler for %s: %v", customTopic, err)
	}

	if retrievedHandler == nil {
		t.Errorf("Expected non-nil handler for %s", customTopic)
	}

	testEvent := &interaction.Event{EventID: "test"}
	err = retrievedHandler(context.Background(), testEvent)
	if err != nil {
		t.Errorf("Unexpected error when calling handler: %v", err)
	}

	if !called {
		t.Error("Expected handler to be called when invoked through GetHandler")
	}
}
