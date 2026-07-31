package reservation

// line_reservation_setting_request_test.go — BE9-2C R⑥: handler/simple_settings_request_test.go に
// 誤同居していた upsertLineReservationSettingRequest のテストを実装と同 package へ移動。

import (
	"testing"

	"github.com/gin-gonic/gin/binding"
)

func TestUpsertLineReservationSettingRequest_ToServiceInput(t *testing.T) {
	dailyLimit := 0
	showNoStaff := true
	closedWeekdays := jsonRawOrEmpty(`[1,2]`)
	req := upsertLineReservationSettingRequest{
		Status:                  "running",
		ClosedWeekdays:          closedWeekdays,
		DailyLimit:              &dailyLimit,
		BookingWindowMaxDays:    30,
		TimeSlotMode:            "minimize_gaps",
		TimeSlotIntervalMinutes: 15,
		NoStaffMode:             "first_available",
		ShowNoStaffOption:       &showNoStaff,
		// R-05 Phase B: line_channel_secret is no longer accepted on this request.
		// Canonical channel secret write owner is clinic_integrations via L-step settings.
		LineAccessToken: "token",
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
	if input.LineAccessToken != req.LineAccessToken {
		t.Errorf("LineAccessToken = %q, want %q", input.LineAccessToken, req.LineAccessToken)
	}
}

// R-05 Phase B: unknown JSON field line_channel_secret must be ignored (not bound).
func TestUpsertLineReservationSettingRequest_IgnoresLegacyChannelSecretJSON(t *testing.T) {
	const body = `{
		"status":"running",
		"booking_window_max_days":30,
		"booking_window_min_days":0,
		"calendar_months":2,
		"time_slot_mode":"minimize_gaps",
		"time_slot_interval_minutes":15,
		"no_staff_mode":"first_available",
		"line_channel_secret":"must-be-ignored"
	}`
	var req upsertLineReservationSettingRequest
	if err := binding.JSON.BindBody([]byte(body), &req); err != nil {
		t.Fatalf("bind: %v", err)
	}
	input := req.toServiceInput()
	// Input type no longer carries LineChannelSecret; binding success without field is the contract.
	if input.Status != "running" {
		t.Errorf("Status = %q, want running", input.Status)
	}
	if input.LineAccessToken != "" {
		t.Errorf("LineAccessToken unexpectedly set to %q", input.LineAccessToken)
	}
}

func TestUpsertLineReservationSettingRequest_ShowNoStaffOptionFalse(t *testing.T) {
	showNoStaff := false
	req := upsertLineReservationSettingRequest{ShowNoStaffOption: &showNoStaff}
	input := req.toServiceInput()
	if input.ShowNoStaffOption {
		t.Error("explicit false must resolve to false")
	}
}

func TestUpsertLineReservationSettingRequest_ShowNoStaffOptionOmitted(t *testing.T) {
	req := upsertLineReservationSettingRequest{}
	input := req.toServiceInput()
	if !input.ShowNoStaffOption {
		t.Error("omitted show_no_staff_option must resolve to true")
	}
}

// RSV-04 / U-X02-RESERVATION-SETTINGS: binding tags must reject enum/range/email violations.
func TestUpsertLineReservationSettingRequest_BindingRejectsInvalidValues(t *testing.T) {
	valid := func() upsertLineReservationSettingRequest {
		return upsertLineReservationSettingRequest{
			Status:                  "running",
			BookingWindowMaxDays:    30,
			BookingWindowMinDays:    2,
			CalendarMonths:          2,
			TimeSlotMode:            "minimize_gaps",
			TimeSlotIntervalMinutes: 15,
			NoStaffMode:             "first_available",
		}
	}

	if err := binding.Validator.ValidateStruct(valid()); err != nil {
		t.Fatalf("valid request rejected: %v", err)
	}

	badStatus := valid()
	badStatus.Status = "open"
	if err := binding.Validator.ValidateStruct(badStatus); err == nil {
		t.Fatal("expected invalid status rejection")
	}

	negWindow := valid()
	negWindow.BookingWindowMaxDays = -1
	if err := binding.Validator.ValidateStruct(negWindow); err == nil {
		t.Fatal("expected negative booking_window_max_days rejection")
	}

	hugeWindow := valid()
	hugeWindow.BookingWindowMaxDays = 367
	if err := binding.Validator.ValidateStruct(hugeWindow); err == nil {
		t.Fatal("expected booking_window_max_days > 366 rejection")
	}

	badMode := valid()
	badMode.TimeSlotMode = "unknown"
	if err := binding.Validator.ValidateStruct(badMode); err == nil {
		t.Fatal("expected invalid time_slot_mode rejection")
	}

	badNoStaff := valid()
	badNoStaff.NoStaffMode = "hide"
	if err := binding.Validator.ValidateStruct(badNoStaff); err == nil {
		t.Fatal("expected invalid no_staff_mode rejection")
	}

	badEmail := valid()
	badEmail.NotificationEmail = "not-an-email"
	if err := binding.Validator.ValidateStruct(badEmail); err == nil {
		t.Fatal("expected invalid notification_email rejection")
	}

	okEmail := valid()
	okEmail.NotificationEmail = "desk@example.com"
	if err := binding.Validator.ValidateStruct(okEmail); err != nil {
		t.Fatalf("valid notification_email rejected: %v", err)
	}
}
