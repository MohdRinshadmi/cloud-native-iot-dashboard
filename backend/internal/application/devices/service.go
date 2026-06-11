// Package devices contains the device-management use-cases. The service
// depends only on domain ports — no GORM, no Gin, no Redis.
package devices

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ioss/iot-dashboard/backend/internal/domain/device"
	"github.com/ioss/iot-dashboard/backend/internal/shared/apperror"
)

// Service implements device CRUD use-cases. All operations are tenant-scoped:
// the tenant id comes from the authenticated principal, never from the client
// payload, so cross-tenant access is structurally impossible.
type Service struct {
	repo device.Repository
	now  func() time.Time
}

// NewService wires the service. `now` is injected for deterministic tests.
func NewService(repo device.Repository, now func() time.Time) *Service {
	return &Service{repo: repo, now: now}
}

// CreateInput carries validated-at-the-edge fields for device registration.
type CreateInput struct {
	TenantID string
	Name     string
	Model    string
	Firmware string
}

// Create registers a new device in the provisioning state.
func (s *Service) Create(ctx context.Context, in CreateInput) (*device.Device, error) {
	d, err := device.NewDevice(uuid.NewString(), in.TenantID, in.Name, in.Model, s.now())
	if err != nil {
		return nil, err
	}
	d.Firmware = in.Firmware
	if err := s.repo.Create(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

// Get fetches one device within the tenant boundary.
func (s *Service) Get(ctx context.Context, tenantID, id string) (*device.Device, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

// ListInput narrows and paginates a device listing.
type ListInput struct {
	TenantID string
	Q        string
	Status   string
	Limit    int
	Offset   int
}

// List returns a page of devices plus the total match count.
func (s *Service) List(ctx context.Context, in ListInput) ([]*device.Device, int64, error) {
	f := device.Filter{
		Q:    in.Q,
		Page: device.Page{Limit: in.Limit, Offset: in.Offset}.Normalize(),
	}
	if in.Status != "" {
		st := device.Status(in.Status)
		if !st.Valid() {
			return nil, 0, apperror.InvalidInput("unknown status filter")
		}
		f.Status = st
	}
	return s.repo.List(ctx, in.TenantID, f)
}

// UpdateInput uses pointers for PATCH semantics: nil = leave unchanged.
type UpdateInput struct {
	Name     *string
	Model    *string
	Firmware *string
}

// Update applies a partial update to descriptive fields.
func (s *Service) Update(ctx context.Context, tenantID, id string, in UpdateInput) (*device.Device, error) {
	d, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	name := d.Name
	if in.Name != nil {
		name = *in.Name
	}
	model, firmware := "", ""
	if in.Model != nil {
		model = *in.Model
	}
	if in.Firmware != nil {
		firmware = *in.Firmware
	}
	if err := d.Rename(name, model, firmware, s.now()); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

// Delete removes a device permanently. (Soft-delete/decommission flows arrive
// with the device-lifecycle work in Phase 7.)
func (s *Service) Delete(ctx context.Context, tenantID, id string) error {
	return s.repo.Delete(ctx, tenantID, id)
}
