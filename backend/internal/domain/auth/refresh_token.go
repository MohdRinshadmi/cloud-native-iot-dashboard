// Package auth holds session/credential domain objects. Refresh tokens are
// stored HASHED (sha-256) — a database leak must never yield usable tokens.
package auth

import (
	"context"
	"time"
)

// RefreshToken is one rotation link in a refresh chain. Tokens are single-use:
// each refresh revokes the presented token and issues a successor. Presenting
// a revoked token is treated as theft and revokes the whole user session set.
type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string // sha-256 hex of the raw token; raw value is never stored
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

// Active reports whether the token can still be exchanged.
func (t *RefreshToken) Active(now time.Time) bool {
	return t.RevokedAt == nil && now.Before(t.ExpiresAt)
}

// RefreshTokenRepository is the persistence port for refresh tokens.
type RefreshTokenRepository interface {
	Create(ctx context.Context, t *RefreshToken) error
	GetByHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	Revoke(ctx context.Context, id string, at time.Time) error
	// RevokeAllForUser is the breach response: kill every active session.
	RevokeAllForUser(ctx context.Context, userID string, at time.Time) error
}
