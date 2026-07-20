package medicalrecord

import "testing"

func TestCreateExaminationTypeRequest_ToServiceInput(t *testing.T) {
	price := int64(3300)
	parentID := uint64(12)

	req := createExaminationTypeRequest{
		Name:           "Blood test",
		Price:          &price,
		IsActive:       true,
		Description:    "CBC and chemistry",
		ParentID:       &parentID,
		SortOrder:      4,
		IsNonInsurance: true,
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
	if input.ParentID != &parentID {
		t.Fatalf("ParentID pointer was not preserved")
	}
	if input.SortOrder != req.SortOrder {
		t.Fatalf("SortOrder = %d, want %d", input.SortOrder, req.SortOrder)
	}
	if input.IsNonInsurance != req.IsNonInsurance {
		t.Fatalf("IsNonInsurance = %v, want %v", input.IsNonInsurance, req.IsNonInsurance)
	}
}

func TestUpdateExaminationTypeRequest_ToServiceInput(t *testing.T) {
	name := ""
	price := int64(0)
	isActive := false
	description := ""
	parentID := uint64(0)
	sortOrder := 0
	isNonInsurance := false

	req := updateExaminationTypeRequest{
		Name:           &name,
		Price:          &price,
		IsActive:       &isActive,
		Description:    &description,
		ParentID:       &parentID,
		ClearParentID:  true,
		SortOrder:      &sortOrder,
		IsNonInsurance: &isNonInsurance,
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
	if input.ParentID != &parentID {
		t.Fatalf("ParentID pointer was not preserved")
	}
	if !input.ClearParentID {
		t.Fatalf("ClearParentID = false, want true")
	}
	if input.SortOrder != &sortOrder {
		t.Fatalf("SortOrder pointer was not preserved")
	}
	if input.IsNonInsurance != &isNonInsurance {
		t.Fatalf("IsNonInsurance pointer was not preserved")
	}
}

func TestUpdateExaminationTypeRequest_ToServiceInput_NilFields(t *testing.T) {
	input := (&updateExaminationTypeRequest{}).toServiceInput()

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
	if input.ParentID != nil {
		t.Fatalf("ParentID = %v, want nil", input.ParentID)
	}
	if input.ClearParentID {
		t.Fatalf("ClearParentID = true, want false")
	}
	if input.SortOrder != nil {
		t.Fatalf("SortOrder = %v, want nil", input.SortOrder)
	}
	if input.IsNonInsurance != nil {
		t.Fatalf("IsNonInsurance = %v, want nil", input.IsNonInsurance)
	}
}
