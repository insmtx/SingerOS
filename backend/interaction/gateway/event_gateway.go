package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/SingerOS/backend/clientmgr"
	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/interaction"
	"github.com/insmtx/SingerOS/backend/interaction/connectors/client"
	"github.com/insmtx/SingerOS/backend/interaction/connectors/github"
	"github.com/insmtx/SingerOS/backend/interaction/connectors/gitlab"
	"github.com/insmtx/SingerOS/backend/interaction/eventbus"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

func SetupRouter(r gin.IRouter, cfg config.Config, publisher eventbus.Publisher, db *gorm.DB) {
	registry := interaction.NewRegistry()

	// Check if GitHub configuration is provided and enabled
	if cfg.Github != nil {
		logs.Info("Setting up GitHub connector")
		githubConnector := github.NewConnector(*cfg.Github, publisher, db)
		registry.Register(githubConnector)
		logs.Info("GitHub connector registered successfully")
	} else {
		logs.Debug("No GitHub configuration provided, skipping GitHub connector setup")
	}

	// Check if GitLab configuration is provided and enabled
	if cfg.Gitlab != nil {
		logs.Info("Setting up GitLab connector")
		gitlabConnector := gitlab.NewConnector(*cfg.Gitlab, publisher)
		registry.Register(gitlabConnector)
		logs.Info("GitLab connector registered successfully")
	} else {
		logs.Debug("No GitLab configuration provided, skipping GitLab connector setup")
	}

	// Register client WebSocket connector
	clientConnector := client.NewConnector(publisher)
	// Type assert to get the actual connector implementation
	actualConnector, ok := clientConnector.(*client.ClientConnector)
	if !ok {
		logs.Errorf("Failed to type assert client connector")
	} else {
		// Initialize client manager with the connector
		clientManager := clientmgr.GetDefaultManager()
		clientManager.SetClientConnector(actualConnector)
	}
	registry.Register(clientConnector)
	logs.Info("Client WebSocket connector registered successfully")

	registry.RegisterRoutes(r)
	logs.Info("Event gateway routes registered successfully")
}
