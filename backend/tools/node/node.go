// Package nodetools provides Docker-backed tools for operating an assistant node.
package nodetools

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/insmtx/SingerOS/backend/tools"
)

const (
	// ProviderNode identifies tools that operate against an assistant work node.
	ProviderNode = "node"

	// ToolNameNodeShell executes shell commands inside a node container.
	ToolNameNodeShell = "node_shell"
	// ToolNameNodeFileRead reads files from a node container.
	ToolNameNodeFileRead = "node_file_read"
	// ToolNameNodeFileWrite writes files into a node container.
	ToolNameNodeFileWrite = "node_file_write"
)

const (
	defaultWorkingDir     = "/workspace"
	defaultShellTimeout   = 120
	minShellTimeout       = 5
	maxShellTimeout       = 600
	defaultReadLimit      = 200
	maxReadLimit          = 2000
	defaultOutputMaxLines = 50
)

type nodeExecutor interface {
	Exec(ctx context.Context, req nodeExecRequest) (nodeExecResult, error)
}

type nodeExecRequest struct {
	ContainerID string
	Args        []string
	Stdin       *string
	Timeout     time.Duration
}

type nodeExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	TimedOut bool
}

type dockerCLIExecutor struct{}

func (e dockerCLIExecutor) Exec(ctx context.Context, req nodeExecRequest) (nodeExecResult, error) {
	if strings.TrimSpace(req.ContainerID) == "" {
		return nodeExecResult{}, fmt.Errorf("container_id is required")
	}
	if len(req.Args) == 0 {
		return nodeExecResult{}, fmt.Errorf("docker exec command is required")
	}

	execCtx := ctx
	cancel := func() {}
	if req.Timeout > 0 {
		execCtx, cancel = context.WithTimeout(ctx, req.Timeout)
	}
	defer cancel()

	args := []string{"exec"}
	if req.Stdin != nil {
		args = append(args, "-i")
	}
	args = append(args, req.ContainerID)
	args = append(args, req.Args...)

	cmd := exec.CommandContext(execCtx, "docker", args...)
	if req.Stdin != nil {
		cmd.Stdin = strings.NewReader(*req.Stdin)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := nodeExecResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if ctx.Err() != nil {
		return result, ctx.Err()
	}
	if errors.Is(execCtx.Err(), context.DeadlineExceeded) {
		result.ExitCode = -1
		result.TimedOut = true
		return result, nil
	}
	if err == nil {
		return result, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		result.ExitCode = exitErr.ExitCode()
		return result, nil
	}

	return result, err
}

// NewTools returns all Docker node tools for registration.
func NewTools() []tools.Tool {
	return []tools.Tool{
		NewNodeShellTool(),
		NewNodeFileReadTool(),
		NewNodeFileWriteTool(),
	}
}

// Register registers all Docker node tools into the provided registry.
func Register(registry *tools.Registry) error {
	if registry == nil {
		return fmt.Errorf("tool registry is required")
	}

	for _, tool := range NewTools() {
		if err := registry.Register(tool); err != nil {
			return err
		}
	}

	return nil
}

func stringValue(input map[string]interface{}, key string) string {
	value, _ := input[key].(string)
	return strings.TrimSpace(value)
}

func intValue(value interface{}) (int, error) {
	switch typed := value.(type) {
	case nil:
		return 0, nil
	case int:
		return typed, nil
	case int32:
		return int(typed), nil
	case int64:
		return int(typed), nil
	case float64:
		return int(typed), nil
	default:
		return 0, fmt.Errorf("invalid integer value")
	}
}

func boolValue(value interface{}) (bool, error) {
	switch typed := value.(type) {
	case nil:
		return false, nil
	case bool:
		return typed, nil
	default:
		return false, fmt.Errorf("invalid boolean value")
	}
}

func clampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func truncateOutput(output string, maxLines int) (string, bool, int) {
	output = strings.TrimSpace(output)
	if output == "" {
		return "", false, 0
	}

	lines := strings.Split(output, "\n")
	if len(lines) <= maxLines {
		return output, false, len(lines)
	}

	return fmt.Sprintf("[输出共 %d 行，显示最后 %d 行]\n%s", len(lines), maxLines, strings.Join(lines[len(lines)-maxLines:], "\n")), true, len(lines)
}

func combineOutput(stdout string, stderr string) string {
	stdout = strings.TrimSpace(stdout)
	stderr = strings.TrimSpace(stderr)
	switch {
	case stdout != "" && stderr != "":
		return stdout + "\n" + stderr
	case stdout != "":
		return stdout
	default:
		return stderr
	}
}

func parentDir(path string) string {
	index := strings.LastIndex(path, "/")
	if index <= 0 {
		return ""
	}
	return path[:index]
}

func countContentLines(content string) int {
	if content == "" {
		return 1
	}

	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	trimmed := strings.TrimSuffix(normalized, "\n")
	if trimmed == "" {
		return 1
	}

	return len(strings.Split(trimmed, "\n"))
}
