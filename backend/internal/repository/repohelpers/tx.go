// Package repohelpers holds shared GORM helpers for repository domain subpackages.
// Parent package repository re-exports thin wrappers so existing call sites stay stable.
// Subpackages import repohelpers instead of forking clinicScope / dbOrTx copies
// (which would create an import cycle if they imported repository).
package repohelpers

import (
	"context"

	"gorm.io/gorm"
)

// txKey is the ambient transaction context key shared by Transactor.WithTx and DBOrTx.
// It must live in this package so parent repository and domain subpackages observe the same value.
type txKey struct{}

// WithTxValue attaches an ambient transaction to ctx for DBOrTx consumers.
func WithTxValue(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// TxFromContext returns the ambient transaction or nil.
func TxFromContext(ctx context.Context) *gorm.DB {
	tx, _ := ctx.Value(txKey{}).(*gorm.DB)
	return tx
}

// DBOrTx returns the ambient transaction when present; otherwise db with context.
func DBOrTx(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx := TxFromContext(ctx); tx != nil {
		return tx.WithContext(ctx)
	}
	return db.WithContext(ctx)
}

// DetachTx returns a context that no longer carries an ambient transaction
// (and is detached from parent cancellation via WithoutCancel).
func DetachTx(ctx context.Context) context.Context {
	return context.WithValue(context.WithoutCancel(ctx), txKey{}, (*gorm.DB)(nil))
}
