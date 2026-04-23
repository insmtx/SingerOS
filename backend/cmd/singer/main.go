// main 包是 SingerOS 后端服务的主入口
//
// SingerOS 是一个 AI 驱动的操作系统，提供事件驱动的交互能力、
// 技能系统、数字助手等功能。该服务负责处理 API 请求、事件路由
// 和业务逻辑的协调。
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	auth "github.com/insmtx/SingerOS/backend/auth"
	authgithub "github.com/insmtx/SingerOS/backend/auth/providers/github"
	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/database"
	"github.com/insmtx/SingerOS/backend/gateway/trace"
	"github.com/insmtx/SingerOS/backend/interaction/eventbus/rabbitmq"
	gateway "github.com/insmtx/SingerOS/backend/interaction/gateway"
	orchestrator "github.com/insmtx/SingerOS/backend/orchestrator"
	agentruntime "github.com/insmtx/SingerOS/backend/runtime"
	"github.com/insmtx/SingerOS/backend/tools"
	skilltools "github.com/insmtx/SingerOS/backend/tools/skill"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/apis/runtime/middleware"
	ygconfig "github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

var (
	configPath string
	httpAddr   string
)

var rootCmd = &cobra.Command{
	Use:   "singer",
	Short: "Backend service for the SingerOS Backend",
	Long:  `This is the backend service for the SingerOS Backend, responsible for handling API requests and business logic.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration from file
		cfg, err := loadConfig()
		if err != nil {
			logs.Fatalf("Failed to load config: %v", err)
			return
		}

		// Initialize event bus - require URL from config file
		rmqUrl := "amqp://singer_user:singer_password@rabbitmq:5672/" // default for docker-compose
		if cfg.RabbitMQ != nil && cfg.RabbitMQ.URL != "" {
			rmqUrl = cfg.RabbitMQ.URL // override with config file value
		}

		rmqCfg := ygconfig.RabbitMQConfig{URL: rmqUrl}
		publisher, _, err := rabbitmq.NewPublisher(rmqCfg)
		if err != nil {
			logs.Fatalf("Failed to create event publisher: %v", err)
			return
		}

		if cfg.LLM == nil || cfg.LLM.APIKey == "" {
			logs.Fatalf("LLM configuration is required for Eino runtime")
			return
		}

		authService := buildAuthService(cfg)

		runtimeConfig, err := buildRuntimeConfig()
		if err != nil {
			logs.Fatalf("Failed to build runtime config: %v", err)
			return
		}

		runner, err := buildRuntimeRunner(context.Background(), cfg, runtimeConfig)
		if err != nil {
			logs.Fatalf("Failed to create agent runtime: %v", err)
			return
		}

		// Create orchestrator to consume events through the runtime boundary.
		orchestratorInstance := orchestrator.NewOrchestrator(publisher, runner)

		// Initialize database if configuration is provided
		var db *gorm.DB
		if cfg.Database != nil && cfg.Database.URL != "" {
			var err error
			db, err = database.InitDB(*cfg.Database)
			if err != nil {
				logs.Fatalf("Failed to initialize database: %v", err)
				return
			}
			logs.Info("Database initialized successfully")
		} else {
			logs.Warn("No database configuration provided")
			logs.Warn("  - Database-dependent features (user persistence, etc.) will be unavailable")
			logs.Warn("  - To enable database, add database.url to your config file")
			logs.Warn("  - See example-config.yaml for database configuration example")
		}

		// Set up the HTTP router
		r := gin.New()
		{
			r.Use(middleware.CORS())
			r.Use(trace.CustomerHeader())
			r.Use(trace.Logger(".Ping", "metrics"))
			r.Use(middleware.Recovery())
		}

		// Set up gateway with connectors
		gateway.SetupRouter(r, *cfg, publisher, db, authService)

		// Create HTTP server
		srv := &http.Server{
			Addr:    httpAddr,
			Handler: r,
		}

		logs.Info("Starting SingerOS backend service...")
		logs.Infof("Listening on %s", httpAddr)

		// Start orchestrator to consume events
		ctx := context.Background()
		if err := orchestratorInstance.Start(ctx); err != nil {
			logs.Errorf("Failed to start orchestrator: %v", err)
		} else {
			logs.Info("Orchestrator started successfully")
		}

		// Start the server in a goroutine
		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logs.Fatalf("Failed to start server: %v", err)
			}
		}()

		// Wait for interrupt signal to gracefully shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logs.Info("Shutting down server...")

		// Gracefully shutdown the server
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logs.Errorf("Server forced to shutdown: %v", err)
		}

		// Close publisher connection
		if publisher != nil {
			publisher.Close()
		}

		logs.Info("Server exited")
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Configuration file path")
	rootCmd.PersistentFlags().StringVar(&httpAddr, "addr", ":8080", "HTTP server address")
}

func loadConfig() (*config.Config, error) {
	var cfg config.Config

	if configPath != "" {
		// Load config from specified path
		err := ygconfig.LoadYamlLocalFile(configPath, &cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to load config from %s: %v", configPath, err)
		}
	} else {
		// Try to load from default locations
		pathsToTry := []string{"./config.yaml", "/app/config.yaml"}

		err := fmt.Errorf("config file not found in any location")
		for _, path := range pathsToTry {
			if err = ygconfig.LoadYamlLocalFile(path, &cfg); err == nil {
				logs.Infof("Loaded config from: %s", path)
				break
			}
		}

		if err != nil {
			logs.Warnf("Could not load config from any path (%v), will proceed without config", err)
			return &config.Config{}, nil
		}
	}

	logs.Info("Configuration loaded successfully")
	return &cfg, nil
}

func buildRuntimeConfig() (agentruntime.Config, error) {
	catalog, skillDir, err := skilltools.LoadDefaultCatalog()
	if err != nil {
		return agentruntime.Config{}, fmt.Errorf("load skills: %w", err)
	}

	logs.Infof("Loaded %d skills from %s for runtime", len(catalog.List()), skillDir)

	toolRegistry, err := buildTooling(catalog)
	if err != nil {
		return agentruntime.Config{}, err
	}

	return agentruntime.Config{
		SkillsCatalog: catalog,
		ToolRegistry:  toolRegistry,
	}, nil
}

func buildAuthService(cfg *config.Config) *auth.Service {
	accountStore := auth.NewInMemoryStore()
	accountResolver := auth.NewAccountResolver(accountStore)
	authService := auth.NewService(accountStore, accountResolver)

	if cfg != nil && cfg.Github != nil {
		authService.RegisterProvider(authgithub.NewOAuthProvider(*cfg.Github))
	}

	return authService
}

func buildTooling(catalog *skilltools.Catalog) (*tools.Registry, error) {
	registry := tools.NewRegistry()

	if err := skilltools.Register(registry, catalog); err != nil {
		return nil, fmt.Errorf("register skill use tool: %w", err)
	}

	logs.Infof("Loaded %d tools for runtime", len(registry.List()))

	return registry, nil
}

func buildRuntimeRunner(ctx context.Context, cfg *config.Config, runtimeConfig agentruntime.Config) (agentruntime.Runner, error) {
	if cfg == nil || cfg.LLM == nil || cfg.LLM.APIKey == "" {
		return nil, fmt.Errorf("llm config is required")
	}

	switch cfg.LLM.Provider {
	case "", "openai":
		logs.Info("Using SingerOS agent runtime")
		return agentruntime.NewAgent(ctx, cfg.LLM, runtimeConfig)
	default:
		return nil, fmt.Errorf("unsupported Eino chat model provider: %s", cfg.LLM.Provider)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logs.Errorf("Error executing command: %v", err)
		os.Exit(1)
	}
}
