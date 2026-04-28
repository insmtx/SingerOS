// api 包提供 SingerOS 的 HTTP API 层
//
// 该包负责设置和管理 HTTP 路由，处理外部 API 请求，
// 并注册各种渠道的连接器。
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/SingerOS/backend/config"
	auth "github.com/insmtx/SingerOS/backend/internal/api/auth"
	"github.com/insmtx/SingerOS/backend/internal/api/connectors/github"
	"github.com/insmtx/SingerOS/backend/internal/api/connectors/gitlab"
	"github.com/insmtx/SingerOS/backend/internal/api/handler"
	"github.com/insmtx/SingerOS/backend/internal/api/middleware"
	eventbus "github.com/insmtx/SingerOS/backend/internal/infra/mq"
	githubprovider "github.com/insmtx/SingerOS/backend/internal/infra/providers/github"
	"github.com/insmtx/SingerOS/backend/internal/infra/websocket"
	"github.com/insmtx/SingerOS/backend/internal/service"
	workerserver "github.com/insmtx/SingerOS/backend/internal/worker/server"
	singerMCP "github.com/insmtx/SingerOS/backend/mcp"
	ygmiddleware "github.com/ygpkg/yg-go/apis/runtime/middleware"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"

	_ "github.com/insmtx/SingerOS/docs/swagger" // Swagger 文档生成的导入
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter 设置事件网关的路由，注册所有连接器
//
// 根据配置初始化并注册 GitHub、GitLab 等渠道连接器，
// 同时设置客户端 WebSocket 连接器，并将所有连接器的路由注册到 HTTP 服务器。
func SetupRouter(cfg config.Config, publisher eventbus.Publisher, db *gorm.DB) *gin.Engine {
	r := gin.New()
	{
		r.Use(ygmiddleware.CORS())
		r.Use(middleware.CallerMiddleware())
		r.Use(middleware.Logger(".Ping", "metrics"))
		r.Use(ygmiddleware.Recovery())
	}
	v1 := r.Group("/v1")

	if cfg.Github != nil {
		logs.Info("Setting up GitHub connector")
		authService := initThirdPartyAuthService(&cfg)
		github.RegisterGitHubRoutes(v1, *cfg.Github, publisher, db, authService)
		logs.Info("GitHub connector registered successfully")
	} else {
		logs.Debug("No GitHub configuration provided, skipping GitHub connector setup")
	}

	if cfg.Gitlab != nil {
		logs.Info("Setting up GitLab connector")
		gitlab.RegisterGitLabRoutes(v1, *cfg.Gitlab, publisher)
		logs.Info("GitLab connector registered successfully")
	} else {
		logs.Debug("No GitLab configuration provided, skipping GitLab connector setup")
	}

	websocket.RegisterWebSocketRoutes(v1, publisher)
	logs.Info("WebSocket connector registered successfully")

	digitalAssistantService := service.NewDigitalAssistantService(db)
	handler.RegisterDigitalAssistantRoutes(v1, digitalAssistantService)
	logs.Info("Digital assistant routes registered successfully")

	workerServer := workerserver.NewServer()
	workerServer.RegisterRoutes(v1)
	logs.Info("Worker server routes registered successfully")

	singerMCP.RegisterRoutes(v1, singerMCP.NewServer())
	logs.Info("MCP routes registered successfully")

	// Swagger UI 路由
	v1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return r
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
