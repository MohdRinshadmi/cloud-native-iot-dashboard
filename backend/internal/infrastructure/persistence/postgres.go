// Package persistence is the PostgreSQL adapter layer: connection management,
// GORM models, migrations and repository implementations of the domain ports.
// GORM never leaks above this package.
package persistence

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/ioss/iot-dashboard/backend/internal/infrastructure/config"
)

// Connect opens the database with production pool settings, retrying briefly
// so `docker compose up` ordering doesn't matter.
func Connect(cfg config.PostgresConfig, log *slog.Logger) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
		// Map driver-specific failures (e.g. unique violations) onto GORM's
		// portable error values so repos can translate them cleanly.
		TranslateError: true,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	var db *gorm.DB
	var err error
	for attempt := 1; attempt <= 10; attempt++ {
		db, err = gorm.Open(postgres.Open(cfg.DSN()), gormCfg)
		if err == nil {
			break
		}
		log.Warn("postgres not ready, retrying",
			slog.Int("attempt", attempt), slog.String("error", err.Error()))
		time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
	}
	if err != nil {
		return nil, fmt.Errorf("postgres connect: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("postgres pool: %w", err)
	}
	// Pool sizing: bounded so a traffic spike degrades gracefully instead of
	// exhausting Postgres connections (max_connections defaults to 100).
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	return db, nil
}

// HealthChecker implements application/health.Checker for Postgres.
type HealthChecker struct{ db *gorm.DB }

// NewHealthChecker wraps the connection for readiness probing.
func NewHealthChecker(db *gorm.DB) *HealthChecker { return &HealthChecker{db: db} }

func (h *HealthChecker) Name() string { return "postgres" }

func (h *HealthChecker) Check(ctx context.Context) error {
	sqlDB, err := h.db.DB()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return sqlDB.PingContext(ctx)
}
