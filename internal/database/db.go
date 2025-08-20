package database

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/mocbotau/api-join-sound/internal/models"
)

// DB is a wrapper around the GORM database connection.
type DB struct {
	*gorm.DB
}

// NewSQLiteDB creates a new SQLite database connection with GORM.
func NewSQLiteDB(dataSource string) (*DB, error) {
	db, err := gorm.Open(sqlite.Open(dataSource), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	dbInstance := &DB{db}

	db.Exec("PRAGMA foreign_keys = ON;")

	if err := db.AutoMigrate(&models.User{}, &models.Sound{}, &models.Setting{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return dbInstance, nil
}
