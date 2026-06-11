// Package hash is the bcrypt adapter for the PasswordHasher port.
package hash

import (
	"golang.org/x/crypto/bcrypt"

	appauth "github.com/ioss/iot-dashboard/backend/internal/application/auth"
)

// BcryptHasher hashes with bcrypt at the given cost (work factor).
type BcryptHasher struct{ cost int }

// NewBcryptHasher uses cost 12 — ~250ms per hash on modern hardware, the
// accepted balance between login latency and brute-force resistance.
func NewBcryptHasher() *BcryptHasher { return &BcryptHasher{cost: 12} }

var _ appauth.PasswordHasher = (*BcryptHasher)(nil)

func (h *BcryptHasher) Hash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), h.cost)
	return string(b), err
}

func (h *BcryptHasher) Compare(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
