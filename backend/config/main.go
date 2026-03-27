package config

type RabbitMQConfig struct {
	URL string `yaml:"url,omitempty" json:"url,omitempty"`
}

type Config struct {
	Github   *GithubAppConfig `yaml:"github,omitempty"`
	Gitlab   *GitlabAppConfig `yaml:"gitlab,omitempty"`
	RabbitMQ *RabbitMQConfig  `yaml:"rabbitmq,omitempty"`
	Database *DatabaseConfig  `yaml:"database,omitempty"`
}

type DatabaseConfig struct {
	URL   string `yaml:"url,omitempty"`
	Debug bool   `yaml:"debug,omitempty"`
}
