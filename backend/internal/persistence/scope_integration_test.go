package persistence

// scope_integration_test.go — ClinicScope / ClinicScopeIn / DBOrTx の統合テスト。
//
// 保護する不変条件:
//   - ClinicScope は指定 clinic_id のみに絞り込む WHERE スコープを付与する。
//   - ClinicScopeIn は複数 clinic_id (IN) に絞り込む WHERE スコープを付与し、
//     空スライスは 0 件（IN (NULL)）を返す。
//   - DBOrTx はアンビエントなトランザクションが無い場合 db.WithContext(ctx) と同等に動作し、
//     Transactor.WithTx 内から呼ばれた場合は自動的に同一トランザクションに参加する。

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupBaseTestDB は ClinicScope / DBOrTx 検証用に clinic_integrations テーブルを用意する。
func setupBaseTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.ClinicIntegration{}))
	require.NoError(t, db.Exec("TRUNCATE TABLE clinic_integrations CASCADE").Error)
	return db
}

// setupMedicalRecordTenantScopeTestDB は MedicalRecordTenantScope 検証用に
// medical_records / treatments を用意する（他リポジトリの CountUsageByXxxID と同じ JOIN 形状）。
func setupMedicalRecordTenantScopeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, db.Exec("TRUNCATE TABLE treatments CASCADE").Error)
	require.NoError(t, db.Exec("TRUNCATE TABLE medical_records CASCADE").Error)
	return db
}

func TestMedicalRecordTenantScope(t *testing.T) {
	db := setupMedicalRecordTenantScopeTestDB(t)
	ctx := context.Background()

	makeMedicalRecord := func(clinicID uint64, deleted bool) *model.MedicalRecord {
		mr := &model.MedicalRecord{ClinicID: clinicID, RecordNo: "R-1", Date: time.Now()}
		require.NoError(t, db.WithContext(ctx).Create(mr).Error)
		if deleted {
			require.NoError(t, db.WithContext(ctx).Delete(mr).Error)
		}
		return mr
	}
	makeTreatment := func(medicalRecordID uint64) *model.Treatment {
		tr := &model.Treatment{MedicalRecordID: medicalRecordID}
		require.NoError(t, db.WithContext(ctx).Create(tr).Error)
		return tr
	}

	// clinic 1: 有効な medical_record 1件 + ソフトデリート済み medical_record 1件
	mrA := makeMedicalRecord(1, false)
	makeTreatment(mrA.ID)
	mrADeleted := makeMedicalRecord(1, true)
	makeTreatment(mrADeleted.ID)
	// clinic 2: 有効な medical_record 1件
	mrB := makeMedicalRecord(2, false)
	makeTreatment(mrB.ID)

	countFor := func(clinicID uint64) int64 {
		var count int64
		require.NoError(t, db.WithContext(ctx).Model(&model.Treatment{}).
			Scopes(MedicalRecordTenantScope("treatments", clinicID)).
			Count(&count).Error)
		return count
	}

	t.Run("指定clinicの有効なmedical_recordに紐づくtreatmentのみJOINで返る", func(t *testing.T) {
		assert.Equal(t, int64(1), countFor(1), "clinic1のtreatment2件のうち、親medical_recordsが論理削除済みの1件はJOINから除外される")
	})

	t.Run("別clinicを指定すると当該clinicの分だけ返る", func(t *testing.T) {
		assert.Equal(t, int64(1), countFor(2))
	})

	t.Run("該当clinicが無ければ0件", func(t *testing.T) {
		assert.Equal(t, int64(0), countFor(999))
	})
}

func TestClinicScope(t *testing.T) {
	db := setupBaseTestDB(t)
	ctx := context.Background()

	require.NoError(t, db.WithContext(ctx).Create(&model.ClinicIntegration{ClinicID: 1, Service: "lstep", KeyName: "k1", KeyValue: "v1"}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.ClinicIntegration{ClinicID: 1, Service: "lstep", KeyName: "k2", KeyValue: "v2"}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.ClinicIntegration{ClinicID: 2, Service: "lstep", KeyName: "k1", KeyValue: "other"}).Error)

	t.Run("指定clinic_idのみ返す", func(t *testing.T) {
		var records []model.ClinicIntegration
		require.NoError(t, db.WithContext(ctx).Scopes(ClinicScope(1)).Find(&records).Error)
		assert.Len(t, records, 2)
		for _, r := range records {
			assert.Equal(t, uint64(1), r.ClinicID)
		}
	})

	t.Run("該当clinic_idが無ければ空になる", func(t *testing.T) {
		var records []model.ClinicIntegration
		require.NoError(t, db.WithContext(ctx).Scopes(ClinicScope(999)).Find(&records).Error)
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
		require.NoError(t, db.WithContext(ctx).Scopes(ClinicScopeIn([]uint64{10, 30})).Find(&records).Error)
		assert.Len(t, records, 2)
		ids := []uint64{records[0].ClinicID, records[1].ClinicID}
		assert.ElementsMatch(t, []uint64{10, 30}, ids)
	})

	t.Run("空スライスは0件を返す（IN (NULL)）", func(t *testing.T) {
		var records []model.ClinicIntegration
		require.NoError(t, db.WithContext(ctx).Scopes(ClinicScopeIn([]uint64{})).Find(&records).Error)
		assert.Empty(t, records)
	})
}

func TestDbOrTx_NoAmbientTx_UsesBaseDBDirectly(t *testing.T) {
	db := setupBaseTestDB(t)
	ctx := context.Background()

	// アンビエントtxが無いコンテキストでは db.WithContext(ctx) と同等の直接書き込みになる。
	require.NoError(t, DBOrTx(ctx, db).Create(&model.ClinicIntegration{ClinicID: 40, Service: "lstep", KeyName: "k", KeyValue: "direct"}).Error)

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
		// DBOrTx がアンビエントtxを検出して同一トランザクションに書き込む。
		require.NoError(t, DBOrTx(txCtx, db).Create(&model.ClinicIntegration{
			ClinicID: 41, Service: "lstep", KeyName: "k", KeyValue: "in-tx",
		}).Error)
		return sentinel
	})
	require.Error(t, err)

	// txの外側からの直接読み込みでも、ロールバックされているため行は存在しないはず。
	var count int64
	require.NoError(t, db.Model(&model.ClinicIntegration{}).Where("clinic_id = ?", 41).Count(&count).Error)
	assert.Equal(t, int64(0), count, "DBOrTx がアンビエントtxに参加していればロールバックと共に消えるはず")
}
