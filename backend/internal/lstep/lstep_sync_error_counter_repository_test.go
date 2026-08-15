package lstep

// lstep_sync_error_counter_repository_test.go — LstepSyncErrorCounterRepository 統合テスト。
//
// 保護する不変条件:
//   - IncrementFailure は (clinic_id, owner_id) 単位で raw SQL ON CONFLICT により
//     新規作成時 failure_count=1、以後は失敗のたびにアトミックに +1 する。
//   - ResetFailure は行を削除する（行不在=失敗0回）。
//   - FindByOwner は clinic_id + owner_id で分離され、該当なしは NotFound を返す。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupLstepSyncErrorCounterTestDB は lstep_sync_error_counters テーブルを用意する。
// 本番マイグレーション（001_init.sql）では UNIQUE(clinic_id, owner_id) が定義されているが、
// GORM AutoMigrate はモデルに uniqueIndex タグが無いため複合 UNIQUE を再現しない。
// IncrementFailure の raw SQL ON CONFLICT を意味のある形で検証するため、明示的に追加する。
func setupLstepSyncErrorCounterTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.LstepSyncErrorCounter{}))
	db.Exec("TRUNCATE TABLE lstep_sync_error_counters CASCADE")
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_test_lstep_sync_error_counter_conflict
		ON lstep_sync_error_counters (clinic_id, owner_id)`)
	return db
}

func TestLstepSyncErrorCounterRepository_IncrementFailure(t *testing.T) {
	db := setupLstepSyncErrorCounterTestDB(t)
	repo := NewLstepSyncErrorCounterRepository(db)
	ctx := context.Background()

	t.Run("初回は failure_count=1 で新規作成される", func(t *testing.T) {
		count, err := repo.IncrementFailure(ctx, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("2回目以降は failure_count がアトミックに加算される", func(t *testing.T) {
		_, err := repo.IncrementFailure(ctx, 1, 200)
		require.NoError(t, err)
		count, err := repo.IncrementFailure(ctx, 1, 200)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
		count, err = repo.IncrementFailure(ctx, 1, 200)
		require.NoError(t, err)
		assert.Equal(t, 3, count)

		var rowCount int64
		require.NoError(t, db.Model(&model.LstepSyncErrorCounter{}).Where("clinic_id = ? AND owner_id = ?", 1, 200).Count(&rowCount).Error)
		assert.Equal(t, int64(1), rowCount, "重複行が作られず1行のみ更新されること")
	})

	t.Run("clinic_id が異なれば同じ owner_id でも独立してカウントされる", func(t *testing.T) {
		const ownerID = uint64(300)
		countClinicA, err := repo.IncrementFailure(ctx, 10, ownerID)
		require.NoError(t, err)
		assert.Equal(t, 1, countClinicA)

		countClinicB, err := repo.IncrementFailure(ctx, 20, ownerID)
		require.NoError(t, err)
		assert.Equal(t, 1, countClinicB, "別クリニックは独立したカウンターを持つ")
	})
}

func TestLstepSyncErrorCounterRepository_ResetFailure(t *testing.T) {
	db := setupLstepSyncErrorCounterTestDB(t)
	repo := NewLstepSyncErrorCounterRepository(db)
	ctx := context.Background()

	t.Run("既存カウンターを削除する", func(t *testing.T) {
		_, err := repo.IncrementFailure(ctx, 1, 400)
		require.NoError(t, err)

		require.NoError(t, repo.ResetFailure(ctx, 1, 400))

		var count int64
		require.NoError(t, db.Model(&model.LstepSyncErrorCounter{}).Where("clinic_id = ? AND owner_id = ?", 1, 400).Count(&count).Error)
		assert.Equal(t, int64(0), count)
	})

	t.Run("存在しないカウンターに対しても no-op でエラーにならない", func(t *testing.T) {
		require.NoError(t, repo.ResetFailure(ctx, 1, 999999))
	})

	t.Run("別クリニックのカウンターには影響しない", func(t *testing.T) {
		const ownerID = uint64(500)
		_, err := repo.IncrementFailure(ctx, 1, ownerID)
		require.NoError(t, err)
		_, err = repo.IncrementFailure(ctx, 2, ownerID)
		require.NoError(t, err)

		require.NoError(t, repo.ResetFailure(ctx, 1, ownerID))

		var countClinic2 int64
		require.NoError(t, db.Model(&model.LstepSyncErrorCounter{}).Where("clinic_id = ? AND owner_id = ?", 2, ownerID).Count(&countClinic2).Error)
		assert.Equal(t, int64(1), countClinic2, "別クリニックのカウンターは削除されない")
	})
}

func TestLstepSyncErrorCounterRepository_FindByOwner(t *testing.T) {
	db := setupLstepSyncErrorCounterTestDB(t)
	repo := NewLstepSyncErrorCounterRepository(db)
	ctx := context.Background()

	t.Run("存在するカウンターを返す", func(t *testing.T) {
		_, err := repo.IncrementFailure(ctx, 1, 600)
		require.NoError(t, err)
		_, err = repo.IncrementFailure(ctx, 1, 600)
		require.NoError(t, err)

		found, err := repo.FindByOwner(ctx, 1, 600)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, 2, found.FailureCount)
	})

	t.Run("存在しない場合は NotFound を返す", func(t *testing.T) {
		_, err := repo.FindByOwner(ctx, 1, 700)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("別クリニックのカウンターは見えない（clinic_id 分離）", func(t *testing.T) {
		const ownerID = uint64(800)
		_, err := repo.IncrementFailure(ctx, 30, ownerID)
		require.NoError(t, err)

		_, err = repo.FindByOwner(ctx, 40, ownerID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}
