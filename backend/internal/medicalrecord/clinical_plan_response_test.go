package medicalrecord

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-495 A: 診療計画レスポンスは diagnoses[]（契約 §0.3 の closed shape + id / version）を返す。
// 旧 diagnosis_type_id / diagnosis_name_id / diagnosis_2_* と埋め込みマスタは廃止した。
func TestToClinicalPlanResponse(t *testing.T) {
	typeID := uint64(11)
	nameID := uint64(12)
	snapName := "猫風邪(保存時)"
	snapCategory := "感染症(保存時)"

	tests := []struct {
		name  string
		input *ClinicalPlanView
		check func(t *testing.T, resp clinicalPlanResponse)
	}{
		{
			name:  "plan without diagnosis rows returns an empty array, not null",
			input: &ClinicalPlanView{Plan: &model.ClinicalPlan{ID: 1, MedicalRecordID: 5}},
			check: func(t *testing.T, resp clinicalPlanResponse) {
				assert.Equal(t, "1", resp.ID)
				assert.Equal(t, "5", resp.MedicalRecordID)
				require.NotNil(t, resp.Diagnoses)
				assert.Empty(t, resp.Diagnoses)
			},
		},
		{
			name: "rows carry stored snapshots and nullable master references as stored",
			input: &ClinicalPlanView{
				Plan: &model.ClinicalPlan{
					ID: 2, MedicalRecordID: 6, PhysicalExam: "元気消失", DiagnosisDetails: "詳細", TreatmentPolicy: "経過観察", Version: 3,
				},
				Diagnoses: []model.ClinicalPlanDiagnosis{
					{ID: 21, Version: 2, DiagnosisSlot: 1, DiagnosisTypeID: &typeID, DiagnosisNameID: &nameID,
						DiagnosisNameSnapshot: &snapName, CategoryNameSnapshot: &snapCategory, SortOrder: 0},
					{ID: 22, Version: 1, DiagnosisSlot: 2, SortOrder: 1},
				},
			},
			check: func(t *testing.T, resp clinicalPlanResponse) {
				assert.Equal(t, "元気消失", resp.PhysicalExam)
				assert.Equal(t, "詳細", resp.DiagnosisDetails)
				assert.Equal(t, "経過観察", resp.TreatmentPolicy)
				assert.Equal(t, 3, resp.Version)
				require.Len(t, resp.Diagnoses, 2)
				first := resp.Diagnoses[0]
				assert.Equal(t, "21", first.ID)
				assert.Equal(t, int64(2), first.Version)
				assert.Equal(t, int16(1), first.DiagnosisSlot)
				require.NotNil(t, first.DiagnosisTypeID)
				assert.Equal(t, "11", *first.DiagnosisTypeID)
				require.NotNil(t, first.DiagnosisNameID)
				assert.Equal(t, "12", *first.DiagnosisNameID)
				require.NotNil(t, first.DiagnosisNameSnapshot)
				assert.Equal(t, snapName, *first.DiagnosisNameSnapshot)
				require.NotNil(t, first.CategoryNameSnapshot)
				assert.Equal(t, snapCategory, *first.CategoryNameSnapshot)
				second := resp.Diagnoses[1]
				assert.Nil(t, second.DiagnosisTypeID)
				assert.Nil(t, second.DiagnosisNameID)
				assert.Nil(t, second.DiagnosisNameSnapshot, "no snapshot is derived from the current master")
				assert.Equal(t, 1, second.SortOrder)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := toClinicalPlanResponse(tt.input)
			tt.check(t, resp)
		})
	}
}
