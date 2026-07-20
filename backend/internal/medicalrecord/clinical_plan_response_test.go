package medicalrecord

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToClinicalPlanResponse(t *testing.T) {
	diagnosisTypeID := uint64(11)
	diagnosisNameID := uint64(12)

	tests := []struct {
		name  string
		input *model.ClinicalPlan
		check func(t *testing.T, resp clinicalPlanResponse)
	}{
		{
			name: "zero-value plan without diagnosis references",
			input: &model.ClinicalPlan{
				ID:              1,
				MedicalRecordID: 5,
			},
			check: func(t *testing.T, resp clinicalPlanResponse) {
				assert.Equal(t, "1", resp.ID)
				assert.Equal(t, "5", resp.MedicalRecordID)
				assert.Nil(t, resp.DiagnosisTypeID)
				assert.Nil(t, resp.DiagnosisNameID)
				assert.Nil(t, resp.DiagnosisType)
				assert.Nil(t, resp.DiagnosisName)
			},
		},
		{
			name: "plan with diagnosis IDs but no preloaded relations",
			input: &model.ClinicalPlan{
				ID:               2,
				MedicalRecordID:  6,
				PhysicalExam:     "元気消失",
				DiagnosisTypeID:  &diagnosisTypeID,
				DiagnosisNameID:  &diagnosisNameID,
				DiagnosisDetails: "詳細",
				TreatmentPolicy:  "経過観察",
			},
			check: func(t *testing.T, resp clinicalPlanResponse) {
				require.NotNil(t, resp.DiagnosisTypeID)
				assert.Equal(t, "11", *resp.DiagnosisTypeID)
				require.NotNil(t, resp.DiagnosisNameID)
				assert.Equal(t, "12", *resp.DiagnosisNameID)
				assert.Nil(t, resp.DiagnosisType, "DiagnosisType relation not preloaded should stay nil")
				assert.Nil(t, resp.DiagnosisName, "DiagnosisName relation not preloaded should stay nil")
				assert.Equal(t, "元気消失", resp.PhysicalExam)
				assert.Equal(t, "詳細", resp.DiagnosisDetails)
				assert.Equal(t, "経過観察", resp.TreatmentPolicy)
			},
		},
		{
			name: "plan with preloaded DiagnosisType and DiagnosisName relations",
			input: &model.ClinicalPlan{
				ID:              3,
				MedicalRecordID: 7,
				DiagnosisTypeID: &diagnosisTypeID,
				DiagnosisNameID: &diagnosisNameID,
				DiagnosisType:   &model.DiagnosisType{ID: diagnosisTypeID, Name: "感染症"},
				DiagnosisName:   &model.DiagnosisName{ID: diagnosisNameID, Name: "猫風邪", DiagnosisTypeID: diagnosisTypeID},
			},
			check: func(t *testing.T, resp clinicalPlanResponse) {
				require.NotNil(t, resp.DiagnosisType)
				assert.Equal(t, "感染症", resp.DiagnosisType.Name)
				require.NotNil(t, resp.DiagnosisName)
				assert.Equal(t, "猫風邪", resp.DiagnosisName.Name)
			},
		},
		{
			name: "plan with diagnosis_2 IDs and preloaded relations",
			input: &model.ClinicalPlan{
				ID:               4,
				MedicalRecordID:  8,
				Diagnosis2TypeID: &diagnosisTypeID,
				Diagnosis2NameID: &diagnosisNameID,
				Diagnosis2Type:   &model.DiagnosisType{ID: diagnosisTypeID, Name: "外科"},
				Diagnosis2Name:   &model.DiagnosisName{ID: diagnosisNameID, Name: "骨折", DiagnosisTypeID: diagnosisTypeID},
			},
			check: func(t *testing.T, resp clinicalPlanResponse) {
				require.NotNil(t, resp.Diagnosis2TypeID)
				assert.Equal(t, "11", *resp.Diagnosis2TypeID)
				require.NotNil(t, resp.Diagnosis2NameID)
				assert.Equal(t, "12", *resp.Diagnosis2NameID)
				require.NotNil(t, resp.Diagnosis2Type)
				assert.Equal(t, "外科", resp.Diagnosis2Type.Name)
				require.NotNil(t, resp.Diagnosis2Name)
				assert.Equal(t, "骨折", resp.Diagnosis2Name.Name)
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
