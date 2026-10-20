package medicalrecord

// treatment_dose_three_values_test.go — EMR-488: 薬剤3値（1回製剤量・用法・処方総量）の
// request→保存→response→再編集 往復 fixture。
// - 未入力(nil) は NULL、0 は明示 0 として区別して保存・応答する（master・投与日数・数量からの補完はしない）
// - 既存の mg 計算値列（dose_amount_mg / dose_amount_unit / dose_param_snapshot）とは別列で保持し混同しない
// - PATCH は既存方式（nil = 未送信 = 変更なし）に従う

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// exactFromJSON は wire と同一の JSON number から ExactNumeric を作る。
// float64 経由の丸めを介さず、保存契約（numeric(12,6)）の期待値をそのまま固定する。
func exactFromJSON(t *testing.T, src string) *model.ExactNumeric {
	t.Helper()
	var n model.ExactNumeric
	require.NoError(t, json.Unmarshal([]byte(src), &n))
	require.True(t, n.Valid)
	return &n
}

// exactText は ExactNumeric を wire 表現（JSON number text）へ戻す。
func exactText(t *testing.T, n *model.ExactNumeric) string {
	t.Helper()
	raw, err := json.Marshal(n)
	require.NoError(t, err)
	return string(raw)
}

func TestCreateTreatmentRequest_ThreeValues_DecodeAndMap(t *testing.T) {
	body := `{
		"item_type": "medicine",
		"medicine_id": 7,
		"unit_price": 500,
		"quantity": 1,
		"is_selected": true,
		"administration_dose_amount": 2.5,
		"administration_dose_unit": "錠",
		"administration_instructions": "1日2回朝夕食後に",
		"prescription_total_amount": 14,
		"prescription_total_unit": "錠"
	}`
	var req createTreatmentRequest
	require.NoError(t, json.Unmarshal([]byte(body), &req))

	input := req.toServiceInput()

	require.NotNil(t, input.AdministrationDoseAmount)
	assert.Equal(t, "2.5", exactText(t, input.AdministrationDoseAmount))
	require.NotNil(t, input.AdministrationDoseUnit)
	assert.Equal(t, "錠", *input.AdministrationDoseUnit)
	require.NotNil(t, input.AdministrationInstructions)
	assert.Equal(t, "1日2回朝夕食後に", *input.AdministrationInstructions)
	require.NotNil(t, input.PrescriptionTotalAmount)
	assert.Equal(t, "14", exactText(t, input.PrescriptionTotalAmount))
	require.NotNil(t, input.PrescriptionTotalUnit)
	assert.Equal(t, "錠", *input.PrescriptionTotalUnit)
}

func TestCreateTreatmentRequest_ThreeValues_OmittedStaysNil(t *testing.T) {
	// 未入力は NULL 保存。master や quantity からの推測補完が入っていないことを request 段で固定する。
	body := `{"item_type": "medicine", "unit_price": 500, "quantity": 3}`
	var req createTreatmentRequest
	require.NoError(t, json.Unmarshal([]byte(body), &req))

	input := req.toServiceInput()

	assert.Nil(t, input.AdministrationDoseAmount)
	assert.Nil(t, input.AdministrationDoseUnit)
	assert.Nil(t, input.AdministrationInstructions)
	assert.Nil(t, input.PrescriptionTotalAmount)
	assert.Nil(t, input.PrescriptionTotalUnit)
}

func TestCreateTreatmentRequest_ThreeValues_ZeroIsExplicitZero(t *testing.T) {
	// 0 は「未入力」ではなく明示 0。nil へ変換・省略されない。
	body := `{
		"item_type": "medicine",
		"administration_dose_amount": 0,
		"prescription_total_amount": 0
	}`
	var req createTreatmentRequest
	require.NoError(t, json.Unmarshal([]byte(body), &req))

	input := req.toServiceInput()

	require.NotNil(t, input.AdministrationDoseAmount)
	assert.Equal(t, "0", exactText(t, input.AdministrationDoseAmount))
	require.NotNil(t, input.PrescriptionTotalAmount)
	assert.Equal(t, "0", exactText(t, input.PrescriptionTotalAmount))
}

func TestUpdateTreatmentRequest_ThreeValues_PartialAbsentIsNoChange(t *testing.T) {
	// PATCH の既存方式に従う: 未送信 = 変更なし。送られた3値だけが map に入る。
	version := 4
	body := `{
		"administration_instructions": "1日3回毎食後",
		"prescription_total_amount": 21,
		"version": 4
	}`
	var req updateTreatmentRequest
	require.NoError(t, json.Unmarshal([]byte(body), &req))

	input := req.toServiceInput()

	require.NotNil(t, input.AdministrationInstructions)
	assert.Equal(t, "1日3回毎食後", *input.AdministrationInstructions)
	require.NotNil(t, input.PrescriptionTotalAmount)
	assert.Equal(t, "21", exactText(t, input.PrescriptionTotalAmount))
	assert.Nil(t, input.AdministrationDoseAmount)
	assert.Nil(t, input.AdministrationDoseUnit)
	assert.Nil(t, input.PrescriptionTotalUnit)
	assert.Equal(t, &version, input.Version)
}

