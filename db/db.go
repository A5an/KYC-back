package db

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/lib/pq"

	"github.com/Sinbad-HQ/kyc/config"
	"github.com/Sinbad-HQ/kyc/db/migrations"
)

// Connect establishes a connection to the database using the provided database configuration.
func Connect(logger *slog.Logger, cfg config.DatabaseConfig) (*sqlx.DB, error) {
	db, err := openConnection(logger, cfg)
	if err != nil {
		return nil, err
	}

	if err := setupDatabase(logger, cfg, db); err != nil {
		return nil, err
	}

	return db, nil
}

func openConnection(logger *slog.Logger, cfg config.DatabaseConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", cfg.URL())
	if err != nil {
		logger.Error(fmt.Sprintf("Could not connect to address: %v due to error: %v ", cfg.URL(), err))
		return nil, err
	}
	logger.Info("Connected to postgres on address: " + cfg.URL())
	db.SetMaxIdleConns(20)
	db.SetMaxOpenConns(100)

	// Create a new mapper which will use the struct field tag "json" instead of "db"
	db.Mapper = reflectx.NewMapperFunc("json", strings.ToLower)
	return db, nil
}

func setupDatabase(logger *slog.Logger, cfg config.DatabaseConfig, dbConn *sqlx.DB) error {
	logger.Info("Creating database with name " + cfg.DbName)
	if err := UpsertDB(dbConn, cfg.DbName); err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	logger.Info("Migrating database to latest")
	if err := migrations.Migrate(logger, dbConn.DB); err != nil {
		return err
	}

	return nil
}

// UpsertDB creates a new db from a name.
func UpsertDB(db *sqlx.DB, name string) error {
	_, err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", name))
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "42P04" {
			// report no error if db already exists
			return nil
		}
		return err
	}
	return nil
}
