// Package tenant is the multi-tenancy bounded context. Every device and user
// belongs to exactly one tenant; isolation is enforced at every query.
package tenant

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/ioss/iot-dashboard/backend/internal/shared/apperror"
)

var slugRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// Tenant is an isolated customer workspace.
type Tenant struct {
	ID        string
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// New constructs a valid Tenant.
func New(id, name, slug string, now time.Time) (*Tenant, error) {
	switch {
	case id == "":
		return nil, apperror.InvalidInput("tenant id is required")
	case strings.TrimSpace(name) == "":
		return nil, apperror.InvalidInput("tenant name is required")
	case !slugRe.MatchString(slug):
		return nil, apperror.InvalidInput("tenant slug must be lowercase kebab-case")
	}
	return &Tenant{ID: id, Name: strings.TrimSpace(name), Slug: slug, CreatedAt: now, UpdatedAt: now}, nil
}

// Slugify derives a URL-safe slug from a display name.
func Slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// Repository is the persistence port for tenants.
type Repository interface {
	Create(ctx context.Context, t *Tenant) error
	GetByID(ctx context.Context, id string) (*Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*Tenant, error)
}
