package medicalrecord

// clinical_plan_repository_test.go — ClinicalPlanRepository の統合テスト（内部カバレッジ向上）。
// 移動元 internal/repository（BE9-2D sub-batch④a）。setupTestDB/ensureAutoMigrated は
// testdb.SetupTestDB/EnsureAutoMigrated 直呼びへ書き換え（vaccination_repository_test.go 先例）。
//
// 対象: FindByMedicalRecordID / Create / Update / Delete
// 検証観点: 正常系、DiagnosisType/DiagnosisName の clinic_id 述語付き Preload、
//           medical_records JOIN/サブクエリ経由の clinic_id 隔離、ソフトデリート除外、NotFound ラップ。
//
// 注: DiagnosisType/DiagnosisName の Preload クロステナント混入防止テストは
//     master_preload_clinic_isolation_test.go に既存のため、本ファイルでは重複させず
//     Create/Update/Delete と自クリニック内の Preload happy path に焦点を当てる。

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

// setupClinicalPlanTestDB は clinical_plans と、その Preload 先である
// diagnosis_types/diagnosis_names を整備する（medical_records は core AutoMigrate 済み）。
func setupClinicalPlanTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.DiagnosisType{}, &model.DiagnosisName{}, &model.ClinicalPlan{}))
	db.Exec("TRUNCATE TABLE clinical_plans CASCADE")
	db.Exec("TRUNCATE TABLE diagnosis_names CASCADE")
	db.Exec("TRUNCATE TABLE diagnosis_types CASCADE")
	return db
}

func makeClinicalPlanMedicalRecord(t *testing.T, db *gorm.DB, clinicID uint64, recordNo string) *model.MedicalRecord {
	t.Helper()
	mr := &model.MedicalRecord{
		ClinicID: clinicID,
		RecordNo: recordNo,
		Date:     time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		Status:   model.MedicalRecordStatusDraft,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(mr).Error)
	return mr
}

