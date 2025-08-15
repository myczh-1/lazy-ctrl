package database

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/myczh-1/lazy-ctrl-cloud/internal/config"
	"github.com/myczh-1/lazy-ctrl-cloud/internal/model"
)

// Connect establishes a database connection
func Connect(cfg config.DatabaseConfig) (*gorm.DB, error) {
	// Use SQLite for simplicity
	dbPath := "data/lazy_ctrl_cloud.db"
	
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	
	// Auto migrate schemas
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	
	return db, nil
}

// autoMigrate performs automatic database migration
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Device{},
		&model.DeviceCommand{},
		&model.UserDevice{},
		&model.ExecutionLog{},
	)
}