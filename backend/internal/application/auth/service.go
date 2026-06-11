// Package auth implements registration, login and the refresh-token rotation
// scheme. Security properties enforced here:
//
//   - refresh tokens are single-use; every exchange rotates the token
//   - presenting an already-revoked token is treated as theft → ALL sessions
//     for that user are revoked (reuse detection)
//   - login failures return one generic error (no user enumeration)
//   - raw refresh tokens are never persisted, only sha-256 hashes
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/google/uuid"

	domauth "github.com/ioss/iot-dashboard/backend/internal/domain/auth"
	"github.com/ioss/iot-dashboard/backend/internal/domain/tenant"
	"github.com/ioss/iot-dashboard/backend/internal/domain/user"
	"github.com/ioss/iot-dashboard/backend/internal/shared/apperror"
)

// Service orchestrates identity use-cases over domain ports.
type Service struct {
	users      user.Repository
	tenants    tenant.Repository
	refresh    domauth.RefreshTokenRepository
	tx         TxManager
	hasher     PasswordHasher
	issuer     AccessTokenIssuer
	refreshTTL time.Duration
	now        func() time.Time
}

// NewService wires the auth service.
func NewService(
	users user.Repository,
	tenants tenant.Repository,
	refresh domauth.RefreshTokenRepository,
	tx TxManager,
	hasher PasswordHasher,
	issuer AccessTokenIssuer,
	refreshTTL time.Duration,
	now func() time.Time,
) *Service {
	return &Service{
		users: users, tenants: tenants, refresh: refresh, tx: tx,
		hasher: hasher, issuer: issuer, refreshTTL: refreshTTL, now: now,
	}
}

// Session is the result of a successful authentication.
type Session struct {
	User             *user.User
	AccessToken      string
	ExpiresIn        time.Duration
	RefreshToken     string // raw value — sent to the client once, never stored
	RefreshExpiresAt time.Time
}

// RegisterInput creates a new tenant workspace with its first admin.
type RegisterInput struct {
	TenantName string
	Email      string
	Name       string
	Password   string
}

// Register provisions tenant + admin user atomically and signs them in.
func (s *Service) Register(ctx context.Context, in RegisterInput) (*Session, error) {
	if len(in.Password) < 8 {
		return nil, apperror.InvalidInput("password must be at least 8 characters")
	}

	hash, err := s.hasher.Hash(in.Password)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	now := s.now()
	t, err := tenant.New(uuid.NewString(), in.TenantName, tenant.Slugify(in.TenantName), now)
	if err != nil {
		return nil, err
	}
	u, err := user.New(uuid.NewString(), t.ID, in.Email, in.Name, hash, user.RoleAdmin, now)
	if err != nil {
		return nil, err
	}

	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.tenants.Create(ctx, t); err != nil {
			return err
		}
		return s.users.Create(ctx, u)
	})
	if err != nil {
		return nil, err
	}

	return s.startSession(ctx, u)
}

// Login authenticates by email + password.
func (s *Service) Login(ctx context.Context, email, password string) (*Session, error) {
	invalid := apperror.New(apperror.CodeUnauthorized, "invalid email or password")

	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		// Burn comparable time even when the user doesn't exist so response
		// timing doesn't reveal account existence.
		_ = s.hasher.Compare("$2a$10$invalidinvalidinvalidinvalidinvalidinvalid1234567890", password)
		return nil, invalid
	}
	if err := s.hasher.Compare(u.PasswordHash, password); err != nil {
		return nil, invalid
	}
	return s.startSession(ctx, u)
}

// Refresh exchanges a valid refresh token for a fresh session (rotation).
func (s *Service) Refresh(ctx context.Context, rawToken string) (*Session, error) {
	invalid := apperror.New(apperror.CodeUnauthorized, "invalid refresh token")
	if rawToken == "" {
		return nil, invalid
	}

	t, err := s.refresh.GetByHash(ctx, hashToken(rawToken))
	if err != nil {
		return nil, invalid
	}

	now := s.now()
	if t.RevokedAt != nil {
		// Reuse of a rotated token ⇒ assume compromise; kill all sessions.
		_ = s.refresh.RevokeAllForUser(ctx, t.UserID, now)
		return nil, invalid
	}
	if !t.Active(now) {
		return nil, invalid
	}

	u, err := s.users.GetByID(ctx, t.UserID)
	if err != nil {
		return nil, invalid
	}

	if err := s.refresh.Revoke(ctx, t.ID, now); err != nil {
		return nil, apperror.Internal(err)
	}
	return s.startSession(ctx, u)
}

// Logout revokes the presented refresh token (access tokens simply expire).
func (s *Service) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	t, err := s.refresh.GetByHash(ctx, hashToken(rawToken))
	if err != nil {
		return nil // already gone — logout is idempotent
	}
	return s.refresh.Revoke(ctx, t.ID, s.now())
}

// Me returns the current principal's user record.
func (s *Service) Me(ctx context.Context, userID string) (*user.User, error) {
	return s.users.GetByID(ctx, userID)
}

// startSession mints the access token + a new refresh token for u.
func (s *Service) startSession(ctx context.Context, u *user.User) (*Session, error) {
	access, expiresIn, err := s.issuer.Issue(u)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	raw, err := newRawToken()
	if err != nil {
		return nil, apperror.Internal(err)
	}
	now := s.now()
	rt := &domauth.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    u.ID,
		TokenHash: hashToken(raw),
		ExpiresAt: now.Add(s.refreshTTL),
		CreatedAt: now,
	}
	if err := s.refresh.Create(ctx, rt); err != nil {
		return nil, apperror.Internal(err)
	}

	return &Session{
		User:             u,
		AccessToken:      access,
		ExpiresIn:        expiresIn,
		RefreshToken:     raw,
		RefreshExpiresAt: rt.ExpiresAt,
	}, nil
}

// newRawToken returns 256 bits of crypto-random, URL-safe text.
func newRawToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// hashToken derives the storage key for a raw refresh token.
func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
