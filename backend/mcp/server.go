// Package mcp exposes SingerOS capabilities through the Model Context Protocol.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/insmtx/SingerOS/backend/tools"
	mcpsdk "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

const (
	serverName    = "SingerOS"
	serverVersion = "0.1.0"
)

// Server owns the MCP SDK server and HTTP transport.
type Server struct {
	sdk  *mcpserver.MCPServer
	http http.Handler
}

// NewServer creates a SingerOS MCP server with the currently public tools.
func NewServer() *Server {
	return NewServerWithTools(NewTools()...)
}

// NewServerWithTools creates a SingerOS MCP server from SingerOS internal tools.
func NewServerWithTools(publicTools ...tools.Tool) *Server {
	sdk := mcpserver.NewMCPServer(
		serverName,
		serverVersion,
		mcpserver.WithRecovery(),
	)

	registerTools(sdk, publicTools)

	return &Server{
		sdk:  sdk,
		http: mcpserver.NewStreamableHTTPServer(sdk),
	}
}

// Handler returns the streamable HTTP MCP transport handler.
func (s *Server) Handler() http.Handler {
	if s == nil {
		return http.NotFoundHandler()
	}
	return s.http
}

// GetTool returns a registered MCP tool by name. It is intended for tests and diagnostics.
func (s *Server) GetTool(name string) *mcpserver.ServerTool {
	if s == nil || s.sdk == nil {
		return nil
	}
	return s.sdk.GetTool(name)
}

func registerTools(s *mcpserver.MCPServer, publicTools []tools.Tool) {
	for _, tool := range publicTools {
		s.AddTool(toMCPTool(tool), toMCPHandler(tool))
	}
}

func toMCPTool(tool tools.Tool) mcpsdk.Tool {
	schema, err := json.Marshal(tool.InputSchema())
	if err != nil {
		schema = []byte(`{"type":"object"}`)
	}
	return mcpsdk.NewToolWithRawSchema(tool.Name(), tool.Description(), schema)
}

func toMCPHandler(tool tools.Tool) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, request mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
		args := request.GetArguments()
		if args == nil {
			args = map[string]any{}
		}

		if validator, ok := tool.(tools.Validator); ok {
			if err := validator.Validate(args); err != nil {
				return mcpsdk.NewToolResultError(err.Error()), nil
			}
		}

		output, err := tool.Execute(ctx, args)
		if err != nil {
			return mcpsdk.NewToolResultError(err.Error()), nil
		}
		if output == "" {
			return mcpsdk.NewToolResultText("{}"), nil
		}

		var structured any
		if err := json.Unmarshal([]byte(output), &structured); err == nil {
			return mcpsdk.NewToolResultStructured(structured, output), nil
		}

		return mcpsdk.NewToolResultText(fmt.Sprintf("%s", output)), nil
	}
}
