package claude

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
	logPath := filepath.Join(t.TempDir(), "claude-time.jsonl")

	adapter := NewAdapter(claudePath, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	handle, err := adapter.Run(ctx, engines.RunRequest{
		WorkDir: workDir,
		Prompt:  "请查询当前系统时间，并用一句中文回答。不要修改任何文件。",
		LogPath: logPath,
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
	for event := range handle.Events {
		t.Logf("received event: type=%s, content=%s", event.Type, event.Content)
		finalEvent = event
	}
	if finalEvent.Type == engines.EventError {
		t.Fatalf("claude execution failed: %s", finalEvent.Content)
	}
	if finalEvent.Type != engines.EventDone {
		t.Fatalf("unexpected final event: %#v", finalEvent)
	}

	result := strings.TrimSpace(handle.ExtractResult(logPath))
	if result == "" {
		t.Fatalf("expected non-empty claude result, log path: %s", logPath)
	}
	t.Logf("claude current time result: %s", result)
}

func TestExtractResultFromLogPrefersResultEvent(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "claude.jsonl")
	content := `{"type":"assistant","message":{"content":[{"type":"text","text":"draft"}]}}` + "\n" +
		`{"type":"result","result":"final","is_error":false}` + "\n"
	if err := os.WriteFile(logPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write log: %v", err)
	}

	if got := ExtractResultFromLog(logPath); got != "final" {
		t.Fatalf("got %q, want final", got)
	}
}

func TestExtractResultFromLogFallsBackToAssistantText(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "claude.jsonl")
	content := `{"type":"assistant","message":{"content":[{"type":"text","text":"answer"}]}}` + "\n"
	if err := os.WriteFile(logPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write log: %v", err)
	}

	if got := ExtractResultFromLog(logPath); got != "answer" {
		t.Fatalf("got %q, want answer", got)
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
