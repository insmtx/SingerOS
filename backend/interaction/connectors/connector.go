// connectors 包提供不同交互渠道的连接器实现
//
// 该包定义了 Connector 接口，并提供了 GitHub、GitLab、企业微信等
// 多种渠道的连接器实现，用于接收和处理来自不同平台的事件。
package connectors

import "github.com/gin-gonic/gin"

// Connector 是渠道连接器的接口，定义了与外部渠道交互的标准方法
//
// 每个渠道（如 GitHub、GitLab、企业微信等）都需要实现该接口，
// 以便能够接收和处理来自该渠道的事件。
type Connector interface {
	// ChannelCode 返回该连接器对应的渠道代码（如 "github"、"gitlab"）
	ChannelCode() string

	// RegisterRoutes 为该连接器注册 HTTP 路由，用于接收 Webhook 等事件
	RegisterRoutes(r gin.IRouter)
}
