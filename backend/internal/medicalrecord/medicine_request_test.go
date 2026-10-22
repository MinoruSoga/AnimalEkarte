package medicalrecord

import (
	"encoding/json"
	"testing"

	"github.com/animal-ekarte/backend/internal/model"
)

func decodeMedicineRequest[T any](t *testing.T, body string) T {
	t.Helper()
	var req T
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	return req
}

func createInputFrom(t *testing.T, body string) (*CreateMedicineInput, error) {
	t.Helper()
	req := decodeMedicineRequest[createMedicineRequest](t, body)
	return req.toServiceInput()
}

func updateInputFrom(t *testing.T, body string) (*UpdateMedicineInput, error) {
	t.Helper()
	req := decodeMedicineRequest[updateMedicineRequest](t, body)
	return req.toServiceInput()
}

func exactString(t *testing.T, n *model.ExactNumeric) string {
	t.Helper()
	if n == nil {
		t.Fatal("ExactNumeric = nil, want value")
	}
	return exactNumericCanonical(*n)
}

func TestCreateMedicineRequest_ToServiceInput(t *testing.T) {
	req := decodeMedicineRequest[createMedicineRequest](t, `{
		"name":"medicine","parent_id":10,"price":1200,"is_active":true,"description":"description",
		"dosage_form":"tablet","medicine_unit":"per_tablet","inventory_id":20,
		"default_quantity":"1.5","sort_order":2,"tax_type":"excluded","tax_rate":"0.10",
		"is_non_insurance":true,"strength":"12.5"}`)

	input, err := req.toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput: %v", err)
	}

	if input.Name != "medicine" || input.ParentID == nil || *input.ParentID != 10 {
		t.Fatalf("name/parent = (%q, %v)", input.Name, input.ParentID)
	}
	if input.Price == nil || *input.Price != 1200 || !input.IsActive || input.Description != "description" {
		t.Fatalf("price/active/description = (%v, %v, %q)", input.Price, input.IsActive, input.Description)
	}
	if input.DosageForm == nil || *input.DosageForm != "tablet" || input.MedicineUnit == nil || *input.MedicineUnit != "per_tablet" {
		t.Fatalf("dosage/unit = (%v, %v)", input.DosageForm, input.MedicineUnit)
	}
	if input.InventoryID == nil || *input.InventoryID != 20 || input.SortOrder != 2 {
		t.Fatalf("inventory/order = (%v, %d)", input.InventoryID, input.SortOrder)
	}
	if got := exactString(t, input.DefaultQuantity); got != "1.5" {
		t.Fatalf("DefaultQuantity = %q, want 1.5", got)
	}
	if input.TaxType == nil || *input.TaxType != "excluded" {
		t.Fatalf("TaxType = %v, want excluded", input.TaxType)
	}
	if got := exactString(t, input.TaxRate); got != "0.1" {
		t.Fatalf("TaxRate = %q, want 0.1", got)
	}
	if got := exactString(t, input.Strength); got != "12.5" {
		t.Fatalf("Strength = %q, want 12.5", got)
	}
	if !input.IsNonInsurance {
		t.Fatal("IsNonInsurance = false, want true")
	}
}

// EMR-447: POST の省略と null はどちらも既定（default_quantity=DB既定1、tax_rate=0.10）へ委ね、明示 "0" は 0 を保持する。
func TestCreateMedicineRequest_DecimalPresence(t *testing.T) {
	for _, body := range []string{`{"name":"m"}`, `{"name":"m","default_quantity":null,"tax_rate":null,"strength":null}`} {
		input, err := createInputFrom(t, body)
		if err != nil {
			t.Fatalf("%s: %v", body, err)
		}
		if input.DefaultQuantity != nil || input.TaxRate != nil || input.Strength != nil {
			t.Fatalf("%s: want nil decimals, got (%v, %v, %v)", body, input.DefaultQuantity, input.TaxRate, input.Strength)
		}
	}

	input, err := createInputFrom(t, `{"name":"m","default_quantity":"0","tax_rate":"0"}`)
	if err != nil {
		t.Fatalf("explicit zero: %v", err)
	}
	if exactString(t, input.DefaultQuantity) != "0" || exactString(t, input.TaxRate) != "0" {
		t.Fatalf("explicit zero not preserved: (%v, %v)", input.DefaultQuantity, input.TaxRate)
	}
}

