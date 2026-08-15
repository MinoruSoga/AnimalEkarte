package reservation

import (
	"net/url"
	"testing"
	"time"
)

func TestNewListReservationSchedulesQuery(t *testing.T) {
	tests := []struct {
		name   string
		values url.Values
		now    time.Time
		want   string
	}{
		{
			name:   "uses provided month",
			values: url.Values{"month": []string{"2026-03"}},
			now:    time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
			want:   "2026-03",
		},
		{
			name:   "defaults to current month when missing",
			values: url.Values{},
			now:    time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
			want:   "2026-05",
		},
		{
			name:   "defaults to current month when empty string",
			values: url.Values{"month": []string{""}},
			now:    time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC),
			want:   "2026-12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newListReservationSchedulesQuery(tt.values, tt.now)
			if got.Month != tt.want {
				t.Errorf("Month = %q, want %q", got.Month, tt.want)
			}
		})
	}
}

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
