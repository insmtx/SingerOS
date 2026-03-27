package types

import (
	"gorm.io/gorm"
)

const TableNameUser = "singer_users"

type User struct {
	gorm.Model
	GithubID    int64  `gorm:"column:github_id;type:bigint;unique_index"`
	GithubLogin string `gorm:"column:github_login;type:varchar(255);not null;unique_index"`
	Name        string `gorm:"column:name;type:varchar(255)"`
	Email       string `gorm:"column:email;type:varchar(255)"`
	AvatarURL   string `gorm:"column:avatar_url;type:varchar(500)"`
	Bio         string `gorm:"column:bio;type:text"`
	Company     string `gorm:"column:company;type:varchar(255)"`
	Location    string `gorm:"column:location;type:varchar(255)"`
	PublicRepos int    `gorm:"column:public_repos;type:integer"`
	Followers   int    `gorm:"column:followers;type:integer"`
}

// TableName overrides the table name
func (User) TableName() string {
	return TableNameUser
}
