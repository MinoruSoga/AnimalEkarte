package reservation

// line_reservation_setting_request_test.go — BE9-2C R⑥: handler/simple_settings_request_test.go に
// 誤同居していた upsertLineReservationSettingRequest のテストを実装と同 package へ移動。

import (
	"testing"
)

func TestUpsertLineReservationSettingRequest_ToServiceInput(t *testing.T) {
	dailyLimit := 0
	showNoStaff := true
	closedWeekdays := jsonRawOrEmpty(`[1,2]`)
	req := upsertLineReservationSettingRequest{
		Status:                  "active",
		ClosedWeekdays:          closedWeekdays,
		DailyLimit:              &dailyLimit,
		BookingWindowMaxDays:    30,
		TimeSlotIntervalMinutes: 15,
		ShowNoStaffOption:       &showNoStaff,
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
