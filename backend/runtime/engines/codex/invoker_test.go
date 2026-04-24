package codex

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
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
	logPath := filepath.Join(t.TempDir(), "codex-time.jsonl")

	adapter := NewAdapter(codexPath, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	handle, err := adapter.Run(ctx, engines.RunRequest{
		WorkDir: workDir,
		Prompt:  "用一句中文回答当前系统时间。不要修改任何文件。",
		LogPath: logPath,
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
	for event := range handle.Events {
		t.Logf("received event: type=%s, content=%s", event.Type, event.Content)
		finalEvent = event
	}
	if finalEvent.Type == engines.EventError {
		t.Fatalf("codex execution failed: %s", finalEvent.Content)
	}
	if finalEvent.Type != engines.EventDone {
		t.Fatalf("unexpected final event: %#v", finalEvent)
	}

	result := strings.TrimSpace(handle.ExtractResult(logPath))
	if result == "" {
		t.Fatalf("expected non-empty codex result, log path: %s", logPath)
	}
	t.Logf("codex current time result: %s", result)
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}
