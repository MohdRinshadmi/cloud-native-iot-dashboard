package persistence

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	appauth "github.com/ioss/iot-dashboard/backend/internal/application/auth"
	"github.com/ioss/iot-dashboard/backend/internal/domain/device"
	"github.com/ioss/iot-dashboard/backend/internal/domain/tenant"
	"github.com/ioss/iot-dashboard/backend/internal/domain/user"
)

// SeedDev provisions a demo workspace for local development ONLY (guarded by
// the caller on APP_ENV). Idempotent: if the demo tenant exists, it no-ops.
//
//	admin@demo.local    / Password123!  (admin)
//	operator@demo.local / Password123!  (operator)
//	viewer@demo.local   / Password123!  (viewer)
func SeedDev(ctx context.Context, db *gorm.DB, hasher appauth.PasswordHasher, log *slog.Logger) error {
	tenants := NewTenantRepository(db)
	if _, err := tenants.GetBySlug(ctx, "demo"); err == nil {
		return nil // already seeded
	}

	now := time.Now().UTC()
	t, err := tenant.New(uuid.NewString(), "Demo Industries", "demo", now)
	if err != nil {
		return err
	}
	if err := tenants.Create(ctx, t); err != nil {
		return err
	}

	hash, err := hasher.Hash("Password123!")
	if err != nil {
		return err
	}

	users := NewUserRepository(db)
	for _, spec := range []struct {
		email string
		name  string
		role  user.Role
	}{
		{"admin@demo.local", "Demo Admin", user.RoleAdmin},
		{"operator@demo.local", "Demo Operator", user.RoleOperator},
		{"viewer@demo.local", "Demo Viewer", user.RoleViewer},
	} {
		u, err := user.New(uuid.NewString(), t.ID, spec.email, spec.name, hash, spec.role, now)
		if err != nil {
			return err
		}
		if err := users.Create(ctx, u); err != nil {
			return err
		}
	}

	// A small fleet with realistic variety so dashboards aren't empty.
	devices := NewDeviceRepository(db)
	specs := []struct {
		name, model, fw string
		status          device.Status
		seenAgo         time.Duration // 0 = never seen
	}{
		{"pump-station-01", "PX-1000", "2.4.1", device.StatusOnline, 30 * time.Second},
		{"pump-station-02", "PX-1000", "2.4.1", device.StatusOnline, 45 * time.Second},
		{"chiller-unit-north", "CH-500", "1.9.0", device.StatusDegraded, 2 * time.Minute},
		{"conveyor-line-a", "CV-220", "3.1.2", device.StatusOnline, 15 * time.Second},
		{"conveyor-line-b", "CV-220", "3.0.8", device.StatusOffline, 6 * time.Hour},
		{"hvac-rooftop-1", "HV-800", "5.2.0", device.StatusOnline, 50 * time.Second},
		{"gateway-warehouse", "GW-X2", "4.0.3", device.StatusOffline, 26 * time.Hour},
		{"sensor-array-dock", "SA-50", "1.2.7", device.StatusProvisioning, 0},
	}
	for _, s := range specs {
		d, err := device.NewDevice(uuid.NewString(), t.ID, s.name, s.model, now)
		if err != nil {
			return err
		}
		d.Firmware = s.fw
		d.Status = s.status
		if s.seenAgo > 0 {
			seen := now.Add(-s.seenAgo)
			d.LastSeenAt = &seen
		}
		if err := devices.Create(ctx, d); err != nil {
			return fmt.Errorf("seed device %s: %w", s.name, err)
		}
	}

	log.Info("dev data seeded",
		slog.String("tenant", "demo"),
		slog.String("admin", "admin@demo.local / Password123!"),
		slog.Int("devices", len(specs)),
	)
	return nil
}
