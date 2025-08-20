package db

import (
	"database/sql"
	"fmt"

	migrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

// OpenAndMigrate opens SQLite database and runs migrations from a directory path.
func OpenAndMigrate(dsn string, migrationsDir string, log *zap.Logger) (*sql.DB, error) {
	database, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Ensure connection works
	if err = database.Ping(); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	driver, err := sqlite.WithInstance(database, &sqlite.Config{})
	if err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("migrate driver: %w", err)
	}

	migrationsPath := fmt.Sprintf("file://%s", migrationsDir)
	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "sqlite", driver)
	if err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("migrate init: %w", err)
	}

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		_ = database.Close()
		return nil, fmt.Errorf("migrate up: %w", err)
	}

	log.Info("database ready")
	return database, nil
}


