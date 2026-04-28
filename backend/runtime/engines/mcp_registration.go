package engines

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const mcpRegisterTimeout = 10 * time.Second
const singerOSMCPTokenEnvVar = "SINGEROS_MCP_TOKEN"

// MCPServerConfig describes the SingerOS MCP endpoint registered with an external CLI client.
type MCPServerConfig struct {
	Name        string
	URL         string
	BearerToken string
}

// NormalizeMCPServerConfig fills defaults for an MCP server registration.
func NormalizeMCPServerConfig(cfg MCPServerConfig) MCPServerConfig {
	cfg.Name = strings.TrimSpace(cfg.Name)
	if cfg.Name == "" {
		cfg.Name = "singeros"
	}
	cfg.URL = strings.TrimSpace(cfg.URL)
	cfg.BearerToken = strings.TrimSpace(cfg.BearerToken)
	return cfg
}

// SingerOSMCPTokenEnvVar returns the env var name used for CLI MCP bearer token registration.
func SingerOSMCPTokenEnvVar() string {
	return singerOSMCPTokenEnvVar
}

// RunCLICommand runs a CLI command with a bounded timeout.
func RunCLICommand(ctx context.Context, cliPath string, args []string, extraEnv []string) error {
	if strings.TrimSpace(cliPath) == "" {
		return fmt.Errorf("cli path is required")
	}
	execCtx, cancel := context.WithTimeout(ctx, mcpRegisterTimeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, cliPath, args...)
	if len(extraEnv) > 0 {
		cmd.Env = append(os.Environ(), extraEnv...)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			return execCtx.Err()
		}
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
