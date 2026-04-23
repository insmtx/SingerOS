package nodetools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/insmtx/SingerOS/backend/tools"
)

// NodeShellTool executes shell commands in a node container.
type NodeShellTool struct {
	tools.BaseTool
	executor nodeExecutor
}

// NewNodeShellTool creates a Docker-backed node shell tool.
func NewNodeShellTool() *NodeShellTool {
	return newNodeShellToolWithExecutor(dockerCLIExecutor{})
}

func newNodeShellToolWithExecutor(executor nodeExecutor) *NodeShellTool {
	return &NodeShellTool{
		BaseTool: tools.NewBaseTool(
			ToolNameNodeShell,
			"Execute a shell command inside an assistant node Docker container",
			tools.Schema{
				Type:     "object",
				Required: []string{"command"},
				Properties: map[string]*tools.Property{
					"command": {
						Type:        "string",
						Description: "Shell command to execute",
					},
					"working_dir": {
						Type:        "string",
						Description: "Working directory inside the container; defaults to /workspace",
					},
					"timeout": {
						Type:        "integer",
						Description: "Timeout in seconds; defaults to 120 and is clamped to 5-600",
					},
				},
			},
		),
		executor: executor,
	}
}

// Validate checks node shell tool input.
func (t *NodeShellTool) Validate(input map[string]interface{}) error {
	if input == nil {
		return fmt.Errorf("input is required")
	}
	if stringValue(input, "command") == "" {
		return fmt.Errorf("command is required")
	}
	if _, err := intValue(input["timeout"]); err != nil {
		return fmt.Errorf("timeout must be an integer")
	}
	if workingDir, ok := input["working_dir"].(string); ok && strings.TrimSpace(workingDir) == "" {
		return fmt.Errorf("working_dir must be a non-empty string")
	}
	return nil
}

// Execute runs the shell command inside the target node container.
func (t *NodeShellTool) Execute(ctx context.Context, input map[string]interface{}) (string, error) {
	if t.executor == nil {
		return "", fmt.Errorf("node executor is required")
	}

	toolCtx, err := tools.RequireToolContext(ctx)
	if err != nil {
		return "", err
	}
	nodeInfo, err := nodeInfoForAssistant(toolCtx)
	if err != nil {
		return "", err
	}
	containerID := nodeInfo.ContainerID

	command := stringValue(input, "command")
	workingDir := stringValue(input, "working_dir")
	if workingDir == "" {
		workingDir = defaultWorkingDir
	}

	timeoutSeconds, _ := intValue(input["timeout"])
	if timeoutSeconds == 0 {
		timeoutSeconds = defaultShellTimeout
	}
	timeoutSeconds = clampInt(timeoutSeconds, minShellTimeout, maxShellTimeout)

	shellCommand := fmt.Sprintf("cd %s && %s", shellQuote(workingDir), command)
	result, err := t.executor.Exec(ctx, nodeExecRequest{
		ContainerID: containerID,
		Args:        []string{"sh", "-c", shellCommand},
		Timeout:     time.Duration(timeoutSeconds) * time.Second,
	})
	if err != nil {
		return "", fmt.Errorf("execute node shell command: %w", err)
	}
	if result.TimedOut {
		return tools.JSONString(map[string]interface{}{
			"container_id": containerID,
			"command":      command,
			"working_dir":  workingDir,
			"timeout":      timeoutSeconds,
			"timed_out":    true,
			"message":      fmt.Sprintf("command timed out after %ds", timeoutSeconds),
		})
	}

	combined := combineOutput(result.Stdout, result.Stderr)
	output, truncated, totalLines := truncateOutput(combined, defaultOutputMaxLines)
	display := fmt.Sprintf("[exit_code=%d]", result.ExitCode)
	if output != "" {
		display += "\n" + output
	}

	return tools.JSONString(map[string]interface{}{
		"container_id": containerID,
		"command":      command,
		"working_dir":  workingDir,
		"timeout":      timeoutSeconds,
		"exit_code":    result.ExitCode,
		"stdout":       result.Stdout,
		"stderr":       result.Stderr,
		"output":       output,
		"display":      display,
		"truncated":    truncated,
		"total_lines":  totalLines,
	})
}
