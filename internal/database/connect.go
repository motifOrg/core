package database

import (
	config "core/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	return gorm.Open(
		postgres.Open(cfg.GetDatabaseDSN()),
		&gorm.Config{TranslateError: true},
	)
}
