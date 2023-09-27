package migrations

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var migrationsFS embed.FS

// newMigrator new Postgres based DB migrator.
// migrationsFS is the embedded filesystem using 'go:embed *.sql' (no subdirectories).
func newMigrator(migrationsFS embed.FS, db *sql.DB) (*migrate.Migrate, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	sourceDriver, err := iofs.New(migrationsFS, ".")
	if err != nil {
		return nil, fmt.Errorf("unable to create 'iofs' source driver: %w", err)
	}

	migrator, err := migrate.NewWithInstance("iofs", sourceDriver, "", driver)
	if err != nil {
		return nil, fmt.Errorf("unable to instantiate migration instance: %w", err)
	}
	return migrator, nil
}

// Migrate the database to a new version.
func Migrate(logger *slog.Logger, db *sql.DB) error {
	migrator, err := newMigrator(migrationsFS, db)
	if err != nil {
		return err
	}
	//defer migrator.Close()

	if err := migrator.Up(); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return convertErrNotExist(migrator)
		}

		if !errors.Is(err, migrate.ErrNoChange) {
			return err
		}

		logger.Info("Database schema is up to date")
		return nil
	}
	logger.Info("Database schema migrated to the latest")

	return nil
}

func convertErrNotExist(migration *migrate.Migrate) error {
	version, _, err := migration.Version()
	if err != nil {
		return errors.New("schema version migration file is not found")
	}
	return fmt.Errorf("schema version '%d' migration file is not found", version)
}
