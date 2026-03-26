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
	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/interaction/eventbus/rabbitmq"
	gateway "github.com/insmtx/SingerOS/backend/interaction/gateway"
	orchestrator "github.com/insmtx/SingerOS/backend/orchestrator"
	"github.com/spf13/cobra"
	ygconfig "github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/logs"
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

		// Initialize event bus using environment variable or docker-compose defaults
		rmqUrl := os.Getenv("RABBITMQ_URL")
		if rmqUrl == "" {
			rmqUrl = "amqp://singer_user:singer_password@rabbitmq:5672/"
		}

		rmqCfg := ygconfig.RabbitMQConfig{URL: rmqUrl}
		publisher, _, err := rabbitmq.NewPublisher(rmqCfg)
		if err != nil {
			logs.Fatalf("Failed to create event publisher: %v", err)
			return
		}

		// Create orchestrator to consume events
		orchestratorInstance := orchestrator.NewOrchestrator(publisher)

		// Set up the HTTP router
		r := gin.Default()

		// Set up gateway with connectors
		gateway.SetupRouter(r, *cfg, publisher)

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
		// Try to load from default path or environment
		err := ygconfig.LoadYamlLocalFile("./config.yaml", &cfg)
		if err != nil {
			logs.Warnf("Could not load config from './config.yaml': %v, will proceed without config", err)
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
