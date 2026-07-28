package medicalrecord

import "testing"

func TestCreateCheckupTypeRequest_ToServiceInput(t *testing.T) {
	price := int64(4400)
	parentID := uint64(9)

	req := createCheckupTypeRequest{
		Name:        "Annual checkup",
		Price:       &price,
		IsActive:    true,
		Description: "Yearly wellness check",
		Interval:    "12 months",
		TargetAge:   "adult",
		ParentID:    &parentID,
		SortOrder:   6,
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
	if input.Interval != req.Interval {
		t.Fatalf("Interval = %q, want %q", input.Interval, req.Interval)
	}
	if input.TargetAge != req.TargetAge {
		t.Fatalf("TargetAge = %q, want %q", input.TargetAge, req.TargetAge)
	}
	if input.ParentID != &parentID {
		t.Fatalf("ParentID pointer was not preserved")
	}
	if input.SortOrder != req.SortOrder {
		t.Fatalf("SortOrder = %d, want %d", input.SortOrder, req.SortOrder)
	}
}

func TestUpdateCheckupTypeRequest_ToServiceInput(t *testing.T) {
	name := ""
	price := int64(0)
	isActive := false
	description := ""
	interval := ""
	targetAge := ""
	parentID := uint64(0)
	sortOrder := 0

	req := updateCheckupTypeRequest{
		Name:          &name,
		Price:         &price,
		IsActive:      &isActive,
		Description:   &description,
		Interval:      &interval,
		TargetAge:     &targetAge,
		ParentID:      &parentID,
		ClearParentID: true,
		SortOrder:     &sortOrder,
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
	if input.Interval != &interval {
		t.Fatalf("Interval pointer was not preserved")
	}
	if input.TargetAge != &targetAge {
		t.Fatalf("TargetAge pointer was not preserved")
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
}

func TestUpdateCheckupTypeRequest_ToServiceInput_NilFields(t *testing.T) {
	input := (&updateCheckupTypeRequest{}).toServiceInput()

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
	if input.Interval != nil {
		t.Fatalf("Interval = %v, want nil", input.Interval)
	}
	if input.TargetAge != nil {
		t.Fatalf("TargetAge = %v, want nil", input.TargetAge)
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
}
