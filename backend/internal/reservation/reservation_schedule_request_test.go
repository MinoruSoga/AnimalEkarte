package reservation

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
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
	// Arrange
	workStart := "2026-10-23T09:00:00+09:00"
	workEnd := "2026-10-23T18:00:00+09:00"
	req := upsertReservationScheduleRequest{
		ShiftType: "full",
		WorkStart: &workStart,
		WorkEnd:   &workEnd,
		Breaks: []breakInputRequest{
			{Start: "2026-10-23T12:00:00+09:00", End: "2026-10-23T13:00:00+09:00"},
		},
	}

	// Act
	input, err := req.toServiceInput()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, req.ShiftType, input.ShiftType)
	require.NotNil(t, input.StartsAt)
	require.NotNil(t, input.EndsAt)
	assert.True(t, input.StartsAt.Equal(time.Date(2026, 10, 23, 9, 0, 0, 0, config.JST)))
	assert.True(t, input.EndsAt.Equal(time.Date(2026, 10, 23, 18, 0, 0, 0, config.JST)))
	require.Len(t, input.Breaks, 1)
	assert.True(t, input.Breaks[0].StartsAt.Equal(time.Date(2026, 10, 23, 12, 0, 0, 0, config.JST)))
	assert.True(t, input.Breaks[0].EndsAt.Equal(time.Date(2026, 10, 23, 13, 0, 0, 0, config.JST)))
}

func TestUpsertReservationScheduleRequest_ToServiceInput_NilBreaks(t *testing.T) {
	req := upsertReservationScheduleRequest{ShiftType: "off"}

	input, err := req.toServiceInput()

	require.NoError(t, err)
	assert.Nil(t, input.StartsAt)
	assert.Nil(t, input.EndsAt)
	assert.NotNil(t, input.Breaks, "Breaks = nil, want empty slice")
	assert.Empty(t, input.Breaks)
}

func TestUpsertReservationScheduleRequest_ToServiceInput_KeepsOvernightNextDay(t *testing.T) {
	// 日跨ぎ勤務: ends_at は翌日の明示日時のまま保持する。
	workStart := "2026-10-23T22:00:00+09:00"
	workEnd := "2026-10-24T07:00:00+09:00"
	req := upsertReservationScheduleRequest{ShiftType: "full", WorkStart: &workStart, WorkEnd: &workEnd}

	input, err := req.toServiceInput()

	require.NoError(t, err)
	assert.Equal(t, "2026-10-24", input.EndsAt.In(config.JST).Format(time.DateOnly))
}

func TestUpsertReservationScheduleRequest_ToServiceInput_RejectsInvalidTimestamps(t *testing.T) {
	legacyClock := "09:00"
	noOffset := "2026-10-23T09:00:00"
	tests := []struct {
		name string
		req  upsertReservationScheduleRequest
	}{
		{name: "legacy HH:MM work_start", req: upsertReservationScheduleRequest{ShiftType: "full", WorkStart: &legacyClock}},
		{name: "work_end without an offset", req: upsertReservationScheduleRequest{ShiftType: "full", WorkEnd: &noOffset}},
		{name: "legacy HH:MM break", req: upsertReservationScheduleRequest{ShiftType: "full", Breaks: []breakInputRequest{{Start: "12:00", End: "13:00"}}}},
		{name: "break end that is not a timestamp", req: upsertReservationScheduleRequest{ShiftType: "full", Breaks: []breakInputRequest{{Start: "2026-10-23T12:00:00+09:00", End: "ab:cd"}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := tt.req.toServiceInput()

			require.Error(t, err)
			assert.Nil(t, input)
			assert.True(t, apperrors.IsInvalidInput(err))
		})
	}
}
