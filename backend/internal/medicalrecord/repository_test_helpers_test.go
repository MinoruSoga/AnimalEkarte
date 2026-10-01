package medicalrecord

import (
	"context"
	"testing"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func makeTestOwner(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Owner {
	t.Helper()
	return testdb.MakeTestOwner(t, db, clinicID, name)
}

// withTx mirrors repository.Transactor.WithTx without importing the flat repository package.
func withTx(ctx context.Context, db *gorm.DB, fn func(ctx context.Context) error) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(persistence.WithTxValue(ctx, tx))
	})
}
