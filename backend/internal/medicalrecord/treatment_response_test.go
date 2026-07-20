package medicalrecord

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// #201 フォローアップ: dose_* 5 フィールドは永続化済み treatment のみ応答に含み、
// 未計算 treatment（processedure/consultation 等）では omitempty で欠落すること。
func TestToTreatmentResponse_DoseFieldsPresentWhenSaved(t *testing.T) {
	weightKg := 4.2
	weightSource := "vital_record:123"
	amountMg := 12.345
	amountUnit := "mg"
	snapshot := json.RawMessage(`{"species":"dog","dose_per_kg":2.5}`)

	treatment := &model.Treatment{
		ID:                1,
		MedicalRecordID:   10,
		ItemType:          model.TreatmentItemTypeMedicine,
		DoseWeightKg:      &weightKg,
		DoseWeightSource:  &weightSource,
		DoseAmountMg:      &amountMg,
		DoseAmountUnit:    &amountUnit,
		DoseParamSnapshot: snapshot,
	}

	resp := toTreatmentResponse(treatment)

	require.NotNil(t, resp.DoseWeightKg)
	assert.Equal(t, weightKg, *resp.DoseWeightKg)
	require.NotNil(t, resp.DoseWeightSource)
	assert.Equal(t, weightSource, *resp.DoseWeightSource)
	require.NotNil(t, resp.DoseAmountMg)
	assert.Equal(t, amountMg, *resp.DoseAmountMg)
	require.NotNil(t, resp.DoseAmountUnit)
	assert.Equal(t, amountUnit, *resp.DoseAmountUnit)
	assert.JSONEq(t, string(snapshot), string(resp.DoseParamSnapshot))

	body, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"dose_weight_kg":4.2`)
	assert.Contains(t, string(body), `"dose_amount_mg":12.345`)
	assert.Contains(t, string(body), `"dose_param_snapshot":{"species":"dog","dose_per_kg":2.5}`)
}

func TestToTreatmentResponse_DoseFieldsOmittedWhenUncalculated(t *testing.T) {
	treatment := &model.Treatment{
		ID:              2,
		MedicalRecordID: 10,
		ItemType:        model.TreatmentItemTypeProcedure,
	}

	resp := toTreatmentResponse(treatment)

	assert.Nil(t, resp.DoseWeightKg)
	assert.Nil(t, resp.DoseWeightSource)
	assert.Nil(t, resp.DoseAmountMg)
	assert.Nil(t, resp.DoseAmountUnit)
	assert.Empty(t, resp.DoseParamSnapshot)

	body, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.NotContains(t, string(body), "dose_weight_kg")
	assert.NotContains(t, string(body), "dose_weight_source")
	assert.NotContains(t, string(body), "dose_amount_mg")
	assert.NotContains(t, string(body), "dose_amount_unit")
	assert.NotContains(t, string(body), "dose_param_snapshot")
}
