package migrations

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migration_files/*.sql
var migrationsDir embed.FS

func RunMigration(dsn string) error {
	d, err := iofs.New(migrationsDir, "migration_files")
	if err != nil {
		return fmt.Errorf("failed to return a FS drive: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return  fmt.Errorf("failed to return a new migrate: %w", err)
	}

	if err := m.Up(); err != nil {
		if !errors.Is(err,  migrate.ErrNoChange) {
			return fmt.Errorf("failed to load migrations: %w", err)
		}
	}

	return nil
}