// Package authtoken is the JWT adapter implementing the application's
// AccessTokenIssuer and TokenVerifier ports (HS256, short-lived).
package authtoken

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	appauth "github.com/ioss/iot-dashboard/backend/internal/application/auth"
	"github.com/ioss/iot-dashboard/backend/internal/domain/user"
)

const issuer = "iot-api"

// claims is the access-token payload. Tenant and role ride in the token so
// request handling needs zero identity lookups on the hot path.
type claims struct {
	jwt.RegisteredClaims
	TenantID string `json:"ten"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// Manager issues and verifies access tokens.
type Manager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

// NewManager wires the JWT manager. `now` injected for deterministic tests.
func NewManager(secret string, ttl time.Duration, now func() time.Time) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl, now: now}
}

var (
	_ appauth.AccessTokenIssuer = (*Manager)(nil)
	_ appauth.TokenVerifier     = (*Manager)(nil)
)

// Issue mints a signed access token for u.
func (m *Manager) Issue(u *user.User) (string, time.Duration, error) {
	now := m.now()
	c := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   u.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
			ID:        uuid.NewString(),
		},
		TenantID: u.TenantID,
		Email:    u.Email,
		Role:     string(u.Role),
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(m.secret)
	if err != nil {
		return "", 0, fmt.Errorf("sign token: %w", err)
	}
	return signed, m.ttl, nil
}

// VerifyAccess validates signature, expiry and issuer, returning the principal.
func (m *Manager) VerifyAccess(token string) (*appauth.Principal, error) {
	var c claims
	parsed, err := jwt.ParseWithClaims(
		token,
		&c,
		func(t *jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), // forbid alg confusion
		jwt.WithIssuer(issuer),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(m.now),
	)
	if err != nil || !parsed.Valid {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	return &appauth.Principal{
		UserID:   c.Subject,
		TenantID: c.TenantID,
		Email:    c.Email,
		Role:     user.Role(c.Role),
	}, nil
}
