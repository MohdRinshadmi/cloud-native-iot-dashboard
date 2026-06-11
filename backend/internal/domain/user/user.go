// Package user is the identity bounded context: users, roles and their
// invariants. Password HASHING is infrastructure; the domain only ever sees
// the resulting hash.
package user

import (
	"regexp"
	"strings"
	"time"

	"github.com/ioss/iot-dashboard/backend/internal/shared/apperror"
)

// Role is the RBAC role assigned to a user. One role per user — simple,
// auditable, and sufficient for admin/operator/viewer separation.
type Role string

const (
	RoleAdmin    Role = "admin"    // full control incl. destructive actions
	RoleOperator Role = "operator" // manage devices, ack alerts
	RoleViewer   Role = "viewer"   // read-only
)

// Valid reports whether r is a known role.
func (r Role) Valid() bool {
	switch r {
	case RoleAdmin, RoleOperator, RoleViewer:
		return true
	default:
		return false
	}
}

var emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// User is the aggregate root. Email is globally unique (it is the login
// identifier); tenant scoping happens through TenantID.
type User struct {
	ID           string
	TenantID     string
	Email        string
	Name         string
	PasswordHash string
	Role         Role
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// New constructs a valid User, enforcing invariants at the boundary.
func New(id, tenantID, email, name, passwordHash string, role Role, now time.Time) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	switch {
	case id == "":
		return nil, apperror.InvalidInput("user id is required")
	case tenantID == "":
		return nil, apperror.InvalidInput("tenant id is required")
	case !emailRe.MatchString(email):
		return nil, apperror.InvalidInput("invalid email address")
	case strings.TrimSpace(name) == "":
		return nil, apperror.InvalidInput("name is required")
	case passwordHash == "":
		return nil, apperror.InvalidInput("password hash is required")
	case !role.Valid():
		return nil, apperror.InvalidInput("invalid role")
	}
	return &User{
		ID:           id,
		TenantID:     tenantID,
		Email:        email,
		Name:         strings.TrimSpace(name),
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// CanManageDevices reports whether the role may create/update devices.
func (u *User) CanManageDevices() bool {
	return u.Role == RoleAdmin || u.Role == RoleOperator
}
