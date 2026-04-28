package agent

import (
	"context"
	"fmt"
	"strings"
)

const (
	// RuntimeKindSingerOS is the built-in SingerOS agent runtime.
	RuntimeKindSingerOS = "singeros"
)

// RuntimeRouter dispatches normalized agent requests to a concrete runtime.
type RuntimeRouter struct {
	defaultKind string
	runners     map[string]Runner
}

// NewRuntimeRouter creates a runtime router with a default runtime kind.
func NewRuntimeRouter(defaultKind string) *RuntimeRouter {
	return &RuntimeRouter{
		defaultKind: normalizeRuntimeKind(defaultKind),
		runners:     make(map[string]Runner),
	}
}

// Register adds or replaces one runtime runner.
func (r *RuntimeRouter) Register(kind string, runner Runner) error {
	kind = normalizeRuntimeKind(kind)
	if kind == "" {
		return fmt.Errorf("runtime kind is required")
	}
	if runner == nil {
		return fmt.Errorf("runtime %q runner is nil", kind)
	}
	if r.runners == nil {
		r.runners = make(map[string]Runner)
	}
	r.runners[kind] = runner
	return nil
}

// SetDefault updates the fallback runtime kind.
func (r *RuntimeRouter) SetDefault(kind string) {
	if r == nil {
		return
	}
	r.defaultKind = normalizeRuntimeKind(kind)
}

// Run executes the request with the selected runtime.
func (r *RuntimeRouter) Run(ctx context.Context, req *RequestContext) (*RunResult, error) {
	if r == nil {
		return nil, fmt.Errorf("runtime router is nil")
	}

	kind := r.defaultKind
	if req != nil && strings.TrimSpace(req.Runtime.Kind) != "" {
		kind = normalizeRuntimeKind(req.Runtime.Kind)
	}
	if kind == "" {
		return nil, fmt.Errorf("runtime kind is required")
	}

	runner := r.runners[kind]
	if runner == nil {
		return nil, fmt.Errorf("runtime %q is not available", kind)
	}
	return runner.Run(ctx, req)
}

func normalizeRuntimeKind(kind string) string {
	return strings.ToLower(strings.TrimSpace(kind))
}
