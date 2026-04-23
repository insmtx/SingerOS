package nodetools

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/insmtx/SingerOS/backend/tools"
)

// NodeFileReadTool reads files from a node container.
type NodeFileReadTool struct {
	tools.BaseTool
	executor nodeExecutor
}

// NewNodeFileReadTool creates a Docker-backed node file read tool.
func NewNodeFileReadTool() *NodeFileReadTool {
	return newNodeFileReadToolWithExecutor(dockerCLIExecutor{})
}

func newNodeFileReadToolWithExecutor(executor nodeExecutor) *NodeFileReadTool {
	return &NodeFileReadTool{
		BaseTool: tools.NewBaseTool(
			ToolNameNodeFileRead,
			"Read a file from an assistant node Docker container with optional line ranges",
			tools.Schema{
				Type:     "object",
				Required: []string{"container_id", "path"},
				Properties: map[string]*tools.Property{
					"container_id": {
						Type:        "string",
						Description: "Docker container id for the assistant node",
					},
					"path": {
						Type:        "string",
						Description: "File path inside the container",
					},
					"offset": {
						Type:        "integer",
						Description: "Starting line number, beginning at 1",
					},
					"limit": {
						Type:        "integer",
						Description: "Number of lines to read; defaults to 200 and is clamped to 1-2000",
					},
				},
			},
		),
		executor: executor,
	}
}

// Validate checks node file read tool input.
func (t *NodeFileReadTool) Validate(input map[string]interface{}) error {
	if input == nil {
		return fmt.Errorf("input is required")
	}
	if stringValue(input, "container_id") == "" {
		return fmt.Errorf("container_id is required")
	}
	if stringValue(input, "path") == "" {
		return fmt.Errorf("path is required")
	}
	if offset, err := intValue(input["offset"]); err != nil {
		return fmt.Errorf("offset must be an integer")
	} else if offset < 0 {
		return fmt.Errorf("offset must be greater than or equal to 0")
	}
	if limit, err := intValue(input["limit"]); err != nil {
		return fmt.Errorf("limit must be an integer")
	} else if limit < 0 {
		return fmt.Errorf("limit must be greater than or equal to 0")
	}
	return nil
}

// Execute reads a file from the target node container.
func (t *NodeFileReadTool) Execute(ctx context.Context, input map[string]interface{}) (string, error) {
	if err := t.Validate(input); err != nil {
		return "", err
	}
	if t.executor == nil {
		return "", fmt.Errorf("node executor is required")
	}

	containerID := stringValue(input, "container_id")
	path := stringValue(input, "path")
	offset, _ := intValue(input["offset"])
	limit, _ := intValue(input["limit"])
	if limit == 0 {
		limit = defaultReadLimit
	}
	limit = clampInt(limit, 1, maxReadLimit)

	checkResult, err := t.executor.Exec(ctx, nodeExecRequest{
		ContainerID: containerID,
		Args:        []string{"sh", "-c", fmt.Sprintf("test -f %s && echo EXISTS || echo NOTFOUND", shellQuote(path))},
		Timeout:     5 * time.Second,
	})
	if err != nil {
		return "", fmt.Errorf("check node file: %w", err)
	}
	if strings.Contains(checkResult.Stdout, "NOTFOUND") {
		return tools.JSONString(map[string]interface{}{
			"container_id": containerID,
			"path":         path,
			"exists":       false,
			"message":      fmt.Sprintf("file does not exist: %s", path),
		})
	}
	if checkResult.ExitCode != 0 {
		return "", fmt.Errorf("check node file failed: %s", strings.TrimSpace(combineOutput(checkResult.Stdout, checkResult.Stderr)))
	}

	shownStart := offset
	if shownStart <= 0 {
		shownStart = 1
	}

	readCommand := fmt.Sprintf("head -n %d %s", limit, shellQuote(path))
	if offset > 0 {
		endLine := offset + limit - 1
		readCommand = fmt.Sprintf("sed -n '%d,%dp' %s", offset, endLine, shellQuote(path))
	}

	readResult, err := t.executor.Exec(ctx, nodeExecRequest{
		ContainerID: containerID,
		Args:        []string{"sh", "-c", readCommand},
		Timeout:     15 * time.Second,
	})
	if err != nil {
		return "", fmt.Errorf("read node file: %w", err)
	}
	if readResult.TimedOut {
		return tools.JSONString(map[string]interface{}{
			"container_id": containerID,
			"path":         path,
			"timed_out":    true,
			"message":      fmt.Sprintf("read file timed out: %s", path),
		})
	}
	if readResult.ExitCode != 0 {
		return "", fmt.Errorf("read node file failed: %s", strings.TrimSpace(combineOutput(readResult.Stdout, readResult.Stderr)))
	}

	content := strings.TrimRight(readResult.Stdout, "\n")
	lines := []string{}
	if content != "" {
		lines = strings.Split(content, "\n")
	}
	numbered := make([]string, 0, len(lines))
	for index, line := range lines {
		numbered = append(numbered, fmt.Sprintf("%6d|%s", shownStart+index, line))
	}

	totalLines := 0
	if totalResult, err := t.executor.Exec(ctx, nodeExecRequest{
		ContainerID: containerID,
		Args:        []string{"sh", "-c", fmt.Sprintf("wc -l < %s", shellQuote(path))},
		Timeout:     5 * time.Second,
	}); err == nil && totalResult.ExitCode == 0 {
		totalLines, _ = strconv.Atoi(strings.TrimSpace(totalResult.Stdout))
	}

	shownEnd := shownStart + len(lines) - 1
	hasMore := totalLines > 0 && len(lines) > 0 && shownEnd < totalLines
	numberedContent := strings.Join(numbered, "\n")
	if hasMore {
		numberedContent += fmt.Sprintf("\n\n[file has %d lines, shown %d-%d]", totalLines, shownStart, shownEnd)
	}

	return tools.JSONString(map[string]interface{}{
		"container_id":       containerID,
		"path":               path,
		"exists":             true,
		"offset":             shownStart,
		"limit":              limit,
		"content":            content,
		"numbered_content":   numberedContent,
		"total_lines":        totalLines,
		"shown_start":        shownStart,
		"shown_end":          shownEnd,
		"has_more":           hasMore,
		"display":            numberedContent,
		"display_line_count": len(lines),
	})
}
