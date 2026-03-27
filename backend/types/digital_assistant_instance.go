// types 包提供 SingerOS 的核心数据类型定义
//
// 该包定义了数字助手、事件、用户、技能等核心领域模型，
// 以及相关的常量和数据库表名定义。
package types

import "gorm.io/gorm"

// DigitalAssistantInstance 表示数字助手的运行实例
//
// 该结构存储数字助手实例的信息，每个实例代表一个运行中的数字助手。
type DigitalAssistantInstance struct {
	gorm.Model
}

// TableName 指定DigitalAssistantInstance结构体对应的数据库表名
func (DigitalAssistantInstance) TableName() string {
	return TableNameDigitalAssistantInstance
}
