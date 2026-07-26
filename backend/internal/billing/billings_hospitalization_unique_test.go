package billing

// billings_hospitalization_unique_test.go — active hospitalization_id の二重 billing 永続化拒否。
//
// 本番 migration 004 の partial UNIQUE をテスト DB に明示適用し、AccountingRepository.Create が
// 同一 hospitalization_id の 2 行目で AlreadyExists を返すこと、NULL / soft-deleted は対象外
// であることを回帰固定する（shift_entry_repository_test.go と同型の INDEX 明示パターン）。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupBillingsHospitalizationUniqueTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	// setupTestDB already AutoMigrates Billing and TRUNCATEs billings.
	// AutoMigrate does not emit the production FK to hospitalizations(id), so synthetic
	// hospitalization_id values are sufficient to exercise the partial UNIQUE only.
	db := testdb.SetupTestDB(t)
	require.NoError(t, db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_billings_hospitalization_id_unique
		  ON billings(hospitalization_id)
		  WHERE hospitalization_id IS NOT NULL AND deleted_at IS NULL
	`).Error)
	return db
}

func makeHospBilling(t *testing.T, hospID *uint64) *model.Billing {
	t.Helper()
	return &model.Billing{
		ClinicID:          1,
		HospitalizationID: hospID,
		Status:            model.BillingStatusWaiting,
		ScheduledDate:     time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC),
		TotalAmount:       1000,
	}
}

func TestBillingsHospitalizationUnique_DualCreateRejected(t *testing.T) {
	db := setupBillingsHospitalizationUniqueTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	hospID := uint64(9001)

	require.NoError(t, repo.Create(ctx, 1, makeHospBilling(t, &hospID)))

	err := repo.Create(ctx, 1, makeHospBilling(t, &hospID))
	require.Error(t, err)
	assert.True(t, apperrors.IsAlreadyExists(err), "duplicate active hospitalization billing must be AlreadyExists: %v", err)
}

func TestBillingsHospitalizationUnique_NullHospitalizationAllowsMultiple(t *testing.T) {
	db := setupBillingsHospitalizationUniqueTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, 1, makeHospBilling(t, nil)))
	require.NoError(t, repo.Create(ctx, 1, makeHospBilling(t, nil)),
		"hospitalization_id NULL (medical_record-style) must allow multiple active billings")
}

func TestBillingsHospitalizationUnique_SoftDeletedExcluded(t *testing.T) {
	db := setupBillingsHospitalizationUniqueTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	hospID := uint64(9002)

	first := makeHospBilling(t, &hospID)
	require.NoError(t, repo.Create(ctx, 1, first))
	require.NoError(t, db.WithContext(ctx).Delete(first).Error)

	require.NoError(t, repo.Create(ctx, 1, makeHospBilling(t, &hospID)),
		"soft-deleted row must be outside UNIQUE; re-create with same hospitalization_id must succeed")
}
