// config 包提供 SingerOS 的配置加载和配置类型定义
//
// 该包负责从配置文件加载各种配置项，包括 GitHub 应用配置、
// GitLab 应用配置、RabbitMQ 消息队列配置和数据库配置等。
package config

// RabbitMQConfig 是 RabbitMQ 消息队列的配置结构
type RabbitMQConfig struct {
	URL string `yaml:"url,omitempty" json:"url,omitempty"` // RabbitMQ 服务地址
}

// Config 是 SingerOS 的主配置结构，包含所有子系统的配置
type Config struct {
	Github   *GithubAppConfig `yaml:"github,omitempty"`   // GitHub 应用配置
	Gitlab   *GitlabAppConfig `yaml:"gitlab,omitempty"`   // GitLab 应用配置
	RabbitMQ *RabbitMQConfig  `yaml:"rabbitmq,omitempty"` // RabbitMQ 消息队列配置
	Database *DatabaseConfig  `yaml:"database,omitempty"` // 数据库配置
}

// DatabaseConfig 是数据库的配置结构
type DatabaseConfig struct {
	URL   string `yaml:"url,omitempty"`   // 数据库连接地址
	Debug bool   `yaml:"debug,omitempty"` // 是否启用调试模式
}
