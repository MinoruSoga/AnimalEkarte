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
		Quantity:                   1, // Create は quantity > 0 を必須とする（成功経路 fixture として妥当化）
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

func TestValidateDosePairsForCreate(t *testing.T) {
	amount := exactFromJSON(t, "2.5")

	tests := []struct {
		name    string
		amount  *model.ExactNumeric
		unit    *string
		wantErr bool
	}{
		{name: "both nil is ok", amount: nil, unit: nil},
		{name: "valid pair is ok", amount: amount, unit: strPtr("錠")},
		{name: "amount only is rejected", amount: amount, unit: nil, wantErr: true},
		{name: "unit only is rejected", amount: nil, unit: strPtr("錠"), wantErr: true},
		{name: "blank unit is rejected", amount: amount, unit: strPtr("   "), wantErr: true},
		{name: "zero with unit is ok", amount: exactFromJSON(t, "0"), unit: strPtr("錠")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDosePairsForCreate(&CreateTreatmentInput{
				AdministrationDoseAmount: tt.amount,
				AdministrationDoseUnit:   tt.unit,
				PrescriptionTotalAmount:  nil,
				PrescriptionTotalUnit:    nil,
			})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	// 処方総量ペアも同一条件で検証される
	err := validateDosePairsForCreate(&CreateTreatmentInput{
		PrescriptionTotalAmount: amount,
	})
	assert.Error(t, err)
}

func TestValidateDosePairsForUpdate(t *testing.T) {
	amount := exactFromJSON(t, "2.5")

	// 未確定ペア（両NULL）の行への 量のみ 指定は 400（DB CHECK 違反を事前に拒否）
	err := validateDosePairsForUpdate(&model.Treatment{}, &UpdateTreatmentInput{
		AdministrationDoseAmount: amount,
	})
	assert.Error(t, err)

	// 未確定ペアの行への 量+単位 同時指定は OK
	err = validateDosePairsForUpdate(&model.Treatment{}, &UpdateTreatmentInput{
		AdministrationDoseAmount: amount,
		AdministrationDoseUnit:   strPtr("錠"),
	})
	assert.NoError(t, err)

	// 有効ペアの行への 量のみ 変更は OK（単位は維持されペアが壊れない）
	current := &model.Treatment{
		AdministrationDoseAmount: amount,
		AdministrationDoseUnit:   strPtr("錠"),
	}
	err = validateDosePairsForUpdate(current, &UpdateTreatmentInput{
		AdministrationDoseAmount: exactFromJSON(t, "3"),
	})
	assert.NoError(t, err)

	// 有効ペアの行への 空白単位 への変更は 400
	err = validateDosePairsForUpdate(current, &UpdateTreatmentInput{
		AdministrationDoseUnit: strPtr("  "),
	})
	assert.Error(t, err)
}

func TestBuildTreatmentUpdate_ThreeValues_DoNotTouchExistingDerivedColumns(t *testing.T) {
	// P2-F2（サイクル1・レビューB）: 既存行が dose_param_snapshot / mg 値 / quantity を持っていても、
	// 3値だけを送った PATCH の writer map は派生列・会計列を一切含まない（旧snapshot 保護）。
	amount := exactFromJSON(t, "1.5")
	fields := buildTreatmentUpdate(&UpdateTreatmentInput{
		AdministrationDoseAmount:   amount,
		AdministrationDoseUnit:     strPtr("錠"),
		AdministrationInstructions: strPtr("1日2回"),
		PrescriptionTotalAmount:    exactFromJSON(t, "7"),
		PrescriptionTotalUnit:      strPtr("錠"),
	})
	for _, forbidden := range []string{
		"dose_amount_mg", "dose_amount_unit", "dose_param_snapshot", "dose_weight_kg", "dose_weight_source",
		"quantity", "unit_price", "discount_rate", "discount_amount", "status", "item_type",
	} {
		assert.NotContains(t, fields, forbidden)
	}
	assert.Len(t, fields, 5)
}

func TestCreateTreatmentRequest_ThreeValues_ExplicitNullAndNaNAreRejected(t *testing.T) {
	// 明示 null は 400（省略=NULL 保存と区別する。checkupTypeInsuranceField と同一の家契約）
	body := `{"item_type": "medicine", "administration_dose_amount": null}`
	var req createTreatmentRequest
	err := json.Unmarshal([]byte(body), &req)
	assert.Error(t, err)

	// NaN 文字列は 400（PostgreSQL numeric は NaN を格納できてしまうため wire で拒否）
	body = `{"item_type": "medicine", "administration_dose_amount": "NaN"}`
	err = json.Unmarshal([]byte(body), &req)
	assert.Error(t, err)

	// Infinity / -Infinity も 400（pgtype は InfinityModifier 付きで受理するため wire で拒否。C5）
	for _, inf := range []string{"Infinity", "-Infinity"} {
		body = `{"item_type": "medicine", "administration_dose_amount": "` + inf + `"}`
		err = json.Unmarshal([]byte(body), &req)
		assert.Error(t, err)
	}

	body = `{"item_type": "medicine", "administration_instructions": null}`
	err = json.Unmarshal([]byte(body), &req)
	assert.Error(t, err)
}

func TestValidateDoseAmountUnitPair_RejectsNegativeAmount(t *testing.T) {
	negative := exactFromJSON(t, "-1")
	blank := "  "
	err := validateDoseAmountUnitPair(negative, &blank)
	assert.Error(t, err)

	validUnit := "錠"
	err = validateDoseAmountUnitPair(negative, &validUnit)
	assert.Error(t, err)
}

func TestValidateDoseFieldsMedicineOnlyForCreate(t *testing.T) {
	amount := exactFromJSON(t, "2.5")
	unit := "錠"

	// 非medicine 明細への3値指定は 400
	err := validateDoseFieldsMedicineOnlyForCreate(model.TreatmentItemTypeConsultation, &CreateTreatmentInput{
		ItemType:                 model.TreatmentItemTypeConsultation,
		AdministrationDoseAmount: amount,
		AdministrationDoseUnit:   &unit,
	})
	assert.Error(t, err)

	// medicine 明細は OK
	err = validateDoseFieldsMedicineOnlyForCreate(model.TreatmentItemTypeMedicine, &CreateTreatmentInput{
		ItemType:                 model.TreatmentItemTypeMedicine,
		AdministrationDoseAmount: amount,
		AdministrationDoseUnit:   &unit,
	})
	assert.NoError(t, err)

	// 3値を指定しない非medicine 明細は OK
	err = validateDoseFieldsMedicineOnlyForCreate(model.TreatmentItemTypeConsultation, &CreateTreatmentInput{
		ItemType: model.TreatmentItemTypeConsultation,
	})
	assert.NoError(t, err)
}

func TestValidateDoseFieldsMedicineOnlyForUpdate(t *testing.T) {
	amount := exactFromJSON(t, "2.5")

	// 適用後 item_type が非medicine なのに3値入力がある更新は 400
	err := validateDoseFieldsMedicineOnlyForUpdate(model.TreatmentItemTypeConsultation, &UpdateTreatmentInput{
		AdministrationDoseAmount: amount,
	})
	assert.Error(t, err)

	// medicine 行への3値入力は OK
	err = validateDoseFieldsMedicineOnlyForUpdate(model.TreatmentItemTypeMedicine, &UpdateTreatmentInput{
		AdministrationDoseAmount: amount,
	})
	assert.NoError(t, err)
}

func TestValidateDoseAmountUnitPair_Numeric12_6Range(t *testing.T) {
	unit := "錠"

	// numeric(12,6) の上限 999999.999999 は OK
	err := validateDoseAmountUnitPair(exactFromJSON(t, "999999.999999"), &unit)
	assert.NoError(t, err)

	// 上限超過（1000000）は 400（DB 22003→500 を事前に拒否）
	err = validateDoseAmountUnitPair(exactFromJSON(t, "1000000"), &unit)
	assert.Error(t, err)

	// 小数点以下7桁は 400
	err = validateDoseAmountUnitPair(exactFromJSON(t, "0.0000001"), &unit)
	assert.Error(t, err)
}

func TestValidateDoseFieldsMedicineOnlyForUpdate_RejectsTypeChangeHidingStoredValues(t *testing.T) {
	// Codex review P2: 3値が保存済みの medicine 行を item_type だけ非medicine へ替える更新は 400
	// （保存済み臨床値が FE で非表示になる退避を防ぐ）。
	current := &model.Treatment{
		ItemType:                 model.TreatmentItemTypeMedicine,
		AdministrationDoseAmount: exactFromJSON(t, "2.5"),
		AdministrationDoseUnit:   strPtr("錠"),
	}
	otherType := model.TreatmentItemTypeOther
	err := validateDoseFieldsMedicineOnlyForUpdate(model.TreatmentItemTypeOther, current, &UpdateTreatmentInput{
		ItemType: &otherType,
	})
	assert.Error(t, err)

	// medicine 行のままの 3値なし更新は OK
	err = validateDoseFieldsMedicineOnlyForUpdate(model.TreatmentItemTypeMedicine, current, &UpdateTreatmentInput{})
	assert.NoError(t, err)
}

func TestCreateTreatmentRequest_ThreeValues_ExponentFormIsAccepted(t *testing.T) {
	// Codex review P2: Python クライアント等が送る指数表記（1e-06）も妥当な numeric として受理する。
	body := `{"item_type": "medicine", "administration_dose_amount": 1e-06, "administration_dose_unit": "mL"}`
	var req createTreatmentRequest
	require.NoError(t, json.Unmarshal([]byte(body), &req))

	input := req.toServiceInput()
	require.NotNil(t, input.AdministrationDoseAmount)
	assert.Equal(t, "0.000001", exactText(t, input.AdministrationDoseAmount))
}
