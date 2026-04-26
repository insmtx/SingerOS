package client

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/internal/agent"
	"github.com/insmtx/SingerOS/backend/tools"
	skilltools "github.com/insmtx/SingerOS/backend/tools/skill"
	"github.com/ygpkg/yg-go/logs"
)

type Worker struct {
	runtime    agent.AgentRuntime
	config     *WorkerConfig
	workerID   string
	startedAt  time.Time
	status     string
	wsClient   *WSClient
}

type WorkerConfig struct {
	Runtime      agent.AgentRuntime
	LLMConfig    *config.LLMConfig
	SkillsDir    string
	ToolsEnabled bool
	ServerAddr   string
}

func NewWorker(ctx context.Context, cfg *WorkerConfig) (*Worker, error) {
	if cfg == nil {
		return nil, fmt.Errorf("worker config is required")
	}

	runtime := cfg.Runtime
	if runtime == nil {
		runtime = buildDefaultRuntime(ctx, cfg)
	}

	if runtime == nil {
		return nil, fmt.Errorf("either Runtime or LLMConfig must be provided")
	}

	workerID := fmt.Sprintf("worker_%d", time.Now().UnixNano())

	w := &Worker{
		runtime:   runtime,
		config:    cfg,
		workerID:  workerID,
		startedAt: time.Now(),
		status:    "initialized",
	}

	if cfg.ServerAddr != "" {
		w.wsClient = NewWSClient(cfg.ServerAddr, workerID)
	}

	return w, nil
}

func buildDefaultRuntime(ctx context.Context, cfg *WorkerConfig) agent.AgentRuntime {
	if cfg.LLMConfig == nil {
		return nil
	}

	catalog, err := loadSkillsCatalog(cfg.SkillsDir)
	if err != nil {
		logs.Errorf("load skills catalog: %v", err)
		return nil
	}

	toolRegistry := tools.NewRegistry()

	if cfg.ToolsEnabled {
		if err := skilltools.Register(toolRegistry, catalog); err != nil {
			logs.Errorf("register tools: %v", err)
			return nil
		}
	}

	agentConfig := agent.Config{
		SkillsCatalog: catalog,
		ToolRegistry:  toolRegistry,
	}

	agentInstance, err := agent.NewAgent(ctx, cfg.LLMConfig, agentConfig)
	if err != nil {
		logs.Errorf("create agent: %v", err)
		return nil
	}

	return agentInstance
}

func (w *Worker) Run(ctx context.Context, req *agent.RequestContext) (*agent.RunResult, error) {
	if w == nil || w.runtime == nil {
		return nil, fmt.Errorf("worker runtime is not initialized")
	}
	
	w.status = "processing"
	result, err := w.runtime.Run(ctx, req)
	if err != nil {
		w.status = "error"
		return nil, err
	}
	
	w.status = "idle"
	return result, nil
}

func (w *Worker) Start(ctx context.Context) error {
	w.status = "running"
	logs.Infof("Worker %s started", w.workerID)

	if w.wsClient != nil {
		if err := w.wsClient.Connect(ctx); err != nil {
			logs.Warnf("Failed to connect to server WebSocket: %v", err)
		} else {
			logs.Info("Connected to server via WebSocket")
		}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		logs.Info("Worker context cancelled")
		return w.Shutdown(ctx)
	case sig := <-sigChan:
		logs.Infof("Received signal %v, shutting down", sig)
		return w.Shutdown(ctx)
	}
}

func (w *Worker) Shutdown(ctx context.Context) error {
	logs.Info("Worker shutting down...")
	w.status = "stopping"

	if w.wsClient != nil {
		w.wsClient.Close()
	}

	return nil
}

func (w *Worker) GetWorkerID() string {
	return w.workerID
}

func (w *Worker) GetStartedAt() time.Time {
	return w.startedAt
}

func (w *Worker) GetStatus() string {
	return w.status
}

func loadSkillsCatalog(skillsDir string) (*skilltools.Catalog, error) {
	if skillsDir == "" {
		return skilltools.NewEmptyCatalog(), nil
	}

	catalog, _, err := skilltools.LoadDefaultCatalog()
	if err != nil {
		return nil, err
	}
	return catalog, nil
}
