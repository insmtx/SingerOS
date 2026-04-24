// api 包提供 SingerOS 的 HTTP API 层
//
// 该包负责设置和管理 HTTP 路由，处理外部 API 请求，
// 并注册各种渠道的连接器。
package api

import (
	"github.com/gin-gonic/gin"
	auth "github.com/insmtx/SingerOS/backend/internal/api/auth"
	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/internal/api/connectors"
	"github.com/insmtx/SingerOS/backend/internal/api/connectors/github"
	"github.com/insmtx/SingerOS/backend/internal/api/connectors/gitlab"
	githubprovider "github.com/insmtx/SingerOS/backend/internal/infra/providers/github"
	"github.com/insmtx/SingerOS/backend/internal/infra/websocket"
	eventbus "github.com/insmtx/SingerOS/backend/internal/infra/mq"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// SetupRouter 设置事件网关的路由，注册所有连接器
//
// 根据配置初始化并注册 GitHub、GitLab 等渠道连接器，
// 同时设置客户端 WebSocket 连接器，并将所有连接器的路由注册到 HTTP 服务器。
func SetupRouter(r gin.IRouter, cfg config.Config, publisher eventbus.Publisher, db *gorm.DB) {
	router := connectors.NewRouter()

	authService := initThirdPartyAuthService(&cfg)

	if cfg.Github != nil {
		logs.Info("Setting up GitHub connector")
		githubConnector := github.NewConnector(*cfg.Github, publisher, db, authService)
		router.Register(githubConnector)
		logs.Info("GitHub connector registered successfully")
	} else {
		logs.Debug("No GitHub configuration provided, skipping GitHub connector setup")
	}

	if cfg.Gitlab != nil {
		logs.Info("Setting up GitLab connector")
		gitlabConnector := gitlab.NewConnector(*cfg.Gitlab, publisher)
		router.Register(gitlabConnector)
		logs.Info("GitLab connector registered successfully")
	} else {
		logs.Debug("No GitLab configuration provided, skipping GitLab connector setup")
	}

	wsConnector := websocket.NewConnector(publisher)
	router.Register(wsConnector)
	websocket.GetManager().SetConnector(wsConnector)
	logs.Info("WebSocket connector registered successfully")

	router.RegisterRoutes(r)
	logs.Info("Event gateway routes registered successfully")
}

// initThirdPartyAuthService 初始化第三方平台授权服务并注册 provider
func initThirdPartyAuthService(cfg *config.Config) *auth.ThirdPartyAuthService {
	accountStore := auth.NewInMemoryStore()
	accountResolver := auth.NewAccountResolver(accountStore)
	authService := auth.NewThirdPartyAuthService(accountStore, accountResolver)

	if cfg != nil && cfg.Github != nil {
		authService.RegisterProvider(githubprovider.NewOAuthProvider(*cfg.Github))
	}

	return authService
}
