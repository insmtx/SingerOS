package codex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

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
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// ExtractResultFromLog 从 JSON 日志中返回最后的 Codex 代理消息。
func ExtractResultFromLog(logPath string) string {
	f, err := os.Open(logPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	var lastText string
	engines.ScanJSONLines(f, func(line string) bool {
		var event codexEvent
		if json.Unmarshal([]byte(line), &event) != nil {
			return true
		}
		if event.Type == "item.completed" && event.Item != nil &&
			event.Item.Type == "agent_message" && event.Item.Text != "" {
			lastText = event.Item.Text
		}
		return true
	})
	return lastText
}

// Run 启动 Codex CLI 进程并将 stdout/stderr 写入 req.LogPath。
func (inv *Invoker) Run(ctx context.Context, req engines.RunRequest) (engines.Process, <-chan engines.Event, error) {
	if req.LogPath == "" {
		return nil, nil, fmt.Errorf("log path is required")
	}
	threadID, resume := inv.resolveThread(req.SessionID, req.Resume)
	args := buildArgs(threadID, resume, req)

	logFile, err := os.OpenFile(req.LogPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("open log file: %w", err)
	}

	pr, pw := io.Pipe()
	writer := io.MultiWriter(logFile, pw)

	execCtx := ctx
	cancel := func() {}
	if req.Timeout > 0 {
		execCtx, cancel = context.WithTimeout(ctx, req.Timeout)
	}

	cmd := exec.CommandContext(execCtx, inv.binary, args...)
	cmd.Dir = req.WorkDir
	cmd.Stdout = writer
	cmd.Stderr = logFile
	cmd.Env = engines.BuildRunEnv(inv.baseEnv, req.ExtraEnv, req.Model)
	if !resume {
		cmd.Stdin = strings.NewReader(req.Prompt)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		_ = logFile.Close()
		_ = pr.Close()
		_ = pw.Close()
		return nil, nil, fmt.Errorf("start codex: %w", err)
	}

	events := make(chan engines.Event, 2)
	proc := engines.NewCmdProcess(cmd)
	events <- engines.Event{Type: engines.EventStarted}

	go func() {
		defer close(events)
		defer logFile.Close()
		defer cancel()

		waitCh := make(chan error, 1)
		go func() {
			waitCh <- cmd.Wait()
			_ = pw.Close()
		}()

		if !resume && req.SessionID != "" {
			if newThreadID := extractSessionID(pr); newThreadID != "" {
				inv.store.Set(req.SessionID, newThreadID)
			}
		}
		_, _ = io.Copy(io.Discard, pr)

		if err := <-waitCh; err != nil {
			events <- engines.Event{Type: engines.EventError, Content: err.Error()}
			return
		}
		events <- engines.Event{Type: engines.EventDone}
	}()

	return proc, events, nil
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

func extractSessionID(r io.Reader) string {
	var threadID string
	engines.ScanJSONLines(r, func(line string) bool {
		var event codexEvent
		if json.Unmarshal([]byte(line), &event) != nil {
			return true
		}
		if event.Type == "thread.started" && event.ThreadID != "" {
			threadID = event.ThreadID
			return false
		}
		return true
	})
	return threadID
}

func (inv *Invoker) resolveThread(sessionID string, resume bool) (string, bool) {
	if !resume {
		return "", false
	}
	threadID, ok := inv.store.Get(sessionID)
	return threadID, ok
}
