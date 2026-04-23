package nodetools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/insmtx/SingerOS/backend/tools"
)

type fakeNodeExecutor struct {
	calls   []nodeExecRequest
	results []nodeExecResult
	err     error
}

func (e *fakeNodeExecutor) Exec(ctx context.Context, req nodeExecRequest) (nodeExecResult, error) {
	e.calls = append(e.calls, req)
	if e.err != nil {
		return nodeExecResult{}, e.err
	}
	if len(e.results) == 0 {
		return nodeExecResult{}, nil
	}
	result := e.results[0]
	e.results = e.results[1:]
	return result, nil
}

func TestNodeShellToolExecute(t *testing.T) {
	executor := &fakeNodeExecutor{
		results: []nodeExecResult{{
			Stdout:   "ok\n",
			ExitCode: 0,
		}},
	}
	tool := newNodeShellToolWithExecutor(executor)

	rawOutput, err := tool.Execute(testNodeToolContext(), map[string]interface{}{
		"command":     "pwd",
		"working_dir": "/workspace/repo",
		"timeout":     1,
	})
	if err != nil {
		t.Fatalf("execute node shell tool: %v", err)
	}
	output := decodeNodeToolOutput(t, rawOutput)

	if output["exit_code"] != float64(0) {
		t.Fatalf("expected exit code 0, got %#v", output["exit_code"])
	}
	if output["timeout"] != float64(minShellTimeout) {
		t.Fatalf("expected clamped timeout %d, got %#v", minShellTimeout, output["timeout"])
	}
	if len(executor.calls) != 1 {
		t.Fatalf("expected 1 executor call, got %d", len(executor.calls))
	}
	call := executor.calls[0]
	if len(call.Args) != 3 || call.Args[0] != "sh" || call.Args[1] != "-c" {
		t.Fatalf("unexpected command args: %#v", call.Args)
	}
	if !strings.Contains(call.Args[2], "cd '/workspace/repo' && pwd") {
		t.Fatalf("unexpected shell command: %s", call.Args[2])
	}
}

func TestNodeFileReadToolExecute(t *testing.T) {
	executor := &fakeNodeExecutor{
		results: []nodeExecResult{
			{Stdout: "EXISTS\n", ExitCode: 0},
			{Stdout: "alpha\nbeta\n", ExitCode: 0},
			{Stdout: "10\n", ExitCode: 0},
		},
	}
	tool := newNodeFileReadToolWithExecutor(executor)

	rawOutput, err := tool.Execute(testNodeToolContext(), map[string]interface{}{
		"path":   "/workspace/app/main.go",
		"offset": 3,
		"limit":  2,
	})
	if err != nil {
		t.Fatalf("execute node file read tool: %v", err)
	}
	output := decodeNodeToolOutput(t, rawOutput)

	if output["content"] != "alpha\nbeta" {
		t.Fatalf("unexpected content: %#v", output["content"])
	}
	if output["shown_start"] != float64(3) || output["shown_end"] != float64(4) {
		t.Fatalf("unexpected shown range: %v-%v", output["shown_start"], output["shown_end"])
	}
	numbered := output["numbered_content"].(string)
	if !strings.Contains(numbered, "     3|alpha") || !strings.Contains(numbered, "     4|beta") {
		t.Fatalf("unexpected numbered content: %s", numbered)
	}
	if !output["has_more"].(bool) {
		t.Fatalf("expected has_more to be true")
	}
	if len(executor.calls) != 3 {
		t.Fatalf("expected 3 executor calls, got %d", len(executor.calls))
	}
}

func TestNodeFileWriteToolExecute(t *testing.T) {
	executor := &fakeNodeExecutor{
		results: []nodeExecResult{
			{ExitCode: 0},
			{ExitCode: 0},
		},
	}
	tool := newNodeFileWriteToolWithExecutor(executor)

	rawOutput, err := tool.Execute(testNodeToolContext(), map[string]interface{}{
		"path":    "/workspace/app/main.go",
		"content": "package main\n",
		"append":  true,
	})
	if err != nil {
		t.Fatalf("execute node file write tool: %v", err)
	}
	output := decodeNodeToolOutput(t, rawOutput)

	if output["action"] != "appended" {
		t.Fatalf("unexpected action: %#v", output["action"])
	}
	if output["line_count"] != float64(1) {
		t.Fatalf("unexpected line count: %#v", output["line_count"])
	}
	if len(executor.calls) != 2 {
		t.Fatalf("expected mkdir and write calls, got %d", len(executor.calls))
	}
	if executor.calls[1].Stdin == nil || *executor.calls[1].Stdin != "package main\n" {
		t.Fatalf("unexpected stdin: %#v", executor.calls[1].Stdin)
	}
	if !strings.Contains(executor.calls[1].Args[2], "tee -a '/workspace/app/main.go'") {
		t.Fatalf("unexpected tee command: %s", executor.calls[1].Args[2])
	}
}

func TestNodeToolValidateDoesNotRequireContainerID(t *testing.T) {
	if err := newNodeShellToolWithExecutor(nil).Validate(map[string]interface{}{
		"command": "pwd",
	}); err != nil {
		t.Fatalf("shell validate should not require container_id: %v", err)
	}
	if err := newNodeFileReadToolWithExecutor(nil).Validate(map[string]interface{}{
		"path": "/workspace/app/main.go",
	}); err != nil {
		t.Fatalf("file read validate should not require container_id: %v", err)
	}
	if err := newNodeFileWriteToolWithExecutor(nil).Validate(map[string]interface{}{
		"path":    "/workspace/app/main.go",
		"content": "package main\n",
	}); err != nil {
		t.Fatalf("file write validate should not require container_id: %v", err)
	}
}

func testNodeToolContext() context.Context {
	return tools.ContextWithToolContext(context.Background(), tools.ToolContext{
		AssistantID: "assistant-1",
	})
}

func decodeNodeToolOutput(t *testing.T, output string) map[string]interface{} {
	t.Helper()

	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("decode node tool output: %v\n%s", err, output)
	}
	return decoded
}
