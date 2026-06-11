package persistence

import (
	"time"

	domauth "github.com/ioss/iot-dashboard/backend/internal/domain/auth"
	"github.com/ioss/iot-dashboard/backend/internal/domain/device"
	"github.com/ioss/iot-dashboard/backend/internal/domain/tenant"
	"github.com/ioss/iot-dashboard/backend/internal/domain/user"
)

// GORM models mirror the SQL schema (migrations are the source of truth —
// AutoMigrate is never used). Each model has explicit converters to/from the
// domain type so persistence concerns can't bleed into business logic.

type tenantModel struct {
	ID        string `gorm:"primaryKey"`
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (tenantModel) TableName() string { return "tenants" }

func tenantToModel(t *tenant.Tenant) *tenantModel {
	return &tenantModel{ID: t.ID, Name: t.Name, Slug: t.Slug, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt}
}

func (m *tenantModel) toDomain() *tenant.Tenant {
	return &tenant.Tenant{ID: m.ID, Name: m.Name, Slug: m.Slug, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}
}

type userModel struct {
	ID           string `gorm:"primaryKey"`
	TenantID     string
	Email        string
	Name         string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (userModel) TableName() string { return "users" }

func userToModel(u *user.User) *userModel {
	return &userModel{
		ID: u.ID, TenantID: u.TenantID, Email: u.Email, Name: u.Name,
		PasswordHash: u.PasswordHash, Role: string(u.Role),
		CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
	}
}

func (m *userModel) toDomain() *user.User {
	return &user.User{
		ID: m.ID, TenantID: m.TenantID, Email: m.Email, Name: m.Name,
		PasswordHash: m.PasswordHash, Role: user.Role(m.Role),
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

type deviceModel struct {
	ID         string `gorm:"primaryKey"`
	TenantID   string
	Name       string
	Model      string
	Firmware   string
	Status     string
	LastSeenAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (deviceModel) TableName() string { return "devices" }

func deviceToModel(d *device.Device) *deviceModel {
	return &deviceModel{
		ID: d.ID, TenantID: d.TenantID, Name: d.Name, Model: d.Model,
		Firmware: d.Firmware, Status: string(d.Status), LastSeenAt: d.LastSeenAt,
		CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
	}
}

func (m *deviceModel) toDomain() *device.Device {
	return &device.Device{
		ID: m.ID, TenantID: m.TenantID, Name: m.Name, Model: m.Model,
		Firmware: m.Firmware, Status: device.Status(m.Status), LastSeenAt: m.LastSeenAt,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

type refreshTokenModel struct {
	ID        string `gorm:"primaryKey"`
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

func (refreshTokenModel) TableName() string { return "refresh_tokens" }

func refreshToModel(t *domauth.RefreshToken) *refreshTokenModel {
	return &refreshTokenModel{
		ID: t.ID, UserID: t.UserID, TokenHash: t.TokenHash,
		ExpiresAt: t.ExpiresAt, RevokedAt: t.RevokedAt, CreatedAt: t.CreatedAt,
	}
}

func (m *refreshTokenModel) toDomain() *domauth.RefreshToken {
	return &domauth.RefreshToken{
		ID: m.ID, UserID: m.UserID, TokenHash: m.TokenHash,
		ExpiresAt: m.ExpiresAt, RevokedAt: m.RevokedAt, CreatedAt: m.CreatedAt,
	}
}
