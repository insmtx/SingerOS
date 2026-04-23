package runtime

import (
	"context"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/insmtx/SingerOS/backend/config"
	runtimeevents "github.com/insmtx/SingerOS/backend/runtime/events"
	"github.com/insmtx/SingerOS/backend/tools"
	nodetools "github.com/insmtx/SingerOS/backend/tools/node"
	skilltools "github.com/insmtx/SingerOS/backend/tools/skill"
	"github.com/ygpkg/yg-go/logs"
	"go.uber.org/zap/zapcore"
)

const defaultTestNodeContainerID = "b327e241316c2a2f62cbee986edd0e71235205f0fde5dc7a4543f5344396b351"

func TestAgentBuildSystemPromptIncludesSkills(t *testing.T) {
	catalog, err := skilltools.NewCatalog(fstest.MapFS{
		"code-review/SKILL.md": {
			Data: []byte(`---
name: code-review
description: Review code.
metadata:
  singeros:
    always: true
---
Always inspect diffs first.`),
		},
	})
	if err != nil {
		t.Fatalf("new skills catalog: %v", err)
	}

	agent := &Agent{
		systemPrompt:  "Base runtime prompt.",
		skillsCatalog: catalog,
	}

	prompt, err := agent.buildSystemPrompt(&RequestContext{
		Assistant: AssistantContext{
			SystemPrompt: "Assistant-specific prompt.",
		},
	})
	if err != nil {
		t.Fatalf("build system prompt: %v", err)
	}

	for _, expected := range []string{
		"Base runtime prompt.",
		"Assistant-specific prompt.",
		"Available skills:",
		"## Skill: code-review",
	} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("expected prompt to contain %q, got %s", expected, prompt)
		}
	}
	if strings.Contains(prompt, "Available tools:") {
		t.Fatalf("tool summary should not be injected into system prompt: %s", prompt)
	}
}

func TestAgentRunRealModel(t *testing.T) {
	logs.SetLevel(zapcore.DebugLevel)

	apiKey := firstNonEmptyEnv("SINGEROS_LLM_API_KEY")
	if apiKey == "" {
		t.Skip("set SINGEROS_LLM_API_KEY to run the real model agent test")
	}

	ctx, cancel := realModelTestContext(t)
	defer cancel()

	registry := tools.NewRegistry()
	agent, err := NewAgent(ctx, &config.LLMConfig{
		Provider: "openai",
		APIKey:   apiKey,
		Model:    firstNonEmptyEnv("SINGEROS_LLM_MODEL"),
		BaseURL:  firstNonEmptyEnv("SINGEROS_LLM_BASE_URL"),
	}, Config{
		ToolRegistry: registry,
	})
	if err != nil {
		t.Fatalf("new agent: %v", err)
	}

	result, err := agent.Run(ctx, &RequestContext{
		RunID: "run_real_model_message",
		Actor: ActorContext{
			UserID:  "test-user",
			Channel: "test",
		},
		Input: InputContext{
			Type: InputTypeMessage,
			Text: "Reply with exactly this text: SingerOS agent runtime ok",
		},
		Runtime:   RuntimeOptions{MaxStep: 2},
		EventSink: runtimeevents.NewLogSink(),
	})
	if err != nil {
		t.Fatalf("run agent: %v", err)
	}
	if result == nil {
		t.Fatalf("expected result")
	}
	if result.Status != RunStatusCompleted {
		t.Fatalf("expected completed result, got %+v", result)
	}
	if strings.TrimSpace(result.Message) == "" {
		t.Fatalf("expected non-empty model response")
	}
	if !strings.Contains(result.Message, "SingerOS agent runtime ok") {
		t.Fatalf("unexpected model response: %s", result.Message)
	}
}

