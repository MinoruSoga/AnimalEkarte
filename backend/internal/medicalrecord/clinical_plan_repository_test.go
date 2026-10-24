package medicalrecord

// clinical_plan_repository_test.go — ClinicalPlanRepository の統合テスト（内部カバレッジ向上）。
// 移動元 internal/repository（BE9-2D sub-batch④a）。setupTestDB/ensureAutoMigrated は
// testdb.SetupTestDB/EnsureAutoMigrated 直呼びへ書き換え（vaccination_repository_test.go 先例）。
//
// 対象: FindByMedicalRecordID / Create / Update / Delete / ListDiagnoses / ReplaceDiagnoses / DeleteDiagnoses
// 検証観点: 正常系、clinic_id（clinical_plans.clinic_id は NOT NULL）による隔離、ソフトデリート除外、
//           NotFound ラップ、診断行（clinical_plan_diagnoses）の置換・同値 no-op・クリア・論理削除。
//
// EMR-495: clinical_plans の診断 4 列と medical_records.status は 001 で廃止された。診断は行で保持し、
// 親カルテの旧 finalized による Update / Delete の拒否（SD-2 系）は廃止した。

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
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.DiagnosisType{}, &model.DiagnosisName{}, &model.ClinicalPlan{}, &model.ClinicalPlanDiagnosis{}))
	db.Exec("TRUNCATE TABLE clinical_plan_diagnoses CASCADE")
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

