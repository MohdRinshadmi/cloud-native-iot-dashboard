package auth

import (
	"context"
	"time"

	"github.com/ioss/iot-dashboard/backend/internal/domain/user"
)

// PasswordHasher abstracts password hashing (bcrypt adapter in infrastructure).
type PasswordHasher interface {
	Hash(plain string) (string, error)
	// Compare returns nil when plain matches hash.
	Compare(hash, plain string) error
}

// AccessTokenIssuer mints short-lived access tokens (JWT adapter in infra).
type AccessTokenIssuer interface {
	Issue(u *user.User) (token string, expiresIn time.Duration, err error)
}

// TokenVerifier validates an access token and returns the principal.
// Implemented by the same JWT adapter; consumed by the HTTP middleware.
type TokenVerifier interface {
	VerifyAccess(token string) (*Principal, error)
}

// TxManager runs a function inside a storage transaction. Repositories
// participating in the same ctx share the transaction (unit-of-work).
type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// Principal is the authenticated identity attached to every request.
type Principal struct {
	UserID   string
	TenantID string
	Email    string
	Role     user.Role
}
