package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/persistence"
)

// Transactor is a temporary compatibility alias for persistence.Transactor.
type Transactor = persistence.Transactor

// NewTransactor delegates to persistence.
func NewTransactor(db *gorm.DB) Transactor {
	return persistence.NewTransactor(db)
}

func txFromContext(ctx context.Context) *gorm.DB {
	return persistence.TxFromContext(ctx)
}

// DetachTx delegates to persistence.
func DetachTx(ctx context.Context) context.Context {
	return persistence.DetachTx(ctx)
}
