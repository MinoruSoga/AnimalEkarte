package trimming

import "testing"

func TestCreateTrimmingOptionRequest_ToServiceInput(t *testing.T) {
	price := int64(1200)
	duration := 15
	active := true
	combinable := true

	req := createTrimmingOptionRequest{
		Name:         "Nail trim",
		Price:        &price,
		IsActive:     &active,
		Description:  "Basic nail trimming",
		Duration:     &duration,
		IsCombinable: &combinable,
		SortOrder:    3,
	}

	input := req.toServiceInput()

	if input.Name != req.Name {
		t.Fatalf("Name = %q, want %q", input.Name, req.Name)
	}
	if input.Price != &price {
		t.Fatalf("Price pointer was not preserved")
	}
	if !input.IsActive {
		t.Fatalf("IsActive = false, want true")
	}
	if input.Description != req.Description {
		t.Fatalf("Description = %q, want %q", input.Description, req.Description)
	}
	if input.Duration != &duration {
		t.Fatalf("Duration pointer was not preserved")
	}
	if !input.IsCombinable {
		t.Fatalf("IsCombinable = false, want true")
	}
	if input.SortOrder != req.SortOrder {
		t.Fatalf("SortOrder = %d, want %d", input.SortOrder, req.SortOrder)
	}
}

func TestCreateTrimmingOptionRequest_BoolFalseAndOmitted(t *testing.T) {
	active := false
	combinable := false
	if (createTrimmingOptionRequest{Name: "x", IsActive: &active, IsCombinable: &combinable}).toServiceInput().IsActive {
		t.Fatal("explicit is_active false must resolve to false")
	}
	if (createTrimmingOptionRequest{Name: "x", IsActive: &active, IsCombinable: &combinable}).toServiceInput().IsCombinable {
		t.Fatal("explicit is_combinable false must resolve to false")
	}
	omitted := (createTrimmingOptionRequest{Name: "x"}).toServiceInput()
	if !omitted.IsActive || !omitted.IsCombinable {
		t.Fatal("omitted bools must resolve to true")
	}
}

func TestUpdateTrimmingOptionRequest_ToServiceInput(t *testing.T) {
	name := ""
	price := int64(0)
	isActive := false
	description := ""
	duration := 0
	isCombinable := false
	sortOrder := 0

	req := updateTrimmingOptionRequest{
		Name:         &name,
		Price:        &price,
		IsActive:     &isActive,
		Description:  &description,
		Duration:     &duration,
		IsCombinable: &isCombinable,
		SortOrder:    &sortOrder,
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
	if input.IsCombinable != &isCombinable {
		t.Fatalf("IsCombinable pointer was not preserved")
	}
	if input.SortOrder != &sortOrder {
		t.Fatalf("SortOrder pointer was not preserved")
	}
}

func TestUpdateTrimmingOptionRequest_ToServiceInput_NilFields(t *testing.T) {
	input := (&updateTrimmingOptionRequest{}).toServiceInput()

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
	if input.IsCombinable != nil {
		t.Fatalf("IsCombinable = %v, want nil", input.IsCombinable)
	}
	if input.SortOrder != nil {
		t.Fatalf("SortOrder = %v, want nil", input.SortOrder)
	}
}
