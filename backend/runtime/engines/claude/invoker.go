package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/insmtx/SingerOS/backend/runtime/engines"
)

// Invoker 启动 Claude Code 进程。
type Invoker struct {
	binary  string
	baseEnv []string
}

// NewInvoker 创建 Claude Code 调用器。
func NewInvoker(binary string, extraEnv map[string]string) *Invoker {
	return &Invoker{
		binary:  binary,
		baseEnv: engines.BuildBaseEnv(extraEnv),
	}
}

type streamEvent struct {
	Type    string         `json:"type"`
	Message *streamMessage `json:"message,omitempty"`
	Result  string         `json:"result,omitempty"`
	IsError bool           `json:"is_error,omitempty"`
}

type streamMessage struct {
	Content []streamContent `json:"content"`
}

type streamContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	Name string `json:"name,omitempty"`
}

// Run 启动 Claude Code 进程并将 stdout/stderr 直接转换为引擎事件。
func (inv *Invoker) Run(ctx context.Context, req engines.RunRequest) (engines.Process, <-chan engines.Event, error) {
	args := buildArgs(req)

	execCtx := ctx
	cancel := func() {}
	if req.Timeout > 0 {
		execCtx, cancel = context.WithTimeout(ctx, req.Timeout)
	}

	cmd := exec.CommandContext(execCtx, inv.binary, args...)
	cmd.Dir = req.WorkDir
	cmd.Stdin = strings.NewReader(req.Prompt)
	cmd.Env = engines.BuildRunEnv(inv.baseEnv, req.ExtraEnv, req.Model)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("open claude stdout: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("open claude stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, nil, fmt.Errorf("start claude: %w", err)
	}

	events := make(chan engines.Event, 16)
	proc := engines.NewCmdProcess(cmd)
	events <- engines.Event{Type: engines.EventStarted}

	go func() {
		defer close(events)
		defer cancel()

		parseState := &claudeStreamState{}
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			scanClaudeStdout(ctx, stdout, events, parseState)
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
		if parseState.isError {
			if parseState.result == "" {
				parseState.result = "claude execution failed"
			}
			events <- engines.Event{Type: engines.EventError, Content: parseState.result}
			return
		}
		if parseState.result == "" && parseState.lastAssistantText != "" {
			if !sendEvent(ctx, events, engines.Event{Type: engines.EventResult, Content: parseState.lastAssistantText}) {
				return
			}
		}
		events <- engines.Event{Type: engines.EventDone}
	}()

	return proc, events, nil
}

type claudeStreamState struct {
	result            string
	isError           bool
	lastAssistantText string
}

func scanClaudeStdout(ctx context.Context, r interface{ Read([]byte) (int, error) }, events chan<- engines.Event, state *claudeStreamState) {
	engines.ScanJSONLines(r, func(line string) bool {
		event := parseClaudeLine(line, state)
		if event.Type == "" {
			return true
		}
		return sendEvent(ctx, events, event)
	})
}

func parseClaudeLine(line string, state *claudeStreamState) engines.Event {
	line = strings.TrimSpace(line)
	if line == "" {
		return engines.Event{}
	}
	var event streamEvent
	if json.Unmarshal([]byte(line), &event) != nil {
		return engines.Event{Type: engines.EventMessageDelta, Content: line}
	}
	switch event.Type {
	case "assistant":
		if event.Message == nil {
			return engines.Event{}
		}
		var b strings.Builder
		for _, block := range event.Message.Content {
			switch block.Type {
			case "text":
				if block.Text != "" {
					state.lastAssistantText = block.Text
					b.WriteString(block.Text)
				}
			case "tool_use":
				if block.Name != "" {
					b.WriteString("[调用工具: ")
					b.WriteString(block.Name)
					b.WriteString("]")
				}
			}
		}
		if b.Len() == 0 {
			return engines.Event{}
		}
		return engines.Event{Type: engines.EventMessageDelta, Content: b.String()}
	case "result":
		state.result = event.Result
		state.isError = event.IsError
		if event.IsError || event.Result == "" {
			return engines.Event{}
		}
		return engines.Event{Type: engines.EventResult, Content: event.Result}
	}
	return engines.Event{}
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

func buildArgs(req engines.RunRequest) []string {
	args := []string{
		"--dangerously-skip-permissions",
		"--verbose",
		"--output-format", "stream-json",
	}
	if req.Model.Model != "" {
		args = append(args, "--model", req.Model.Model)
	}
	if req.SessionID != "" {
		if req.Resume {
			args = append(args, "--resume", req.SessionID)
		} else {
			args = append(args, "--session-id", req.SessionID)
		}
	}
	return append(args, "--print")
}