func TestUpdateMedicineRequest_ToServiceInput(t *testing.T) {
	req := decodeMedicineRequest[updateMedicineRequest](t, `{
		"name":"","parent_id":10,"clear_parent_id":true,"price":0,"is_active":false,"description":"",
		"dosage_form":"","medicine_unit":"","inventory_id":20,"default_quantity":"0","sort_order":0,
		"tax_type":"exempt","tax_rate":"0","is_non_insurance":false,"strength":"0.0005","clear_strength":false}`)

	input, err := req.toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput: %v", err)
	}

	if input.Name == nil || *input.Name != "" || input.ParentID == nil || *input.ParentID != 10 || !input.ClearParentID {
		t.Fatalf("name/parent = (%v, %v, %v)", input.Name, input.ParentID, input.ClearParentID)
	}
	if input.Price == nil || *input.Price != 0 || input.IsActive == nil || *input.IsActive {
		t.Fatalf("price/active = (%v, %v)", input.Price, input.IsActive)
	}
	if input.Description == nil || *input.Description != "" || input.DosageForm == nil || *input.DosageForm != "" || input.MedicineUnit == nil || *input.MedicineUnit != "" {
		t.Fatalf("description/dosage/unit = (%v, %v, %v)", input.Description, input.DosageForm, input.MedicineUnit)
	}
	if exactString(t, input.DefaultQuantity) != "0" || exactString(t, input.TaxRate) != "0" {
		t.Fatalf("explicit zero not preserved: (%v, %v)", input.DefaultQuantity, input.TaxRate)
	}
	if got := exactString(t, input.Strength); got != "0.0005" {
		t.Fatalf("Strength = %q, want 0.0005", got)
	}
	if input.SortOrder == nil || *input.SortOrder != 0 || input.IsNonInsurance == nil || *input.IsNonInsurance {
		t.Fatalf("order/insurance = (%v, %v)", input.SortOrder, input.IsNonInsurance)
	}
}

// EMR-447: PATCH の省略と null は保持（nil）。null を strength の新しいクリア手段にしない（clear_strength を使う）。
func TestUpdateMedicineRequest_ToServiceInput_NilFields(t *testing.T) {
	for _, body := range []string{`{}`, `{"default_quantity":null,"tax_rate":null,"strength":null}`} {
		input, err := updateInputFrom(t, body)
		if err != nil {
			t.Fatalf("%s: %v", body, err)
		}
		if input.Name != nil || input.ParentID != nil || input.ClearParentID || input.Price != nil || input.IsActive != nil || input.TaxType != nil {
			t.Fatalf("%s: unexpected non-decimal fields set", body)
		}
		if input.DefaultQuantity != nil || input.TaxRate != nil || input.Strength != nil || input.ClearStrength {
			t.Fatalf("%s: want decimals untouched, got (%v, %v, %v, %v)", body, input.DefaultQuantity, input.TaxRate, input.Strength, input.ClearStrength)
		}
	}
}

// EMR-447: JSON number・符号・指数・保存尺度超過・範囲外は丸めず 400（invalid input）にする。
func TestMedicineRequest_RejectsNonExactDecimal(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"tax_rate JSON number", `{"tax_rate":0.1}`},
		{"default_quantity JSON number", `{"default_quantity":1}`},
		{"strength JSON number", `{"strength":12.5}`},
		{"negative", `{"default_quantity":"-1"}`},
		{"exponent", `{"strength":"1e2"}`},
		{"strength without integer part", `{"strength":".5"}`},
		{"strength leading zero", `{"strength":"05"}`},
		{"strength five fraction digits", `{"strength":"12.34567"}`},
		{"leading plus", `{"tax_rate":"+0.1"}`},
		{"empty string", `{"tax_rate":""}`},
		{"NaN", `{"strength":"NaN"}`},
		{"leading zero", `{"default_quantity":"01"}`},
		{"default_quantity scale over", `{"default_quantity":"1.234"}`},
		{"default_quantity integer over", `{"default_quantity":"123456789"}`},
		{"strength scale over", `{"strength":"0.00001"}`},
		{"strength integer over", `{"strength":"1234567"}`},
		{"strength zero", `{"strength":"0"}`},
		{"tax_rate over one", `{"tax_rate":"1.01"}`},
		{"tax_rate percent", `{"tax_rate":"10"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := updateInputFrom(t, tc.body); err == nil {
				t.Fatalf("update %s: want error", tc.body)
			}
			create := `{"name":"m",` + tc.body[1:]
			if _, err := createInputFrom(t, create); err == nil {
				t.Fatalf("create %s: want error", create)
			}
		})
	}
}

// EMR-447: 保存尺度内の末尾0は値として受理し、境界値 1 / numeric 最大桁を拒否しない。
func TestMedicineRequest_AcceptsBoundaryDecimals(t *testing.T) {
	input, err := updateInputFrom(t, `{"default_quantity":"99999999.990","tax_rate":"1.000","strength":"999999.9999"}`)
	if err != nil {
		t.Fatalf("boundary: %v", err)
	}
	if exactString(t, input.DefaultQuantity) != "99999999.99" || exactString(t, input.TaxRate) != "1" || exactString(t, input.Strength) != "999999.9999" {
		t.Fatalf("boundary values = (%v, %v, %v)", input.DefaultQuantity, input.TaxRate, input.Strength)
	}
}

// SPEC-QA-20261022-447-EXACT Q2: 末尾0だけが異なる同じ十進値は numeric(10,4) に収まれば受理する。
func TestMedicineRequest_StrengthTrailingZerosWithinScale(t *testing.T) {
	input, err := updateInputFrom(t, `{"strength":"1.23000"}`)
	if err != nil {
		t.Fatalf("1.23000: %v", err)
	}
	if got := exactString(t, input.Strength); got != "1.23" {
		t.Fatalf("Strength = %q, want 1.23", got)
	}
}

// exactNumPtr は test fixture 用に decimal 文字列から ExactNumeric を作る（不正な literal は test の誤り）。
func exactNumPtr(s string) *model.ExactNumeric {
	var n model.ExactNumeric
	if err := n.Scan(s); err != nil {
		panic("invalid decimal fixture: " + s)
	}
	return &n
}
