package codex

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/insmtx/SingerOS/backend/runtime/engines"
)

// Invoker 启动 Codex CLI 进程。
type Invoker struct {
	binary  string        // codex 可执行文件路径
	baseEnv []string      // 基础环境变量
	store   *SessionStore // 会话存储，用于恢复对话
}

// NewInvoker 创建 Codex CLI 调用器。
func NewInvoker(binary string, store *SessionStore, extraEnv map[string]string) *Invoker {
	if store == nil {
		store = NewSessionStore()
	}
	return &Invoker{
		binary:  binary,
		baseEnv: engines.BuildBaseEnv(extraEnv),
		store:   store,
	}
}

type codexEvent struct {
	Type     string     `json:"type"`
	ThreadID string     `json:"thread_id,omitempty"`
	Item     *codexItem `json:"item,omitempty"`
}

type codexItem struct {
	Type        string          `json:"type"`
	Text        json.RawMessage `json:"text,omitempty"`
	Command     string          `json:"command,omitempty"`
	CommandLine string          `json:"command_line,omitempty"`
	Name        string          `json:"name,omitempty"`
	Output      string          `json:"output,omitempty"`
}

// Run 启动 Codex CLI 进程并将 stdout/stderr 直接转换为引擎事件。
func (inv *Invoker) Run(ctx context.Context, req engines.RunRequest) (engines.Process, <-chan engines.Event, error) {
	threadID, resume := inv.resolveThread(req.SessionID, req.Resume)
	args := buildArgs(threadID, resume, req)

	execCtx := ctx
	cancel := func() {}
	if req.Timeout > 0 {
		execCtx, cancel = context.WithTimeout(ctx, req.Timeout)
	}

	cmd := exec.CommandContext(execCtx, inv.binary, args...)
	cmd.Dir = req.WorkDir
	cmd.Env = engines.BuildRunEnv(inv.baseEnv, req.ExtraEnv, req.Model)
	if !resume {
		cmd.Stdin = strings.NewReader(req.Prompt)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("open codex stdout: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("open codex stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, nil, fmt.Errorf("start codex: %w", err)
	}

	events := make(chan engines.Event, 16)
	proc := engines.NewCmdProcess(cmd)
	events <- engines.Event{Type: engines.EventStarted}

	go func() {
		defer close(events)
		defer cancel()

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			inv.scanStdout(ctx, stdout, events, req.SessionID, !resume)
		}()
		go func() {
			defer wg.Done()
			scanPlainOutput(ctx, stderr, events, engines.EventMessageDelta)
		}()

		err := cmd.Wait()
		wg.Wait()
		if err != nil {
			events <- engines.Event{Type: engines.EventError, Content: err.Error()}
			return
		}
		events <- engines.Event{Type: engines.EventDone}
	}()

	return proc, events, nil
}

func (inv *Invoker) scanStdout(ctx context.Context, r interface{ Read([]byte) (int, error) }, events chan<- engines.Event, sessionID string, captureSession bool) {
	engines.ScanJSONLines(r, func(line string) bool {
		event, threadID := parseCodexLine(line)
		if captureSession && sessionID != "" && threadID != "" {
			inv.store.Set(sessionID, threadID)
		}
		if event.Type == "" {
			return true
		}
		return sendEvent(ctx, events, event)
	})
}

func parseCodexLine(line string) (engines.Event, string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return engines.Event{}, ""
	}
	var event codexEvent
	if json.Unmarshal([]byte(line), &event) != nil {
		return engines.Event{Type: engines.EventMessageDelta, Content: line}, ""
	}
	if event.Type == "thread.started" && event.ThreadID != "" {
		return engines.Event{}, event.ThreadID
	}
	if event.Item == nil {
		return engines.Event{}, ""
	}

	item := event.Item
	switch item.Type {
	case "agent_message":
		text := decodeCodexText(item.Text)
		if text == "" {
			return engines.Event{}, ""
		}
		eventType := engines.EventMessageDelta
		if event.Type == "item.completed" {
			eventType = engines.EventResult
		}
		return engines.Event{Type: eventType, Content: text}, ""
	case "command_execution", "tool_call", "shell_command":
		command := firstNonEmptyString(item.Command, item.CommandLine, item.Name)
		if command != "" {
			return engines.Event{Type: engines.EventMessageDelta, Content: "$ " + command}, ""
		}
	case "command_output", "tool_output", "shell_output":
		output := firstNonEmptyString(item.Output, decodeCodexText(item.Text))
		if output != "" {
			return engines.Event{Type: engines.EventMessageDelta, Content: truncateOutput(output, 300)}, ""
		}
	}
	return engines.Event{}, ""
}

func decodeCodexText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text
	}
	var parts []any
	if json.Unmarshal(raw, &parts) == nil {
		var b strings.Builder
		for _, part := range parts {
			if value, ok := part.(string); ok {
				b.WriteString(value)
			}
		}
		return b.String()
	}
	return ""
}

func scanPlainOutput(ctx context.Context, r interface{ Read([]byte) (int, error) }, events chan<- engines.Event, eventType engines.EventType) {
	engines.ScanJSONLines(r, func(line string) bool {
		line = strings.TrimSpace(line)
		if line == "" {
			return true
		}
		return sendEvent(ctx, events, engines.Event{Type: eventType, Content: line})
	})
}

func sendEvent(ctx context.Context, events chan<- engines.Event, event engines.Event) bool {
	select {
	case <-ctx.Done():
		return false
	case events <- event:
		return true
	}
}

func truncateOutput(value string, maxLen int) string {
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen] + "..."
}

func buildArgs(threadID string, resume bool, req engines.RunRequest) []string {
	args := []string{"exec"}
	args = append(args, singerProviderConfigArgs(req)...)
	if req.Model.Model != "" {
		args = append(args, "--model", req.Model.Model)
	}
	if resume && threadID != "" {
		args = append(args, "resume", threadID, "--json", "--skip-git-repo-check", "--dangerously-bypass-approvals-and-sandbox")
		if req.Prompt != "" {
			args = append(args, req.Prompt)
		}
		return args
	}
	return append(args, "-", "--json", "--skip-git-repo-check", "--dangerously-bypass-approvals-and-sandbox")
}

func singerProviderConfigArgs(req engines.RunRequest) []string {
	baseURL := firstNonEmptyString(
		req.Model.BaseURL,
		envValue(req.ExtraEnv, "OPENAI_API_BASE"),
		envValue(req.ExtraEnv, "OPENAI_BASE_URL"),
		os.Getenv("OPENAI_API_BASE"),
		os.Getenv("OPENAI_BASE_URL"),
	)
	return []string{
		"-c", `model_provider="singer"`,
		"-c", `model_providers.singer.name="singer"`,
		"-c", fmt.Sprintf(`model_providers.singer.base_url=%q`, baseURL),
		"-c", `model_providers.singer.env_key="OPENAI_API_KEY"`,
		"-c", `model_providers.singer.wire_api="chat"`,
	}
}

func envValue(entries []string, key string) string {
	prefix := key + "="
	for _, entry := range entries {
		if strings.HasPrefix(entry, prefix) {
			return strings.TrimPrefix(entry, prefix)
		}
	}
	return ""
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func (inv *Invoker) resolveThread(sessionID string, resume bool) (string, bool) {
	if !resume {
		return "", false
	}
	threadID, ok := inv.store.Get(sessionID)
	return threadID, ok
}
