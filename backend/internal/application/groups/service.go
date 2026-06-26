// Package groups contains device-group (fleet) use-cases: CRUD plus the
// device-count rollup shown in the management UI.
package groups

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ioss/iot-dashboard/backend/internal/domain/group"
)

// Service implements group management use-cases over the group port.
type Service struct {
	repo group.Repository
	now  func() time.Time
}

// NewService wires the service.
func NewService(repo group.Repository, now func() time.Time) *Service {
	return &Service{repo: repo, now: now}
}

// GroupWithCount pairs a group with its device count for listing.
type GroupWithCount struct {
	Group       *group.Group
	DeviceCount int64
}

// Create registers a new group.
func (s *Service) Create(ctx context.Context, tenantID, name, description string) (*group.Group, error) {
	g, err := group.New(uuid.NewString(), tenantID, name, description, s.now())
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

// List returns all groups for a tenant with device counts.
func (s *Service) List(ctx context.Context, tenantID string) ([]GroupWithCount, error) {
	groupsList, err := s.repo.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	counts, err := s.repo.DeviceCounts(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]GroupWithCount, len(groupsList))
	for i, g := range groupsList {
		out[i] = GroupWithCount{Group: g, DeviceCount: counts[g.ID]}
	}
	return out, nil
}

// Get fetches one group.
func (s *Service) Get(ctx context.Context, tenantID, id string) (*group.Group, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

// Update renames/redescribes a group.
func (s *Service) Update(ctx context.Context, tenantID, id, name, description string) (*group.Group, error) {
	g, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if err := g.Rename(name, description, s.now()); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

// Delete removes a group (devices are detached via ON DELETE SET NULL).
func (s *Service) Delete(ctx context.Context, tenantID, id string) error {
	return s.repo.Delete(ctx, tenantID, id)
}
