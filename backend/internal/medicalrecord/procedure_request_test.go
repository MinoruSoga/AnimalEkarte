package medicalrecord

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestCreateProcedureRequest_ToServiceInput(t *testing.T) {
	price := int64(5500)
	duration := 30
	parentID := uint64(7)
	taxRate := 0.1

	req := createProcedureRequest{
		Name:        "Dental scaling",
		Price:       &price,
		IsActive:    true,
		Description: "Routine dental procedure",
		Duration:    &duration,
		Anesthesia:  "general",
		ParentID:    &parentID,
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
	if input.Duration != &duration {
		t.Fatalf("Duration pointer was not preserved")
	}
	if input.Anesthesia != req.Anesthesia {
		t.Fatalf("Anesthesia = %q, want %q", input.Anesthesia, req.Anesthesia)
	}
	if input.ParentID != &parentID {
		t.Fatalf("ParentID pointer was not preserved")
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

func TestUpdateProcedureRequest_ToServiceInput(t *testing.T) {
	name := ""
	price := int64(0)
	isActive := false
	description := ""
	duration := 0
	anesthesia := "none"
	parentID := uint64(0)
	sortOrder := 0
	taxType := "exempt"
	taxRate := 0.0

	req := updateProcedureRequest{
		Name:          &name,
		Price:         &price,
		IsActive:      &isActive,
		Description:   &description,
		Duration:      &duration,
		Anesthesia:    &anesthesia,
		ParentID:      &parentID,
		ClearParentID: true,
		SortOrder:     &sortOrder,
		TaxType:       &taxType,
		TaxRate:       &taxRate,
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
	if input.Duration != &duration {
		t.Fatalf("Duration pointer was not preserved")
	}
	if input.Anesthesia != &anesthesia {
		t.Fatalf("Anesthesia pointer was not preserved")
	}
	if input.ParentID != &parentID {
		t.Fatalf("ParentID pointer was not preserved")
	}
	if !input.ClearParentID {
		t.Fatalf("ClearParentID = false, want true")
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

func TestUpdateProcedureRequest_ToServiceInput_NilFields(t *testing.T) {
	input := (&updateProcedureRequest{}).toServiceInput()

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
	if input.Duration != nil {
		t.Fatalf("Duration = %v, want nil", input.Duration)
	}
	if input.Anesthesia != nil {
		t.Fatalf("Anesthesia = %v, want nil", input.Anesthesia)
	}
	if input.ParentID != nil {
		t.Fatalf("ParentID = %v, want nil", input.ParentID)
	}
	if input.ClearParentID {
		t.Fatalf("ClearParentID = true, want false")
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
	if input.IsNonInsurance != nil {
		t.Fatalf("IsNonInsurance = %v, want nil (未指定=変更なし)", *input.IsNonInsurance)
	}
}

// EMR-454: procedure も consultation と同じ wire 契約（create 省略=nil、update 未指定=nil、
// 明示 null=400、true/false=明示更新）を守る。
func TestProcedureRequest_IsNonInsuranceWireContract(t *testing.T) {
	t.Run("create omits is_non_insurance", func(t *testing.T) {
		var req createProcedureRequest
		body := `{"name":"処置","anesthesia":"none","tax_type":"excluded"}`
		if err := json.Unmarshal([]byte(body), &req); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := req.IsNonInsurance.toServiceInput(); got != nil {
			t.Fatalf("IsNonInsurance = %v, want nil", *got)
		}
	})

	t.Run("create explicit null is rejected with 400", func(t *testing.T) {
		var req createProcedureRequest
		body := `{"name":"処置","is_non_insurance":null}`
		if err := json.Unmarshal([]byte(body), &req); err == nil {
			t.Fatal("explicit null must be rejected, got nil error")
		}
	})

	for _, value := range []bool{true, false} {
		t.Run(fmt.Sprintf("create explicit %v", value), func(t *testing.T) {
			var req createProcedureRequest
			body := fmt.Sprintf(`{"name":"処置","is_non_insurance":%v}`, value)
			if err := json.Unmarshal([]byte(body), &req); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := req.IsNonInsurance.toServiceInput()
			if got == nil || *got != value {
				t.Fatalf("IsNonInsurance = %v, want %v", got, value)
			}
		})
	}

	t.Run("update omits is_non_insurance", func(t *testing.T) {
		var req updateProcedureRequest
		if err := json.Unmarshal([]byte(`{"name":"処置"}`), &req); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := req.toServiceInput().IsNonInsurance; got != nil {
			t.Fatalf("IsNonInsurance = %v, want nil", *got)
		}
	})

	for _, value := range []bool{true, false} {
		t.Run(fmt.Sprintf("update explicit %v", value), func(t *testing.T) {
			var req updateProcedureRequest
			body := fmt.Sprintf(`{"is_non_insurance":%v}`, value)
			if err := json.Unmarshal([]byte(body), &req); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := req.toServiceInput().IsNonInsurance
			if got == nil || *got != value {
				t.Fatalf("IsNonInsurance = %v, want %v", got, value)
			}
		})
	}

	t.Run("update explicit null is rejected", func(t *testing.T) {
		var req updateProcedureRequest
		if err := json.Unmarshal([]byte(`{"is_non_insurance":null}`), &req); err == nil {
			t.Fatal("explicit null must be rejected, got nil error")
		}
	})
}
