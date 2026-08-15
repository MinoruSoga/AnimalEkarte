package lstep

// lstep_settings_repository_test.go — LstepSettingsRepository (clinic_integrations) 統合テスト。
//
// 保護する不変条件:
//   - FindByClinicAndService は clinic_id + service で正しく分離される。
//   - Upsert は (clinic_id, service, key_name) で UPSERT する（重複行は増えず key_value/updated_at のみ更新）。
//   - DeleteByClinicAndService は指定 clinic_id + service の行のみ削除し、他 clinic/service には影響しない。

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

// setupLstepSettingsTestDB は clinic_integrations テーブルを用意する。
// 本番マイグレーション（001_init.sql）では UNIQUE(clinic_id, service, key_name) が
// 定義されているが、GORM AutoMigrate はモデル（ClinicIntegration）に uniqueIndex タグが
// 無いため複合 UNIQUE を再現しない。Upsert の ON CONFLICT を意味のある形で検証するため、
// ここで明示的に複合 UNIQUE を追加する（lstep_friend_attribute_snapshot_repository_test.go と同型）。
func setupLstepSettingsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.ClinicIntegration{}))
	db.Exec("TRUNCATE TABLE clinic_integrations CASCADE")
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_test_clinic_integrations_conflict
		ON clinic_integrations (clinic_id, service, key_name)`)
	return db
}

func TestLstepSettingsRepository_Upsert(t *testing.T) {
	db := setupLstepSettingsTestDB(t)
	repo := NewLstepSettingsRepository(db)
	ctx := context.Background()

	t.Run("新規作成できる", func(t *testing.T) {
		require.NoError(t, repo.Upsert(ctx, &model.ClinicIntegration{
			ClinicID: 1,
			Service:  model.IntegrationServiceLstep,
			KeyName:  model.IntegrationKeyLstepAPIKey,
			KeyValue: "original-value",
		}))

		var stored model.ClinicIntegration
		require.NoError(t, db.Where("clinic_id = ? AND service = ? AND key_name = ?", 1, model.IntegrationServiceLstep, model.IntegrationKeyLstepAPIKey).First(&stored).Error)
		assert.Equal(t, "original-value", stored.KeyValue)
	})

	t.Run("同一 (clinic_id, service, key_name) は UPSERT で key_value のみ更新し重複行を作らない", func(t *testing.T) {
		clinicID := uint64(2)
		require.NoError(t, repo.Upsert(ctx, &model.ClinicIntegration{
			ClinicID: clinicID,
			Service:  model.IntegrationServiceLstep,
			KeyName:  model.IntegrationKeyLstepBaseURL,
			KeyValue: "https://old.example.com",
		}))
		require.NoError(t, repo.Upsert(ctx, &model.ClinicIntegration{
			ClinicID: clinicID,
			Service:  model.IntegrationServiceLstep,
			KeyName:  model.IntegrationKeyLstepBaseURL,
			KeyValue: "https://new.example.com",
		}))

		var count int64
		require.NoError(t, db.Model(&model.ClinicIntegration{}).
			Where("clinic_id = ? AND service = ? AND key_name = ?", clinicID, model.IntegrationServiceLstep, model.IntegrationKeyLstepBaseURL).
			Count(&count).Error)
		assert.Equal(t, int64(1), count, "重複行が作られてはならない")

		var stored model.ClinicIntegration
		require.NoError(t, db.Where("clinic_id = ? AND service = ? AND key_name = ?", clinicID, model.IntegrationServiceLstep, model.IntegrationKeyLstepBaseURL).First(&stored).Error)
		assert.Equal(t, "https://new.example.com", stored.KeyValue, "key_value は最新値に更新されるべき")
	})
}

func TestLstepSettingsRepository_Upsert_ClinicIsolation(t *testing.T) {
	db := setupLstepSettingsTestDB(t)
	repo := NewLstepSettingsRepository(db)
	ctx := context.Background()

	t.Run("同一 service+key_name でも clinic_id が異なれば別レコードとして共存する", func(t *testing.T) {
		const clinicA, clinicB = uint64(50), uint64(60)
		require.NoError(t, repo.Upsert(ctx, &model.ClinicIntegration{
			ClinicID: clinicA, Service: model.IntegrationServiceLstep, KeyName: model.IntegrationKeyLstepAPIKey, KeyValue: "clinic-a-key",
		}))
		require.NoError(t, repo.Upsert(ctx, &model.ClinicIntegration{
			ClinicID: clinicB, Service: model.IntegrationServiceLstep, KeyName: model.IntegrationKeyLstepAPIKey, KeyValue: "clinic-b-key",
		}))

		var count int64
		require.NoError(t, db.Model(&model.ClinicIntegration{}).
			Where("service = ? AND key_name = ?", model.IntegrationServiceLstep, model.IntegrationKeyLstepAPIKey).
			Count(&count).Error)
		assert.Equal(t, int64(2), count, "clinic_id が異なれば衝突せず両方作成される")

		var storedA, storedB model.ClinicIntegration
		require.NoError(t, db.Where("clinic_id = ?", clinicA).First(&storedA).Error)
		require.NoError(t, db.Where("clinic_id = ?", clinicB).First(&storedB).Error)
		assert.Equal(t, "clinic-a-key", storedA.KeyValue)
		assert.Equal(t, "clinic-b-key", storedB.KeyValue)
	})
}

func TestLstepSettingsRepository_FindByClinicAndService(t *testing.T) {
	db := setupLstepSettingsTestDB(t)
	repo := NewLstepSettingsRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(10), uint64(20)
	require.NoError(t, db.Create(&model.ClinicIntegration{ClinicID: clinicA, Service: model.IntegrationServiceLstep, KeyName: model.IntegrationKeyLstepAPIKey, KeyValue: "a-key"}).Error)
	require.NoError(t, db.Create(&model.ClinicIntegration{ClinicID: clinicA, Service: model.IntegrationServiceLstep, KeyName: model.IntegrationKeyLstepBaseURL, KeyValue: "a-url"}).Error)
	require.NoError(t, db.Create(&model.ClinicIntegration{ClinicID: clinicA, Service: "other-service", KeyName: "unrelated", KeyValue: "irrelevant"}).Error)
	require.NoError(t, db.Create(&model.ClinicIntegration{ClinicID: clinicB, Service: model.IntegrationServiceLstep, KeyName: model.IntegrationKeyLstepAPIKey, KeyValue: "b-key"}).Error)

	t.Run("同一クリニック・サービスの設定のみ返す", func(t *testing.T) {
		records, err := repo.FindByClinicAndService(ctx, clinicA, model.IntegrationServiceLstep)
		require.NoError(t, err)
		require.Len(t, records, 2)
		for _, r := range records {
			assert.Equal(t, clinicA, r.ClinicID)
			assert.Equal(t, model.IntegrationServiceLstep, r.Service)
		}
	})

	t.Run("別クリニックの設定は含まれない（clinic_id 分離）", func(t *testing.T) {
		records, err := repo.FindByClinicAndService(ctx, clinicB, model.IntegrationServiceLstep)
		require.NoError(t, err)
		require.Len(t, records, 1)
		assert.Equal(t, "b-key", records[0].KeyValue)
	})

	t.Run("該当なしは空スライスを返す", func(t *testing.T) {
		records, err := repo.FindByClinicAndService(ctx, clinicA, "nonexistent-service")
		require.NoError(t, err)
		assert.Empty(t, records)
	})
}

func TestLstepSettingsRepository_FindCredentialByClinicServiceKey(t *testing.T) {
	db := setupLstepSettingsTestDB(t)
	repo := NewLstepSettingsRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(71), uint64(72)
	require.NoError(t, db.Create(&model.ClinicIntegration{
		ClinicID: clinicA, Service: model.IntegrationServiceLstep,
		KeyName: model.IntegrationKeyLineChannelSecret, KeyValue: "canonical-a-placeholder",
	}).Error)
	require.NoError(t, db.Create(&model.ClinicIntegration{
		ClinicID: clinicA, Service: model.IntegrationServiceLstep,
		KeyName: model.IntegrationKeyLstepAPIKey, KeyValue: "other-key-placeholder",
	}).Error)
	require.NoError(t, db.Create(&model.ClinicIntegration{
		ClinicID: clinicB, Service: model.IntegrationServiceLstep,
		KeyName: model.IntegrationKeyLineChannelSecret, KeyValue: "canonical-b-placeholder",
	}).Error)

	got, err := repo.FindCredentialByClinicServiceKey(
		ctx,
		clinicA,
		model.IntegrationServiceLstep,
		model.IntegrationKeyLineChannelSecret,
	)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, clinicA, got.ClinicID)
	assert.Equal(t, model.IntegrationServiceLstep, got.Service)
	assert.Equal(t, model.IntegrationKeyLineChannelSecret, got.KeyName)
	assert.Equal(t, "canonical-a-placeholder", got.KeyValue)

	_, err = repo.FindCredentialByClinicServiceKey(
		ctx,
		clinicA,
		model.IntegrationServiceLstep,
		"missing-key",
	)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestLstepSettingsRepository_FindCredentialByClinicServiceKey_ForeignOnlyMatchIsNotFound(t *testing.T) {
	db := setupLstepSettingsTestDB(t)
	repo := NewLstepSettingsRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(81), uint64(82)
	require.NoError(t, db.Create(&model.ClinicIntegration{
		ClinicID: clinicB,
		Service:  model.IntegrationServiceLstep,
		KeyName:  model.IntegrationKeyLineChannelSecret,
		KeyValue: "foreign-only-placeholder",
	}).Error)

	got, err := repo.FindCredentialByClinicServiceKey(
		ctx,
		clinicA,
		model.IntegrationServiceLstep,
		model.IntegrationKeyLineChannelSecret,
	)

	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, got)
}

func TestLstepSettingsRepository_DeleteByClinicAndService(t *testing.T) {
	db := setupLstepSettingsTestDB(t)
	repo := NewLstepSettingsRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(30), uint64(40)
	require.NoError(t, db.Create(&model.ClinicIntegration{ClinicID: clinicA, Service: model.IntegrationServiceLstep, KeyName: model.IntegrationKeyLstepAPIKey, KeyValue: "a-key"}).Error)
	require.NoError(t, db.Create(&model.ClinicIntegration{ClinicID: clinicA, Service: "other-service", KeyName: "unrelated", KeyValue: "irrelevant"}).Error)
	require.NoError(t, db.Create(&model.ClinicIntegration{ClinicID: clinicB, Service: model.IntegrationServiceLstep, KeyName: model.IntegrationKeyLstepAPIKey, KeyValue: "b-key"}).Error)

	require.NoError(t, repo.DeleteByClinicAndService(ctx, clinicA, model.IntegrationServiceLstep))

	t.Run("指定クリニック・サービスの設定は削除される", func(t *testing.T) {
		var count int64
		require.NoError(t, db.Model(&model.ClinicIntegration{}).Where("clinic_id = ? AND service = ?", clinicA, model.IntegrationServiceLstep).Count(&count).Error)
		assert.Equal(t, int64(0), count)
	})

	t.Run("同一クリニックの別サービスは削除されない", func(t *testing.T) {
		var count int64
		require.NoError(t, db.Model(&model.ClinicIntegration{}).Where("clinic_id = ? AND service = ?", clinicA, "other-service").Count(&count).Error)
		assert.Equal(t, int64(1), count)
	})

	t.Run("別クリニックの同サービスは削除されない（clinic_id 分離）", func(t *testing.T) {
		var count int64
		require.NoError(t, db.Model(&model.ClinicIntegration{}).Where("clinic_id = ? AND service = ?", clinicB, model.IntegrationServiceLstep).Count(&count).Error)
		assert.Equal(t, int64(1), count)
	})
}
