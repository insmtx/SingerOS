// database 包提供 SingerOS 的数据库初始化和管理功能
//
// 该包负责数据库连接的初始化、表结构的自动迁移，
// 以及提供获取数据库实例的方法。
package database

import (
	"fmt"

	"github.com/ygpkg/yg-go/dbtools"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"

	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/types"
)

// dbName 是数据库名称常量
const dbName = "singer"

// InitDB 创建并初始化数据库连接
//
// 使用 dbtools 初始化数据库连接，并根据配置决定是否启用调试模式，
// 最后运行数据库迁移来创建所有必要的表结构。
func InitDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	db, err := dbtools.InitDBConn(dbName, cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	if cfg.Debug {
		db = db.Debug()
	}

	// 运行数据库迁移
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	logs.Info("Database connection initialized successfully")
	return db, nil
}

// runMigrations 为所有模型创建数据库表
//
// 该函数会自动为所有定义的模型创建或更新数据库表结构。
func runMigrations(db *gorm.DB) error {
	models := []interface{}{
		&types.User{},
		&types.Event{},
		&types.DigitalAssistant{},
		&types.Skill{},
		&types.SkillRegistry{},
		&types.SkillExecutionLog{},
		&types.DigitalAssistantInstance{},
	}

	if err := dbtools.InitModel(db, models...); err != nil {
		return err
	}

	logs.Info("Database migrations completed")
	return nil
}

// GetDB 获取默认的数据库实例
func GetDB() *gorm.DB {
	return dbtools.DB(dbName)
}
