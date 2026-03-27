// types 包提供 SingerOS 的核心数据类型定义
//
// 该包定义了数字助手、事件、用户、技能等核心领域模型，
// 以及相关的常量和数据库表名定义。
package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

// DigitalAssistant 数字助手结构体定义了AI数字助手的基本信息与配置
type DigitalAssistant struct {
	gorm.Model
	// digital_assistant - 助手标识符，在组织内唯一标识数字助手，VARCHAR(255)，NOT NULL
	Code string `gorm:"column:code;type:varchar(255);not null;index"`
	// digital_assistant - 所属组织ID，INTEGER，NOT NULL
	OrgID uint `gorm:"column:org_id;type:integer;not null;index"`
	// digital_assistant - 拥有者ID，INTEGER，NOT NULL
	OwnerID uint `gorm:"column:owner_id;type:integer;not null;index"`

	// digital_assistant - 数字助手名称，VARCHAR(255)，NOT NULL
	Name string `gorm:"column:name;type:varchar(255);not null"`

	// digital_assistant - 描述信息，TEXT，允许为空
	Description string `gorm:"column:description;type:text"`
	// digital_assistant - 头像URL地址，VARCHAR(500)，允许为空
	Avatar string `gorm:"column:avatar;type:varchar(500)"`

	// digital_assistant - 状态，表示数字助手当前运行状态，VARCHAR(50)，NOT NULL
	Status string `gorm:"column:status;type:varchar(50);not null"` // 建议使用 types.DigitalAssistantStatus 定义的常量值
	// digital_assistant - 版本号，跟踪配置变动版本，INTEGER，默认值0
	Version int `gorm:"column:version;type:integer;default:0"`

	// digital_assistant - 配置项，包含完整的数字助手配置信息，JSON，NOT NULL
	Config AssistantConfig `gorm:"column:config;type:json;not null"`
}

// TableName 指定DigitalAssistant结构体对应的数据库表名
func (DigitalAssistant) TableName() string {
	return TableNameDigitalAssistant
}

// AssistantConfig 数字助手配置结构定义了数字助手的运行时、LLM、技能等配置
type AssistantConfig struct {
	// 运行时配置 - 定义执行环境类型和参数
	Runtime RuntimeConfig `json:"runtime_config"`
	// LLM配置 - 定义大型语言模型的类型和参数
	LLM LLMConfig `json:"llm_config"`
	// 技能引用列表 - 数字助手能够使用的技能集
	Skills []SkillRef `json:"skills"`
	// 渠道引用列表 - 数字助手集成的通信渠道
	Channels []ChannelRef `json:"channels"`
	// 知识库引用列表 - 数字助手可访问的知识资源
	Knowledge []KnowledgeRef `json:"knowledge"`
	// 记忆配置 - 定义数字助手的记忆类型和参数
	Memory MemoryConfig `json:"memory_config"`
	// 策略配置 - 定义数字助手的安全策略
	Policies PolicyConfig `json:"policies_config"`
}

var _ driver.Valuer = (*AssistantConfig)(nil)

// Scan 实现 sql.Scanner 接口，用于从数据库中读取 JSON 数据并解析为 AssistantConfig 结构体
func (ac *AssistantConfig) Scan(value any) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into AssistantConfig", value)
	}

	return json.Unmarshal(bytes, ac)
}

// Value 实现 driver.Valuer 接口，用于将 AssistantConfig 结构体转换为 JSON 数据存储到数据库中
func (ac AssistantConfig) Value() (driver.Value, error) {
	return json.Marshal(ac)
}

// SkillRef 技能引用定义了数字助手所使用的技能信息
type SkillRef struct {
	// 技能引用 - 技能代码，技能的唯一标识符
	SkillCode string `json:"skill_code"`
	// 技能引用 - 版本号，使用技能的指定版本
	Version string `json:"version"`
	// 技能引用 - 自定义配置，针对特定技能的配置选项
	Config map[string]any `json:"config"`
}

// ChannelRef 渠道引用定义了数字助手所使用的交互渠道信息
type ChannelRef struct {
	// 渠道引用 - 类型，渠道类型标识 (如：GitHub, GitLab, WeChat等)
	Type string `json:"type"` // 建议使用 types.ChannelType 定义的常量值
	// 渠道引用 - 配置，渠道的自定义配置选项
	Config map[string]any `json:"config"`
}

// KnowledgeRef 知识库引用定义了数字助手可访问的知识资源信息
type KnowledgeRef struct {
	// 知识库引用 - 类型，知识库类型标识
	Type string `json:"type"` // 建议使用 types.KnowledgeType 定义的常量值
	// 知识库引用 - 数据集ID，目标数据集的唯一标识符
	DatasetID string `json:"dataset_id"`
	// 知识库引用 - 仓库信息，仓库路径或关联数据源
	Repo string `json:"repo"`
}

// RuntimeConfig 运行时配置定义了执行环境的类型和参数
type RuntimeConfig struct {
	// 运行时配置 - 类型，运行时环境类型标识 (如：docker, process等)
	Type string `json:"type"` // 建议使用 types.RuntimeType 定义的常量值
	// 运行时配置 - 配置，运行时的自定义配置选项
	Config map[string]any `json:"config"`
}

// LLMConfig LLM配置定义了大型语言模型的类型和参数
type LLMConfig struct {
	// LLM配置 - 类型，LLM提供商类型标识 (如：openai, claude, deepseek等)
	Type string `json:"type"` // 建议使用 types.LLMProviderType 定义的常量值
	// LLM配置 - 配置，LLM相关自定义配置选项
	Config map[string]any `json:"config"`
}

// MemoryConfig 记忆配置定义了记忆功能的类型和参数
type MemoryConfig struct {
	// 记忆配置 - 类型，记忆存储类型标识 (如：redis, postgres等)
	Type string `json:"type"` // 建议使用 types.MemoryType 定义的常量值
	// 记忆配置 - 配置，记忆相关的自定义配置选项
	Config map[string]any `json:"config"`
}

// PolicyConfig 策略配置定义了权限与安全策略的功能类型和参数
type PolicyConfig struct {
	// 策略配置 - 类型，策略类型标识
	Type string `json:"type"`
	// 策略配置 - 配置，策略的自定义配置选项
	Config map[string]any `json:"config"`
}
