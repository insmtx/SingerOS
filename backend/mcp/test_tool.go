package mcp

import (
	"github.com/insmtx/SingerOS/backend/tools"
	testtools "github.com/insmtx/SingerOS/backend/tools/test"
)

// NewTools returns the SingerOS tools that are currently exposed through MCP.
func NewTools() []tools.Tool {
	return []tools.Tool{
		testtools.NewEchoTool(),
	}
}
