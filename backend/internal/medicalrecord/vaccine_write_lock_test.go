package medicalrecord

import (
	"context"
	"strings"
	"testing"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SQL generation and callback-controlled rows only; no PostgreSQL connection.
func TestVaccineRepository_WriteLockSQL(t *testing.T) {
	db, err := gorm.Open(postgres.Open("host=localhost user=unused dbname=unused sslmode=disable"), &gorm.Config{DisableAutomaticPing: true, DryRun: true, SkipDefaultTransaction: true})
	require.NoError(t, err)
	conn, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	type statement struct {
		sql  string
		vars []any
		ctx  context.Context
	}
	var statements []statement
	require.NoError(t, db.Callback().Query().After("gorm:query").Register("capture_vaccine_locks", func(tx *gorm.DB) {
		statements = append(statements, statement{tx.Statement.SQL.String(), append([]any(nil), tx.Statement.Vars...), tx.Statement.Context})
		if row, ok := tx.Statement.Dest.(*model.Vaccine); ok {
			row.ID = 17
			row.ClinicID = 2
			tx.RowsAffected = 1
		}
	}))
	require.NoError(t, db.Callback().Delete().After("gorm:delete").Register("capture_vaccine_delete", func(tx *gorm.DB) {
		statements = append(statements, statement{tx.Statement.SQL.String(), append([]any(nil), tx.Statement.Vars...), tx.Statement.Context})
		tx.RowsAffected = 1
	}))
	repo := &vaccineRepository{db: nil} // all operations must route through ambient DB
	require.Error(t, repo.LockWriteRows(t.Context(), 2, []uint64{17}))
	ctx := persistence.WithTxValue(t.Context(), db)
	require.NoError(t, repo.LockWriteRows(ctx, 2, []uint64{17, 10, 17}))
	require.Len(t, statements, 2)
	for i, id := range []uint64{10, 17} {
		require.Contains(t, statements[i].sql, "FOR UPDATE")
		require.Contains(t, statements[i].sql, "clinic_id")
		require.Contains(t, statements[i].sql, "deleted_at")
		require.Equal(t, id, statements[i].vars[0])
		require.Contains(t, statements[i].vars, uint64(2))
		require.Equal(t, ctx, statements[i].ctx)
	}
	statements = nil
	require.NoError(t, repo.Delete(ctx, 2, 17))
	require.Len(t, statements, 2, "target lock and guard must be separate SQL statements")
	require.Contains(t, statements[0].sql, "FOR UPDATE")
	require.True(t, strings.HasPrefix(statements[1].sql, "UPDATE"))
	require.Contains(t, statements[1].sql, "NOT EXISTS")
	require.Contains(t, statements[1].sql, "children.parent_id")
	require.Contains(t, statements[1].sql, "vaccinations.vaccine_id")
	statements = nil
	_, err = repo.FindByID(ctx, 2, 17)
	require.NoError(t, err)
	require.Contains(t, statements[0].sql, "FOR SHARE", "existing vaccination references retain stability")
	require.Error(t, repo.LockWriteRows(ctx, 2, []uint64{0}))
}

func TestVaccineService_DeleteRechecksAfterWriterCommit(t *testing.T) {
	// Deterministic lock-wait model: a competing writer commits its child before
	// the delete lock returns. This is a service-order harness, not a DB race test.
	var events []string
	children := int64(0)
	repo := &mockVaccineRepository{
		lockWriteRowsFn: func(ctx context.Context, cid uint64, ids []uint64) error {
			require.Equal(t, true, ctx.Value(vaccineAuditTXKey{}))
			require.Equal(t, []uint64{17}, ids)
			events = append(events, "writer-commits", "lock-acquired")
			children = 1
			return nil
		},
		findByIDFn: func(context.Context, uint64, uint64) (*model.Vaccine, error) {
			events = append(events, "read")
			return &model.Vaccine{ID: 17}, nil
		},
		countChildrenByParentIDFn: func(ctx context.Context, _, _ uint64) (int64, error) {
			require.Equal(t, true, ctx.Value(vaccineAuditTXKey{}))
			events = append(events, "fresh-count")
			return children, nil
		},
		deleteFn: func(context.Context, uint64, uint64) error {
			t.Fatal("committed child must prevent deletion")
			return nil
		},
	}
	svc := NewVaccineServiceWithAudit(repo, MasterInsuranceWriteDependencies{Transactor: vaccineTestTransactorFn(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(context.WithValue(ctx, vaccineAuditTXKey{}, true))
	})})
	require.Error(t, svc.Delete(t.Context(), 2, 17))
	require.Equal(t, []string{"writer-commits", "lock-acquired", "read", "fresh-count"}, events)
}
