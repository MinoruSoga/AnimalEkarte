package medicalrecord

import "testing"

func TestCreateVaccineRequest_ToServiceInput(t *testing.T) {
	price := int64(1200)
	parentID := uint64(10)
	req := createVaccineRequest{
		Name:        "vaccine",
		Price:       &price,
		IsActive:    true,
		Description: "description",
		Species:     "dog",
		Interval:    "1year",
		ParentID:    &parentID,
		SortOrder:   2,
	}

	input := req.toServiceInput()

	if input.Name != req.Name {
		t.Fatalf("Name = %q, want %q", input.Name, req.Name)
	}
	if input.Price == nil || *input.Price != price {
		t.Fatalf("Price = %v, want %d", input.Price, price)
	}
	if !input.IsActive {
		t.Fatal("IsActive = false, want true")
	}
	if input.Description != req.Description {
		t.Fatalf("Description = %q, want %q", input.Description, req.Description)
	}
	if input.Species == nil || *input.Species != req.Species {
		t.Fatalf("Species = %v, want %q", input.Species, req.Species)
	}
	if input.Interval != req.Interval {
		t.Fatalf("Interval = %q, want %q", input.Interval, req.Interval)
	}
	if input.ParentID == nil || *input.ParentID != parentID {
		t.Fatalf("ParentID = %v, want %d", input.ParentID, parentID)
	}
	if input.SortOrder != req.SortOrder {
		t.Fatalf("SortOrder = %d, want %d", input.SortOrder, req.SortOrder)
	}
}

func TestCreateVaccineRequest_ToServiceInput_EmptySpecies(t *testing.T) {
	req := createVaccineRequest{Name: "vaccine"}

	input := req.toServiceInput()

	if input.Species != nil {
		t.Fatalf("Species = %v, want nil", input.Species)
	}
}

func TestUpdateVaccineRequest_ToServiceInput(t *testing.T) {
	name := ""
	price := int64(0)
	isActive := false
	description := ""
	species := "cat"
	interval := ""
	parentID := uint64(0)
	sortOrder := 0
	req := updateVaccineRequest{
		Name:          &name,
		Price:         &price,
		IsActive:      &isActive,
		Description:   &description,
		Species:       &species,
		Interval:      &interval,
		ParentID:      &parentID,
		ClearParentID: true,
		SortOrder:     &sortOrder,
	}

	input := req.toServiceInput()

	if input.Name == nil || *input.Name != name {
		t.Fatalf("Name = %v, want explicit empty string", input.Name)
	}
	if input.Price == nil || *input.Price != 0 {
		t.Fatalf("Price = %v, want explicit zero", input.Price)
	}
	if input.IsActive == nil || *input.IsActive {
		t.Fatalf("IsActive = %v, want explicit false", input.IsActive)
	}
	if input.Description == nil || *input.Description != description {
		t.Fatalf("Description = %v, want explicit empty string", input.Description)
	}
	if input.Species == nil || *input.Species != species {
		t.Fatalf("Species = %v, want %q", input.Species, species)
	}
	if input.Interval == nil || *input.Interval != interval {
		t.Fatalf("Interval = %v, want explicit empty string", input.Interval)
	}
	if input.ParentID == nil || *input.ParentID != 0 {
		t.Fatalf("ParentID = %v, want explicit zero", input.ParentID)
	}
	if !input.ClearParentID {
		t.Fatal("ClearParentID = false, want true")
	}
	if input.SortOrder == nil || *input.SortOrder != 0 {
		t.Fatalf("SortOrder = %v, want explicit zero", input.SortOrder)
	}
}

func TestUpdateVaccineRequest_ToServiceInput_NilFields(t *testing.T) {
	req := updateVaccineRequest{}

	input := req.toServiceInput()

	if input.Name != nil {
		t.Fatalf("Name = %v, want nil", input.Name)
	}
	if input.Price != nil {
		t.Fatalf("Price = %v, want nil", input.Price)
	}
	if input.IsActive != nil {
		t.Fatalf("IsActive = %v, want nil", input.IsActive)
	}
	if input.Species != nil {
		t.Fatalf("Species = %v, want nil", input.Species)
	}
	if input.ClearParentID {
		t.Fatal("ClearParentID = true, want false")
	}
}
