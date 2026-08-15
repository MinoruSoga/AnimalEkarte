package medicalrecord

import "testing"

func TestCreateCageRequest_ToServiceInput(t *testing.T) {
	price := int64(2200)

	req := createCageRequest{
		Name:        "ICU 1",
		CageType:    "icu",
		CageSize:    "large",
		Price:       &price,
		IsActive:    true,
		Description: "Intensive care cage",
		SortOrder:   2,
	}

	input := req.toServiceInput()

	if input.Name != req.Name {
		t.Fatalf("Name = %q, want %q", input.Name, req.Name)
	}
	if input.CageType != req.CageType {
		t.Fatalf("CageType = %q, want %q", input.CageType, req.CageType)
	}
	if input.CageSize != req.CageSize {
		t.Fatalf("CageSize = %q, want %q", input.CageSize, req.CageSize)
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
	if input.SortOrder != req.SortOrder {
		t.Fatalf("SortOrder = %d, want %d", input.SortOrder, req.SortOrder)
	}
}

func TestUpdateCageRequest_ToServiceInput(t *testing.T) {
	name := ""
	cageType := "general"
	cageSize := "small"
	price := int64(0)
	isActive := false
	description := ""
	sortOrder := 0

	req := updateCageRequest{
		Name:        &name,
		CageType:    &cageType,
		CageSize:    &cageSize,
		Price:       &price,
		IsActive:    &isActive,
		Description: &description,
		SortOrder:   &sortOrder,
	}

	input := req.toServiceInput()

	if input.Name != &name {
		t.Fatalf("Name pointer was not preserved")
	}
	if input.CageType != &cageType {
		t.Fatalf("CageType pointer was not preserved")
	}
	if input.CageSize != &cageSize {
		t.Fatalf("CageSize pointer was not preserved")
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
	if input.SortOrder != &sortOrder {
		t.Fatalf("SortOrder pointer was not preserved")
	}
}

func TestUpdateCageRequest_ToServiceInput_NilFields(t *testing.T) {
	input := (&updateCageRequest{}).toServiceInput()

	if input.Name != nil {
		t.Fatalf("Name = %v, want nil", input.Name)
	}
	if input.CageType != nil {
		t.Fatalf("CageType = %v, want nil", input.CageType)
	}
	if input.CageSize != nil {
		t.Fatalf("CageSize = %v, want nil", input.CageSize)
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
	if input.SortOrder != nil {
		t.Fatalf("SortOrder = %v, want nil", input.SortOrder)
	}
}
