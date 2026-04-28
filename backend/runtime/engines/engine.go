// Package engines 定义了外部 AI CLI 引擎的执行边界。
// 包含引擎接口定义、运行请求/结果数据结构、进程生命周期事件等核心类型。
package engines

import (
	"context"
	"time"

	runtimeevents "github.com/insmtx/SingerOS/backend/runtime/events"
)

const (
	// EngineClaude is the registry name for Claude Code.
	EngineClaude = "claude"
	// EngineCodex is the registry name for Codex CLI.
	EngineCodex = "codex"
)

const (
	// EventStarted indicates that the external process has started.
	EventStarted = runtimeevents.EventStarted
	// EventMessageDelta indicates human-readable CLI output that can be streamed to callers.
	EventMessageDelta = runtimeevents.EventMessageDelta
	// EventResult indicates the final assistant result emitted by the CLI.
	EventResult = runtimeevents.EventResult
	// EventDone indicates that the external process completed successfully.
	EventDone = runtimeevents.EventCompleted
	// EventError indicates that the external process failed.
	EventError = runtimeevents.EventFailed
)

// EventType 分类引擎进程发出的生命周期事件类型。
type EventType = runtimeevents.EventType

// Event 引擎进程发出的生命周期事件。
type Event = runtimeevents.Event

// PrepareRequest 包含引擎特定的工作区设置输入。
type PrepareRequest struct {
	WorkDir string
}

// ModelConfig 携带注入到 CLI 进程的模型和提供商设置。
type ModelConfig struct {
	Provider string // 提供商名称（如 openai, anthropic）
	Model    string // 模型名称
	APIKey   string // API 密钥
	BaseURL  string // API 基础 URL
}

// RunRequest 包含执行一次外部 CLI 运行所需的全部输入。
type RunRequest struct {
	ExecutionID string        // 执行唯一标识
	SessionID   string        // 会话 ID，用于恢复对话
	Resume      bool          // 是否恢复之前的会话
	WorkDir     string        // 工作目录
	Prompt      string        // 输入提示
	Model       ModelConfig   // 模型配置
	ExtraEnv    []string      // 额外环境变量
	Timeout     time.Duration // 超时时间
}

// Process 是运行中的外部 CLI 进程的句柄。
type Process interface {
	PID() int    // 获取进程 ID
	Stop() error // 停止进程
}

// RunHandle 引擎进程启动后返回的句柄。
type RunHandle struct {
	Process Process      // 进程控制句柄
	Events  <-chan Event // 事件通道
}

// Engine 通过外部 AI CLI 执行提示。
type Engine interface {
	// Prepare 准备引擎工作区
	Prepare(ctx context.Context, req PrepareRequest) error
	// RegisterMCP registers a Model Context Protocol server with the engine CLI.
	RegisterMCP(ctx context.Context, cfg MCPServerConfig) error
	// Run 运行引擎并返回进程句柄
	Run(ctx context.Context, req RunRequest) (*RunHandle, error)
}
