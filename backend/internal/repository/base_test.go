package repository

// base_test.go — clinicScope / clinicScopeIn / dbOrTx の単体テスト。
//
// 保護する不変条件:
//   - clinicScope は指定 clinic_id のみに絞り込む WHERE スコープを付与する。
//   - clinicScopeIn は複数 clinic_id (IN) に絞り込む WHERE スコープを付与し、
//     空スライスは 0 件（IN (NULL)）を返す。
//   - dbOrTx はアンビエントなトランザクションが無い場合 db.WithContext(ctx) と同等に動作し、
//     Transactor.WithTx 内から呼ばれた場合は自動的に同一トランザクションに参加する。

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// setupBaseTestDB は clinic_scope / dbOrTx 検証用に clinic_integrations テーブルを用意する。
func setupBaseTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.ClinicIntegration{}))
	db.Exec("TRUNCATE TABLE clinic_integrations CASCADE")
	return db
}

func TestClinicScope(t *testing.T) {
	db := setupBaseTestDB(t)
	ctx := context.Background()

	require.NoError(t, db.WithContext(ctx).Create(&model.ClinicIntegration{ClinicID: 1, Service: "lstep", KeyName: "k1", KeyValue: "v1"}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.ClinicIntegration{ClinicID: 1, Service: "lstep", KeyName: "k2", KeyValue: "v2"}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.ClinicIntegration{ClinicID: 2, Service: "lstep", KeyName: "k1", KeyValue: "other"}).Error)

	t.Run("指定clinic_idのみ返す", func(t *testing.T) {
		var records []model.ClinicIntegration
		require.NoError(t, db.WithContext(ctx).Scopes(clinicScope(1)).Find(&records).Error)
		assert.Len(t, records, 2)
		for _, r := range records {
			assert.Equal(t, uint64(1), r.ClinicID)
		}
	})

	t.Run("該当clinic_idが無ければ空になる", func(t *testing.T) {
		var records []model.ClinicIntegration
		require.NoError(t, db.WithContext(ctx).Scopes(clinicScope(999)).Find(&records).Error)
		assert.Empty(t, records)
	})
}

func TestClinicScopeIn(t *testing.T) {
	db := setupBaseTestDB(t)
	ctx := context.Background()

	require.NoError(t, db.WithContext(ctx).Create(&model.ClinicIntegration{ClinicID: 10, Service: "lstep", KeyName: "k", KeyValue: "a"}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.ClinicIntegration{ClinicID: 20, Service: "lstep", KeyName: "k", KeyValue: "b"}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.ClinicIntegration{ClinicID: 30, Service: "lstep", KeyName: "k", KeyValue: "c"}).Error)

	t.Run("複数clinic_idにマッチする行を返す", func(t *testing.T) {
		var records []model.ClinicIntegration
		require.NoError(t, db.WithContext(ctx).Scopes(clinicScopeIn([]uint64{10, 30})).Find(&records).Error)
		assert.Len(t, records, 2)
		ids := []uint64{records[0].ClinicID, records[1].ClinicID}
		assert.ElementsMatch(t, []uint64{10, 30}, ids)
	})

	t.Run("空スライスは0件を返す（IN (NULL)）", func(t *testing.T) {
		var records []model.ClinicIntegration
		require.NoError(t, db.WithContext(ctx).Scopes(clinicScopeIn([]uint64{})).Find(&records).Error)
		assert.Empty(t, records)
	})
}

func TestDbOrTx_NoAmbientTx_UsesBaseDBDirectly(t *testing.T) {
	db := setupBaseTestDB(t)
	ctx := context.Background()

	// アンビエントtxが無いコンテキストでは db.WithContext(ctx) と同等の直接書き込みになる。
	require.NoError(t, dbOrTx(ctx, db).Create(&model.ClinicIntegration{ClinicID: 40, Service: "lstep", KeyName: "k", KeyValue: "direct"}).Error)

	var count int64
	require.NoError(t, db.Model(&model.ClinicIntegration{}).Where("clinic_id = ?", 40).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestDbOrTx_AmbientTx_ParticipatesInSameTransaction(t *testing.T) {
	db := setupBaseTestDB(t)
	tx := NewTransactor(db)
	ctx := context.Background()

	sentinel := errors.New("force rollback")
	err := tx.WithTx(ctx, func(txCtx context.Context) error {
		// dbOrTx がアンビエントtxを検出して同一トランザクションに書き込む。
		require.NoError(t, dbOrTx(txCtx, db).Create(&model.ClinicIntegration{
			ClinicID: 41, Service: "lstep", KeyName: "k", KeyValue: "in-tx",
		}).Error)
		return sentinel
	})
	require.Error(t, err)

	// txの外側からの直接読み込みでも、ロールバックされているため行は存在しないはず。
	var count int64
	require.NoError(t, db.Model(&model.ClinicIntegration{}).Where("clinic_id = ?", 41).Count(&count).Error)
	assert.Equal(t, int64(0), count, "dbOrTx がアンビエントtxに参加していればロールバックと共に消えるはず")
}
