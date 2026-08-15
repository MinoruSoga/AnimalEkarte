package pet

import "testing"

func TestCreateAnimalSpeciesRequest_ToServiceInput(t *testing.T) {
	active := true
	req := createAnimalSpeciesRequest{
		Name:      "dog",
		IsActive:  &active,
		SortOrder: 2,
	}

	input := req.toServiceInput()

	if input.Name != req.Name {
		t.Fatalf("Name = %q, want %q", input.Name, req.Name)
	}
	if !input.IsActive {
		t.Fatalf("IsActive = %t, want true", input.IsActive)
	}
	if input.SortOrder != req.SortOrder {
		t.Fatalf("SortOrder = %d, want %d", input.SortOrder, req.SortOrder)
	}
}

func TestCreateAnimalSpeciesRequest_IsActiveOmitted(t *testing.T) {
	req := createAnimalSpeciesRequest{Name: "dog", SortOrder: 1}
	if req.IsActive != nil {
		t.Fatalf("omitted is_active must remain nil (presence absent), got %v", req.IsActive)
	}
	input := req.toServiceInput()
	if !input.IsActive {
		t.Fatalf("omitted is_active must resolve to true, got false")
	}
}

func TestCreateAnimalSpeciesRequest_IsActiveFalse(t *testing.T) {
	active := false
	req := createAnimalSpeciesRequest{Name: "dog", IsActive: &active}
	input := req.toServiceInput()
	if input.IsActive {
		t.Fatalf("explicit false must resolve to false")
	}
}

func TestUpdateAnimalSpeciesRequest_ToServiceInput(t *testing.T) {
	name := ""
	isActive := false
	sortOrder := 0
	req := updateAnimalSpeciesRequest{
		Name:      &name,
		IsActive:  &isActive,
		SortOrder: &sortOrder,
	}

	input := req.toServiceInput()

	if input.Name == nil || *input.Name != name {
		t.Fatalf("Name = %v, want explicit empty string", input.Name)
	}
	if input.IsActive == nil || *input.IsActive {
		t.Fatalf("IsActive = %v, want explicit false", input.IsActive)
	}
	if input.SortOrder == nil || *input.SortOrder != 0 {
		t.Fatalf("SortOrder = %v, want explicit zero", input.SortOrder)
	}
}

func TestUpdateAnimalSpeciesRequest_ToServiceInput_NilFields(t *testing.T) {
	req := updateAnimalSpeciesRequest{}

	input := req.toServiceInput()

	if input.Name != nil {
		t.Fatalf("Name = %v, want nil", input.Name)
	}
	if input.IsActive != nil {
		t.Fatalf("IsActive = %v, want nil", input.IsActive)
	}
	if input.SortOrder != nil {
		t.Fatalf("SortOrder = %v, want nil", input.SortOrder)
	}
}
