package billing

// cash_register_close_repository_test.go — CashRegisterCloseRepository 統合テスト。
//
// 保護する不変条件:
//   - FindAll / FindByID / HasCloseOnDate / FindByDateAndPeriod は clinic_id でテナント隔離される。
//   - FindAll は close_date DESC, period ASC で返し、startDate/endDate でフィルタする。
//   - FindByID / FindAll は ClosedByStaff を deleted_at IS NULL 付きで Preload する。
//   - FindByDateAndPeriod は未検出時 (nil, nil) を返す（NotFound エラーにしない）。
//   - W-013: Create / CreateAdjustment のみ。Update/Delete/soft-delete 再開は app に存在しない（append-only）。
//   - 同一 (clinic_id, close_date, period) の二重 Create は Conflict になる。

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupCashRegisterCloseTestDB は cash_register_closes / adjustments と ClosedByStaff Preload に必要な
// staffs テーブルを用意する。
func setupCashRegisterCloseTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Staff{},
		&model.CashRegisterClose{},
		&model.CashRegisterCloseAdjustment{},
	))
	db.Exec("TRUNCATE TABLE cash_register_close_adjustments CASCADE")
	db.Exec("TRUNCATE TABLE cash_register_closes CASCADE")
	db.Exec("TRUNCATE TABLE staffs CASCADE")
	return db
}

func makeCashCloseStaff(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Staff {
	t.Helper()
	s := &model.Staff{ClinicID: clinicID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(s).Error)
	return s
}

func makeCashRegisterClose(t *testing.T, db *gorm.DB, clinicID uint64, closeDate time.Time, period string, closedBy *uint64) *model.CashRegisterClose {
	t.Helper()
	c := &model.CashRegisterClose{
		ClinicID:          clinicID,
		CloseDate:         closeDate,
		Period:            period,
		TheoreticalCash:   1000,
		ActualCash:        1000,
		CashDifference:    0,
		CategoryBreakdown: json.RawMessage(`{}`),
		ClosedBy:          closedBy,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(c).Error)
	return c
}

func TestCashRegisterCloseRepository_Create(t *testing.T) {
	db := setupCashRegisterCloseTestDB(t)
	repo := NewCashRegisterCloseRepository(db)
	ctx := context.Background()

	c := &model.CashRegisterClose{
		ClinicID:          uint64(1),
		CloseDate:         time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		Period:            "am",
		CategoryBreakdown: json.RawMessage(`{}`),
	}
	require.NoError(t, repo.Create(ctx, c))
	assert.NotZero(t, c.ID)
}

func TestCashRegisterCloseRepository_Create_DuplicateDatePeriodConflicts(t *testing.T) {
	db := setupCashRegisterCloseTestDB(t)
	repo := NewCashRegisterCloseRepository(db)
	ctx := context.Background()

	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	require.NoError(t, repo.Create(ctx, &model.CashRegisterClose{
		ClinicID:          1,
		CloseDate:         date,
		Period:            "am",
		CategoryBreakdown: json.RawMessage(`{}`),
	}))

	// W-013: soft-delete 再開経路は存在しない。同一キーの再 Create は Conflict。
	err := repo.Create(ctx, &model.CashRegisterClose{
		ClinicID:          1,
		CloseDate:         date,
		Period:            "am",
		CategoryBreakdown: json.RawMessage(`{}`),
	})
	require.Error(t, err)
}

func TestCashRegisterCloseRepository_CreateAdjustment(t *testing.T) {
	db := setupCashRegisterCloseTestDB(t)
	repo := NewCashRegisterCloseRepository(db)
	ctx := context.Background()

	closeRec := makeCashRegisterClose(t, db, 1, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), "am", nil)
	actorID := uint64(9)

	adj := &model.CashRegisterCloseAdjustment{
		ClinicID:           1,
		CloseID:            closeRec.ID,
		BillingID:          42,
		AccountingDelta:    -500,
		CashMovementAmount: 0,
		Reason:             "金額訂正",
		ActorID:            &actorID,
		ExecutedAt:         time.Now(),
	}
	require.NoError(t, repo.CreateAdjustment(ctx, adj))
	assert.NotZero(t, adj.ID)

	var got model.CashRegisterCloseAdjustment
	require.NoError(t, db.First(&got, adj.ID).Error)
	assert.Equal(t, closeRec.ID, got.CloseID)
	assert.Equal(t, uint64(42), got.BillingID)
	assert.Equal(t, int64(-500), got.AccountingDelta)
	assert.Equal(t, "金額訂正", got.Reason)
}

func TestCashRegisterCloseRepository_CreateAdjustment_ReasonRequired(t *testing.T) {
	db := setupCashRegisterCloseTestDB(t)
	repo := NewCashRegisterCloseRepository(db)
	ctx := context.Background()

	closeRec := makeCashRegisterClose(t, db, 1, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), "pm", nil)
	err := repo.CreateAdjustment(ctx, &model.CashRegisterCloseAdjustment{
		ClinicID:  1,
		CloseID:   closeRec.ID,
		BillingID: 1,
		Reason:    "",
	})
	require.Error(t, err)
}

