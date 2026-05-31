package handler

import "testing"

func TestCreateReservationTypeGroupRequest_ToServiceInput(t *testing.T) {
	req := createReservationTypeGroupRequest{
		Name:      "Grooming",
		Color:     "#00AA88",
		SortOrder: 4,
		IsActive:  true,
	}

	input := req.toServiceInput()

	if input.Name != req.Name {
		t.Fatalf("Name = %q, want %q", input.Name, req.Name)
	}
	if input.Color != req.Color {
		t.Fatalf("Color = %q, want %q", input.Color, req.Color)
	}
	if input.SortOrder != req.SortOrder {
		t.Fatalf("SortOrder = %d, want %d", input.SortOrder, req.SortOrder)
	}
	if input.IsActive != req.IsActive {
		t.Fatalf("IsActive = %v, want %v", input.IsActive, req.IsActive)
	}
}

func TestUpdateReservationTypeGroupRequest_ToServiceInput(t *testing.T) {
	name := ""
	color := ""
	sortOrder := 0
	isActive := false

	req := updateReservationTypeGroupRequest{
		Name:      &name,
		Color:     &color,
		SortOrder: &sortOrder,
		IsActive:  &isActive,
	}

	input := req.toServiceInput()

	if input.Name != &name {
		t.Fatalf("Name pointer was not preserved")
	}
	if input.Color != &color {
		t.Fatalf("Color pointer was not preserved")
	}
	if input.SortOrder != &sortOrder {
		t.Fatalf("SortOrder pointer was not preserved")
	}
	if input.IsActive != &isActive {
		t.Fatalf("IsActive pointer was not preserved")
	}
}

func TestUpdateReservationTypeGroupRequest_ToServiceInput_NilFields(t *testing.T) {
	input := (&updateReservationTypeGroupRequest{}).toServiceInput()

	if input.Name != nil || input.Color != nil || input.SortOrder != nil || input.IsActive != nil {
		t.Fatalf("input = %+v, want all nil fields", input)
	}
}
