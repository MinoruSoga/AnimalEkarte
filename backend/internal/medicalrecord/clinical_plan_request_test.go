package medicalrecord

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// EMR-495 A: diagnoses は diagnosis_slot 必須の行配列。未送信 / null = 変更なし、[] = 全クリア。
func TestUpdateClinicalPlanRequest_Diagnoses(t *testing.T) {
	slot1, slot2 := int16(1), int16(2)
	typeID, nameID := uint64(3), uint64(4)
	snap := "胃炎"
	sortOrder := 5

	t.Run("rows are converted with snapshots and sort_order as sent", func(t *testing.T) {
		req := updateClinicalPlanRequest{Diagnoses: &[]clinicalPlanDiagnosisRequest{
			{DiagnosisSlot: &slot1, DiagnosisTypeID: &typeID, DiagnosisNameID: &nameID, DiagnosisNameSnapshot: &snap, SortOrder: &sortOrder},
			{DiagnosisSlot: &slot2},
		}}

		input, err := req.toServiceInput()

		require.NoError(t, err)
		require.NotNil(t, input.Diagnoses)
		rows := *input.Diagnoses
		require.Len(t, rows, 2)
		assert.Equal(t, int16(1), rows[0].Slot)
		assert.Equal(t, typeID, *rows[0].DiagnosisTypeID)
		assert.Equal(t, snap, *rows[0].DiagnosisNameSnapshot)
		assert.Equal(t, 5, rows[0].SortOrder)
		assert.Equal(t, 0, rows[1].SortOrder, "sort_order omitted = 0")
		assert.Nil(t, rows[1].DiagnosisNameSnapshot, "snapshots are not derived from the master")
	})

	t.Run("omitted diagnoses stays nil (rows untouched)", func(t *testing.T) {
		input, err := updateClinicalPlanRequest{}.toServiceInput()
		require.NoError(t, err)
		assert.Nil(t, input.Diagnoses)
	})

	t.Run("empty array stays non-nil (clear all rows)", func(t *testing.T) {
		req := updateClinicalPlanRequest{Diagnoses: &[]clinicalPlanDiagnosisRequest{}}
		input, err := req.toServiceInput()
		require.NoError(t, err)
		require.NotNil(t, input.Diagnoses)
		assert.Empty(t, *input.Diagnoses)
	})

	t.Run("missing diagnosis_slot is invalid input", func(t *testing.T) {
		req := updateClinicalPlanRequest{Diagnoses: &[]clinicalPlanDiagnosisRequest{{DiagnosisNameSnapshot: &snap}}}
		_, err := req.toServiceInput()
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("JSON body binds diagnoses and rejects the retired flat diagnosis fields silently", func(t *testing.T) {
		var req updateClinicalPlanRequest
		require.NoError(t, json.Unmarshal([]byte(`{"diagnoses":[{"diagnosis_slot":2,"diagnosis_name_snapshot":"x"}],"diagnosis_type_id":9}`), &req))
		input, err := req.toServiceInput()
		require.NoError(t, err)
		require.NotNil(t, input.Diagnoses)
		assert.Equal(t, int16(2), (*input.Diagnoses)[0].Slot)
	})
}

// BUG-010: JSON 空文字は明示クリア（pointer to ""）、フィールド欠落は未送信（nil）として区別する。
func TestUpdateClinicalPlanRequest_EmptyStringVsOmitted(t *testing.T) {
	empty := ""
	withEmpty := updateClinicalPlanRequest{
		PhysicalExam:     &empty,
		DiagnosisDetails: &empty,
		TreatmentPolicy:  &empty,
	}
	inputEmpty, err := withEmpty.toServiceInput()
	require.NoError(t, err)
	if inputEmpty.PhysicalExam == nil || *inputEmpty.PhysicalExam != "" {
		t.Fatalf("empty PhysicalExam pointer = %v, want &\"\"", inputEmpty.PhysicalExam)
	}
	if inputEmpty.DiagnosisDetails == nil || *inputEmpty.DiagnosisDetails != "" {
		t.Fatalf("empty DiagnosisDetails pointer = %v, want &\"\"", inputEmpty.DiagnosisDetails)
	}
	if inputEmpty.TreatmentPolicy == nil || *inputEmpty.TreatmentPolicy != "" {
		t.Fatalf("empty TreatmentPolicy pointer = %v, want &\"\"", inputEmpty.TreatmentPolicy)
	}

	omitted := updateClinicalPlanRequest{}
	inputOmitted, err := omitted.toServiceInput()
	require.NoError(t, err)
	if inputOmitted.PhysicalExam != nil {
		t.Fatalf("omitted PhysicalExam = %v, want nil", inputOmitted.PhysicalExam)
	}
	if inputOmitted.DiagnosisDetails != nil {
		t.Fatalf("omitted DiagnosisDetails = %v, want nil", inputOmitted.DiagnosisDetails)
	}
	if inputOmitted.TreatmentPolicy != nil {
		t.Fatalf("omitted TreatmentPolicy = %v, want nil", inputOmitted.TreatmentPolicy)
	}

	fieldsEmpty := buildClinicalPlanUpdate(inputEmpty)
	if fieldsEmpty["physical_exam"] != "" || fieldsEmpty["diagnosis_details"] != "" || fieldsEmpty["treatment_policy"] != "" {
		t.Fatalf("empty clear fields = %#v, want empty strings for all three text columns", fieldsEmpty)
	}
	fieldsOmitted := buildClinicalPlanUpdate(inputOmitted)
	if len(fieldsOmitted) != 0 {
		t.Fatalf("omitted fields = %#v, want empty map (no-op update)", fieldsOmitted)
	}
}
