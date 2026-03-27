// types 包提供 SingerOS 的核心数据类型定义
//
// 该包定义了数字助手、事件、用户、技能等核心领域模型，
// 以及相关的常量和数据库表名定义。
package types

import (
	"gorm.io/gorm"
)

// TableNameUser 用户表名
const TableNameUser = "singer_users"

// User 表示系统中的用户信息
//
// 该结构存储用户的基本信息，包括 GitHub 关联信息、
// 个人资料、头像等。
type User struct {
	gorm.Model
	GithubID    int64  `gorm:"column:github_id;type:bigint;unique_index"`                   // GitHub 用户 ID
	GithubLogin string `gorm:"column:github_login;type:varchar(255);not null;unique_index"` // GitHub 登录名
	Name        string `gorm:"column:name;type:varchar(255)"`                               // 用户姓名
	Email       string `gorm:"column:email;type:varchar(255)"`                              // 用户邮箱
	AvatarURL   string `gorm:"column:avatar_url;type:varchar(500)"`                         // 头像 URL
	Bio         string `gorm:"column:bio;type:text"`                                        // 个人简介
	Company     string `gorm:"column:company;type:varchar(255)"`                            // 公司信息
	Location    string `gorm:"column:location;type:varchar(255)"`                           // 地理位置
	PublicRepos int    `gorm:"column:public_repos;type:integer"`                            // 公开仓库数量
	Followers   int    `gorm:"column:followers;type:integer"`                               // 关注者数量
}

// TableName 重写表名
func (User) TableName() string {
	return TableNameUser
}
