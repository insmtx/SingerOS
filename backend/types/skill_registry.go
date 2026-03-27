// types 包提供 SingerOS 的核心数据类型定义
//
// 该包定义了数字助手、事件、用户、技能等核心领域模型，
// 以及相关的常量和数据库表名定义。
package types

import "gorm.io/gorm"

// SkillRegistry 记录技能在系统中的注册信息
type SkillRegistry struct {
	gorm.Model
	// 关联的技能ID
	SkillID uint `gorm:"column:skill_id;type:integer;not null;uniqueIndex"`
	// 技能服务地址（如果是远程技能）
	ServiceEndpoint string `gorm:"column:service_endpoint;type:varchar(500)"`
	// 技能执行路径
	ExecutionPath string `gorm:"column:execution_path;type:varchar(500)"`
	// 技能认证令牌
	Token string `gorm:"column:token;type:varchar(255)"`
	// 注册状态（registered, unregistered, unhealthy等）
	Status string `gorm:"column:status;type:varchar(50);not null;default:registered"` // 建议使用 types.SkillRegistryStatus 定义的常量值
	// 最后心跳时间
	LastHeartbeat string `gorm:"column:last_heartbeat;type:timestamp"`
	// 健康状态
	IsHealthy bool `gorm:"column:is_healthy;type:boolean;default:true"`
}

// TableName 指定SkillRegistry对应的数据库表名
func (SkillRegistry) TableName() string {
	return TableNameSkillRegistry
}
