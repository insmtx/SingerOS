package main

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/internal/agent"
	"github.com/insmtx/SingerOS/backend/internal/eventengine"
	"github.com/insmtx/SingerOS/backend/internal/infra/mq"
	singerMCP "github.com/insmtx/SingerOS/backend/mcp"
	"github.com/insmtx/SingerOS/backend/runtime/engines"
	"github.com/insmtx/SingerOS/backend/runtime/engines/builtin"
	"github.com/insmtx/SingerOS/backend/tools"
	skilltools "github.com/insmtx/SingerOS/backend/tools/skill"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/lifecycle"
	"github.com/ygpkg/yg-go/logs"
)

var (
	workerConfigPath string
	workerServerAddr string
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Start the SingerOS background worker",
	Long:  `Start the background worker service for processing asynchronous tasks and events.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := loadWorkerConfig(workerConfigPath, workerServerAddr)
		if err != nil {
			logs.Fatalf("Failed to load config: %v", err)
			return
		}

		natsUrl := "nats://nats:4222"
		if cfg.NATS != nil && cfg.NATS.URL != "" {
			natsUrl = cfg.NATS.URL
		}

		subscriber, err := mq.NewPublisher(natsUrl)
		if err != nil {
			logs.Fatalf("Failed to create event subscriber: %v", err)
			return
		}

		runtimeConfig, err := buildRuntimeConfig()
		if err != nil {
			logs.Fatalf("Failed to build runtime config: %v", err)
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		runner, err := buildRuntimeRunner(ctx, cfg, runtimeConfig)
		if err != nil {
			cancel()
			logs.Fatalf("Failed to create agent runtime: %v", err)
			return
		}

		orchestratorInstance := eventengine.NewOrchestrator(subscriber, runner)
		if err := orchestratorInstance.Start(ctx); err != nil {
			cancel()
			logs.Fatalf("Failed to start orchestrator: %v", err)
			return
		}
		logs.Info("Orchestrator started successfully")

		lifecycle.Std().AddCloseFunc(func() error {
			cancel()
			return nil
		})
		lifecycle.Std().AddCloseFunc(subscriber.Close)

		logs.Info("Worker runtime initialized successfully")
		logs.Info("Worker service started")

		lifecycle.Std().WaitExit()

		logs.Info("Worker exited")
	},
}

func init() {
	workerCmd.Flags().StringVar(&workerConfigPath, "config", "", "Configuration file path")
	workerCmd.Flags().StringVar(&workerServerAddr, "server-addr", ":8080", "SingerOS server address for runtime MCP bootstrap")
	rootCmd.AddCommand(workerCmd)
}

func loadWorkerConfig(configPath string, bootstrapAddr string) (*config.Config, error) {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return nil, err
	}

	bootstrapped, err := builtin.BootstrapCLIEngines(context.Background(), cfg.CLI, defaultCLIBootstrapOptions(bootstrapAddr))
	if err != nil {
		logs.Warnf("CLI bootstrap failed: %v", err)
	}
	if bootstrapped != nil {
		cfg.CLI = bootstrapped
	}

	return cfg, nil
}

func defaultCLIBootstrapOptions(addr string) builtin.BootstrapOptions {
	return builtin.BootstrapOptions{
		MCP: engines.MCPServerConfig{
			Name:        "singeros",
			URL:         mcpURLFromAddr(addr),
			BearerToken: singerMCP.DefaultAuthToken(),
		},
	}
}

func mcpURLFromAddr(addr string) string {
	host := "localhost"
	port := "8080"

	if strings.TrimSpace(addr) != "" {
		if splitHost, splitPort, err := net.SplitHostPort(addr); err == nil {
			if splitHost != "" && splitHost != "0.0.0.0" && splitHost != "::" && splitHost != "[::]" {
				host = splitHost
			}
			if splitPort != "" {
				port = splitPort
			}
		} else if strings.HasPrefix(addr, ":") {
			port = strings.TrimPrefix(addr, ":")
		} else {
			host = addr
		}
	}

	return fmt.Sprintf("http://%s:%s/v1/mcp", host, port)
}

func buildRuntimeConfig() (agent.Config, error) {
	catalog, skillDir, err := skilltools.LoadDefaultCatalog()
	if err != nil {
		return agent.Config{}, fmt.Errorf("load skills: %w", err)
	}

	logs.Infof("Loaded %d skills from %s for runtime", len(catalog.List()), skillDir)

	toolRegistry, err := buildTooling(catalog)
	if err != nil {
		return agent.Config{}, err
	}

	return agent.Config{
		SkillsCatalog: catalog,
		ToolRegistry:  toolRegistry,
	}, nil
}

func buildTooling(catalog *skilltools.Catalog) (*tools.Registry, error) {
	registry := tools.NewRegistry()

	if err := skilltools.Register(registry, catalog); err != nil {
		return nil, fmt.Errorf("register skill use tool: %w", err)
	}

	logs.Infof("Loaded %d tools for runtime", len(registry.List()))

	return registry, nil
}

func buildRuntimeRunner(ctx context.Context, cfg *config.Config, runtimeConfig agent.Config) (agent.Runner, error) {
	if cfg == nil || cfg.LLM == nil || cfg.LLM.APIKey == "" {
		return nil, fmt.Errorf("llm config is required")
	}

	switch cfg.LLM.Provider {
	case "", "openai":
		logs.Info("Using SingerOS agent runtime")
		return agent.NewAgent(ctx, cfg.LLM, runtimeConfig)
	default:
		return nil, fmt.Errorf("unsupported Eino chat model provider: %s", cfg.LLM.Provider)
	}
}
