package staff

import (
	"reflect"
	"testing"
)

func TestReservationStaffUpdateFields_KnownKeysOnly(t *testing.T) {
	name := "改ざん"
	staffType := "nurse"
	visible := false
	comment := "注記"
	sortOrder := 3
	isActive := false

	got := reservationStaffUpdateFields(ReservationStaffUpdate{
		Name:               &name,
		StaffType:          &staffType,
		ReservationVisible: &visible,
		ReservationComment: &comment,
		SortOrder:          &sortOrder,
		IsActive:           &isActive,
	})

	want := map[string]any{
		"name":                name,
		"staff_type":          staffType,
		"reservation_visible": visible,
		"reservation_comment": comment,
		"sort_order":          sortOrder,
		"is_active":           isActive,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("reservationStaffUpdateFields() = %#v, want %#v", got, want)
	}
}

func TestReservationStaffUpdateFields_EmptyOmitsKeys(t *testing.T) {
	got := reservationStaffUpdateFields(ReservationStaffUpdate{})
	if len(got) != 0 {
		t.Fatalf("empty command must not emit columns, got %#v", got)
	}
}
