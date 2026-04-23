// gitlab 包提供 GitLab 平台的连接器实现
//
// 该包实现了与 GitLab 平台的集成，包括 Webhook 事件接收、
// OAuth 认证流程等功能。
package gitlab

import (
	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/internal/connectors"
	eventbus "github.com/insmtx/SingerOS/backend/internal/infra/mq"
)

var _ connectors.Connector = (*GitlabConnector)(nil)

// GitlabConnector 是 GitLab 平台的连接器实现
type GitlabConnector struct {
	config    config.GitlabAppConfig // GitLab 应用配置
	publisher eventbus.Publisher     // 事件发布者
}

// NewConnector 创建一个新的 GitLab 连接器实例
func NewConnector(cfg config.GitlabAppConfig, publisher eventbus.Publisher) *GitlabConnector {
	return &GitlabConnector{
		config:    cfg,
		publisher: publisher,
	}
}

// ChannelCode 返回 GitLab 渠道的标识符
func (c *GitlabConnector) ChannelCode() string {
	return "gitlab"
}
