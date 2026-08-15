package clinic

// closing_special_period_repository_test.go — ClosingSpecialPeriodRepository の統合テスト
// （実 Postgres テスト DB）。happy path・not-found・clinic_id 隔離・FindByDate（該当なしは
// nil,nil を返す）・CheckOverlap（重複判定・excludeID）・Update/Delete を対象とする。
//
// model.ClosingSpecialPeriod は soft-delete フィールド（gorm.DeletedAt）を持たないため、
// Delete は物理削除になる（他のマスタ系リポジトリと異なる挙動）。

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

// setupClosingSpecialPeriodRepositoryTestDB は closing_special_periods テーブルを用意する。
//
// AutoMigrate の既知の制約: model.ClosingSpecialPeriod.AmPmBoundary/PmEnd は
// `gorm:"type:time"` タグ付き string フィールドだが、GORM の AutoMigrate はこのタグを無視し
// timestamptz 列として作成する（本番は migrations/001_init.sql の生SQLで time 型を使うため
// 影響を受けない。この差異はテスト専用DBのAutoMigrateパスにのみ存在する）。
// "12:00:00" のような time リテラルを timestamptz 列へ INSERT すると
// invalid input syntax for type timestamp with time zone (SQLSTATE 22007) で失敗するため、
// AutoMigrate 後に明示的に ALTER COLUMN で本番と同じ time 型へ矯正する。
func setupClosingSpecialPeriodRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.ClosingSpecialPeriod{}))
	require.NoError(t, db.Exec(`ALTER TABLE closing_special_periods ALTER COLUMN am_pm_boundary TYPE time USING am_pm_boundary::time`).Error)
	require.NoError(t, db.Exec(`ALTER TABLE closing_special_periods ALTER COLUMN pm_end TYPE time USING pm_end::time`).Error)
	db.Exec("TRUNCATE TABLE closing_special_periods CASCADE")
	return db
}

// makeClosingSpecialPeriod はテスト用の ClosingSpecialPeriod を作成して返す。
func makeClosingSpecialPeriod(t *testing.T, db *gorm.DB, clinicID uint64, start, end time.Time) *model.ClosingSpecialPeriod {
	t.Helper()
	p := &model.ClosingSpecialPeriod{
		ClinicID:     clinicID,
		StartDate:    start,
		EndDate:      end,
		AmPmBoundary: "12:00:00",
		PmEnd:        "18:00:00",
	}
	require.NoError(t, db.WithContext(context.Background()).Create(p).Error)
	return p
}

func TestClosingSpecialPeriodRepository_Create_And_FindByID(t *testing.T) {
	db := setupClosingSpecialPeriodRepositoryTestDB(t)
	repo := NewClosingSpecialPeriodRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	t.Run("happy path: Create してから FindByID で取得できる", func(t *testing.T) {
		start := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)
		end := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
		p := &model.ClosingSpecialPeriod{ClinicID: clinicID, StartDate: start, EndDate: end, AmPmBoundary: "12:00:00", PmEnd: "18:00:00", Note: "夏季休診"}
		created, err := repo.Create(ctx, p)
		require.NoError(t, err)
		require.NotZero(t, created.ID)

		got, err := repo.FindByID(ctx, clinicID, created.ID)
		require.NoError(t, err)
		assert.Equal(t, "夏季休診", got.Note)
	})

	t.Run("存在しない ID は NotFound エラー", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicID, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("別クリニックからは FindByID できない（clinic_id 隔離）", func(t *testing.T) {
		p := makeClosingSpecialPeriod(t, db, clinicID, time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC))
		_, err := repo.FindByID(ctx, uint64(999), p.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "別クリニックからは NotFound であるべき: %v", err)
	})
}

func TestClosingSpecialPeriodRepository_FindAll_ClinicIsolationAndOrder(t *testing.T) {
	db := setupClosingSpecialPeriodRepositoryTestDB(t)
	repo := NewClosingSpecialPeriodRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	late := makeClosingSpecialPeriod(t, db, clinicA, time.Date(2026, 12, 29, 0, 0, 0, 0, time.UTC), time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC))
	early := makeClosingSpecialPeriod(t, db, clinicA, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC))
	makeClosingSpecialPeriod(t, db, clinicB, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC))

	t.Run("clinicA は自院の2件を start_date 昇順で返す", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, early.ID, got[0].ID)
		assert.Equal(t, late.ID, got[1].ID)
	})

	t.Run("clinicB は自院の1件のみ（混入なし）", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicB)
		require.NoError(t, err)
		assert.Len(t, got, 1)
	})
}

