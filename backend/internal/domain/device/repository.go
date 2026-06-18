package device

import (
	"context"
	"time"
)

// Repository is the persistence PORT for the Device aggregate. The domain
// declares the interface it needs; infrastructure (Phase 3, GORM/Postgres)
// provides the adapter. This inversion is what keeps the domain testable
// and storage-agnostic.
//
// All reads are tenant-scoped: callers must pass the tenantID so the adapter
// can enforce row-level isolation.
type Repository interface {
	Create(ctx context.Context, d *Device) error
	GetByID(ctx context.Context, tenantID, id string) (*Device, error)
	Update(ctx context.Context, d *Device) error
	Delete(ctx context.Context, tenantID, id string) error
	List(ctx context.Context, tenantID string, f Filter) ([]*Device, int64, error)
	// CountByStatus returns the device count per status for one tenant, in a
	// single GROUP BY query (the fleet-summary aggregate).
	CountByStatus(ctx context.Context, tenantID string) (map[Status]int64, error)
}

// Filter narrows a device listing. Zero values mean "no constraint".
type Filter struct {
	// Q matches name or model (case-insensitive substring).
	Q string
	// Status restricts to a single connectivity state.
	Status Status
	Page   Page
}

// StatusChange reports a connectivity transition produced by the offline
// sweep — enough information to broadcast the event tenant-scoped.
type StatusChange struct {
	DeviceID   string
	TenantID   string
	Status     Status
	LastSeenAt *time.Time
}

// Page is a simple, validated pagination request.
type Page struct {
	Limit  int
	Offset int
}

// Normalize clamps pagination to safe bounds (defends the DB from abuse).
func (p Page) Normalize() Page {
	if p.Limit <= 0 || p.Limit > 200 {
		p.Limit = 50
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
	return p
}