func TestBuildTreatmentUpdate_ThreeValues(t *testing.T) {
	amount := exactFromJSON(t, "1.25")

	fields := buildTreatmentUpdate(&UpdateTreatmentInput{
		AdministrationDoseAmount:   amount,
		AdministrationDoseUnit:     strPtr("mL"),
		AdministrationInstructions: strPtr("1日1回"),
	})

	assert.Equal(t, *amount, fields["administration_dose_amount"])
	assert.Equal(t, "mL", fields["administration_dose_unit"])
	assert.Equal(t, "1日1回", fields["administration_instructions"])
	// 未送信の列は map に入らない（nil = 変更なし）
	assert.NotContains(t, fields, "prescription_total_amount")
	assert.NotContains(t, fields, "prescription_total_unit")
	// mg 計算値列・会計数量と混同しない
	assert.NotContains(t, fields, "dose_amount_mg")
	assert.NotContains(t, fields, "quantity")
}

func TestTreatmentService_Create_ThreeValues_Persisted(t *testing.T) {
	const clinicID = uint64(1)
	var created *model.Treatment
	repo := &mockTreatmentRepository{
		createFn: func(_ context.Context, treatment *model.Treatment) error {
			created = treatment
			return nil
		},
	}
	svc := newTreatmentSvc(repo, draftMedicalRecordRepo(), &mockInventoryRepository{}, nil)

	amount := exactFromJSON(t, "2")
	total := exactFromJSON(t, "10")
	input := &CreateTreatmentInput{
		ItemType:                   model.TreatmentItemTypeMedicine,
		AdministrationDoseAmount:   amount,
		AdministrationDoseUnit:     strPtr("錠"),
		AdministrationInstructions: strPtr("1日2回"),
		PrescriptionTotalAmount:    total,
		PrescriptionTotalUnit:      strPtr("錠"),
	}
	treatment, err := svc.Create(context.Background(), clinicID, 1, input)
	require.NoError(t, err)
	require.NotNil(t, treatment)

	require.NotNil(t, created)
	assert.Equal(t, amount, created.AdministrationDoseAmount)
	require.NotNil(t, created.AdministrationDoseUnit)
	assert.Equal(t, "錠", *created.AdministrationDoseUnit)
	require.NotNil(t, created.AdministrationInstructions)
	assert.Equal(t, "1日2回", *created.AdministrationInstructions)
	assert.Equal(t, total, created.PrescriptionTotalAmount)
	require.NotNil(t, created.PrescriptionTotalUnit)
	assert.Equal(t, "錠", *created.PrescriptionTotalUnit)
	// 3値の入力だけでは mg 計算値・スナップショットは生成されない（混同しない）
	assert.Nil(t, created.DoseAmountMg)
	assert.Nil(t, created.DoseParamSnapshot)
}

func TestToTreatmentResponse_ThreeValues_NullVsZero(t *testing.T) {
	nilRow := &model.Treatment{ItemType: model.TreatmentItemTypeMedicine}
	resp := toTreatmentResponse(nilRow)
	raw, err := json.Marshal(resp)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(raw, &m))
	// NULL は省略されず null として出力される（未保存と 0 を区別する）
	assert.Contains(t, m, "administration_dose_amount")
	assert.Nil(t, m["administration_dose_amount"])
	assert.Contains(t, m, "prescription_total_amount")
	assert.Nil(t, m["prescription_total_amount"])

	zero := exactFromJSON(t, "0")
	zeroRow := &model.Treatment{ItemType: model.TreatmentItemTypeMedicine,
		AdministrationDoseAmount: zero}
	respZero := toTreatmentResponse(zeroRow)
	rawZero, err := json.Marshal(respZero)
	require.NoError(t, err)
	var mZero map[string]any
	require.NoError(t, json.Unmarshal(rawZero, &mZero))
	// 0 は数値として出力される（null に潰れない）
	assert.Equal(t, 0.0, mZero["administration_dose_amount"])
}

func TestThreeValues_ReEditRoundTrip(t *testing.T) {
	// 保存→応答→再編集: 応答された3値をそのまま PATCH で送り直しても同値が保たれる。
	row := &model.Treatment{ItemType: model.TreatmentItemTypeMedicine,
		AdministrationDoseAmount:   exactFromJSON(t, "2.5"),
		AdministrationDoseUnit:     strPtr("錠"),
		AdministrationInstructions: strPtr("1日2回朝夕食後に"),
		PrescriptionTotalAmount:    exactFromJSON(t, "14"),
		PrescriptionTotalUnit:      strPtr("錠"),
	}
	respRaw, err := json.Marshal(toTreatmentResponse(row))
	require.NoError(t, err)

	var src map[string]any
	require.NoError(t, json.Unmarshal(respRaw, &src))
	patch := map[string]any{
		"administration_dose_amount":  src["administration_dose_amount"],
		"administration_dose_unit":    src["administration_dose_unit"],
		"administration_instructions": src["administration_instructions"],
		"prescription_total_amount":   src["prescription_total_amount"],
		"prescription_total_unit":     src["prescription_total_unit"],
		"version":                     row.Version,
	}
	patchRaw, err := json.Marshal(patch)
	require.NoError(t, err)
	var req updateTreatmentRequest
	require.NoError(t, json.Unmarshal(patchRaw, &req))
	input := req.toServiceInput()

	assert.Equal(t, "2.5", exactText(t, input.AdministrationDoseAmount))
	assert.Equal(t, "錠", *input.AdministrationDoseUnit)
	assert.Equal(t, "1日2回朝夕食後に", *input.AdministrationInstructions)
	assert.Equal(t, "14", exactText(t, input.PrescriptionTotalAmount))
	assert.Equal(t, "錠", *input.PrescriptionTotalUnit)
}
