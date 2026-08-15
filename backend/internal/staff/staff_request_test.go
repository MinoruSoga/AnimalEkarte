package staff

import "testing"

func TestCreateStaffRequest_ToCreateServiceInput(t *testing.T) {
	occupationID := uint64(10)
	visible := false

	input := (&createStaffRequest{
		Name:                   "Dr. Staff",
		LicenseNumber:          "LIC-001",
		OccupationID:           &occupationID,
		SortOrder:              3,
		StaffType:              "doctor",
		ReservationDisplayName: "Dr. S",
		ReservationVisible:     &visible,
		ReservationComment:     "comment",
		ReservationImageURL:    "https://example.test/staff.png",
	}).toCreateServiceInput(5)

	if input.ClinicID != 5 {
		t.Fatalf("ClinicID = %d, want 5", input.ClinicID)
	}
	if input.OccupationID != &occupationID {
		t.Fatalf("OccupationID pointer was not preserved")
	}
	if input.ReservationVisible != &visible {
		t.Fatalf("ReservationVisible pointer was not preserved")
	}
}

func TestCreateStaffRequest_ToCreateWithAccountServiceInput(t *testing.T) {
	input := (&createStaffRequest{
		Name:     "Dr. Staff",
		Email:    " staff@example.test ",
		Password: "password123",
	}).toCreateWithAccountServiceInput(5)

	if input.ClinicID != 5 {
		t.Fatalf("ClinicID = %d, want 5", input.ClinicID)
	}
	if input.Email != "staff@example.test" {
		t.Fatalf("Email = %q, want staff@example.test", input.Email)
	}
	if input.Password != "password123" {
		t.Fatalf("Password = %q, want password123", input.Password)
	}
}

func TestCreateStaffRequest_HasAccountEmail(t *testing.T) {
	if (&createStaffRequest{Email: "   "}).hasAccountEmail() {
		t.Fatalf("hasAccountEmail() = true, want false")
	}
	if !(&createStaffRequest{Email: "staff@example.test"}).hasAccountEmail() {
		t.Fatalf("hasAccountEmail() = false, want true")
	}
}

func TestUpdateStaffRequest_ToServiceInput(t *testing.T) {
	name := ""
	licenseNumber := ""
	occupationID := uint64(10)
	sortOrder := 0
	isActive := false
	password := "password123"
	staffType := "doctor"
	reservationDisplayName := "Dr. S"
	reservationVisible := false
	reservationComment := ""
	reservationImageURL := ""

	input := (&updateStaffRequest{
		Name:                   &name,
		LicenseNumber:          &licenseNumber,
		OccupationID:           &occupationID,
		SortOrder:              &sortOrder,
		IsActive:               &isActive,
		Password:               &password,
		StaffType:              &staffType,
		ReservationDisplayName: &reservationDisplayName,
		ReservationVisible:     &reservationVisible,
		ReservationComment:     &reservationComment,
		ReservationImageURL:    &reservationImageURL,
	}).toServiceInput()

	if input.Name != &name {
		t.Fatalf("Name pointer was not preserved")
	}
	if input.IsActive != &isActive {
		t.Fatalf("IsActive pointer was not preserved")
	}
	if input.ReservationVisible != &reservationVisible {
		t.Fatalf("ReservationVisible pointer was not preserved")
	}
}

func TestUpdateStaffRequest_ToServiceInput_NilFields(t *testing.T) {
	input := (&updateStaffRequest{}).toServiceInput()

	if input.Name != nil || input.IsActive != nil || input.ReservationVisible != nil {
		t.Fatalf("expected nil optional fields, got %#v", input)
	}
}
