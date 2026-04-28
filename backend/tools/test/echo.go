// Package testtools provides simple tools used for connectivity and integration tests.
package testtools

import (
	"context"
	"fmt"
	"strings"

	"github.com/insmtx/SingerOS/backend/tools"
)

const (
	// ToolNameEcho is the stable tool name for the connectivity test tool.
	ToolNameEcho = "singeros_echo"

	serverName = "SingerOS"
)

// EchoTool is a SingerOS internal tool for connectivity tests.
type EchoTool struct {
	tools.BaseTool
}

type echoResult struct {
	Message string `json:"message"`
	Server  string `json:"server"`
}

// NewEchoTool creates the SingerOS connectivity test tool.
func NewEchoTool() *EchoTool {
	return &EchoTool{
		BaseTool: tools.NewBaseTool(
			ToolNameEcho,
			"Echoes a message to verify SingerOS tool connectivity.",
			tools.Schema{
				Type:     "object",
				Required: []string{"message"},
				Properties: map[string]*tools.Property{
					"message": {
						Type:        "string",
						Description: "Message to echo back.",
					},
				},
			},
		),
	}
}

// Validate checks the echo tool input.
func (t *EchoTool) Validate(input map[string]interface{}) error {
	if strings.TrimSpace(stringValue(input, "message")) == "" {
		return fmt.Errorf("message is required")
	}
	return nil
}

// Execute echoes the message as structured JSON.
func (t *EchoTool) Execute(ctx context.Context, input map[string]interface{}) (string, error) {
	message := stringValue(input, "message")

	result := echoResult{
		Message: strings.TrimSpace(message),
		Server:  serverName,
	}

	return tools.JSONString(result)
}

func stringValue(input map[string]interface{}, key string) string {
	value, _ := input[key].(string)
	return strings.TrimSpace(value)
}
