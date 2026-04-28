package externalcli

import (
	"context"
	"testing"

	"github.com/insmtx/SingerOS/backend/internal/agent"
	"github.com/insmtx/SingerOS/backend/runtime/engines"
)

func TestRunnerAdaptsEngineResult(t *testing.T) {
	engine := &fakeEngine{result: "done"}
	runner, err := NewRunner("fake", engine, nil)
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}

	result, err := runner.Run(context.Background(), &agent.RequestContext{
		RunID: "run_cli",
		Input: agent.InputContext{
			Type: agent.InputTypeMessage,
			Text: "hello",
		},
		Runtime: agent.RuntimeOptions{WorkDir: "/tmp"},
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if result.Status != agent.RunStatusCompleted {
		t.Fatalf("expected completed, got %s", result.Status)
	}
	if result.Message != "done" {
		t.Fatalf("expected extracted result, got %q", result.Message)
	}
	if engine.runReq.WorkDir != "/tmp" {
		t.Fatalf("expected work dir /tmp, got %q", engine.runReq.WorkDir)
	}
	if engine.runReq.Prompt == "" {
		t.Fatal("expected prompt to be built")
	}
}

type fakeEngine struct {
	runReq engines.RunRequest
	result string
}

func (e *fakeEngine) Prepare(_ context.Context, _ engines.PrepareRequest) error {
	return nil
}

func (e *fakeEngine) RegisterMCP(_ context.Context, _ engines.MCPServerConfig) error {
	return nil
}

func (e *fakeEngine) Run(_ context.Context, req engines.RunRequest) (*engines.RunHandle, error) {
	e.runReq = req
	events := make(chan engines.Event, 3)
	events <- engines.Event{Type: engines.EventStarted}
	events <- engines.Event{Type: engines.EventResult, Content: e.result}
	events <- engines.Event{Type: engines.EventDone}
	close(events)
	return &engines.RunHandle{
		Events: events,
	}, nil
}
