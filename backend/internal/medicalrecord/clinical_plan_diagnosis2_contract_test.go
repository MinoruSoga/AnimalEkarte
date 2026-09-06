package medicalrecord

// 移動元 internal/repository（BE9-2D sub-batch④a）。setupClinicalPlanTestDB /
// makeClinicalPlanMedicalRecord / makeClinicalPlanDiagnosisType / makeClinicalPlanDiagnosisName は
// clinical_plan_repository_test.go の同 package ヘルパーを再利用する。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestClinicalPlanRepository_Diagnosis2_SetChangeClearAndPreload(t *testing.T) {
	db := setupClinicalPlanTestDB(t)
	repo := NewClinicalPlanRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	mr := makeClinicalPlanMedicalRecord(t, db, clinicA, "MR-CP-D2-001")
	typeA := makeClinicalPlanDiagnosisType(t, db, clinicA, "内科")
	nameA := makeClinicalPlanDiagnosisName(t, db, clinicA, typeA.ID, "胃炎")
	typeB := makeClinicalPlanDiagnosisType(t, db, clinicA, "外科")
	nameB := makeClinicalPlanDiagnosisName(t, db, clinicA, typeB.ID, "骨折")

	plan := &model.ClinicalPlan{MedicalRecordID: mr.ID, PhysicalExam: "初期"}
	require.NoError(t, repo.Create(ctx, plan))

	t.Run("設定", func(t *testing.T) {
		typeID := typeA.ID
		nameID := nameA.ID
		typePtr := &typeID
		namePtr := &nameID
		require.NoError(t, repo.Update(ctx, clinicA, plan.ID, UpdateClinicalPlanInput{
			Diagnosis2TypeID: &typePtr,
			Diagnosis2NameID: &namePtr,
		}, nil))
		got, err := repo.FindByMedicalRecordID(ctx, clinicA, mr.ID)
		require.NoError(t, err)
		require.NotNil(t, got.Diagnosis2TypeID)
		assert.Equal(t, typeA.ID, *got.Diagnosis2TypeID)
		require.NotNil(t, got.Diagnosis2Type)
		assert.Equal(t, "内科", got.Diagnosis2Type.Name)
		require.NotNil(t, got.Diagnosis2Name)
		assert.Equal(t, "胃炎", got.Diagnosis2Name.Name)
	})

	t.Run("変更", func(t *testing.T) {
		typeID := typeB.ID
		nameID := nameB.ID
		typePtr := &typeID
		namePtr := &nameID
		require.NoError(t, repo.Update(ctx, clinicA, plan.ID, UpdateClinicalPlanInput{
			Diagnosis2TypeID: &typePtr,
			Diagnosis2NameID: &namePtr,
		}, nil))
		got, err := repo.FindByMedicalRecordID(ctx, clinicA, mr.ID)
		require.NoError(t, err)
		require.NotNil(t, got.Diagnosis2TypeID)
		assert.Equal(t, typeB.ID, *got.Diagnosis2TypeID)
		assert.Equal(t, "骨折", got.Diagnosis2Name.Name)
	})

	t.Run("クリア", func(t *testing.T) {
		var nilType *uint64
		var nilName *uint64
		require.NoError(t, repo.Update(ctx, clinicA, plan.ID, UpdateClinicalPlanInput{
			Diagnosis2TypeID: &nilType,
			Diagnosis2NameID: &nilName,
		}, nil))
		got, err := repo.FindByMedicalRecordID(ctx, clinicA, mr.ID)
		require.NoError(t, err)
		assert.Nil(t, got.Diagnosis2TypeID)
		assert.Nil(t, got.Diagnosis2NameID)
		assert.Nil(t, got.Diagnosis2Type)
		assert.Nil(t, got.Diagnosis2Name)
	})
}
