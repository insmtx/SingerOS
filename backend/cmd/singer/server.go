// @title SingerOS API
// @version 1.0
// @description SingerOS 数字助手平台 API，提供数字助手管理、技能调用、事件处理等功能
// @host localhost:8080
// @BasePath /v1
// @schemes http https
package main

import (
	"fmt"
	"net/http"

	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/internal/api"
	infradb "github.com/insmtx/SingerOS/backend/internal/infra/db"
	"github.com/insmtx/SingerOS/backend/internal/infra/mq"
	"github.com/spf13/cobra"
	ygconfig "github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/lifecycle"
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
	Long:  `Start the HTTP server that handles API requests and publishes external events.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := loadConfig(serverConfigPath)
		if err != nil {
			logs.Fatalf("Failed to load config: %v", err)
			return
		}

		natsUrl := "nats://nats:4222"
		if cfg.NATS != nil && cfg.NATS.URL != "" {
			natsUrl = cfg.NATS.URL
		}

		publisher, err := mq.NewPublisher(natsUrl)
		if err != nil {
			logs.Fatalf("Failed to create event publisher: %v", err)
			return
		}

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

		r := api.SetupRouter(*cfg, publisher, db)

		srv := &http.Server{
			Addr:    serverHttpAddr,
			Handler: r,
		}

		logs.Info("Starting SingerOS backend service...")
		logs.Infof("Listening on %s", serverHttpAddr)

		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logs.Fatalf("Failed to start server: %v", err)
			}
		}()

		lifecycle.Std().AddCloseFunc(func() error {
			if err := srv.Shutdown(cmd.Context()); err != nil {
				logs.Errorf("Server forced to shutdown: %v", err)
			}
			return nil
		})

		lifecycle.Std().AddCloseFunc(publisher.Close)
		lifecycle.Std().WaitExit()

		logs.Info("Server exited")
	},
}

func init() {
	serverCmd.Flags().StringVar(&serverConfigPath, "config", "", "Configuration file path")
	serverCmd.Flags().StringVar(&serverHttpAddr, "addr", ":8080", "HTTP server address")
	rootCmd.AddCommand(serverCmd)
}

func loadConfig(configPath string) (*config.Config, error) {
	var cfg config.Config

	if configPath != "" {
		err := ygconfig.LoadYamlLocalFile(configPath, &cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to load config from %s: %v", configPath, err)
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
		}
	}

	logs.Info("Configuration loaded successfully")
	return &cfg, nil
}
