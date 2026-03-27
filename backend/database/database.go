package database

import (
	"fmt"

	"github.com/ygpkg/yg-go/dbtools"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"

	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/types"
)

const dbName = "singer"

// InitDB creates and initializes the database connection using dbtools.
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

// runMigrations creates tables for all models.
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

// GetDB returns the default database instance.
func GetDB() *gorm.DB {
	return dbtools.DB(dbName)
}