func TestAgentRunNodeTool(t *testing.T) {
	logs.SetLevel(zapcore.DebugLevel)

	apiKey := firstNonEmptyEnv("SINGEROS_LLM_API_KEY")
	if apiKey == "" {
		t.Skip("set SINGEROS_LLM_API_KEY to run the real model agent tool-call test")
	}

	ctx, cancel := realModelTestContext(t)
	defer cancel()
	containerID := realModelNodeContainerID()

	registry := tools.NewRegistry()
	if err := nodetools.Register(registry); err != nil {
		t.Fatalf("register node tools: %v", err)
	}

	agent, err := NewAgent(ctx, &config.LLMConfig{
		Provider: "openai",
		APIKey:   apiKey,
		Model:    firstNonEmptyEnv("SINGEROS_LLM_MODEL"),
		BaseURL:  firstNonEmptyEnv("SINGEROS_LLM_BASE_URL"),
	}, Config{
		ToolRegistry: registry,
	})
	if err != nil {
		t.Fatalf("new agent: %v", err)
	}

	sink := &recordingEventSink{}
	result, err := agent.Run(ctx, &RequestContext{
		RunID: "run_real_model_node_shell_time",
		Assistant: AssistantContext{
			ID:   "test-assistant",
			Name: "Tool Test Assistant",
			SystemPrompt: strings.Join([]string{
				"你必须使用工具完成用户任务，不能凭空回答。",
				"node_shell 的 container_id 必须使用 " + containerID + "。",
			}, "\n"),
		},
		Actor: ActorContext{
			UserID:  "test-user",
			Channel: "test",
		},
		Input: InputContext{
			Type: InputTypeMessage,
			Text: "使用工具查询当前系统时间。",
		},
		Runtime: RuntimeOptions{MaxStep: 6},
		Capability: CapabilityContext{
			AllowedTools: []string{
				nodetools.ToolNameNodeShell,
				nodetools.ToolNameNodeFileRead,
				nodetools.ToolNameNodeFileWrite,
			},
		},
		EventSink: sink,
	})
	if err != nil {
		t.Fatalf("run agent: %v", err)
	}
	if result == nil {
		t.Fatalf("expected result")
	}
	if result.Status != RunStatusCompleted {
		t.Fatalf("expected completed result, got %+v", result)
	}
	if strings.TrimSpace(result.Message) == "" {
		t.Fatalf("expected non-empty model response")
	}

	toolEvent := sink.firstToolEvent(runtimeevents.RunEventToolCallCompleted, nodetools.ToolNameNodeShell)
	if toolEvent == nil {
		t.Fatalf("expected completed %s tool call, events=%s", nodetools.ToolNameNodeShell, sink.eventSummary())
	}
	if !strings.Contains(toolEvent.Content, nodetools.ToolNameNodeShell) {
		t.Fatalf("expected %s tool event content, got %s", nodetools.ToolNameNodeShell, toolEvent.Content)
	}
	if !strings.Contains(toolEvent.Content, "[exit_code=0]") {
		t.Fatalf("expected %s content to contain exit_code=0, got %s", nodetools.ToolNameNodeShell, toolEvent.Content)
	}
}

