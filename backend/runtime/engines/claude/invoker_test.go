package claude

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/insmtx/SingerOS/backend/runtime/engines"
)

func TestAdapterAskCurrentTime(t *testing.T) {
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		t.Skip("claude CLI not found in PATH")
	}
	apiKey := firstNonEmptyEnv("SINGEROS_LLM_API_KEY")
	if apiKey == "" {
		t.Skip("set SINGEROS_LLM_API_KEY to run the real claude adapter test")
	}

	workDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	adapter := NewAdapter(claudePath, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	handle, err := adapter.Run(ctx, engines.RunRequest{
		WorkDir: workDir,
		Prompt:  "请查询当前系统时间，并用一句中文回答。不要修改任何文件。",
		Model: engines.ModelConfig{
			Provider: "anthropic",
			APIKey:   apiKey,
			Model:    firstNonEmptyEnv("SINGEROS_LLM_MODEL"),
			BaseURL:  firstNonEmptyEnv("SINGEROS_LLM_BASE_URL"),
		},
		Timeout: 2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("run claude adapter: %v", err)
	}

	var finalEvent engines.Event
	var result string
	for event := range handle.Events {
		t.Logf("received event: type=%s, content=%s", event.Type, event.Content)
		if event.Type == engines.EventResult {
			result = strings.TrimSpace(event.Content)
		}
		finalEvent = event
	}
	if finalEvent.Type == engines.EventError {
		t.Fatalf("claude execution failed: %s", finalEvent.Content)
	}
	if finalEvent.Type != engines.EventDone {
		t.Fatalf("unexpected final event: %#v", finalEvent)
	}

	if result == "" {
		t.Fatal("expected non-empty claude result")
	}
	t.Logf("claude current time result: %s", result)
}

func TestParseClaudeLineEmitsResultEvent(t *testing.T) {
	state := &claudeStreamState{}
	event := parseClaudeLine(`{"type":"result","result":"final","is_error":false}`, state)
	if event.Type != engines.EventResult || event.Content != "final" {
		t.Fatalf("unexpected event: %#v", event)
	}
	if state.result != "final" || state.isError {
		t.Fatalf("unexpected state: %#v", state)
	}
}

func TestParseClaudeLineTracksAssistantFallback(t *testing.T) {
	state := &claudeStreamState{}
	event := parseClaudeLine(`{"type":"assistant","message":{"content":[{"type":"text","text":"answer"}]}}`, state)
	if event.Type != engines.EventMessageDelta || event.Content != "answer" {
		t.Fatalf("unexpected event: %#v", event)
	}
	if state.lastAssistantText != "answer" {
		t.Fatalf("got %q, want answer", state.lastAssistantText)
	}
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}
