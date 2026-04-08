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
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/SingerOS/backend/agent/react"
	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/database"
	"github.com/insmtx/SingerOS/backend/interaction/eventbus/rabbitmq"
	gateway "github.com/insmtx/SingerOS/backend/interaction/gateway"
	"github.com/insmtx/SingerOS/backend/llm"
	openai_llm "github.com/insmtx/SingerOS/backend/llm/openai"
	orchestrator "github.com/insmtx/SingerOS/backend/orchestrator"
	skills "github.com/insmtx/SingerOS/backend/skills"
	code_review_skill "github.com/insmtx/SingerOS/backend/skills/tool_skills/code_review_skill"
	echo_skill "github.com/insmtx/SingerOS/backend/skills/tool_skills/echo_skill"
	"github.com/spf13/cobra"
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

		// Initialize skill manager
		skillManager := skills.NewSimpleSkillManager()

		// Register echo skill
		echoSkill := echo_skill.NewEchoSkill()
		if err := skillManager.Register(echoSkill); err != nil {
			logs.Fatalf("Failed to register echo skill: %v", err)
			return
		}

		// Register additional programming-related skills for ReAct agent
		// - Code review skill
		// - PR analysis skill
		// - Documentation search skill
		// This can happen via dynamic skill loading mechanism

		// Create and register the code review skill
		codeReviewSkill := code_review_skill.NewCodeReviewSkill()
		if err := skillManager.Register(codeReviewSkill); err != nil {
			logs.Fatalf("Failed to register code review skill: %v", err)
			return
		}

		// Create and register the PR analysis skill
		prAnalysisSkill := react.NewPRAnalysisSkill()
		if err := skillManager.Register(prAnalysisSkill); err != nil {
			logs.Fatalf("Failed to register PR analysis skill: %v", err)
			return
		}

		// Initialize LLM provider
		var llmProvider llm.Provider
		if cfg.LLM != nil && cfg.LLM.APIKey != "" {
			// Configure LLM provider based on config
			switch cfg.LLM.Provider {
			case "openai", "":
				llmConfig := openai_llm.DefaultConfig(cfg.LLM.APIKey)
				if cfg.LLM.BaseURL != "" {
					llmConfig.BaseURL = cfg.LLM.BaseURL
				}
				if cfg.LLM.Model != "" {
					// We can customize default model if needed
				}
				llmProvider = openai_llm.NewProvider(llmConfig)
			default:
				logs.Fatalf("Unsupported LLM provider: %s", cfg.LLM.Provider)
				return
			}
		} else {
			logs.Warnf("No LLM configuration provided, using mock provider for testing")
			// Create a mock provider for testing/development purposes
			llmProvider = &mockLLMProvider{}
		}

		// Create orchestrator to consume events with skill manager and LLM provider
		orchestratorInstance := orchestrator.NewOrchestrator(publisher, skillManager, llmProvider)

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
		r := gin.Default()

		// Set up gateway with connectors
		gateway.SetupRouter(r, *cfg, publisher, db)

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

func main() {
	if err := rootCmd.Execute(); err != nil {
		logs.Errorf("Error executing command: %v", err)
		os.Exit(1)
	}
}

// Mock LLM provider for development/testing when no API key is provided
type mockLLMProvider struct{}

func (m *mockLLMProvider) Name() string { return "mock" }
func (m *mockLLMProvider) Generate(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
	responseContent := "This is a mock response for development/testing. Actual LLM response would appear here."
	if len(req.Messages) > 0 {
		// Simple example of appending message content to simulate a response
		lastMessage := req.Messages[len(req.Messages)-1]
		responseContent = fmt.Sprintf("Mock processing of: %s", lastMessage.Content)
	}

	return &llm.GenerateResponse{
		Content: responseContent,
		Usage: llm.TokenUsage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
		FinishReason: "stop",
	}, nil
}

func (m *mockLLMProvider) GenerateStream(ctx context.Context, req *llm.GenerateRequest) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)

	// Send a mock response
	go func() {
		defer close(ch)
		ch <- llm.StreamChunk{
			Content: "This is a mock streaming response from development/testing.",
			Done:    true,
		}
	}()

	return ch, nil
}

func (m *mockLLMProvider) CountTokens(text string) int {
	return len(strings.Fields(text)) // Simple estimation
}

func (m *mockLLMProvider) Models() []string {
	return []string{"mock-model"}
}
