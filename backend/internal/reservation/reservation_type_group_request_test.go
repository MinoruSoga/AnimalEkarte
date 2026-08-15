package reservation

import "testing"

func TestCreateReservationTypeGroupRequest_ToServiceInput(t *testing.T) {
	active := true
	req := createReservationTypeGroupRequest{
		Name:      "Grooming",
		Color:     "#00AA88",
		SortOrder: 4,
		IsActive:  &active,
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
	if !input.IsActive {
		t.Fatalf("IsActive = false, want true")
	}
}

func TestCreateReservationTypeGroupRequest_IsActiveFalse(t *testing.T) {
	active := false
	req := createReservationTypeGroupRequest{Name: "x", IsActive: &active}
	if req.toServiceInput().IsActive {
		t.Fatal("explicit false must resolve to false")
	}
}

func TestCreateReservationTypeGroupRequest_IsActiveOmitted(t *testing.T) {
	req := createReservationTypeGroupRequest{Name: "x"}
	if !req.toServiceInput().IsActive {
		t.Fatal("omitted is_active must resolve to true")
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
