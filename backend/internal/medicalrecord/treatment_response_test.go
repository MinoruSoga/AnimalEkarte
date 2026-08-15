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

func TestTreatmentResponsesRejectLoadedMastersOutsideParentClinic(t *testing.T) {
	const (
		clinicA       = uint64(1)
		clinicB       = uint64(2)
		consultation  = uint64(10)
		procedure     = uint64(20)
		medicine      = uint64(30)
		inventory     = uint64(40)
		medicalRecord = uint64(50)
	)
	treatment := &model.Treatment{
		ID:              1,
		MedicalRecordID: medicalRecord,
		ItemType:        model.TreatmentItemTypeOther,
		ConsultationID:  ptrUint64(consultation),
		ProcedureID:     ptrUint64(procedure),
		MedicineID:      ptrUint64(medicine),
		InventoryID:     ptrUint64(inventory),
		MedicalRecord:   &model.MedicalRecord{ID: medicalRecord, ClinicID: clinicA},
		Consultation:    &model.Consultation{ID: consultation, ClinicID: clinicB, Name: "別院診察"},
		Procedure:       &model.Procedure{ID: procedure, ClinicID: clinicB, Name: "別院処置"},
		Medicine:        &model.Medicine{ID: medicine, ClinicID: clinicB, Name: "別院薬剤"},
		Inventory:       &model.InventoryItem{ID: inventory, ClinicID: clinicB, Name: "別院在庫"},
	}

	response := toTreatmentResponse(treatment)
	assert.Nil(t, response.ConsultationID)
	assert.Nil(t, response.ProcedureID)
	assert.Nil(t, response.MedicineID)
	assert.Nil(t, response.InventoryID)

	historyResponse := toPetTreatmentHistoryResponse(treatment)
	assert.Nil(t, historyResponse.ProcedureID)
	assert.Nil(t, historyResponse.ProcedureName)
	assert.Nil(t, historyResponse.MedicineID)
	assert.Nil(t, historyResponse.MedicineName)
	assert.Nil(t, historyResponse.Anesthesia)
	assert.Nil(t, historyResponse.IsSurgery)
}