func TestAgentRunWeatherSkillQuery(t *testing.T) {
	logs.SetLevel(zapcore.DebugLevel)

	apiKey := firstNonEmptyEnv("SINGEROS_LLM_API_KEY")
	if apiKey == "" {
		t.Skip("set SINGEROS_LLM_API_KEY to run the real model agent weather skill test")
	}

	ctx, cancel := realModelTestContext(t)
	defer cancel()
	containerID := realModelNodeContainerID()

	catalog, skillDir := newBundledRuntimeSkillsCatalog(t)
	if _, err := catalog.Get("weather"); err != nil {
		t.Fatalf("weather skill must be available in %s: %v", skillDir, err)
	}

	registry := tools.NewRegistry()
	if err := skilltools.Register(registry, catalog); err != nil {
		t.Fatalf("register skill tools: %v", err)
	}
	if err := nodetools.Register(registry); err != nil {
		t.Fatalf("register node tools: %v", err)
	}

	agent, err := NewAgent(ctx, &config.LLMConfig{
		Provider: "openai",
		APIKey:   apiKey,
		Model:    firstNonEmptyEnv("SINGEROS_LLM_MODEL"),
		BaseURL:  firstNonEmptyEnv("SINGEROS_LLM_BASE_URL"),
	}, Config{
		SkillsCatalog: catalog,
		ToolRegistry:  registry,
	})
	if err != nil {
		t.Fatalf("new agent: %v", err)
	}

	sink := &recordingEventSink{}
	result, err := agent.Run(ctx, &RequestContext{
		RunID: "run_real_model_weather_skill_shanghai",
		Assistant: AssistantContext{
			ID:   "test-weather-assistant",
			Name: "Weather Skill Test Assistant",
			SystemPrompt: strings.Join([]string{
				"你必须使用工具完成用户任务，不能凭空回答。",
				"node_shell 的 container_id 必须使用 " + containerID + "。",
			}, "\n"),
		},
		Actor: ActorContext{
			UserID:  "test-user",
			Channel: "test",
		},
		Input: InputContext{
			Type: InputTypeTaskInstruction,
			Text: "使用 weather 这个 skill 来查询上海的天气。",
		},
		Runtime: RuntimeOptions{MaxStep: 20},
		Capability: CapabilityContext{
			AllowedTools: []string{
				skilltools.ToolNameSkillUse,
				nodetools.ToolNameNodeShell,
			},
		},
		EventSink: sink,
	})
	if err != nil {
		t.Fatalf("run weather skill agent: %v", err)
	}
	if result == nil {
		t.Fatalf("expected result")
	}
	if result.Status != RunStatusCompleted {
		t.Fatalf("expected completed result, got %+v", result)
	}
	if strings.TrimSpace(result.Message) == "" {
		t.Fatalf("expected non-empty model response")
	}

	skillEvent := sink.firstToolEvent(runtimeevents.RunEventToolCallCompleted, skilltools.ToolNameSkillUse)
	if skillEvent == nil {
		t.Fatalf("expected completed %s tool call, events=%s", skilltools.ToolNameSkillUse, sink.eventSummary())
	}
	if !strings.Contains(skillEvent.Content, `"name":"weather"`) {
		t.Fatalf("expected %s output to load weather skill, got %s", skilltools.ToolNameSkillUse, skillEvent.Content)
	}

	shellEvent := sink.firstToolEvent(runtimeevents.RunEventToolCallCompleted, nodetools.ToolNameNodeShell)
	if shellEvent == nil {
		t.Fatalf("expected completed %s tool call, events=%s", nodetools.ToolNameNodeShell, sink.eventSummary())
	}
	if !strings.Contains(shellEvent.Content, "[exit_code=0]") {
		t.Fatalf("expected %s content to contain exit_code=0, got %s", nodetools.ToolNameNodeShell, shellEvent.Content)
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

func realModelNodeContainerID() string {
	if containerID := firstNonEmptyEnv("SINGEROS_TEST_NODE_CONTAINER_ID"); containerID != "" {
		return containerID
	}
	return defaultTestNodeContainerID
}

func newBundledRuntimeSkillsCatalog(t *testing.T) (*skilltools.Catalog, string) {
	t.Helper()

	_, currentFile, _, ok := goruntime.Caller(0)
	if !ok {
		t.Fatalf("resolve current test file")
	}

	skillsDir := filepath.Join(filepath.Dir(currentFile), "..", "skills")
	catalog, err := skilltools.NewCatalog(os.DirFS(skillsDir))
	if err != nil {
		t.Fatalf("load bundled skills catalog from %s: %v", skillsDir, err)
	}

	return catalog, skillsDir
}

func realModelTestContext(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()

	timeoutValue := strings.TrimSpace(os.Getenv("SINGEROS_TEST_TIMEOUT"))
	if timeoutValue == "" {
		timeoutValue = "3m"
	}
	if timeoutValue == "0" || strings.EqualFold(timeoutValue, "none") {
		return context.Background(), func() {}
	}

	timeout, err := time.ParseDuration(timeoutValue)
	if err != nil {
		t.Fatalf("parse SINGEROS_TEST_TIMEOUT: %v", err)
	}
	return context.WithTimeout(context.Background(), timeout)
}

type recordingEventSink struct {
	mu     sync.Mutex
	events []*runtimeevents.RunEvent
}

func (s *recordingEventSink) Emit(ctx context.Context, event *runtimeevents.RunEvent) error {
	if event == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	copied := *event
	logs.DebugContextf(ctx, "recordingEventSink event: type=%s run_id=%s seq=%d content=%s",
		copied.Type, copied.RunID, copied.Seq, copied.Content)
	s.events = append(s.events, &copied)
	return nil
}

func (s *recordingEventSink) firstToolEvent(eventType runtimeevents.RunEventType, toolName string) *runtimeevents.RunEvent {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, event := range s.events {
		if event == nil || event.Type != eventType {
			continue
		}
		if strings.Contains(event.Content, toolName) {
			return event
		}
	}
	return nil
}

func (s *recordingEventSink) eventSummary() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	parts := make([]string, 0, len(s.events))
	for _, event := range s.events {
		if event == nil {
			continue
		}
		if event.Content != "" {
			parts = append(parts, string(event.Type)+":"+event.Content)
			continue
		}
		parts = append(parts, string(event.Type))
	}
	return strings.Join(parts, ", ")
}
