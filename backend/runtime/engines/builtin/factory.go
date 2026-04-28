// Package builtin 连接内置的外部 CLI 引擎适配器。
package builtin

import (
	"fmt"

	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/runtime/engines"
	"github.com/insmtx/SingerOS/backend/runtime/engines/claude"
	"github.com/insmtx/SingerOS/backend/runtime/engines/codex"
)

// NewRegistryFromConfig creates a registry with every detected built-in CLI engine.
func NewRegistryFromConfig(cfg *config.CLIEnginesConfig) (*engines.Registry, error) {
	registry := engines.NewRegistry()
	for _, status := range engines.DiscoverAvailableCLI() {
		if !status.Installed {
			continue
		}
		engine, err := newEngine(status.Name, status.Path)
		if err != nil {
			return nil, err
		}
		if err := registry.Register(status.Name, engine); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

func newEngine(name string, path string) (engines.Engine, error) {
	switch name {
	case engines.EngineClaude:
		return claude.NewAdapter(path, nil), nil
	case engines.EngineCodex:
		return codex.NewAdapter(path, nil), nil
	default:
		return nil, fmt.Errorf("unsupported CLI engine %q", name)
	}
}
