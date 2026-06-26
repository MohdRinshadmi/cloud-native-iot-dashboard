// Package group is the device-fleet bounded context: named collections of
// devices within a tenant, used for organization and fan-out operations
// (group-wide OTA, bulk commands).
package group

import (
	"context"
	"strings"
	"time"

	"github.com/ioss/iot-dashboard/backend/internal/shared/apperror"
)

// Group is a named collection of devices belonging to one tenant.
type Group struct {
	ID          string
	TenantID    string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// New constructs a valid Group.
func New(id, tenantID, name, description string, now time.Time) (*Group, error) {
	switch {
	case id == "":
		return nil, apperror.InvalidInput("group id is required")
	case tenantID == "":
		return nil, apperror.InvalidInput("tenant id is required")
	case strings.TrimSpace(name) == "":
		return nil, apperror.InvalidInput("group name is required")
	}
	return &Group{
		ID:          id,
		TenantID:    tenantID,
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Rename updates the group's descriptive fields.
func (g *Group) Rename(name, description string, now time.Time) error {
	if strings.TrimSpace(name) == "" {
		return apperror.InvalidInput("group name cannot be empty")
	}
	g.Name = strings.TrimSpace(name)
	g.Description = strings.TrimSpace(description)
	g.UpdatedAt = now
	return nil
}

// Repository is the persistence port for groups. All reads are tenant-scoped.
type Repository interface {
	Create(ctx context.Context, g *Group) error
	GetByID(ctx context.Context, tenantID, id string) (*Group, error)
	Update(ctx context.Context, g *Group) error
	Delete(ctx context.Context, tenantID, id string) error
	List(ctx context.Context, tenantID string) ([]*Group, error)
	// DeviceCounts returns deviceCount per groupID for one tenant.
	DeviceCounts(ctx context.Context, tenantID string) (map[string]int64, error)
}
