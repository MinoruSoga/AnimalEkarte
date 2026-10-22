package medicalrecord

import (
	"encoding/json"
	"testing"
)

func TestCreateHospitalizationPlanRequest_ToServiceInput(t *testing.T) {
	price := int64(8800)
	taxRate := 0.1

	req := createHospitalizationPlanRequest{
		Name:        "Standard stay",
		Price:       &price,
		IsActive:    true,
		Description: "Standard hospitalization plan",
		BodySize:    "medium",
		BillingUnit: "per_day",
		SortOrder:   3,
		TaxType:     "included",
		TaxRate:     &taxRate,
	}

	input := req.toServiceInput()

	if input.Name != req.Name {
		t.Fatalf("Name = %q, want %q", input.Name, req.Name)
	}
	if input.Price != &price {
		t.Fatalf("Price pointer was not preserved")
	}
	if input.IsActive != req.IsActive {
		t.Fatalf("IsActive = %v, want %v", input.IsActive, req.IsActive)
	}
	if input.Description != req.Description {
		t.Fatalf("Description = %q, want %q", input.Description, req.Description)
	}
	if input.BodySize != req.BodySize {
		t.Fatalf("BodySize = %q, want %q", input.BodySize, req.BodySize)
	}
	if input.BillingUnit != req.BillingUnit {
		t.Fatalf("BillingUnit = %q, want %q", input.BillingUnit, req.BillingUnit)
	}
	if input.SortOrder != req.SortOrder {
		t.Fatalf("SortOrder = %d, want %d", input.SortOrder, req.SortOrder)
	}
	if input.TaxType != req.TaxType {
		t.Fatalf("TaxType = %q, want %q", input.TaxType, req.TaxType)
	}
	if input.TaxRate != &taxRate {
		t.Fatalf("TaxRate pointer was not preserved")
	}
}

func TestUpdateHospitalizationPlanRequest_ToServiceInput(t *testing.T) {
	name := ""
	price := int64(0)
	isActive := false
	description := ""
	bodySize := "small"
	billingUnit := "per_night"
	sortOrder := 0
	taxType := "exempt"
	taxRate := 0.0

	req := updateHospitalizationPlanRequest{
		Name:        &name,
		Price:       &price,
		IsActive:    &isActive,
		Description: &description,
		SortOrder:   &sortOrder,
		TaxType:     &taxType,
		TaxRate:     &taxRate,
		BodySize:    hospitalizationPlanNullableEnumField{set: true, value: &bodySize},
		BillingUnit: hospitalizationPlanNullableEnumField{set: true, value: &billingUnit},
	}

	input := req.toServiceInput()

	if input.Name != &name {
		t.Fatalf("Name pointer was not preserved")
	}
	if input.Price != &price {
		t.Fatalf("Price pointer was not preserved")
	}
	if input.IsActive != &isActive {
		t.Fatalf("IsActive pointer was not preserved")
	}
	if input.Description != &description {
		t.Fatalf("Description pointer was not preserved")
	}
	if input.BodySize == nil || *input.BodySize != &bodySize {
		t.Fatalf("BodySize value pointer was not preserved")
	}
	if input.BillingUnit == nil || *input.BillingUnit != &billingUnit {
		t.Fatalf("BillingUnit value pointer was not preserved")
	}
	if input.SortOrder != &sortOrder {
		t.Fatalf("SortOrder pointer was not preserved")
	}
	if input.TaxType != &taxType {
		t.Fatalf("TaxType pointer was not preserved")
	}
	if input.TaxRate != &taxRate {
		t.Fatalf("TaxRate pointer was not preserved")
	}
}

func TestUpdateHospitalizationPlanRequest_ToServiceInput_NilFields(t *testing.T) {
	input := (&updateHospitalizationPlanRequest{}).toServiceInput()

	if input.Name != nil {
		t.Fatalf("Name = %v, want nil", input.Name)
	}
	if input.Price != nil {
		t.Fatalf("Price = %v, want nil", input.Price)
	}
	if input.IsActive != nil {
		t.Fatalf("IsActive = %v, want nil", input.IsActive)
	}
	if input.Description != nil {
		t.Fatalf("Description = %v, want nil", input.Description)
	}
	if input.BodySize != nil {
		t.Fatalf("BodySize = %v, want nil", input.BodySize)
	}
	if input.BillingUnit != nil {
		t.Fatalf("BillingUnit = %v, want nil", input.BillingUnit)
	}
	if input.SortOrder != nil {
		t.Fatalf("SortOrder = %v, want nil", input.SortOrder)
	}
	if input.TaxType != nil {
		t.Fatalf("TaxType = %v, want nil", input.TaxType)
	}
	if input.TaxRate != nil {
		t.Fatalf("TaxRate = %v, want nil", input.TaxRate)
	}
}

// EMR-528: PATCH body_size / billing_unit の wire 3値（省略=保持 / null=SQL NULL / 値=設定）を decode で区別する。
func TestUpdateHospitalizationPlanRequest_NullableEnumPresence(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantSet   bool
		wantNull  bool
		wantValue string
	}{
		{name: "omitted keeps existing", body: `{"name":"x"}`, wantSet: false},
		{name: "explicit null clears", body: `{"body_size":null,"billing_unit":null}`, wantSet: true, wantNull: true},
		{name: "value sets", body: `{"body_size":"large","billing_unit":"per_night"}`, wantSet: true, wantValue: "set"},
		{name: "empty string is a value, not null", body: `{"body_size":"","billing_unit":""}`, wantSet: true, wantValue: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req updateHospitalizationPlanRequest
			if err := json.Unmarshal([]byte(tt.body), &req); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			input := req.toServiceInput()
			for field, got := range map[string]**string{"body_size": input.BodySize, "billing_unit": input.BillingUnit} {
				if !tt.wantSet {
					if got != nil {
						t.Fatalf("%s = %v, want omitted (nil)", field, got)
					}
					continue
				}
				if got == nil {
					t.Fatalf("%s omitted, want present", field)
				}
				if tt.wantNull {
					if *got != nil {
						t.Fatalf("%s = %q, want explicit null", field, **got)
					}
					continue
				}
				if *got == nil {
					t.Fatalf("%s = null, want a string value", field)
				}
				if tt.wantValue == "" && **got != "" {
					t.Fatalf("%s = %q, want empty string preserved for service rejection", field, **got)
				}
			}
		})
	}
}

func TestUpdateHospitalizationPlanRequest_NullableEnumRejectsNonString(t *testing.T) {
	for _, body := range []string{`{"body_size":1}`, `{"billing_unit":true}`, `{"body_size":{}}`} {
		var req updateHospitalizationPlanRequest
		if err := json.Unmarshal([]byte(body), &req); err == nil {
			t.Fatalf("Unmarshal(%s) error = nil, want invalid input", body)
		}
	}
}