func TestCashRegisterCloseRepository_AppendOnlyContract_NoDeleteMethod(t *testing.T) {
	// 通常経路に一般 Update は無い。BUG-032 の Void は特権取消の明示 API のみ。
	db := setupCashRegisterCloseTestDB(t)
	repo := NewCashRegisterCloseRepository(db)
	ctx := context.Background()

	c := makeCashRegisterClose(t, db, 1, time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC), "emg", nil)
	got, err := repo.FindByID(ctx, 1, c.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, c.ID, got.ID)

	// Void なしの再 Create は拒否される。
	err = repo.Create(ctx, &model.CashRegisterClose{
		ClinicID:          1,
		CloseDate:         time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC),
		Period:            "emg",
		CategoryBreakdown: json.RawMessage(`{}`),
	})
	require.Error(t, err, "同一 date/period の再 Create は Void なしでは不可")
}

func TestCashRegisterCloseRepository_FindAll(t *testing.T) {
	db := setupCashRegisterCloseTestDB(t)
	repo := NewCashRegisterCloseRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	staffA := makeCashCloseStaff(t, db, clinicA, "スタッフA")

	closeJun1AM := makeCashRegisterClose(t, db, clinicA, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), "am", &staffA.ID)
	closeJun1PM := makeCashRegisterClose(t, db, clinicA, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), "pm", &staffA.ID)
	closeJun2AM := makeCashRegisterClose(t, db, clinicA, time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC), "am", nil)
	makeCashRegisterClose(t, db, clinicB, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), "am", nil)

	t.Run("returns clinic-scoped rows ordered by close_date DESC, period ASC with staff preloaded", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, clinicA, nil, nil, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		require.Len(t, got, 3)
		assert.Equal(t, closeJun2AM.ID, got[0].ID)
		assert.Equal(t, closeJun1AM.ID, got[1].ID)
		assert.Equal(t, closeJun1PM.ID, got[2].ID)
		require.NotNil(t, got[1].ClosedByStaff)
		assert.Equal(t, "スタッフA", got[1].ClosedByStaff.Name)
	})

	t.Run("clinic isolation: clinic B only sees its own row", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, clinicB, nil, nil, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
	})

	t.Run("date range filters by startDate/endDate", func(t *testing.T) {
		start := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)
		got, total, err := repo.FindAll(ctx, clinicA, &start, &start, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, closeJun2AM.ID, got[0].ID)
	})

	t.Run("pagination limits results", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, clinicA, nil, nil, 1, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, got, 2)

		got2, total2, err := repo.FindAll(ctx, clinicA, nil, nil, 2, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total2)
		assert.Len(t, got2, 1)
	})
}

func TestCashRegisterCloseRepository_FindByID(t *testing.T) {
	db := setupCashRegisterCloseTestDB(t)
	repo := NewCashRegisterCloseRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	staffA := makeCashCloseStaff(t, db, clinicA, "スタッフA")
	c := makeCashRegisterClose(t, db, clinicA, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), "am", &staffA.ID)

	t.Run("found with staff preloaded", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, c.ID)
		require.NoError(t, err)
		require.NotNil(t, got.ClosedByStaff)
		assert.Equal(t, "スタッフA", got.ClosedByStaff.Name)
	})

	t.Run("not found for nonexistent id", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, uint64(999999))
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("clinic isolation", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, c.ID)
		assert.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestCashRegisterCloseRepository_HasCloseOnDate(t *testing.T) {
	db := setupCashRegisterCloseTestDB(t)
	repo := NewCashRegisterCloseRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	date := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	makeCashRegisterClose(t, db, clinicA, date, "am", nil)

	t.Run("true when a close exists on the date", func(t *testing.T) {
		got, err := repo.HasCloseOnDate(ctx, clinicA, date)
		require.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("false when no close exists on the date", func(t *testing.T) {
		got, err := repo.HasCloseOnDate(ctx, clinicA, time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		assert.False(t, got)
	})

	t.Run("clinic isolation: clinic B sees no close on that date", func(t *testing.T) {
		got, err := repo.HasCloseOnDate(ctx, clinicB, date)
		require.NoError(t, err)
		assert.False(t, got)
	})
}

func TestCashRegisterCloseRepository_FindByDateAndPeriod(t *testing.T) {
	db := setupCashRegisterCloseTestDB(t)
	repo := NewCashRegisterCloseRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	date := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	c := makeCashRegisterClose(t, db, clinicA, date, "pm", nil)

	t.Run("found", func(t *testing.T) {
		got, err := repo.FindByDateAndPeriod(ctx, clinicA, date, "pm")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, c.ID, got.ID)
	})

	t.Run("not found returns (nil, nil) rather than an error", func(t *testing.T) {
		got, err := repo.FindByDateAndPeriod(ctx, clinicA, date, "am")
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("clinic isolation: clinic B does not see clinic A's record", func(t *testing.T) {
		got, err := repo.FindByDateAndPeriod(ctx, clinicB, date, "pm")
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
