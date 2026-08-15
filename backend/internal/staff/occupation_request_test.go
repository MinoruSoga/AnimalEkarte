package staff

import "testing"

func TestCreateOccupationRequest_ToServiceInput(t *testing.T) {
	active := true
	req := createOccupationRequest{
		Name:        "Veterinarian",
		Description: "Medical staff",
		IsActive:    &active,
		SortOrder:   2,
	}

	input := req.toServiceInput()

	if input.Name != req.Name {
		t.Fatalf("Name = %q, want %q", input.Name, req.Name)
	}
	if input.Description != req.Description {
		t.Fatalf("Description = %q, want %q", input.Description, req.Description)
	}
	if !input.IsActive {
		t.Fatalf("IsActive = false, want true")
	}
	if input.SortOrder != req.SortOrder {
		t.Fatalf("SortOrder = %d, want %d", input.SortOrder, req.SortOrder)
	}
}

func TestCreateOccupationRequest_IsActiveFalse(t *testing.T) {
	active := false
	req := createOccupationRequest{Name: "x", IsActive: &active}
	if req.toServiceInput().IsActive {
		t.Fatal("explicit false must resolve to false")
	}
}

func TestCreateOccupationRequest_IsActiveOmitted(t *testing.T) {
	req := createOccupationRequest{Name: "x"}
	if !req.toServiceInput().IsActive {
		t.Fatal("omitted is_active must resolve to true")
	}
}

func TestUpdateOccupationRequest_ToServiceInput(t *testing.T) {
	name := ""
	description := ""
	isActive := false
	sortOrder := 0

	req := updateOccupationRequest{
		Name:        &name,
		Description: &description,
		IsActive:    &isActive,
		SortOrder:   &sortOrder,
	}

	input := req.toServiceInput()

	if input.Name != &name {
		t.Fatalf("Name pointer was not preserved")
	}
	if input.Description != &description {
		t.Fatalf("Description pointer was not preserved")
	}
	if input.IsActive != &isActive {
		t.Fatalf("IsActive pointer was not preserved")
	}
	if input.SortOrder != &sortOrder {
		t.Fatalf("SortOrder pointer was not preserved")
	}
}

func TestUpdateOccupationRequest_ToServiceInput_NilFields(t *testing.T) {
	input := (&updateOccupationRequest{}).toServiceInput()

	if input.Name != nil {
		t.Fatalf("Name = %v, want nil", input.Name)
	}
	if input.Description != nil {
		t.Fatalf("Description = %v, want nil", input.Description)
	}
	if input.IsActive != nil {
		t.Fatalf("IsActive = %v, want nil", input.IsActive)
	}
	if input.SortOrder != nil {
		t.Fatalf("SortOrder = %v, want nil", input.SortOrder)
	}
}
