package reservation

// line_reservation_setting_request_test.go — BE9-2C R⑥: handler/simple_settings_request_test.go に
// 誤同居していた upsertLineReservationSettingRequest のテストを実装と同 package へ移動。

import (
	"testing"
)

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
