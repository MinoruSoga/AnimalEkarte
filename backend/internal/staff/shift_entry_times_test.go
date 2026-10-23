package staff

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestBuildShiftEntryUpdate_ConvertsClockToTimestampsOnWorkDate(t *testing.T) {
	// Arrange
	start, end := "09:30", "18:00:00"
	input := &UpdateShiftEntryInput{StartTime: &start, EndTime: &end, WorkDate: testShiftDate}

	// Act
	fields, err := buildShiftEntryUpdate(input)

	// Assert
	require.NoError(t, err)
	startsAt, ok := fields["starts_at"].(*time.Time)
	require.True(t, ok)
	endsAt, ok := fields["ends_at"].(*time.Time)
	require.True(t, ok)
	assert.Equal(t, "09:30:00", shiftClock(startsAt))
	assert.Equal(t, "18:00:00", shiftClock(endsAt))
	assert.NotContains(t, fields, "start_time")
	assert.NotContains(t, fields, "end_time")
}

func TestBuildShiftEntryUpdate_EmptyClockClearsTimestamp(t *testing.T) {
	// Arrange
	empty := ""
	input := &UpdateShiftEntryInput{StartTime: &empty, EndTime: &empty, WorkDate: testShiftDate}

	// Act
	fields, err := buildShiftEntryUpdate(input)

	// Assert
	require.NoError(t, err)
	assert.Nil(t, fields["starts_at"])
	assert.Nil(t, fields["ends_at"])
}

func TestBuildShiftEntryUpdate_RequiresWorkDateWhenTimesChange(t *testing.T) {
	// Arrange
	start := "09:00"

	// Act
	fields, err := buildShiftEntryUpdate(&UpdateShiftEntryInput{StartTime: &start})

	// Assert
	require.Error(t, err)
	assert.Nil(t, fields)
}

func TestBuildShiftEntryUpdate_NotesOnlyNeedsNoWorkDate(t *testing.T) {
	// Arrange
	notes := "memo"

	// Act
	fields, err := buildShiftEntryUpdate(&UpdateShiftEntryInput{Notes: &notes})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"notes": "memo"}, fields)
}

func TestBuildShiftEntryBreaks(t *testing.T) {
	t.Run("converts HH:MM breaks on the work date", func(t *testing.T) {
		// Act
		breaks, err := buildShiftEntryBreaks(testShiftDate, []ShiftBreakInput{{BreakStart: "12:00", BreakEnd: "13:00:00"}})

		// Assert
		require.NoError(t, err)
		require.Len(t, breaks, 1)
		assert.Equal(t, "12:00:00", shiftClock(breaks[0].StartsAt))
		assert.Equal(t, "13:00:00", shiftClock(breaks[0].EndsAt))
	})

	t.Run("rejects invalid or empty break times", func(t *testing.T) {
		for _, in := range []ShiftBreakInput{
			{BreakStart: "", BreakEnd: "13:00"},
			{BreakStart: "12:00", BreakEnd: ""},
			{BreakStart: "ab:cd", BreakEnd: "13:00"},
		} {
			// Act
			breaks, err := buildShiftEntryBreaks(testShiftDate, []ShiftBreakInput{in})

			// Assert
			require.Error(t, err)
			assert.Nil(t, breaks)
			assert.True(t, apperrors.IsInvalidInput(err), "input %+v: %v", in, err)
		}
	})
}
