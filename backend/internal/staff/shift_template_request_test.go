package staff

import "testing"

func TestCreateShiftTemplateRequest_ToServiceInput(t *testing.T) {
	isActive := false
	req := createShiftTemplateRequest{
		Name:      "Morning",
		ShiftType: "morning",
		StartTime: "09:00",
		EndTime:   "13:00",
		Notes:     "notes",
		SortOrder: 2,
		IsActive:  &isActive,
		Breaks: []shiftTemplateBreakRequest{
			{BreakStart: "10:00", BreakEnd: "10:15"},
		},
	}

	input := req.toServiceInput()

	if input.Name != req.Name {
		t.Errorf("Name = %q, want %q", input.Name, req.Name)
	}
	if input.IsActive == nil || *input.IsActive {
		t.Errorf("IsActive = %v, want false pointer", input.IsActive)
	}
	if len(input.Breaks) != 1 {
		t.Fatalf("len(Breaks) = %d, want 1", len(input.Breaks))
	}
	if input.Breaks[0].BreakStart != "10:00" {
		t.Errorf("BreakStart = %q, want 10:00", input.Breaks[0].BreakStart)
	}
}

func TestUpdateShiftTemplateRequest_ToServiceInput(t *testing.T) {
	shiftType := "paid_leave"
	startTime := ""
	sortOrder := 0
	breaksReq := []shiftTemplateBreakRequest{{BreakStart: "12:00", BreakEnd: "12:30"}}
	req := updateShiftTemplateRequest{
		ShiftType: &shiftType,
		StartTime: &startTime,
		SortOrder: &sortOrder,
		Breaks:    &breaksReq,
	}

	input := req.toServiceInput()

	if input.ShiftType == nil || string(*input.ShiftType) != shiftType {
		t.Errorf("ShiftType = %v, want %q", input.ShiftType, shiftType)
	}
	if input.StartTime == nil || *input.StartTime != startTime {
		t.Errorf("StartTime = %v, want empty string pointer", input.StartTime)
	}
	if input.SortOrder == nil || *input.SortOrder != sortOrder {
		t.Errorf("SortOrder = %v, want %d", input.SortOrder, sortOrder)
	}
	if input.Breaks == nil || len(*input.Breaks) != 1 {
		t.Fatalf("Breaks = %v, want one item", input.Breaks)
	}
	if (*input.Breaks)[0].BreakEnd != "12:30" {
		t.Errorf("BreakEnd = %q, want 12:30", (*input.Breaks)[0].BreakEnd)
	}
}
