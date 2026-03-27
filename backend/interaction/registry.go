// interaction 包提供事件驱动的交互层功能
//
// 该包负责事件的定义、分发和处理，是 SingerOS 的核心交互层。
// 支持多种渠道的事件接入，并通过事件总线进行分发。
package interaction

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/SingerOS/backend/interaction/connectors"
)

// Registry 是连接器注册表，用于管理不同渠道的连接器
type Registry struct {
	connectors map[string]connectors.Connector // 渠道代码到连接器的映射
}

// NewRegistry 创建一个新的连接器注册表
func NewRegistry() *Registry {

	return &Registry{
		connectors: map[string]connectors.Connector{},
	}
}

// Register 向注册表中注册一个连接器
func (r *Registry) Register(c connectors.Connector) {
	r.connectors[c.ChannelCode()] = c
}

// RegisterRoutes 为所有已注册的连接器注册 HTTP 路由
func (r *Registry) RegisterRoutes(
	router gin.IRouter,
) {

	for _, c := range r.connectors {
		c.RegisterRoutes(router)
	}
}
