package lstep

// lstep_sync_settings_repository_test.go — LstepSyncSettingsRepository 統合テスト。
//
// 保護する不変条件:
//   - FindByClinicID は clinic_id スコープで1件を返し、存在しない clinic_id は NotFound を返す。
//   - Upsert は clinic_id の UNIQUE 制約に対する ON CONFLICT DO UPDATE で、
//     初回は新規作成、2回目以降は既存行を更新する（重複行を作らない）。
//   - Upsert は is_sync_enabled / sync_enabled_at / updated_at のみを更新対象とする。

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

// setupLstepSyncSettingsTestDB は lstep_settings テーブルを用意する。
// LstepSettings.ClinicID には `uniqueIndex` タグがあるため AutoMigrate で
// UNIQUE(clinic_id) が再現され、Upsert の ON CONFLICT(clinic_id) をそのまま検証できる。
func setupLstepSyncSettingsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.LstepSettings{}))
	db.Exec("TRUNCATE TABLE lstep_settings CASCADE")
	return db
}

// makeLstepSettings はテスト用同期設定を作成して返す。
func makeLstepSettings(t *testing.T, db *gorm.DB, clinicID uint64, enabled bool) *model.LstepSettings {
	t.Helper()
	settings := &model.LstepSettings{
		ClinicID:      clinicID,
		IsSyncEnabled: enabled,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(settings).Error)
	return settings
}

func TestLstepSyncSettingsRepository_FindByClinicID(t *testing.T) {
	db := setupLstepSyncSettingsTestDB(t)
	repo := NewLstepSyncSettingsRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)
	makeLstepSettings(t, db, clinicA, true)

	t.Run("存在する clinic_id は設定を返す", func(t *testing.T) {
		found, err := repo.FindByClinicID(ctx, clinicA)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, clinicA, found.ClinicID)
		assert.True(t, found.IsSyncEnabled)
	})

	t.Run("別クリニックの clinic_id は NotFound を返す（clinic_id 分離）", func(t *testing.T) {
		_, err := repo.FindByClinicID(ctx, clinicB)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("未登録の clinic_id は NotFound を返す", func(t *testing.T) {
		_, err := repo.FindByClinicID(ctx, uint64(999999))
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})
}

func TestLstepSyncSettingsRepository_Upsert(t *testing.T) {
	db := setupLstepSyncSettingsTestDB(t)
	repo := NewLstepSyncSettingsRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("既存行がない場合は新規作成する", func(t *testing.T) {
		enabledAt := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
		result, err := repo.Upsert(ctx, &model.LstepSettings{
			ClinicID:      clinicA,
			IsSyncEnabled: true,
			SyncEnabledAt: &enabledAt,
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, clinicA, result.ClinicID)
		assert.True(t, result.IsSyncEnabled)
		require.NotNil(t, result.SyncEnabledAt)
		assert.WithinDuration(t, enabledAt, *result.SyncEnabledAt, time.Second)

		var count int64
		require.NoError(t, db.Model(&model.LstepSettings{}).Where("clinic_id = ?", clinicA).Count(&count).Error)
		assert.Equal(t, int64(1), count, "新規作成は1行のみであるべき")
	})

	t.Run("既存行がある場合は同一行を更新し重複行を作らない", func(t *testing.T) {
		enabledAt := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
		created, err := repo.Upsert(ctx, &model.LstepSettings{
			ClinicID:      clinicB,
			IsSyncEnabled: true,
			SyncEnabledAt: &enabledAt,
		})
		require.NoError(t, err)
		originalID := created.ID

		updated, err := repo.Upsert(ctx, &model.LstepSettings{
			ClinicID:      clinicB,
			IsSyncEnabled: false,
			SyncEnabledAt: nil,
		})
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, originalID, updated.ID, "更新は同一行に対して行われるべき（新規行を作らない）")
		assert.False(t, updated.IsSyncEnabled)
		assert.Nil(t, updated.SyncEnabledAt, "sync_enabled_at は nil に更新されるべき")

		var count int64
		require.NoError(t, db.Model(&model.LstepSettings{}).Where("clinic_id = ?", clinicB).Count(&count).Error)
		assert.Equal(t, int64(1), count, "Upsert 2回実行後も1行のみであるべき（ON CONFLICT DO UPDATE）")
	})

	t.Run("別クリニックの Upsert は他クリニックの行に影響しない（clinic_id 分離）", func(t *testing.T) {
		before, err := repo.FindByClinicID(ctx, clinicA)
		require.NoError(t, err)

		_, err = repo.Upsert(ctx, &model.LstepSettings{
			ClinicID:      clinicB,
			IsSyncEnabled: true,
		})
		require.NoError(t, err)

		after, err := repo.FindByClinicID(ctx, clinicA)
		require.NoError(t, err)
		assert.Equal(t, before.IsSyncEnabled, after.IsSyncEnabled, "clinic A の行は clinic B の Upsert で変化しないべき")
		assert.Equal(t, before.ID, after.ID)
	})
}
