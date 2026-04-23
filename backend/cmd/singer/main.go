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
	"github.com/insmtx/SingerOS/backend/internal/eventengine"
	"github.com/insmtx/SingerOS/backend/internal/execution"
	githubprovider "github.com/insmtx/SingerOS/backend/providers/github"
	agentruntime "github.com/insmtx/SingerOS/backend/runtime"
	bundledskills "github.com/insmtx/SingerOS/backend/skills/bundled"
	skillcatalog "github.com/insmtx/SingerOS/backend/skills/catalog"
	"github.com/insmtx/SingerOS/backend/toolruntime"
	"github.com/insmtx/SingerOS/backend/tools"
	githubtools "github.com/insmtx/SingerOS/backend/tools/github"
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

		// Create Execution Engine
		executionEngine := execution.NewExecutionEngine()

		// Create Event Engine with Execution Engine
		eventEngine := eventengine.NewEventEngine(publisher, executionEngine)

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

		// Start Event Engine to consume events
		ctx := context.Background()
		if err := eventEngine.Start(ctx); err != nil {
			logs.Errorf("Failed to start Event Engine: %v", err)
		} else {
			logs.Info("Event Engine started successfully")
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

func buildRuntimeConfig(cfg *config.Config, authService *auth.Service) (agentruntime.Config, error) {
	catalog, err := skillcatalog.New(bundledskills.FS)
	if err != nil {
		return agentruntime.Config{}, fmt.Errorf("load bundled skills: %w", err)
	}

	logs.Infof("Loaded %d bundled skills for runtime", len(catalog.List()))

	toolRegistry, toolExecRuntime, err := buildTooling(cfg, authService)
	if err != nil {
		return agentruntime.Config{}, err
	}

	return agentruntime.Config{
		SkillsCatalog: catalog,
		ToolRegistry:  toolRegistry,
		ToolRuntime:   toolExecRuntime,
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

func buildTooling(cfg *config.Config, authService *auth.Service) (*tools.Registry, *toolruntime.Runtime, error) {
	registry := tools.NewRegistry()

	var githubFactory *githubprovider.ClientFactory
	if cfg != nil && cfg.Github != nil {
		githubFactory = githubprovider.NewClientFactory(*cfg.Github, authService)
		if err := registry.Register(githubtools.NewAccountInfoTool(nil)); err != nil {
			return nil, nil, fmt.Errorf("register github account info tool: %w", err)
		}
		if err := registry.Register(githubtools.NewPullRequestMetadataTool(nil)); err != nil {
			return nil, nil, fmt.Errorf("register github pr metadata tool: %w", err)
		}
		if err := registry.Register(githubtools.NewPullRequestFilesTool(nil)); err != nil {
			return nil, nil, fmt.Errorf("register github pr files tool: %w", err)
		}
		if err := registry.Register(githubtools.NewRepositoryFileTool(nil)); err != nil {
			return nil, nil, fmt.Errorf("register github repository file tool: %w", err)
		}
		if err := registry.Register(githubtools.NewCompareCommitsTool(nil)); err != nil {
			return nil, nil, fmt.Errorf("register github compare commits tool: %w", err)
		}
		if err := registry.Register(githubtools.NewPullRequestReviewPublishTool(nil)); err != nil {
			return nil, nil, fmt.Errorf("register github pr review publish tool: %w", err)
		}
	}

	logs.Infof("Loaded %d tools for runtime", len(registry.ListInfos()))

	return registry, toolruntime.New(registry, githubFactory), nil
}

func buildRuntimeRunner(ctx context.Context, cfg *config.Config, runtimeConfig agentruntime.Config) (agentruntime.Runner, error) {
	if cfg == nil || cfg.LLM == nil || cfg.LLM.APIKey == "" {
		return nil, fmt.Errorf("llm config is required")
	}

	switch cfg.LLM.Provider {
	case "", "openai":
		logs.Info("Using Eino runtime runner")
		return agentruntime.NewEinoRunner(ctx, cfg.LLM, runtimeConfig)
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