// TestClinicalPlanRepository_Create_FindByMedicalRecordID は作成・取得（clinic_id 必須・
// medical_records 経由の clinic_id 隔離・NotFound）を検証する。
func TestClinicalPlanRepository_Create_FindByMedicalRecordID(t *testing.T) {
	db := setupClinicalPlanTestDB(t)
	repo := NewClinicalPlanRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	mrA := makeClinicalPlanMedicalRecord(t, db, clinicA, "MR-CP-A-001")

	plan := &model.ClinicalPlan{
		ClinicID:         clinicA,
		MedicalRecordID:  mrA.ID,
		PhysicalExam:     "元気消沈",
		DiagnosisDetails: "嘔吐あり",
		TreatmentPolicy:  "対症療法",
	}

	t.Run("Create で新規作成される", func(t *testing.T) {
		require.NoError(t, repo.Create(ctx, plan))
		assert.NotZero(t, plan.ID)
	})

	t.Run("clinic_id 未設定の Create は拒否される（clinical_plans.clinic_id は NOT NULL）", func(t *testing.T) {
		mrNoClinic := makeClinicalPlanMedicalRecord(t, db, clinicA, "MR-CP-A-NOCLINIC")
		err := repo.Create(ctx, &model.ClinicalPlan{MedicalRecordID: mrNoClinic.ID})
		require.Error(t, err)
	})

	t.Run("自医院で取得できる", func(t *testing.T) {
		got, err := repo.FindByMedicalRecordID(ctx, clinicA, mrA.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "対症療法", got.TreatmentPolicy)
		assert.Equal(t, clinicA, got.ClinicID)
	})

	t.Run("別医院からの取得は NotFound（medical_records JOIN 経由の隔離）", func(t *testing.T) {
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
	plan := &model.ClinicalPlan{ClinicID: clinicA, MedicalRecordID: mrA.ID, PhysicalExam: "更新前"}
	require.NoError(t, repo.Create(ctx, plan))

	t.Run("正常系: physical_exam が更新される", func(t *testing.T) {
		exam := "更新後"
		require.NoError(t, repo.Update(ctx, clinicA, plan.ID, UpdateClinicalPlanInput{PhysicalExam: &exam}, nil))

		got, err := repo.FindByMedicalRecordID(ctx, clinicA, mrA.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新後", got.PhysicalExam)
	})

	t.Run("別医院からの更新は NotFound", func(t *testing.T) {
		exam := "乗っ取り"
		err := repo.Update(ctx, clinicB, plan.ID, UpdateClinicalPlanInput{PhysicalExam: &exam}, nil)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDの更新は NotFound", func(t *testing.T) {
		exam := "存在しない"
		err := repo.Update(ctx, clinicA, 9999999, UpdateClinicalPlanInput{PhysicalExam: &exam}, nil)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// TestClinicalPlanRepository_Delete はソフトデリート（JOIN による clinic_id 隔離・NotFound）を検証する。
func TestClinicalPlanRepository_Delete(t *testing.T) {
	db := setupClinicalPlanTestDB(t)
	repo := NewClinicalPlanRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	mrA := makeClinicalPlanMedicalRecord(t, db, clinicA, "MR-CP-A-003")
	plan := &model.ClinicalPlan{ClinicID: clinicA, MedicalRecordID: mrA.ID, PhysicalExam: "削除対象"}
	require.NoError(t, repo.Create(ctx, plan))

	t.Run("別医院からの削除は NotFound", func(t *testing.T) {
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
}

// diagnosisFixture は診断行テスト用の plan と診断マスタを用意する。
func diagnosisFixture(t *testing.T, db *gorm.DB, repo ClinicalPlanRepository, clinicID uint64, recordNo string) (*model.ClinicalPlan, *model.DiagnosisType, *model.DiagnosisName) {
	t.Helper()
	mr := makeClinicalPlanMedicalRecord(t, db, clinicID, recordNo)
	dt := makeClinicalPlanDiagnosisType(t, db, clinicID, "内科-"+recordNo)
	dn := makeClinicalPlanDiagnosisName(t, db, clinicID, dt.ID, "胃炎-"+recordNo)
	plan := &model.ClinicalPlan{ClinicID: clinicID, MedicalRecordID: mr.ID}
	require.NoError(t, repo.Create(context.Background(), plan))
	return plan, dt, dn
}

// EMR-495 A: 診断行の置換・同値 no-op・クリア・clinic 隔離・論理削除。
func TestClinicalPlanRepository_Diagnoses_ReplaceListDelete(t *testing.T) {
	db := setupClinicalPlanTestDB(t)
	repo := NewClinicalPlanRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	plan, dt, dn := diagnosisFixture(t, db, repo, clinicA, "MR-CP-DIAG-001")
	snap := "胃炎(保存時)"
	cat := "内科(保存時)"
	first := []ClinicalPlanDiagnosisInput{
		{Slot: 1, DiagnosisTypeID: &dt.ID, DiagnosisNameID: &dn.ID, DiagnosisNameSnapshot: &snap, CategoryNameSnapshot: &cat, SortOrder: 0},
		{Slot: 2, DiagnosisNameSnapshot: &snap, SortOrder: 0},
	}

	t.Run("置換でスナップショットを送信値のまま保存し、slot/sort_order 昇順で返す", func(t *testing.T) {
		require.NoError(t, repo.ReplaceDiagnoses(ctx, clinicA, plan.ID, first))

		rows, err := repo.ListDiagnoses(ctx, clinicA, plan.ID)
		require.NoError(t, err)
		require.Len(t, rows, 2)
		assert.Equal(t, int16(1), rows[0].DiagnosisSlot)
		require.NotNil(t, rows[0].DiagnosisNameSnapshot)
		assert.Equal(t, snap, *rows[0].DiagnosisNameSnapshot)
		assert.Equal(t, clinicA, rows[0].ClinicID)
		assert.Equal(t, int16(2), rows[1].DiagnosisSlot)
		assert.Nil(t, rows[1].DiagnosisTypeID, "type/name は nullable。マスタから補完しない")
	})

	t.Run("同値の置換は no-op（行 ID が変わらない）", func(t *testing.T) {
		before, err := repo.ListDiagnoses(ctx, clinicA, plan.ID)
		require.NoError(t, err)

		require.NoError(t, repo.ReplaceDiagnoses(ctx, clinicA, plan.ID, first))

		after, err := repo.ListDiagnoses(ctx, clinicA, plan.ID)
		require.NoError(t, err)
		require.Len(t, after, len(before))
		for i := range before {
			assert.Equal(t, before[i].ID, after[i].ID)
		}
	})

	t.Run("変更は旧行を論理削除して新行を挿入する", func(t *testing.T) {
		changed := []ClinicalPlanDiagnosisInput{{Slot: 1, DiagnosisNameSnapshot: &snap}}
		require.NoError(t, repo.ReplaceDiagnoses(ctx, clinicA, plan.ID, changed))

		rows, err := repo.ListDiagnoses(ctx, clinicA, plan.ID)
		require.NoError(t, err)
		require.Len(t, rows, 1)

		var all []model.ClinicalPlanDiagnosis
		require.NoError(t, db.WithContext(ctx).Where("clinical_plan_id = ?", plan.ID).Find(&all).Error)
		assert.Len(t, all, 3, "旧 2 行は物理削除されず deleted_at が設定される")
	})

	t.Run("別医院からの ListDiagnoses は空", func(t *testing.T) {
		rows, err := repo.ListDiagnoses(ctx, clinicB, plan.ID)
		require.NoError(t, err)
		assert.Empty(t, rows)
	})

	t.Run("空配列の置換は全行クリア", func(t *testing.T) {
		require.NoError(t, repo.ReplaceDiagnoses(ctx, clinicA, plan.ID, []ClinicalPlanDiagnosisInput{}))
		rows, err := repo.ListDiagnoses(ctx, clinicA, plan.ID)
		require.NoError(t, err)
		assert.Empty(t, rows)
	})

	t.Run("DeleteDiagnoses は生存行を論理削除する", func(t *testing.T) {
		require.NoError(t, repo.ReplaceDiagnoses(ctx, clinicA, plan.ID, first))
		require.NoError(t, repo.DeleteDiagnoses(ctx, clinicA, plan.ID))
		rows, err := repo.ListDiagnoses(ctx, clinicA, plan.ID)
		require.NoError(t, err)
		assert.Empty(t, rows)
	})
}
