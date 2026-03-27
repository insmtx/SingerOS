// types 包提供 SingerOS 的核心数据类型定义
//
// 该包定义了数字助手、事件、用户、技能等核心领域模型，
// 以及相关的常量和数据库表名定义。
package types

// 数据库表名前缀常量
const (
	tablenamePrefix = "singer_" // 数据库表名统一前缀
)

// 数据库表名常量定义
const (
	// TableNameDigitalAssistant 数字助手表名
	TableNameDigitalAssistant = tablenamePrefix + "digital_assistant"
	// TableNameDigitalAssistantInstance 数字助手实例表名
	TableNameDigitalAssistantInstance = tablenamePrefix + "digital_assistant_instance"

	// TableNameEvent 事件表名
	TableNameEvent = tablenamePrefix + "event"

	// TableNameSkill 技能表名
	TableNameSkill = tablenamePrefix + "skill"
	// TableNameSkillLog 技能执行日志表名
	TableNameSkillLog = tablenamePrefix + "skill_execution_log"
	// TableNameSkillRegistry 技能注册表名
	TableNameSkillRegistry = tablenamePrefix + "skill_registry"
)
