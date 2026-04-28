// Package externalcli adapts external agent CLIs to the SingerOS agent.Runner boundary.
package externalcli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/internal/agent"
	agentevents "github.com/insmtx/SingerOS/backend/internal/agent/events"
	"github.com/insmtx/SingerOS/backend/runtime/engines"
)

// Runner executes one SingerOS agent request through an external CLI engine.
type Runner struct {
	name   string
	engine engines.Engine
	model  engines.ModelConfig
}

// NewRunner creates a SingerOS runner backed by one external CLI engine.
func NewRunner(name string, engine engines.Engine, llmConfig *config.LLMConfig) (*Runner, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("runtime name is required")
	}
	if engine == nil {
		return nil, fmt.Errorf("runtime %q engine is nil", name)
	}
	return &Runner{
		name:   name,
		engine: engine,
		model:  modelFromConfig(llmConfig),
	}, nil
}

// Run executes one normalized request through the configured CLI engine.
func (r *Runner) Run(ctx context.Context, req *agent.RequestContext) (*agent.RunResult, error) {
	startedAt := time.Now().UTC()
	if r == nil || r.engine == nil {
		return nil, fmt.Errorf("external CLI runtime is not initialized")
	}
	if req == nil {
		return nil, fmt.Errorf("request context is required")
	}
	ensureRunDefaults(req)

	emitter := agentevents.NewEmitter(req.RunID, req.TraceID, sinkForRequest(req))
	if err := emitter.Emit(ctx, &agentevents.RunEvent{Type: agentevents.RunEventStarted}); err != nil {
		return nil, err
	}

	workDir := strings.TrimSpace(req.Runtime.WorkDir)
	if workDir == "" {
		workDir = "."
	}
	if err := r.engine.Prepare(ctx, engines.PrepareRequest{WorkDir: workDir}); err != nil {
		return r.failedResult(ctx, emitter, req, startedAt, err, failureMetadata("", workDir)), err
	}

	logPath, err := logPathForRun(req.RunID, r.name)
	if err != nil {
		return r.failedResult(ctx, emitter, req, startedAt, err, failureMetadata("", workDir)), err
	}

	handle, err := r.engine.Run(ctx, engines.RunRequest{
		ExecutionID: req.RunID,
		WorkDir:     workDir,
		Prompt:      buildPrompt(req),
		LogPath:     logPath,
		Model:       modelForRequest(req, r.model),
	})
	if err != nil {
		return r.failedResult(ctx, emitter, req, startedAt, err, failureMetadata(logPath, workDir)), err
	}

	if handle != nil && handle.Process != nil {
		_ = emitter.Emit(ctx, &agentevents.RunEvent{
			Type:    agentevents.RunEventMessageDelta,
			Content: fmt.Sprintf("external runtime %s started with pid %d", r.name, handle.Process.PID()),
		})
	}

	if err := consumeEvents(ctx, emitter, handle); err != nil {
		return r.failedResult(ctx, emitter, req, startedAt, err, failureMetadata(logPath, workDir)), err
	}

	message := ""
	if handle != nil && handle.ExtractResult != nil {
		message = strings.TrimSpace(handle.ExtractResult(logPath))
	}
	result := &agent.RunResult{
		RunID:       req.RunID,
		TraceID:     req.TraceID,
		Status:      agent.RunStatusCompleted,
		Message:     message,
		StartedAt:   startedAt,
		CompletedAt: time.Now().UTC(),
		Metadata: map[string]any{
			"runtime":  r.name,
			"log_path": logPath,
			"work_dir": workDir,
		},
	}
	_ = emitter.Emit(ctx, &agentevents.RunEvent{
		Type:    agentevents.RunEventCompleted,
		Content: result.Message,
	})
	return result, nil
}

