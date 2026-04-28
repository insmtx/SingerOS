// Package claude 将 Claude Code 适配到 SingerOS 外部 CLI 引擎接口。
package claude

import (
	"context"

	"github.com/insmtx/SingerOS/backend/runtime/engines"
)

// Adapter 通过 Claude Code 执行提示。
type Adapter struct {
	invoker *Invoker
}

// NewAdapter 创建 Claude Code 引擎适配器。
func NewAdapter(binary string, extraEnv map[string]string) *Adapter {
	if binary == "" {
		binary = "claude"
	}
	return &Adapter{invoker: NewInvoker(binary, extraEnv)}
}

// Prepare 执行 Claude 工作区设置（当前为空实现）。
func (a *Adapter) Prepare(_ context.Context, _ engines.PrepareRequest) error {
	return nil
}

// RegisterMCP registers a streamable HTTP MCP server with Claude Code.
func (a *Adapter) RegisterMCP(ctx context.Context, cfg engines.MCPServerConfig) error {
	cfg = engines.NormalizeMCPServerConfig(cfg)
	_ = engines.RunCLICommand(ctx, a.invoker.binary, []string{"mcp", "remove", cfg.Name}, nil)

	args := []string{"mcp", "add", "--transport", "http", cfg.Name, "--scope", "user", cfg.URL}
	if cfg.BearerToken != "" {
		args = append(args, "--header", "Authorization: Bearer "+cfg.BearerToken)
	}
	return engines.RunCLICommand(ctx, a.invoker.binary, args, nil)
}

// Run 启动 Claude Code 并返回进程句柄。
func (a *Adapter) Run(ctx context.Context, req engines.RunRequest) (*engines.RunHandle, error) {
	proc, events, err := a.invoker.Run(ctx, req)
	if err != nil {
		return nil, err
	}
	return &engines.RunHandle{
		Process:       proc,
		Events:        events,
		ExtractResult: ExtractResultFromLog,
	}, nil
}

var _ engines.Engine = (*Adapter)(nil)
