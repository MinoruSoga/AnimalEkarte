package handler

import "testing"

func TestUpsertReservationScheduleRequest_ToServiceInput(t *testing.T) {
	workStart := "09:00"
	workEnd := "18:00"
	req := upsertReservationScheduleRequest{
		ShiftType: "full",
		WorkStart: &workStart,
		WorkEnd:   &workEnd,
		Breaks: []breakInputRequest{
			{Start: "12:00", End: "13:00"},
		},
	}

	input := req.toServiceInput()

	if input.ShiftType != req.ShiftType {
		t.Errorf("ShiftType = %q, want %q", input.ShiftType, req.ShiftType)
	}
	if input.WorkStart == nil || *input.WorkStart != workStart {
		t.Errorf("WorkStart = %v, want %q", input.WorkStart, workStart)
	}
	if len(input.Breaks) != 1 {
		t.Fatalf("len(Breaks) = %d, want 1", len(input.Breaks))
	}
	if input.Breaks[0].Start != "12:00" || input.Breaks[0].End != "13:00" {
		t.Errorf("Breaks[0] = %+v, want 12:00-13:00", input.Breaks[0])
	}
}

func TestUpsertReservationScheduleRequest_ToServiceInput_NilBreaks(t *testing.T) {
	req := upsertReservationScheduleRequest{ShiftType: "off"}

	input := req.toServiceInput()

	if len(input.Breaks) != 0 {
		t.Errorf("len(Breaks) = %d, want 0", len(input.Breaks))
	}
	if input.Breaks == nil {
		t.Error("Breaks = nil, want empty slice")
	}
}
