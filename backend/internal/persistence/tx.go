// Package persistence owns the shared GORM transaction and query helpers used
// by the domain packages. It contains no domain business rules.
package persistence

import (
	"context"

	"gorm.io/gorm"
)

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

// DBOrTx returns the ambient transaction when present; otherwise db with ctx.
func DBOrTx(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx := TxFromContext(ctx); tx != nil {
		return tx.WithContext(ctx)
	}
	return db.WithContext(ctx)
}

// DetachTx returns a context detached from cancellation and ambient transaction.
func DetachTx(ctx context.Context) context.Context {
	return context.WithValue(context.WithoutCancel(ctx), txKey{}, (*gorm.DB)(nil))
}
