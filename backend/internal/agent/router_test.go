package agent_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/internal/agent"
	agentevents "github.com/insmtx/SingerOS/backend/internal/agent/events"
	"github.com/insmtx/SingerOS/backend/internal/agent/externalcli"
	"github.com/insmtx/SingerOS/backend/runtime/engines"
	"github.com/insmtx/SingerOS/backend/runtime/engines/claude"
)

func TestRuntimeRouterUsesRequestedRuntime(t *testing.T) {
	router := agent.NewRuntimeRouter(agent.RuntimeKindSingerOS)
	singerRunner := &testRunner{message: "singer"}
	codexRunner := &testRunner{message: "codex"}

	if err := router.Register(agent.RuntimeKindSingerOS, singerRunner); err != nil {
		t.Fatalf("register singeros: %v", err)
	}
	if err := router.Register("codex", codexRunner); err != nil {
		t.Fatalf("register codex: %v", err)
	}

	result, err := router.Run(context.Background(), &agent.RequestContext{
		Runtime: agent.RuntimeOptions{Kind: "codex"},
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if result.Message != "codex" {
		t.Fatalf("expected codex runner, got %q", result.Message)
	}
}

func TestRuntimeRouterUsesDefaultRuntime(t *testing.T) {
	router := agent.NewRuntimeRouter(agent.RuntimeKindSingerOS)
	if err := router.Register(agent.RuntimeKindSingerOS, &testRunner{message: "default"}); err != nil {
		t.Fatalf("register default: %v", err)
	}

	result, err := router.Run(context.Background(), &agent.RequestContext{})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if result.Message != "default" {
		t.Fatalf("expected default runner, got %q", result.Message)
	}
}

type testRunner struct {
	message string
}

func (r *testRunner) Run(_ context.Context, req *agent.RequestContext) (*agent.RunResult, error) {
	return &agent.RunResult{
		RunID:   req.RunID,
		TraceID: req.TraceID,
		Status:  agent.RunStatusCompleted,
		Message: r.message,
	}, nil
}

func TestRuntimeRouterClaudeRunnerCallsSingerOSEchoTool(t *testing.T) {
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		t.Skipf("claude CLI not found in PATH: %v", err)
	}

	repoRoot := findRepoRoot(t)
	llmConfig := loadRealLLMConfig(t)
	t.Logf("using model config: provider=%q model=%q base_url_set=%t api_key_set=%t",
		llmConfig.Provider, llmConfig.Model, llmConfig.BaseURL != "", llmConfig.APIKey != "")

	claudeEngine := claude.NewAdapter(claudePath, nil)
	claudeRunner, err := externalcli.NewRunner(engines.EngineClaude, claudeEngine, llmConfig)
	if err != nil {
		t.Fatalf("new claude runner: %v", err)
	}

	router := agent.NewRuntimeRouter(engines.EngineClaude)
	if err := router.Register(engines.EngineClaude, claudeRunner); err != nil {
		t.Fatalf("register claude runner: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	eventSink := agent.RunEventSink(agentevents.SinkFunc(func(_ context.Context, event *agentevents.RunEvent) error {
		encoded, _ := json.Marshal(event)
		t.Logf("runtime event: %s", string(encoded))
		return nil
	}))

	result, err := router.Run(ctx, &agent.RequestContext{
		RunID:   "run_claude_echo_integration",
		TraceID: "trace_claude_echo_integration",
		Assistant: agent.AssistantContext{
			ID:   "assistant_integration_test",
			Name: "Claude CLI Integration Test",
		},
		Actor: agent.ActorContext{
			UserID:  "integration_test",
			Channel: "go_test",
		},
		Input: agent.InputContext{
			Type: agent.InputTypeTaskInstruction,
			Text: `必须调用已配置的 SingerOS MCP 工具 singeros_echo，参数 message 使用 "hello from claude runner"。
调用完成后，在最终答案中原样返回工具结果 JSON，并说明你已经完成工具调用。`,
		},
		Runtime: agent.RuntimeOptions{
			Kind:    engines.EngineClaude,
			WorkDir: repoRoot,
		},
		EventSink: eventSink,
	})
	if err != nil {
		if result != nil {
			resultJSON, _ := json.MarshalIndent(result, "", "  ")
			t.Logf("failed run result:\n%s", string(resultJSON))
			logProcessOutput(t, result)
		}
		t.Fatalf("run claude runner: %v", err)
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	t.Logf("final run result:\n%s", string(resultJSON))
	logProcessOutput(t, result)

	if result.Status != agent.RunStatusCompleted {
		t.Fatalf("expected completed status, got %s", result.Status)
	}
	if !strings.Contains(result.Message, "hello from claude runner") {
		t.Fatalf("expected final result to include echo message, got %q", result.Message)
	}
	if !strings.Contains(result.Message, "SingerOS") {
		t.Fatalf("expected final result to include echo server name, got %q", result.Message)
	}
}

func loadRealLLMConfig(t *testing.T) *config.LLMConfig {
	t.Helper()

	llmConfig := &config.LLMConfig{
		Provider: strings.TrimSpace(os.Getenv("SINGEROS_LLM_PROVIDER")),
		APIKey:   strings.TrimSpace(os.Getenv("SINGEROS_LLM_API_KEY")),
		Model:    strings.TrimSpace(os.Getenv("SINGEROS_LLM_MODEL")),
		BaseURL:  strings.TrimSpace(os.Getenv("SINGEROS_LLM_BASE_URL")),
	}
	if llmConfig.APIKey == "" {
		llmConfig.APIKey = strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY"))
	}
	if llmConfig.APIKey == "" {
		llmConfig.APIKey = strings.TrimSpace(os.Getenv("ANTHROPIC_AUTH_TOKEN"))
	}
	if llmConfig.Provider == "" && llmConfig.APIKey != "" {
		llmConfig.Provider = "anthropic"
	}
	if llmConfig.APIKey == "" {
		t.Log("no model API key found in env; relying on existing claude CLI authentication")
	}

	return llmConfig
}

func findRepoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found while searching for repository root")
		}
		dir = parent
	}
}

func logProcessOutput(t *testing.T, result *agent.RunResult) {
	t.Helper()
	if result == nil || result.Metadata == nil {
		return
	}
	logPath, _ := result.Metadata["log_path"].(string)
	if strings.TrimSpace(logPath) == "" {
		return
	}
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Logf("read runtime process log %s: %v", logPath, err)
		return
	}
	t.Logf("runtime process log path: %s", logPath)
	t.Logf("runtime process output:\n%s", string(content))
}
