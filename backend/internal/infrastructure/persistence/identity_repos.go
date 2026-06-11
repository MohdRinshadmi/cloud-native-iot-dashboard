package persistence

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"

	domauth "github.com/ioss/iot-dashboard/backend/internal/domain/auth"
	"github.com/ioss/iot-dashboard/backend/internal/domain/tenant"
	"github.com/ioss/iot-dashboard/backend/internal/domain/user"
)

// ---- users ------------------------------------------------------------------

// UserRepository is the GORM adapter for user.Repository.
type UserRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) *UserRepository { return &UserRepository{db: db} }

var _ user.Repository = (*UserRepository)(nil)

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	return translateError(dbFrom(ctx, r.db).Create(userToModel(u)).Error, "user not found")
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var m userModel
	err := dbFrom(ctx, r.db).Where("email = ?", strings.ToLower(email)).First(&m).Error
	if err != nil {
		return nil, translateError(err, "user not found")
	}
	return m.toDomain(), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	var m userModel
	err := dbFrom(ctx, r.db).Where("id = ?", id).First(&m).Error
	if err != nil {
		return nil, translateError(err, "user not found")
	}
	return m.toDomain(), nil
}

// ---- tenants ------------------------------------------------------------------

// TenantRepository is the GORM adapter for tenant.Repository.
type TenantRepository struct{ db *gorm.DB }

func NewTenantRepository(db *gorm.DB) *TenantRepository { return &TenantRepository{db: db} }

var _ tenant.Repository = (*TenantRepository)(nil)

func (r *TenantRepository) Create(ctx context.Context, t *tenant.Tenant) error {
	return translateError(dbFrom(ctx, r.db).Create(tenantToModel(t)).Error, "tenant not found")
}

func (r *TenantRepository) GetByID(ctx context.Context, id string) (*tenant.Tenant, error) {
	var m tenantModel
	err := dbFrom(ctx, r.db).Where("id = ?", id).First(&m).Error
	if err != nil {
		return nil, translateError(err, "tenant not found")
	}
	return m.toDomain(), nil
}

func (r *TenantRepository) GetBySlug(ctx context.Context, slug string) (*tenant.Tenant, error) {
	var m tenantModel
	err := dbFrom(ctx, r.db).Where("slug = ?", slug).First(&m).Error
	if err != nil {
		return nil, translateError(err, "tenant not found")
	}
	return m.toDomain(), nil
}

// ---- refresh tokens -----------------------------------------------------------

// RefreshTokenRepository is the GORM adapter for auth.RefreshTokenRepository.
type RefreshTokenRepository struct{ db *gorm.DB }

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

var _ domauth.RefreshTokenRepository = (*RefreshTokenRepository)(nil)

func (r *RefreshTokenRepository) Create(ctx context.Context, t *domauth.RefreshToken) error {
	return translateError(dbFrom(ctx, r.db).Create(refreshToModel(t)).Error, "token not found")
}

func (r *RefreshTokenRepository) GetByHash(ctx context.Context, hash string) (*domauth.RefreshToken, error) {
	var m refreshTokenModel
	err := dbFrom(ctx, r.db).Where("token_hash = ?", hash).First(&m).Error
	if err != nil {
		return nil, translateError(err, "token not found")
	}
	return m.toDomain(), nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id string, at time.Time) error {
	res := dbFrom(ctx, r.db).Model(&refreshTokenModel{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", at)
	return translateError(res.Error, "token not found")
}

func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string, at time.Time) error {
	res := dbFrom(ctx, r.db).Model(&refreshTokenModel{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", at)
	return translateError(res.Error, "token not found")
}
