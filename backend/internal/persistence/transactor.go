package persistence

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// Transactor owns an application transaction boundary.
type Transactor interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type gormTransactor struct {
	db *gorm.DB
}

// NewTransactor returns a GORM-backed transaction boundary.
func NewTransactor(db *gorm.DB) Transactor {
	return &gormTransactor{db: db}
}

func (t *gormTransactor) WithTx(
	ctx context.Context,
	fn func(ctx context.Context) error,
) error {
	if err := t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(WithTxValue(ctx, tx))
	}); err != nil {
		return apperrors.Wrap(err, "transaction failed")
	}
	return nil
}
