package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/wywyy3cee/tgbot-anime-tracker/pkg/logger"
)

type Database struct {
	DB     *sqlx.DB
	logger *logger.Logger
}

func Connect(databaseURL string, logger *logger.Logger) (*Database, error) {
	logger.Info("Connecting to database...")

	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established")

	return &Database{
		DB:     db,
		logger: logger,
	}, nil
}

func (d *Database) RunMigrations(migrationsDir string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(d.DB.DB, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func (d *Database) Close() error {
	return d.DB.Close()
}
