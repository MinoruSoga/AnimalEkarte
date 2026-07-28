package reservation

import "testing"

func TestCreateReservationTypeLiffRequest_ToServiceInput(t *testing.T) {
	visible := true
	req := createReservationTypeLiffRequest{
		Name:                 "Online booking",
		Color:                "#AA00AA",
		Description:          "LIFF visible course",
		SortOrder:            4,
		DurationMinutes:      30,
		ShortName:            "ONL",
		ShowShortName:        true,
		ReservationVisible:   &visible,
		ReservationComment:   "Comment",
		ReservationDayOption: "anyday",
		IsInternal:           true,
	}

	input := req.toServiceInput()

	if input.Name != req.Name || input.Color != req.Color || input.Description != req.Description {
		t.Fatalf("basic fields = %+v, want request values", input)
	}
	if input.SortOrder != req.SortOrder || input.DurationMinutes != req.DurationMinutes {
		t.Fatalf("numeric fields = %+v, want request values", input)
	}
	if input.ShortName != req.ShortName || input.ShowShortName != req.ShowShortName ||
		!input.ReservationVisible ||
		input.ReservationComment != req.ReservationComment ||
		input.ReservationDayOption != req.ReservationDayOption ||
		input.IsInternal != req.IsInternal {
		t.Fatalf("LIFF fields = %+v, want request values", input)
	}
}

func TestCreateReservationTypeLiffRequest_ReservationVisibleOmittedDefaultsTrue(t *testing.T) {
	// LIFF omit contract: missing reservation_visible resolves to true (not false).
	req := createReservationTypeLiffRequest{Name: "x"}
	if !req.toServiceInput().ReservationVisible {
		t.Fatal("omitted reservation_visible must resolve to true for LIFF create")
	}
}

func TestCreateReservationTypeLiffRequest_ReservationVisibleFalse(t *testing.T) {
	visible := false
	req := createReservationTypeLiffRequest{Name: "x", ReservationVisible: &visible}
	if req.toServiceInput().ReservationVisible {
		t.Fatal("explicit false must resolve to false")
	}
}

func TestUpdateReservationTypeLiffRequest_ToServiceInput(t *testing.T) {
	name := ""
	color := ""
	description := ""
	sortOrder := 0
	duration := 0
	shortName := ""
	showShortName := false
	visible := false
	comment := ""
	dayOption := "none"
	isInternal := false

	req := updateReservationTypeLiffRequest{
		Name:                 &name,
		Color:                &color,
		Description:          &description,
		SortOrder:            &sortOrder,
		DurationMinutes:      &duration,
		ShortName:            &shortName,
		ShowShortName:        &showShortName,
		ReservationVisible:   &visible,
		ReservationComment:   &comment,
		ReservationDayOption: &dayOption,
		IsInternal:           &isInternal,
	}

	input := req.toServiceInput()

	if input.Name != &name || input.Color != &color || input.Description != &description ||
		input.SortOrder != &sortOrder || input.DurationMinutes != &duration ||
		input.ShortName != &shortName || input.ShowShortName != &showShortName ||
		input.ReservationVisible != &visible || input.ReservationComment != &comment ||
		input.ReservationDayOption != &dayOption || input.IsInternal != &isInternal {
		t.Fatalf("input = %+v, want request pointers", input)
	}
}

func TestUpdateReservationTypeLiffRequest_ToServiceInput_NilFields(t *testing.T) {
	input := (&updateReservationTypeLiffRequest{}).toServiceInput()

	if input.Name != nil || input.Color != nil || input.Description != nil || input.SortOrder != nil ||
		input.DurationMinutes != nil || input.ShortName != nil || input.ShowShortName != nil ||
		input.ReservationVisible != nil || input.ReservationComment != nil ||
		input.ReservationDayOption != nil || input.IsInternal != nil {
		t.Fatalf("input = %+v, want all nil fields", input)
	}
}
