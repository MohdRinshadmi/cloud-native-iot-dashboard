package persistence

import (
	"embed"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// Migrate applies all pending SQL migrations. Files are embedded in the
// binary, so the container needs no sidecar volume — the API migrates itself
// at boot (safe: golang-migrate takes a Postgres advisory lock, so concurrent
// replicas don't race).
func Migrate(db *gorm.DB, log *slog.Logger) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	src, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("migrate source: %w", err)
	}
	driver, err := migratepg.WithInstance(sqlDB, &migratepg.Config{})
	if err != nil {
		return fmt.Errorf("migrate driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate init: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}

	version, dirty, _ := m.Version()
	log.Info("database migrated", slog.Uint64("version", uint64(version)), slog.Bool("dirty", dirty))
	return nil
}
