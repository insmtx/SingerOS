package codex

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/insmtx/SingerOS/backend/runtime/engines"
)

func TestSessionStore(t *testing.T) {
	store := NewSessionStore()
	if _, ok := store.Get("session-1"); ok {
		t.Fatal("empty store should not return a mapping")
	}

	store.Set("session-1", "thread-1")
	if got, ok := store.Get("session-1"); !ok || got != "thread-1" {
		t.Fatalf("got %q, %v; want thread-1, true", got, ok)
	}
}

func TestAdapterAskCurrentTime(t *testing.T) {
	codexPath, err := exec.LookPath("codex")
	if err != nil {
		t.Skip("codex CLI not found in PATH")
	}
	apiKey := firstNonEmptyEnv("SINGEROS_LLM_API_KEY")
	if apiKey == "" {
		t.Skip("set SINGEROS_LLM_API_KEY to run the real codex adapter test")
	}

	workDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	adapter := NewAdapter(codexPath, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	handle, err := adapter.Run(ctx, engines.RunRequest{
		WorkDir: workDir,
		Prompt:  "用一句中文回答当前系统时间。不要修改任何文件。",
		Model: engines.ModelConfig{
			Provider: "openai",
			APIKey:   apiKey,
			Model:    firstNonEmptyEnv("SINGEROS_LLM_MODEL"),
			BaseURL:  firstNonEmptyEnv("SINGEROS_LLM_BASE_URL"),
		},
		Timeout: 2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("run codex adapter: %v", err)
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
		t.Fatalf("codex execution failed: %s", finalEvent.Content)
	}
	if finalEvent.Type != engines.EventDone {
		t.Fatalf("unexpected final event: %#v", finalEvent)
	}

	if result == "" {
		t.Fatal("expected non-empty codex result")
	}
	t.Logf("codex current time result: %s", result)
}

func TestParseCodexLineEmitsResult(t *testing.T) {
	event, threadID := parseCodexLine(`{"type":"item.completed","item":{"type":"agent_message","text":"final"}}`)
	if threadID != "" {
		t.Fatalf("unexpected thread id: %s", threadID)
	}
	if event.Type != engines.EventResult || event.Content != "final" {
		t.Fatalf("unexpected event: %#v", event)
	}
}

func TestParseCodexLineCapturesThread(t *testing.T) {
	event, threadID := parseCodexLine(`{"type":"thread.started","thread_id":"thread-1"}`)
	if event.Type != "" {
		t.Fatalf("unexpected event: %#v", event)
	}
	if threadID != "thread-1" {
		t.Fatalf("got thread id %q, want thread-1", threadID)
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
