package builtin

import (
	"context"
	"errors"
	"fmt"

	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/runtime/engines"
	"github.com/ygpkg/yg-go/logs"
)

// BootstrapOptions controls host-level CLI bootstrap side effects.
type BootstrapOptions struct {
	SkillsSourceDir string
	SkillTargetDirs []string
	MCP             engines.MCPServerConfig
}

// BootstrapCLIEngines discovers built-in CLIs, syncs skills, and registers SingerOS MCP.
func BootstrapCLIEngines(ctx context.Context, cfg *config.CLIEnginesConfig, opts BootstrapOptions) (*config.CLIEnginesConfig, error) {
	if cfg == nil {
		cfg = &config.CLIEnginesConfig{}
	}
	if ctx == nil {
		ctx = context.Background()
	}

	var bootstrapErr error

	if err := engines.SyncBuiltinSkills(opts.SkillsSourceDir, opts.SkillTargetDirs); err != nil {
		bootstrapErr = errors.Join(bootstrapErr, err)
		logs.Warnf("Sync built-in skills failed: %v", err)
	}

	logs.Info("Starting CLI engine auto-detection...")
	available := engines.DiscoverAvailableCLI()

	if len(available) == 0 {
		logs.Warn("No CLI tools detected")
		return cfg, bootstrapErr
	}

	hasAvailable := false
	for _, s := range available {
		if s.Installed {
			hasAvailable = true
			logs.Infof("  - %s: %s (v%s) @ %s",
				s.DisplayName, s.Name, s.Version, s.Path)
		} else {
			logs.Infof("  - %s: not installed (install: %s)",
				s.DisplayName, s.InstallCmd)
		}
	}

	if !hasAvailable {
		logs.Warn("No CLI engines available, cannot determine default engine")
		return cfg, bootstrapErr
	}

	if cfg.Default == "" {
		defaultName := engines.GetDefaultEngineName(available)
		if defaultName != "" {
			cfg.Default = defaultName
			logs.Infof("Auto-detected default engine: %s", defaultName)
		}
	}

	if err := registerMCPForAvailableCLI(ctx, available, opts.MCP); err != nil {
		bootstrapErr = errors.Join(bootstrapErr, err)
		logs.Warnf("Register MCP server for CLI engines failed: %v", err)
	}

	logs.Info("CLI engine auto-detection complete")
	return cfg, bootstrapErr
}

func registerMCPForAvailableCLI(ctx context.Context, available []engines.CLIToolStatus, cfg engines.MCPServerConfig) error {
	cfg = engines.NormalizeMCPServerConfig(cfg)
	if cfg.URL == "" {
		logs.Debug("No MCP URL provided, skipping CLI MCP registration")
		return nil
	}

	var registerErr error
	for _, status := range available {
		if !status.Installed {
			continue
		}

		engine, err := newEngine(status.Name, status.Path)
		if err != nil {
			registerErr = errors.Join(registerErr, fmt.Errorf("%s: %w", status.Name, err))
			continue
		}
		if err := engine.RegisterMCP(ctx, cfg); err != nil {
			registerErr = errors.Join(registerErr, fmt.Errorf("%s: %w", status.Name, err))
			logs.Warnf("Failed to register SingerOS MCP server for %s: %v", status.Name, err)
			continue
		}
		logs.Infof("Registered SingerOS MCP server for %s", status.Name)
	}
	return registerErr
}
