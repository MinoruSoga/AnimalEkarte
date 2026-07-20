package handler

import "testing"

func TestUpdateCompanyRequest_ToServiceInput(t *testing.T) {
	name := ""
	email := "clinic@example.com"
	req := updateCompanyRequest{Name: &name, Email: &email}

	input := req.toServiceInput()

	if input.Name == nil || *input.Name != name {
		t.Errorf("Name = %v, want empty string pointer", input.Name)
	}
	if input.Email == nil || *input.Email != email {
		t.Errorf("Email = %v, want %q", input.Email, email)
	}
}

// TestUpdateClinicalPlanRequest_ToServiceInput moved to
// internal/medicalrecord/clinical_plan_request_test.go (BE9-2D sub-batch④a — updateClinicalPlanRequest moved there).

// TestUpdateInquiryRequest_ToServiceInput moved to
// internal/medicalrecord/inquiry_request_test.go (BE9-2D — updateInquiryRequest moved there).

func TestUpsertLineReservationSettingRequest_ToServiceInput(t *testing.T) {
	dailyLimit := 0
	closedWeekdays := jsonRawOrEmpty(`[1,2]`)
	req := upsertLineReservationSettingRequest{
		Status:                  "active",
		ClosedWeekdays:          closedWeekdays,
		DailyLimit:              &dailyLimit,
		BookingWindowMaxDays:    30,
		TimeSlotIntervalMinutes: 15,
		ShowNoStaffOption:       true,
		LineChannelSecret:       "secret",
	}

	input := req.toServiceInput()

	if input.Status != req.Status {
		t.Errorf("Status = %q, want %q", input.Status, req.Status)
	}
	if string(input.ClosedWeekdays) != string(closedWeekdays) {
		t.Errorf("ClosedWeekdays = %s, want %s", input.ClosedWeekdays, closedWeekdays)
	}
	if input.DailyLimit == nil || *input.DailyLimit != dailyLimit {
		t.Errorf("DailyLimit = %v, want %d", input.DailyLimit, dailyLimit)
	}
	if !input.ShowNoStaffOption {
		t.Error("ShowNoStaffOption = false, want true")
	}
	if input.LineChannelSecret != req.LineChannelSecret {
		t.Errorf("LineChannelSecret = %q, want %q", input.LineChannelSecret, req.LineChannelSecret)
	}
}
