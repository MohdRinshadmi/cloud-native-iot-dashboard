package persistence

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/ioss/iot-dashboard/backend/internal/shared/apperror"
)

// txKey carries an open transaction through context (unit-of-work pattern).
type txKey struct{}

// TxManager implements application/auth.TxManager over GORM.
type TxManager struct{ db *gorm.DB }

// NewTxManager wraps the root connection.
func NewTxManager(db *gorm.DB) *TxManager { return &TxManager{db: db} }

// WithinTx runs fn inside a transaction. Repositories called with the
// returned ctx automatically join it via dbFrom.
func (m *TxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

// dbFrom returns the ambient transaction if one is open, else the root DB.
func dbFrom(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return fallback.WithContext(ctx)
}

// translateError maps GORM/Postgres failures to transport-agnostic apperrors.
func translateError(err error, notFoundMsg string) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return apperror.NotFound(notFoundMsg)
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return apperror.New(apperror.CodeConflict, "resource already exists")
	default:
		return apperror.Internal(err)
	}
}