func consumeEvents(ctx context.Context, emitter *agentevents.Emitter, handle *engines.RunHandle) error {
	if handle == nil || handle.Events == nil {
		return nil
	}
	for event := range handle.Events {
		switch event.Type {
		case engines.EventStarted:
			continue
		case engines.EventDone:
			return nil
		case engines.EventError:
			if strings.TrimSpace(event.Content) == "" {
				return fmt.Errorf("external runtime failed")
			}
			return fmt.Errorf("%s", event.Content)
		default:
			if strings.TrimSpace(event.Content) != "" {
				_ = emitter.Emit(ctx, &agentevents.RunEvent{
					Type:    agentevents.RunEventMessageDelta,
					Content: event.Content,
				})
			}
		}
	}
	return nil
}

func (r *Runner) failedResult(ctx context.Context, emitter *agentevents.Emitter, req *agent.RequestContext, startedAt time.Time, err error, metadata map[string]any) *agent.RunResult {
	status := agent.RunStatusFailed
	eventType := agentevents.RunEventFailed
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		status = agent.RunStatusCancelled
		eventType = agentevents.RunEventCancelled
	}
	message := ""
	if err != nil {
		message = err.Error()
	}
	_ = emitter.Emit(ctx, &agentevents.RunEvent{
		Type:    eventType,
		Content: message,
	})
	return &agent.RunResult{
		RunID:       req.RunID,
		TraceID:     req.TraceID,
		Status:      status,
		Error:       message,
		StartedAt:   startedAt,
		CompletedAt: time.Now().UTC(),
		Metadata:    metadataWithRuntime(metadata, r.name),
	}
}

func failureMetadata(logPath string, workDir string) map[string]any {
	metadata := map[string]any{}
	if strings.TrimSpace(logPath) != "" {
		metadata["log_path"] = logPath
	}
	if strings.TrimSpace(workDir) != "" {
		metadata["work_dir"] = workDir
	}
	return metadata
}

func metadataWithRuntime(metadata map[string]any, runtimeName string) map[string]any {
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadata["runtime"] = runtimeName
	return metadata
}

func ensureRunDefaults(req *agent.RequestContext) {
	if req.RunID == "" {
		req.RunID = fmt.Sprintf("run_%d", time.Now().UTC().UnixNano())
	}
	if req.Input.Type == "" {
		req.Input.Type = agent.InputTypeMessage
	}
}

func sinkForRequest(req *agent.RequestContext) agentevents.EventSink {
	if req == nil || req.EventSink == nil {
		return agentevents.NewNoopSink()
	}
	return req.EventSink
}

func modelFromConfig(cfg *config.LLMConfig) engines.ModelConfig {
	if cfg == nil {
		return engines.ModelConfig{}
	}
	return engines.ModelConfig{
		Provider: cfg.Provider,
		Model:    cfg.Model,
		APIKey:   cfg.APIKey,
		BaseURL:  cfg.BaseURL,
	}
}

func modelForRequest(req *agent.RequestContext, fallback engines.ModelConfig) engines.ModelConfig {
	model := fallback
	if req == nil {
		return model
	}
	if strings.TrimSpace(req.Model.Provider) != "" {
		model.Provider = req.Model.Provider
	}
	if strings.TrimSpace(req.Model.Model) != "" {
		model.Model = req.Model.Model
	}
	return model
}

func logPathForRun(runID string, runtimeName string) (string, error) {
	dir := filepath.Join(os.TempDir(), "singeros-runtime")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create runtime log dir: %w", err)
	}
	name := sanitizePathPart(runID)
	if name == "" {
		name = fmt.Sprintf("run_%d", time.Now().UTC().UnixNano())
	}
	runtimeName = sanitizePathPart(runtimeName)
	if runtimeName == "" {
		runtimeName = "external"
	}
	return filepath.Join(dir, fmt.Sprintf("%s-%s.jsonl", name, runtimeName)), nil
}

func sanitizePathPart(value string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '-', r == '_', r == '.':
			return r
		default:
			return '_'
		}
	}, strings.TrimSpace(value))
}

var _ agent.Runner = (*Runner)(nil)
