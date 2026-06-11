package user

import "context"

// Repository is the persistence port for users.
type Repository interface {
	Create(ctx context.Context, u *User) error
	// GetByEmail is the login lookup; email is globally unique.
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
}
