package clinic

import "testing"

func TestUpdateClinicSettingsRequest_ToServiceInput(t *testing.T) {
	boundary := "12:00"
	req := UpdateClinicSettingsRequest{
		ClosingAmPmBoundary: &boundary,
		ClosedWeekdays:      []int64{0, 6},
	}

	input := req.ToServiceInput()

	if input.ClosingAmPmBoundary == nil || *input.ClosingAmPmBoundary != boundary {
		t.Errorf("ClosingAmPmBoundary = %v, want %q", input.ClosingAmPmBoundary, boundary)
	}
	if len(input.ClosedWeekdays) != 2 {
		t.Errorf("len(ClosedWeekdays) = %d, want 2", len(input.ClosedWeekdays))
	}
}

func TestCreateSpecialPeriodRequest_ToServiceInput(t *testing.T) {
	req := CreateSpecialPeriodRequest{
		StartDate:    "2026-05-28",
		EndDate:      "2026-05-29",
		AmPmBoundary: "12:00",
		PmEnd:        "18:00",
		Note:         "holiday",
	}

	input := req.ToServiceInput()

	if input.StartDate != req.StartDate {
		t.Errorf("StartDate = %q, want %q", input.StartDate, req.StartDate)
	}
	if input.PmEnd != req.PmEnd {
		t.Errorf("PmEnd = %q, want %q", input.PmEnd, req.PmEnd)
	}
}

func TestUpdateSpecialPeriodRequest_ToServiceInput(t *testing.T) {
	startDate := "2026-05-28"
	note := ""
	req := UpdateSpecialPeriodRequest{
		StartDate: &startDate,
		Note:      &note,
	}

	input := req.ToServiceInput()

	if input.StartDate == nil || *input.StartDate != startDate {
		t.Errorf("StartDate = %v, want %q", input.StartDate, startDate)
	}
	if input.Note == nil || *input.Note != note {
		t.Errorf("Note = %v, want empty string pointer", input.Note)
	}
}
