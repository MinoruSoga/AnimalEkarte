// Package repohelpers is a temporary BE9-2F compatibility surface.
// New code imports internal/persistence directly.
package repohelpers

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/persistence"
)

// WithTxValue delegates to persistence so the ambient transaction key has one owner.
func WithTxValue(ctx context.Context, tx *gorm.DB) context.Context {
	return persistence.WithTxValue(ctx, tx)
}

// TxFromContext delegates to persistence.
func TxFromContext(ctx context.Context) *gorm.DB {
	return persistence.TxFromContext(ctx)
}

// DBOrTx delegates to persistence.
func DBOrTx(ctx context.Context, db *gorm.DB) *gorm.DB {
	return persistence.DBOrTx(ctx, db)
}

// DetachTx delegates to persistence.
func DetachTx(ctx context.Context) context.Context {
	return persistence.DetachTx(ctx)
}
