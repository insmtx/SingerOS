package config

type GithubAppConfig struct {
	AppID             int64  `yaml:"app_id"`
	PrivateKey        string `yaml:"private_key"`
	WebhookSecret     string `yaml:"webhook_secret"`
	BaseURL           string `yaml:"base_url"`
	ClientID          string `yaml:"client_id"`
	ClientSecret      string `yaml:"client_secret"`
	SkipWebhookVerify bool   `yaml:"skip_webhook_verify"` // 跳过webhook签名验证（仅用于开发）
}
