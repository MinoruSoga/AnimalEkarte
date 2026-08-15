package medicalrecord

import "testing"

func TestCreateChiefComplaintRequest_ToServiceInput(t *testing.T) {
	active := true
	req := createChiefComplaintRequest{
		Name:        "Cough",
		Description: "Respiratory symptom",
		IsActive:    &active,
		SortOrder:   5,
	}

	input := req.toServiceInput()

	if input.Name != req.Name {
		t.Fatalf("Name = %q, want %q", input.Name, req.Name)
	}
	if input.Description != req.Description {
		t.Fatalf("Description = %q, want %q", input.Description, req.Description)
	}
	if !input.IsActive {
		t.Fatalf("IsActive = %v, want true", input.IsActive)
	}
	if input.SortOrder != req.SortOrder {
		t.Fatalf("SortOrder = %d, want %d", input.SortOrder, req.SortOrder)
	}
}

func TestCreateChiefComplaintRequest_IsActiveOmitted(t *testing.T) {
	req := createChiefComplaintRequest{Name: "Cough", Description: "desc", SortOrder: 1}
	if req.IsActive != nil {
		t.Fatalf("omitted is_active must remain nil (presence absent), got %v", req.IsActive)
	}
	input := req.toServiceInput()
	if !input.IsActive {
		t.Fatalf("omitted is_active must resolve to true, got false")
	}
}

func TestCreateChiefComplaintRequest_IsActiveFalse(t *testing.T) {
	active := false
	req := createChiefComplaintRequest{Name: "Cough", IsActive: &active}
	input := req.toServiceInput()
	if input.IsActive {
		t.Fatalf("explicit false must resolve to false")
	}
}

func TestUpdateChiefComplaintRequest_ToServiceInput(t *testing.T) {
	name := ""
	description := ""
	isActive := false
	sortOrder := 0

	req := updateChiefComplaintRequest{
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

func TestUpdateChiefComplaintRequest_ToServiceInput_NilFields(t *testing.T) {
	input := (&updateChiefComplaintRequest{}).toServiceInput()

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
