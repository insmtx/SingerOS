package engines

import (
	"os"
	"sort"
	"strings"
)

// BuildBaseEnv 返回当前环境并附加额外的环境变量。
func BuildBaseEnv(extraEnv map[string]string) []string {
	builder := newEnvBuilder(os.Environ())
	builder.applyMap(extraEnv)
	return builder.slice()
}

// BuildRunEnv 为 CLI 进程组装环境变量条目，根据不同的模型提供商设置对应的 API 密钥环境变量。
func BuildRunEnv(baseEnv []string, extraEnv []string, model ModelConfig) []string {
	builder := newEnvBuilder(baseEnv)
	builder.applyEntries(extraEnv)

	switch strings.ToLower(model.Provider) {
	case "anthropic", "claude":
		builder.setIfNotEmpty("ANTHROPIC_API_KEY", model.APIKey)
		builder.setIfNotEmpty("ANTHROPIC_AUTH_TOKEN", model.APIKey)
		builder.setIfNotEmpty("ANTHROPIC_BASE_URL", model.BaseURL)
	case "openai", "codex", "deepseek", "moonshot", "qwen", "zhipu", "":
		builder.setIfNotEmpty("OPENAI_API_KEY", model.APIKey)
		builder.setIfNotEmpty("OPENAI_API_BASE", model.BaseURL)
		builder.setIfNotEmpty("OPENAI_BASE_URL", model.BaseURL)
	default:
		builder.setIfNotEmpty("OPENAI_API_KEY", model.APIKey)
		builder.setIfNotEmpty("OPENAI_BASE_URL", model.BaseURL)
	}

	return builder.slice()
}

type envBuilder struct {
	values map[string]string
}

func newEnvBuilder(entries []string) *envBuilder {
	builder := &envBuilder{
		values: make(map[string]string, len(entries)),
	}
	builder.applyEntries(entries)
	return builder
}

func (b *envBuilder) applyMap(entries map[string]string) {
	for key, value := range entries {
		key = strings.TrimSpace(key)
		if key == "" || value == "" {
			continue
		}
		b.values[key] = value
	}
}

func (b *envBuilder) applyEntries(entries []string) {
	for _, entry := range entries {
		key, value, ok := splitEnvEntry(entry)
		if !ok {
			continue
		}
		b.values[key] = value
	}
}

func (b *envBuilder) setIfNotEmpty(key string, value string) {
	key = strings.TrimSpace(key)
	if key == "" || value == "" {
		return
	}
	b.values[key] = value
}

func (b *envBuilder) slice() []string {
	keys := make([]string, 0, len(b.values))
	for key := range b.values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	env := make([]string, 0, len(keys))
	for _, key := range keys {
		env = append(env, key+"="+b.values[key])
	}
	return env[:len(env):len(env)]
}

func splitEnvEntry(entry string) (key string, value string, ok bool) {
	parts := strings.SplitN(entry, "=", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	key = strings.TrimSpace(parts[0])
	if key == "" {
		return "", "", false
	}
	return key, parts[1], true
}
