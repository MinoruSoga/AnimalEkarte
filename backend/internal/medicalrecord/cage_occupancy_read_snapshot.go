package medicalrecord

import (
	"context"
	"database/sql"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type cageReadOnlySnapshotKey struct{}

type cageOccupancySnapshotRunner interface {
	WithCageOccupancyReadSnapshot(context.Context, func(context.Context) error) error
}

func cageReadOnlySnapshot(ctx context.Context) bool {
	tx, _ := ctx.Value(cageReadOnlySnapshotKey{}).(*gorm.DB)
	return tx != nil && tx == persistence.TxFromContext(ctx)
}

// withCageOccupancyReadSnapshot scopes all candidate queries to one snapshot.
// It never starts a transaction/savepoint when an ambient transaction exists.
// An incompatible ambient transaction is refused rather than silently replaced.
func withCageOccupancyReadSnapshot(ctx context.Context, db *gorm.DB, read func(context.Context) error) error {
	if persistence.TxFromContext(ctx) != nil {
		if cageReadOnlySnapshot(ctx) {
			return read(ctx)
		}
		var settings struct {
			Isolation string
			ReadOnly  bool
		}
		if err := persistence.DBOrTx(ctx, db).Raw(
			"SELECT current_setting('transaction_isolation') AS isolation, current_setting('transaction_read_only')::boolean AS read_only",
		).Scan(&settings).Error; err != nil {
			return apperrors.Wrap(err, "failed to inspect cage read transaction")
		}
		if settings.Isolation != "repeatable read" && settings.Isolation != "serializable" {
			return apperrors.WrapInternalServerError("cage candidates require a consistent ambient snapshot")
		}
		if settings.ReadOnly {
			ctx = context.WithValue(ctx, cageReadOnlySnapshotKey{}, persistence.TxFromContext(ctx))
		}
		return read(ctx)
	}
	if db == nil {
		return apperrors.WrapInternalServerError("cage snapshot database is not configured")
	}
	if _, bound := db.Statement.ConnPool.(gorm.TxCommitter); bound {
		return apperrors.WrapInternalServerError("cage snapshot requires the existing transaction context")
	}
	return persistence.DBOrTx(ctx, db).Transaction(func(tx *gorm.DB) error {
		readCtx := persistence.WithTxValue(ctx, tx)
		return read(context.WithValue(readCtx, cageReadOnlySnapshotKey{}, tx))
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
}