func makeClinicalPlanDiagnosisType(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.DiagnosisType {
	t.Helper()
	dt := &model.DiagnosisType{ClinicID: clinicID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(dt).Error)
	return dt
}

func makeClinicalPlanDiagnosisName(t *testing.T, db *gorm.DB, clinicID, typeID uint64, name string) *model.DiagnosisName {
	t.Helper()
	dn := &model.DiagnosisName{ClinicID: clinicID, DiagnosisTypeID: typeID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(dn).Error)
	return dn
}

// TestClinicalPlanRepository_Create_FindByMedicalRecordID は作成・取得（診断マスタ Preload の
// happy path・medical_records 経由の clinic_id 隔離・NotFound）を検証する。
func TestClinicalPlanRepository_Create_FindByMedicalRecordID(t *testing.T) {
	db := setupClinicalPlanTestDB(t)
	repo := NewClinicalPlanRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	mrA := makeClinicalPlanMedicalRecord(t, db, clinicA, "MR-CP-A-001")
	dtA := makeClinicalPlanDiagnosisType(t, db, clinicA, "内科")
	dnA := makeClinicalPlanDiagnosisName(t, db, clinicA, dtA.ID, "急性胃腸炎")

	plan := &model.ClinicalPlan{
		MedicalRecordID:  mrA.ID,
		PhysicalExam:     "元気消沈",
		DiagnosisTypeID:  &dtA.ID,
		DiagnosisNameID:  &dnA.ID,
		DiagnosisDetails: "嘔吐あり",
		TreatmentPolicy:  "対症療法",
	}

	t.Run("Create で新規作成される", func(t *testing.T) {
		require.NoError(t, repo.Create(ctx, plan))
		assert.NotZero(t, plan.ID)
	})

	t.Run("自クリニックで取得すると診断マスタが Preload される", func(t *testing.T) {
		got, err := repo.FindByMedicalRecordID(ctx, clinicA, mrA.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "対症療法", got.TreatmentPolicy)
		require.NotNil(t, got.DiagnosisType)
		assert.Equal(t, "内科", got.DiagnosisType.Name)
		require.NotNil(t, got.DiagnosisName)
		assert.Equal(t, "急性胃腸炎", got.DiagnosisName.Name)
	})

	t.Run("別クリニックからの取得は NotFound（medical_records JOIN 経由の隔離）", func(t *testing.T) {
		_, err := repo.FindByMedicalRecordID(ctx, clinicB, mrA.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("存在しない medical_record_id は NotFound", func(t *testing.T) {
		_, err := repo.FindByMedicalRecordID(ctx, clinicA, 9999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// TestClinicalPlanRepository_Update は部分更新（サブクエリによる clinic_id 隔離・NotFound）を検証する。
func TestClinicalPlanRepository_Update(t *testing.T) {
	db := setupClinicalPlanTestDB(t)
	repo := NewClinicalPlanRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	mrA := makeClinicalPlanMedicalRecord(t, db, clinicA, "MR-CP-A-002")
	plan := &model.ClinicalPlan{MedicalRecordID: mrA.ID, PhysicalExam: "更新前"}
	require.NoError(t, repo.Create(ctx, plan))

	t.Run("正常系: physical_exam が更新される", func(t *testing.T) {
		require.NoError(t, repo.Update(ctx, clinicA, plan.ID, map[string]any{"physical_exam": "更新後"}, nil))

		got, err := repo.FindByMedicalRecordID(ctx, clinicA, mrA.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新後", got.PhysicalExam)
	})

	t.Run("別クリニックからの更新は NotFound", func(t *testing.T) {
		err := repo.Update(ctx, clinicB, plan.ID, map[string]any{"physical_exam": "乗っ取り"}, nil)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDの更新は NotFound", func(t *testing.T) {
		err := repo.Update(ctx, clinicA, 9999999, map[string]any{"physical_exam": "存在しない"}, nil)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("親カルテが確定済みの場合は Conflict（SD-2 系ガード監査）", func(t *testing.T) {
		mrFinalized := makeClinicalPlanMedicalRecord(t, db, clinicA, "MR-CP-A-FINALIZED")
		require.NoError(t, db.WithContext(ctx).Model(&model.MedicalRecord{}).
			Where("id = ?", mrFinalized.ID).Update("status", model.MedicalRecordStatusFinalized).Error)
		finalizedPlan := &model.ClinicalPlan{MedicalRecordID: mrFinalized.ID, PhysicalExam: "確定前"}
		require.NoError(t, repo.Create(ctx, finalizedPlan))

		err := repo.Update(ctx, clinicA, finalizedPlan.ID, map[string]any{"physical_exam": "確定後の書込"}, nil)

		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "確定済みカルテの clinical_plan 更新は Conflict であるべき: %v", err)
	})
}

// TestClinicalPlanRepository_Delete はソフトデリート（JOIN による clinic_id 隔離・NotFound）を検証する。
func TestClinicalPlanRepository_Delete(t *testing.T) {
	db := setupClinicalPlanTestDB(t)
	repo := NewClinicalPlanRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	mrA := makeClinicalPlanMedicalRecord(t, db, clinicA, "MR-CP-A-003")
	plan := &model.ClinicalPlan{MedicalRecordID: mrA.ID, PhysicalExam: "削除対象"}
	require.NoError(t, repo.Create(ctx, plan))

	t.Run("別クリニックからの削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, plan.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("正常系: ソフトデリートされ FindByMedicalRecordID から除外されるが行自体は残る", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, plan.ID))

		_, err := repo.FindByMedicalRecordID(ctx, clinicA, mrA.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		var raw model.ClinicalPlan
		require.NoError(t, db.WithContext(ctx).Unscoped().Where("id = ?", plan.ID).First(&raw).Error)
		assert.True(t, raw.DeletedAt.Valid, "物理削除ではなくソフトデリートであるべき")
	})

	t.Run("存在しないIDの削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 9999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("親カルテが確定済みの場合は Conflict（SD-2 系ガード監査）", func(t *testing.T) {
		mrFinalized := makeClinicalPlanMedicalRecord(t, db, clinicA, "MR-CP-A-DEL-FINALIZED")
		require.NoError(t, db.WithContext(ctx).Model(&model.MedicalRecord{}).
			Where("id = ?", mrFinalized.ID).Update("status", model.MedicalRecordStatusFinalized).Error)
		finalizedPlan := &model.ClinicalPlan{MedicalRecordID: mrFinalized.ID, PhysicalExam: "確定済み"}
		require.NoError(t, repo.Create(ctx, finalizedPlan))

		err := repo.Delete(ctx, clinicA, finalizedPlan.ID)

		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "確定済みカルテの clinical_plan 削除は Conflict であるべき: %v", err)
	})
}
