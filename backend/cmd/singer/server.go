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
	infradb "github.com/insmtx/SingerOS/backend/internal/infra/db"
	"github.com/insmtx/SingerOS/backend/internal/service/trace"
	"github.com/insmtx/SingerOS/backend/internal/infra/mq"
	"github.com/insmtx/SingerOS/backend/internal/service"
	"github.com/insmtx/SingerOS/backend/internal/eventengine"
	"github.com/insmtx/SingerOS/backend/internal/agent"
	"github.com/insmtx/SingerOS/backend/tools"
	skilltools "github.com/insmtx/SingerOS/backend/tools/skill"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/apis/runtime/middleware"
	ygconfig "github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

var (
	serverConfigPath string
	serverHttpAddr   string
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the SingerOS backend HTTP server",
	Long:  `Start the HTTP server that handles API requests, events, and orchestrator services.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := loadConfig()
		if err != nil {
			logs.Fatalf("Failed to load config: %v", err)
			return
		}

		rmqUrl := "amqp://singer_user:singer_password@rabbitmq:5672/"
		if cfg.RabbitMQ != nil && cfg.RabbitMQ.URL != "" {
			rmqUrl = cfg.RabbitMQ.URL
		}

		rmqCfg := ygconfig.RabbitMQConfig{URL: rmqUrl}
		publisher, _, err := mq.NewPublisher(rmqCfg)
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

		orchestratorInstance := eventengine.NewOrchestrator(publisher, runner)

		var db *gorm.DB
		if cfg.Database != nil && cfg.Database.URL != "" {
			var err error
			db, err = infradb.InitDB(*cfg.Database)
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

		r := gin.New()
		{
			r.Use(middleware.CORS())
			r.Use(trace.CustomerHeader())
			r.Use(trace.Logger(".Ping", "metrics"))
			r.Use(middleware.Recovery())
		}

		service.SetupRouter(r, *cfg, publisher, db, authService)

		srv := &http.Server{
			Addr:    serverHttpAddr,
			Handler: r,
		}

		logs.Info("Starting SingerOS backend service...")
		logs.Infof("Listening on %s", serverHttpAddr)

		ctx := context.Background()
		if err := orchestratorInstance.Start(ctx); err != nil {
			logs.Errorf("Failed to start orchestrator: %v", err)
		} else {
			logs.Info("Orchestrator started successfully")
		}

		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logs.Fatalf("Failed to start server: %v", err)
			}
		}()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logs.Info("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logs.Errorf("Server forced to shutdown: %v", err)
		}

		if publisher != nil {
			publisher.Close()
		}

		logs.Info("Server exited")
	},
}

func init() {
	serverCmd.Flags().StringVar(&serverConfigPath, "config", "", "Configuration file path")
	serverCmd.Flags().StringVar(&serverHttpAddr, "addr", ":8080", "HTTP server address")
	rootCmd.AddCommand(serverCmd)
}

func loadConfig() (*config.Config, error) {
	var cfg config.Config

	if serverConfigPath != "" {
		err := ygconfig.LoadYamlLocalFile(serverConfigPath, &cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to load config from %s: %v", serverConfigPath, err)
		}
	} else {
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
