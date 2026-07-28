package reservation

import "testing"

func TestCreateReservationTypeRequest_ToServiceInput(t *testing.T) {
	duration := 30
	visible := false
	active := true
	groupID := uint64(7)

	req := createReservationTypeRequest{
		Name:                   "Consultation",
		Color:                  "#123456",
		IsActive:               &active,
		Description:            "General consultation",
		SortOrder:              3,
		GroupID:                &groupID,
		Category:               "general",
		ReservationDisplayName: "Visit",
		DurationMinutes:        &duration,
		ShortName:              "VIS",
		ShowShortName:          true,
		ReservationVisible:     &visible,
		ReservationComment:     "Bring records",
		ReservationImageURL:    "https://example.test/image.png",
		ReservationDayOption:   "weekday",
		IsInternal:             true,
	}

	input := req.toServiceInput()

	if input.Name != req.Name || input.Color != req.Color || input.Category != req.Category {
		t.Fatalf("basic fields = %+v, want request values", input)
	}
	if !input.IsActive {
		t.Fatalf("IsActive = false, want true")
	}
	if input.GroupID != &groupID || input.DurationMinutes != &duration || input.ReservationVisible != &visible {
		t.Fatalf("pointer fields were not preserved")
	}
	if input.ReservationDisplayName != req.ReservationDisplayName || input.ShortName != req.ShortName ||
		input.ReservationDayOption != req.ReservationDayOption || input.IsInternal != req.IsInternal {
		t.Fatalf("LINE fields = %+v, want request values", input)
	}
}

func TestCreateReservationTypeRequest_IsActiveFalse(t *testing.T) {
	active := false
	req := createReservationTypeRequest{Name: "x", IsActive: &active}
	if req.toServiceInput().IsActive {
		t.Fatal("explicit false must resolve to false")
	}
}

func TestCreateReservationTypeRequest_IsActiveOmitted(t *testing.T) {
	req := createReservationTypeRequest{Name: "x"}
	if !req.toServiceInput().IsActive {
		t.Fatal("omitted is_active must resolve to true")
	}
}

func TestUpdateReservationTypeRequest_ToServiceInput(t *testing.T) {
	name := ""
	color := ""
	isActive := false
	description := ""
	sortOrder := 0
	groupID := uint64(0)
	category := "trimming"
	displayName := ""
	duration := 0
	shortName := ""
	showShortName := false
	visible := false
	comment := ""
	imageURL := ""
	dayOption := "none"
	isInternal := false

	req := updateReservationTypeRequest{
		Name:                   &name,
		Color:                  &color,
		IsActive:               &isActive,
		Description:            &description,
		SortOrder:              &sortOrder,
		GroupID:                &groupID,
		ClearGroupID:           true,
		Category:               &category,
		ReservationDisplayName: &displayName,
		DurationMinutes:        &duration,
		ShortName:              &shortName,
		ShowShortName:          &showShortName,
		ReservationVisible:     &visible,
		ReservationComment:     &comment,
		ReservationImageURL:    &imageURL,
		ReservationDayOption:   &dayOption,
		IsInternal:             &isInternal,
	}

	input := req.toServiceInput()

	if input.Name != &name || input.Color != &color || input.IsActive != &isActive || input.Description != &description {
		t.Fatalf("basic pointer fields were not preserved")
	}
	if input.SortOrder != &sortOrder || input.GroupID != &groupID || !input.ClearGroupID || input.Category != &category {
		t.Fatalf("group/category fields were not preserved")
	}
	if input.ReservationDisplayName != &displayName || input.DurationMinutes != &duration || input.ShortName != &shortName ||
		input.ShowShortName != &showShortName || input.ReservationVisible != &visible ||
		input.ReservationComment != &comment || input.ReservationImageURL != &imageURL ||
		input.ReservationDayOption != &dayOption || input.IsInternal != &isInternal {
		t.Fatalf("LINE pointer fields were not preserved")
	}
}

func TestCreateUnavailableTimeRequest_ToServiceInput(t *testing.T) {
	dayOfWeek := int8(1)
	req := createUnavailableTimeRequest{
		UnavailableType: "weekly",
		DayOfWeek:       &dayOfWeek,
		StartTime:       "09:00",
		EndTime:         "10:00",
	}

	input, err := req.toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}

	if input.UnavailableType != req.UnavailableType {
		t.Errorf("UnavailableType = %q, want %q", input.UnavailableType, req.UnavailableType)
	}
	if input.DayOfWeek == nil || *input.DayOfWeek != dayOfWeek {
		t.Errorf("DayOfWeek = %v, want %d", input.DayOfWeek, dayOfWeek)
	}
	if input.StartTime != req.StartTime {
		t.Errorf("StartTime = %q, want %q", input.StartTime, req.StartTime)
	}
}

func TestCreateUnavailableTimeRequest_ToServiceInput_SpecificDate(t *testing.T) {
	specificDate := "2026-05-28"
	req := createUnavailableTimeRequest{
		UnavailableType: "specific",
		SpecificDate:    &specificDate,
		StartTime:       "09:00",
		EndTime:         "10:00",
	}

	input, err := req.toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}

	if input.SpecificDate == nil || input.SpecificDate.Format("2006-01-02") != specificDate {
		t.Errorf("SpecificDate = %v, want %q", input.SpecificDate, specificDate)
	}
}

func TestUpdateReservationTypeRequest_ToServiceInput_NilFields(t *testing.T) {
	input := (&updateReservationTypeRequest{}).toServiceInput()

	if input.Name != nil || input.Color != nil || input.IsActive != nil || input.Description != nil ||
		input.SortOrder != nil || input.GroupID != nil || input.ClearGroupID || input.Category != nil ||
		input.ReservationDisplayName != nil || input.DurationMinutes != nil || input.ShortName != nil ||
		input.ShowShortName != nil || input.ReservationVisible != nil || input.ReservationComment != nil ||
		input.ReservationImageURL != nil || input.ReservationDayOption != nil || input.IsInternal != nil {
		t.Fatalf("input = %+v, want all nil fields and ClearGroupID false", input)
	}
}