func TestClosingSpecialPeriodRepository_FindByDate(t *testing.T) {
	db := setupClosingSpecialPeriodRepositoryTestDB(t)
	repo := NewClosingSpecialPeriodRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	p := makeClosingSpecialPeriod(t, db, clinicA, time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC), time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC))

	t.Run("期間内の日付はヒットする", func(t *testing.T) {
		got, err := repo.FindByDate(ctx, clinicA, time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, p.ID, got.ID)
	})

	t.Run("期間外の日付は nil,nil を返す（エラーにしない）", func(t *testing.T) {
		got, err := repo.FindByDate(ctx, clinicA, time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("別クリニックの期間はヒットしない（clinic_id 隔離）", func(t *testing.T) {
		got, err := repo.FindByDate(ctx, clinicB, time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestClosingSpecialPeriodRepository_Update(t *testing.T) {
	db := setupClosingSpecialPeriodRepositoryTestDB(t)
	repo := NewClosingSpecialPeriodRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	p := makeClosingSpecialPeriod(t, db, clinicA, time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC), time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC))

	t.Run("同一クリニックでは Update が反映される", func(t *testing.T) {
		got, err := repo.Update(ctx, clinicA, p.ID, map[string]any{"note": "更新後メモ"})
		require.NoError(t, err)
		assert.Equal(t, "更新後メモ", got.Note)
	})

	t.Run("別クリニックからの Update は NotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicB, p.ID, map[string]any{"note": "改ざん試行"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		got, err := repo.FindByID(ctx, clinicA, p.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新後メモ", got.Note, "別クリニックからの Update でメモが変わってはならない")
	})

	t.Run("存在しない ID の Update は NotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicA, 999999, map[string]any{"note": "x"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestClosingSpecialPeriodRepository_Delete(t *testing.T) {
	db := setupClosingSpecialPeriodRepositoryTestDB(t)
	repo := NewClosingSpecialPeriodRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	p := makeClosingSpecialPeriod(t, db, clinicA, time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC), time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC))

	t.Run("別クリニックからの Delete は NotFound で行が残る", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, p.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		got, err := repo.FindByID(ctx, clinicA, p.ID)
		require.NoError(t, err)
		assert.Equal(t, p.ID, got.ID, "別クリニックからの Delete で行が消えてはならない")
	})

	t.Run("同一クリニックでは Delete が成功し物理削除される（soft-delete 列なし）", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, p.ID))
		_, err := repo.FindByID(ctx, clinicA, p.ID)
		assert.True(t, apperrors.IsNotFound(err))

		var rawCount int64
		db.Unscoped().Model(&model.ClosingSpecialPeriod{}).Where("id = ?", p.ID).Count(&rawCount)
		assert.Equal(t, int64(0), rawCount, "ClosingSpecialPeriod は soft-delete 列を持たないため物理削除される")
	})

	t.Run("存在しない ID の Delete は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestClosingSpecialPeriodRepository_CreateCheckingOverlap(t *testing.T) {
	db := setupClosingSpecialPeriodRepositoryTestDB(t)
	repo := NewClosingSpecialPeriodRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	first := &model.ClosingSpecialPeriod{
		ClinicID:     clinicID,
		StartDate:    time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC),
		AmPmBoundary: "12:00",
		PmEnd:        "19:00",
		Note:         "first",
	}
	created, err := repo.CreateCheckingOverlap(ctx, first)
	require.NoError(t, err)
	require.NotZero(t, created.ID)

	overlap := &model.ClosingSpecialPeriod{
		ClinicID:     clinicID,
		StartDate:    time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC),
		AmPmBoundary: "12:00",
		PmEnd:        "19:00",
		Note:         "overlap",
	}
	_, err = repo.CreateCheckingOverlap(ctx, overlap)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "overlapping create must conflict: %v", err)

	var count int64
	require.NoError(t, db.Model(&model.ClosingSpecialPeriod{}).Where("clinic_id = ?", clinicID).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestClosingSpecialPeriodRepository_CheckOverlap(t *testing.T) {
	db := setupClosingSpecialPeriodRepositoryTestDB(t)
	repo := NewClosingSpecialPeriodRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	existing := makeClosingSpecialPeriod(t, db, clinicA, time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC), time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC))

	t.Run("重複する期間は true", func(t *testing.T) {
		overlap, err := repo.CheckOverlap(ctx, clinicA, time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC), time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC), nil)
		require.NoError(t, err)
		assert.True(t, overlap)
	})

	t.Run("重複しない期間は false", func(t *testing.T) {
		overlap, err := repo.CheckOverlap(ctx, clinicA, time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC), nil)
		require.NoError(t, err)
		assert.False(t, overlap)
	})

	t.Run("excludeID で自身を除外すると false", func(t *testing.T) {
		overlap, err := repo.CheckOverlap(ctx, clinicA, time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC), time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC), &existing.ID)
		require.NoError(t, err)
		assert.False(t, overlap)
	})

	t.Run("別クリニックの重複期間は影響しない（clinic_id 隔離）", func(t *testing.T) {
		overlap, err := repo.CheckOverlap(ctx, clinicB, time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC), time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC), nil)
		require.NoError(t, err)
		assert.False(t, overlap)
	})
}
