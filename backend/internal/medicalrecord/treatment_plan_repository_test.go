package medicalrecord

// treatment_plan_repository_test.go — TreatmentPlanRepository 統合テスト。
//
// 保護する不変条件:
//   - FindByMedicalRecordID / FindByHospitalizationID / FindByID / Update / Delete は
//     clinic_id でテナント隔離される（clinicScopeQuery 経由）。
//   - FindByMedicalRecordID / FindByHospitalizationID は sort_order ASC で返す。
//   - Update / Delete は対象なしで NotFound を返す。
//   - Delete はソフトデリートであり、以後 FindByID から除外される。

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

// setupTreatmentPlanTestDB は treatment_plans / hospitalizations / medical_records を整備する。
// Hospitalization.CageID / DoctorID は本テストでは使わない（nil のまま）ため cages / staffs は
// AutoMigrate しない（master_preload_clinic_isolation_test.go の既存パターンを踏襲）。
func setupTreatmentPlanTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}, &model.Hospitalization{}, &model.TreatmentPlan{}))
	db.Exec("TRUNCATE TABLE treatment_plans CASCADE")
	db.Exec("TRUNCATE TABLE hospitalizations CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

// makeTreatmentPlanMedicalRecord はテスト用 MedicalRecord を作成する（pet_id なしで良い場合の軽量版）。
func makeTreatmentPlanMedicalRecord(t *testing.T, db *gorm.DB, clinicID uint64, recordNo string) *model.MedicalRecord {
	t.Helper()
	mr := &model.MedicalRecord{
		ClinicID: clinicID,
		RecordNo: recordNo,
		Date:     time.Now(),
		Status:   model.MedicalRecordStatusFinalized,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(mr).Error)
	return mr
}

// makeTreatmentPlan はテスト用 TreatmentPlan を作成して返す。
func makeTreatmentPlan(t *testing.T, db *gorm.DB, clinicID uint64, medicalRecordID, hospitalizationID *uint64, content string, sortOrder int) *model.TreatmentPlan {
	t.Helper()
	tp := &model.TreatmentPlan{
		ClinicID:          clinicID,
		MedicalRecordID:   medicalRecordID,
		HospitalizationID: hospitalizationID,
		TreatmentContent:  content,
		SortOrder:         sortOrder,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(tp).Error)
	return tp
}

func TestTreatmentPlanRepository_FindByMedicalRecordID(t *testing.T) {
	db := setupTreatmentPlanTestDB(t)
	repo := NewTreatmentPlanRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	mrA := makeTreatmentPlanMedicalRecord(t, db, clinicA, "TP-MR-A")
	mrOther := makeTreatmentPlanMedicalRecord(t, db, clinicA, "TP-MR-OTHER")
	mrB := makeTreatmentPlanMedicalRecord(t, db, clinicB, "TP-MR-B")

	planSecond := makeTreatmentPlan(t, db, clinicA, &mrA.ID, nil, "後・治療計画", 2)
	planFirst := makeTreatmentPlan(t, db, clinicA, &mrA.ID, nil, "先・治療計画", 1)
	makeTreatmentPlan(t, db, clinicA, &mrOther.ID, nil, "別カルテの計画", 1)
	makeTreatmentPlan(t, db, clinicB, &mrB.ID, nil, "別クリニックの計画", 1)

	t.Run("returns plans for the medical record ordered by sort_order ASC", func(t *testing.T) {
		got, err := repo.FindByMedicalRecordID(ctx, clinicA, mrA.ID)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, planFirst.ID, got[0].ID)
		assert.Equal(t, planSecond.ID, got[1].ID)
	})

	t.Run("clinic isolation: different clinic scope returns empty", func(t *testing.T) {
		got, err := repo.FindByMedicalRecordID(ctx, clinicB, mrA.ID)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("empty for medical record with no plans", func(t *testing.T) {
		got, err := repo.FindByMedicalRecordID(ctx, clinicA, uint64(999999))
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestTreatmentPlanRepository_FindByHospitalizationID(t *testing.T) {
	db := setupTreatmentPlanTestDB(t)
	repo := NewTreatmentPlanRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	ownerA := makeTestOwner(t, db, clinicA, "入院飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "入院ポチA")
	hospA := makeHospitalizationRec(t, db, clinicA, ownerA.ID, petA.ID, nil)
	hospOther := makeHospitalizationRec(t, db, clinicA, ownerA.ID, petA.ID, nil)

	ownerB := makeTestOwner(t, db, clinicB, "入院飼主B")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "入院ポチB")
	hospB := makeHospitalizationRec(t, db, clinicB, ownerB.ID, petB.ID, nil)

	planSecond := makeTreatmentPlan(t, db, clinicA, nil, &hospA.ID, "後・入院治療計画", 2)
	planFirst := makeTreatmentPlan(t, db, clinicA, nil, &hospA.ID, "先・入院治療計画", 1)
	makeTreatmentPlan(t, db, clinicA, nil, &hospOther.ID, "別入院の計画", 1)
	makeTreatmentPlan(t, db, clinicB, nil, &hospB.ID, "別クリニック入院の計画", 1)

	t.Run("returns plans for the hospitalization ordered by sort_order ASC", func(t *testing.T) {
		got, err := repo.FindByHospitalizationID(ctx, clinicA, hospA.ID)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, planFirst.ID, got[0].ID)
		assert.Equal(t, planSecond.ID, got[1].ID)
	})

	t.Run("clinic isolation: different clinic scope returns empty", func(t *testing.T) {
		got, err := repo.FindByHospitalizationID(ctx, clinicB, hospA.ID)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("empty for hospitalization with no plans", func(t *testing.T) {
		got, err := repo.FindByHospitalizationID(ctx, clinicA, uint64(999999))
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestTreatmentPlanRepository_FindByID(t *testing.T) {
	db := setupTreatmentPlanTestDB(t)
	repo := NewTreatmentPlanRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	mrA := makeTreatmentPlanMedicalRecord(t, db, clinicA, "TP-FIND-A")
	plan := makeTreatmentPlan(t, db, clinicA, &mrA.ID, nil, "対象計画", 1)

	t.Run("found", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, plan.ID)
		require.NoError(t, err)
		assert.Equal(t, plan.ID, got.ID)
	})

	t.Run("not found for nonexistent id", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, uint64(999999))
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("clinic isolation", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, plan.ID)
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("soft-deleted plan is not found", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, plan.ID, nil, nil))
		got, err := repo.FindByID(ctx, clinicA, plan.ID)
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestTreatmentPlanRepository_Create(t *testing.T) {
	db := setupTreatmentPlanTestDB(t)
	repo := NewTreatmentPlanRepository(db)
	ctx := context.Background()

	const clinicA = uint64(1)
	mrA := makeTreatmentPlanMedicalRecord(t, db, clinicA, "TP-CREATE-A")

	plan := &model.TreatmentPlan{ClinicID: clinicA, MedicalRecordID: &mrA.ID, TreatmentContent: "新規治療計画"}
	require.NoError(t, repo.Create(ctx, plan))
	assert.NotZero(t, plan.ID)
}

func TestTreatmentPlanRepository_Update(t *testing.T) {
	db := setupTreatmentPlanTestDB(t)
	repo := NewTreatmentPlanRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	mrA := makeTreatmentPlanMedicalRecord(t, db, clinicA, "TP-UPDATE-A")
	plan := makeTreatmentPlan(t, db, clinicA, &mrA.ID, nil, "更新前", 1)

	t.Run("updates successfully", func(t *testing.T) {
		require.NoError(t, repo.Update(ctx, clinicA, plan.ID, nil, nil, map[string]any{"treatment_content": "更新後"}))
		got, err := repo.FindByID(ctx, clinicA, plan.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新後", got.TreatmentContent)
	})

	t.Run("not found for nonexistent id", func(t *testing.T) {
		err := repo.Update(ctx, clinicA, uint64(999999), nil, nil, map[string]any{"treatment_content": "x"})
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("clinic isolation: wrong clinic returns NotFound", func(t *testing.T) {
		err := repo.Update(ctx, clinicB, plan.ID, nil, nil, map[string]any{"treatment_content": "乗っ取り"})
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("parent bind mismatch returns NotFound (MRD-03)", func(t *testing.T) {
		wrongMR := uint64(999999)
		err := repo.Update(ctx, clinicA, plan.ID, &wrongMR, nil, map[string]any{"treatment_content": "parent mismatch"})
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("parent bind match updates successfully (MRD-03)", func(t *testing.T) {
		mrID := mrA.ID
		require.NoError(t, repo.Update(ctx, clinicA, plan.ID, &mrID, nil, map[string]any{"treatment_content": "parent ok"}))
		got, err := repo.FindByID(ctx, clinicA, plan.ID)
		require.NoError(t, err)
		assert.Equal(t, "parent ok", got.TreatmentContent)
	})
}

func TestTreatmentPlanRepository_Delete(t *testing.T) {
	db := setupTreatmentPlanTestDB(t)
	repo := NewTreatmentPlanRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	mrA := makeTreatmentPlanMedicalRecord(t, db, clinicA, "TP-DELETE-A")
	plan := makeTreatmentPlan(t, db, clinicA, &mrA.ID, nil, "削除対象", 1)

	t.Run("clinic isolation: wrong clinic cannot delete", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, plan.ID, nil, nil)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("soft-deletes successfully", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, plan.ID, nil, nil))

		var unscoped int64
		require.NoError(t, db.Unscoped().Model(&model.TreatmentPlan{}).Where("id = ?", plan.ID).Count(&unscoped).Error)
		assert.Equal(t, int64(1), unscoped, "物理的には行が残る（ソフトデリート）")

		var scoped int64
		require.NoError(t, db.Model(&model.TreatmentPlan{}).Where("id = ?", plan.ID).Count(&scoped).Error)
		assert.Equal(t, int64(0), scoped)
	})

	t.Run("not found for already-deleted id", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, plan.ID, nil, nil)
		assert.True(t, apperrors.IsNotFound(err))
	})
}
