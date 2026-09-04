package reservation

import "testing"

func TestCreateReservationStaffRequest_ToServiceInput(t *testing.T) {
	visible := true
	req := createReservationStaffRequest{
		Name:               "Dr. A",
		StaffType:          "doctor",
		ReservationVisible: &visible,
		ReservationComment: "comment",
		SortOrder:          3,
	}

	input := req.toServiceInput()

	if input.Name != req.Name {
		t.Errorf("Name = %q, want %q", input.Name, req.Name)
	}
	if !input.ReservationVisible {
		t.Error("ReservationVisible = false, want true")
	}
	if input.SortOrder != 3 {
		t.Errorf("SortOrder = %d, want 3", input.SortOrder)
	}
}
func TestCreateReservationStaffRequest_ReservationVisibleOmittedDefaultsTrue(t *testing.T) {
	req := createReservationStaffRequest{Name: "Dr. A"}
	if !req.toServiceInput().ReservationVisible {
		t.Error("omitted reservation_visible must resolve to true")
	}
}

func TestCreateReservationStaffRequest_ReservationVisibleFalse(t *testing.T) {
	visible := false
	req := createReservationStaffRequest{Name: "Dr. A", ReservationVisible: &visible}
	if req.toServiceInput().ReservationVisible {
		t.Error("explicit false must resolve to false")
	}
}

func TestUpdateReservationStaffRequest_ToServiceInput(t *testing.T) {
	name := ""
	visible := false
	sortOrder := 0
	req := updateReservationStaffRequest{
		Name:               &name,
		ReservationVisible: &visible,
		SortOrder:          &sortOrder,
	}

	input := req.toServiceInput()

	if input.Name == nil || *input.Name != name {
		t.Errorf("Name = %v, want empty string pointer", input.Name)
	}
	if input.ReservationVisible == nil || *input.ReservationVisible {
		t.Errorf("ReservationVisible = %v, want false pointer", input.ReservationVisible)
	}
	if input.SortOrder == nil || *input.SortOrder != sortOrder {
		t.Errorf("SortOrder = %v, want %d", input.SortOrder, sortOrder)
	}
}
