package database

import (
	"database/sql"
	"fmt"

	config "core/internal/config"
	logger "core/pkg/log"

	_ "github.com/lib/pq"
	"gorm.io/gorm"
)

func Init(cfg *config.DatabaseConfig) *gorm.DB {
	logger.Info("Ensuring database exists")
	err := Ensure(cfg)
	logger.Fatal(err, "Failed to ensure database exists")

	db, err := Connect(cfg)
	logger.Fatal(err, "Failed to connect to database")

	logger.Info("Running migrations")
	err = Migrate(cfg)
	logger.Fatal(err, "Failed to migrate database")

	return db
}

func Ensure(cfg *config.DatabaseConfig) error {
	db, err := sql.Open("postgres", cfg.GetAdminDSN())
	if err != nil {
		return err
	}
	defer db.Close()

	var exists bool
	err = db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
		cfg.Name,
	).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		logger.Info("Database '%s' already exists", cfg.Name)
		return nil
	}

	logger.Info("Creating database '%s'", cfg.Name)
	_, err = db.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, cfg.Name))
	return err
}
