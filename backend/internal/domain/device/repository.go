package device

import "context"

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
}

// Filter narrows a device listing. Zero values mean "no constraint".
type Filter struct {
	// Q matches name or model (case-insensitive substring).
	Q string
	// Status restricts to a single connectivity state.
	Status Status
	Page   Page
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
