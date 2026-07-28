package medicalrecord

import "testing"

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
		BodySize:    &bodySize,
		BillingUnit: &billingUnit,
		SortOrder:   &sortOrder,
		TaxType:     &taxType,
		TaxRate:     &taxRate,
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
	if input.BodySize != &bodySize {
		t.Fatalf("BodySize pointer was not preserved")
	}
	if input.BillingUnit != &billingUnit {
		t.Fatalf("BillingUnit pointer was not preserved")
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
