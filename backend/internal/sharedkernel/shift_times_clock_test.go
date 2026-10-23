package sharedkernel

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func strPtr(s string) *string { return &s }

func TestShiftClockToTimestamp(t *testing.T) {
	// Arrange: DATE 列は UTC 0 時で読まれうるため、time.Local の暦日で組み立てる契約を検証する。
	localDate := time.Date(2026, 10, 23, 0, 0, 0, 0, time.Local)

	tests := []struct {
		name  string
		date  time.Time
		clock *string
		want  *time.Time
	}{
		{
			name:  "HH:MM:SS on local midnight date",
			date:  localDate,
			clock: strPtr("09:30:15"),
			want:  ptrTime(time.Date(2026, 10, 23, 9, 30, 15, 0, time.Local)),
		},
		{
			name:  "HH:MM accepted and normalized",
			date:  localDate,
			clock: strPtr("18:05"),
			want:  ptrTime(time.Date(2026, 10, 23, 18, 5, 0, 0, time.Local)),
		},
		{
			name:  "date in another zone uses the local calendar day",
			date:  localDate.In(time.UTC),
			clock: strPtr("00:00:00"),
			want:  ptrTime(time.Date(localDate.In(time.Local).Year(), localDate.In(time.Local).Month(), localDate.In(time.Local).Day(), 0, 0, 0, 0, time.Local)),
		},
		{name: "nil clock", date: localDate, clock: nil, want: nil},
		{name: "empty clock", date: localDate, clock: strPtr(""), want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got, err := ShiftClockToTimestamp(tt.date, tt.clock)

			// Assert
			require.NoError(t, err)
			if tt.want == nil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.True(t, tt.want.Equal(*got), "want %s got %s", tt.want, got)
			assert.Equal(t, time.Local, got.Location())
		})
	}
}

func TestShiftClockToTimestamp_InvalidClock(t *testing.T) {
	for _, bad := range []string{"25:00", "9:5", "abc", "12-30"} {
		t.Run(bad, func(t *testing.T) {
			// Act
			got, err := ShiftClockToTimestamp(time.Date(2026, 10, 23, 0, 0, 0, 0, time.Local), strPtr(bad))

			// Assert
			require.Error(t, err)
			assert.Nil(t, got)
			assert.True(t, errors.Is(err, apperrors.ErrInvalidInput))
		})
	}
}

func TestShiftTimestampToClock(t *testing.T) {
	// Arrange
	ts := time.Date(2026, 10, 23, 9, 5, 7, 0, time.Local)

	// Act / Assert
	require.NotNil(t, ShiftTimestampToClock(&ts))
	assert.Equal(t, "09:05:07", *ShiftTimestampToClock(&ts))
	assert.Equal(t, "09:05:07", *ShiftTimestampToClock(ptrTime(ts.In(time.UTC))), "reads in time.Local regardless of source zone")
	assert.Nil(t, ShiftTimestampToClock(nil))
}

func TestShiftClockRoundTrip(t *testing.T) {
	date := time.Date(2026, 10, 23, 0, 0, 0, 0, time.Local)
	for _, clock := range []string{"00:00:00", "09:00:00", "23:59:59"} {
		ts, err := ShiftClockToTimestamp(date, strPtr(clock))
		require.NoError(t, err)
		assert.Equal(t, clock, *ShiftTimestampToClock(ts))
	}
}

func ptrTime(t time.Time) *time.Time { return &t }
