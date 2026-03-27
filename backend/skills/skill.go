// skills 包提供 SingerOS 的技能系统定义和实现
//
// 技能是 SingerOS 的核心能力单元，可以是本地实现的功能，
// 也可以是远程 API 服务。该包定义了技能接口、基础实现和相关类型。
package skills

import (
	"context"
)

// SkillInputValidator 定义输入验证器接口，用于验证传入的参数
type SkillInputValidator interface {
	Validate(input map[string]interface{}) error
}

// InputSchema 定义输入结构模式信息
type InputSchema struct {
	Type       string               `json:"type"`       // schema 类型，通常为 object
	Required   []string             `json:"required"`   // 必需字段列表
	Properties map[string]*Property `json:"properties"` // 属性映射
}

// Property 定义单个属性的模式描述
type Property struct {
	Type        string      `json:"type"`                  // 数据类型 (string, number, integer, boolean, array, object)
	Title       string      `json:"title,omitempty"`       // 标题（可选）
	Description string      `json:"description,omitempty"` // 描述（可选）
	Default     interface{} `json:"default,omitempty"`     // 默认值（可选）
	Items       *Property   `json:"items,omitempty"`       // 数组子项定义（当type为array时）
	Enum        []string    `json:"enum,omitempty"`        // 枚举值清单（当type为enum时）
}

// OutputSchema 定义输出结构模式信息
type OutputSchema struct {
	Type       string               `json:"type"`       // schema 类型，通常为 object
	Required   []string             `json:"required"`   // 必需字段列表
	Properties map[string]*Property `json:"properties"` // 属性映射
}

// SkillType 定义技能类型的枚举
type SkillType string

const (
	LocalSkill  SkillType = "local"  // 本地技能
	RemoteSkill SkillType = "remote" // 远程API技能
)

// Permission 定义单个权限项的结构
type Permission struct {
	Resource string `json:"resource"` // 资源标识符
	Action   string `json:"action"`   // 动作权限 (read, write, execute, etc.)
}

// SkillInfo 定义技能基本信息结构体
type SkillInfo struct {
	ID           string       `json:"id"`               // 技能唯一标识符
	Name         string       `json:"name"`             // 技能名称
	Description  string       `json:"description"`      // 技能描述信息
	Author       string       `json:"author,omitempty"` // 技能作者（可选）
	Version      string       `json:"version"`          // 版本号
	Category     string       `json:"category"`         // 技能类别
	Icon         string       `json:"icon,omitempty"`   // 图标URL（可选）
	SkillType    SkillType    `json:"skill_type"`       // 技能类型: 本地或远程API (default: local)
	Permissions  []Permission `json:"permissions"`      // 所需权限列表
	InputSchema  InputSchema  `json:"input_schema"`     // 输入参数schema
	OutputSchema OutputSchema `json:"output_schema"`    // 输出结果schema
}

// Skill 定义技能的基本接口，提供所有实现必须遵循的方法签名
type Skill interface {
	// Info 返回技能的元信息
	Info() *SkillInfo

	// Execute 执行技能逻辑，接收上下文和输入参数，返回输出结果和错误
	Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)

	// Validate 验证输入参数是否符合预期格式
	Validate(input map[string]interface{}) error

	// GetID 获取技能的唯一标识符
	GetID() string

	// GetName 获取技能的显示名称
	GetName() string

	// GetDescription 获取对技能功能的描述
	GetDescription() string
}

// BaseSkill 提供Skill接口的基础实现，子类可嵌入此结构体来减少样板代码
type BaseSkill struct {
	InfoData *SkillInfo
}

// Info 返回技能的元信息
func (bs *BaseSkill) Info() *SkillInfo {
	return bs.InfoData
}

// GetID 获取技能的唯一标识符
func (bs *BaseSkill) GetID() string {
	return bs.InfoData.ID
}

// GetName 获取技能的显示名称
func (bs *BaseSkill) GetName() string {
	return bs.InfoData.Name
}

// GetDescription 获取对技能功能的描述
func (bs *BaseSkill) GetDescription() string {
	return bs.InfoData.Description
}

// Validate 实现基本的输入验证机制
func (bs *BaseSkill) Validate(input map[string]interface{}) error {
	// 可在此基础上扩充实现基于InputSchema的验证逻辑
	return nil
}

// Execute 方法留给具体的技能实现，这里定义默认错误
func (bs *BaseSkill) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// 基础类不直接执行操作，需由具体实现覆盖
	return nil, NotImplementedError{"BaseSkill does not implement Execute method"}
}

// NotImplementedError 当某个方法未被正确实现时抛出错误
type NotImplementedError struct {
	Message string
}

func (e NotImplementedError) Error() string {
	return e.Message
}

// SkillManager 定义技能管理器接口，用于注册、管理和执行技能
type SkillManager interface {
	Register(skill Skill) error
	Get(skillID string) (Skill, error)
	List() []Skill
	Execute(ctx context.Context, skillID string, input map[string]interface{}) (map[string]interface{}, error)
}

// ExecutionContext 表示技能执行上下文，包含一些公共上下文信息
type ExecutionContext struct {
	Context    context.Context
	SessionID  string
	UserID     string
	Assistant  string
	Parameters map[string]string
}
