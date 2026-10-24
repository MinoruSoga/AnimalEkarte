package staff

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildShiftEntryUpdate_WritesExplicitTimestampsAsIs(t *testing.T) {
	// Arrange: 日跨ぎ（22:00 から翌日 07:00）の明示日時
	startsAt := shiftTS(t, testShiftDate, "22:00")
	endsAt := shiftTS(t, testShiftDate.AddDate(0, 0, 1), "07:00")
	input := &UpdateShiftEntryInput{
		StartsAt: OptionalShiftTime{Set: true, Value: startsAt},
		EndsAt:   OptionalShiftTime{Set: true, Value: endsAt},
	}

	// Act
	fields := buildShiftEntryUpdate(input)

	// Assert
	gotStart, ok := fields["starts_at"].(*time.Time)
	require.True(t, ok)
	gotEnd, ok := fields["ends_at"].(*time.Time)
	require.True(t, ok)
	assert.True(t, startsAt.Equal(*gotStart))
	assert.True(t, endsAt.Equal(*gotEnd), "ends_at は翌日の明示日時のまま保存する")
	assert.NotContains(t, fields, "start_time")
	assert.NotContains(t, fields, "end_time")
}

func TestBuildShiftEntryUpdate_ExplicitNullClearsTimestamps(t *testing.T) {
	// Arrange
	input := &UpdateShiftEntryInput{
		StartsAt: OptionalShiftTime{Set: true},
		EndsAt:   OptionalShiftTime{Set: true},
	}

	// Act
	fields := buildShiftEntryUpdate(input)

	// Assert
	require.Contains(t, fields, "starts_at")
	require.Contains(t, fields, "ends_at")
	assert.Nil(t, fields["starts_at"])
	assert.Nil(t, fields["ends_at"])
}

func TestBuildShiftEntryUpdate_UnspecifiedFieldsAreNotWritten(t *testing.T) {
	// Arrange
	notes := "memo"

	// Act
	fields := buildShiftEntryUpdate(&UpdateShiftEntryInput{Notes: &notes})

	// Assert
	assert.Equal(t, map[string]any{"notes": "memo"}, fields)
}

func TestBuildShiftEntryBreaks(t *testing.T) {
	// Arrange
	breakStart := *shiftTS(t, testShiftDate.AddDate(0, 0, 1), "01:00")
	breakEnd := *shiftTS(t, testShiftDate.AddDate(0, 0, 1), "02:00")

	// Act
	breaks := buildShiftEntryBreaks([]ShiftBreakInput{{StartsAt: breakStart, EndsAt: breakEnd}})

	// Assert
	require.Len(t, breaks, 1)
	require.NotNil(t, breaks[0].StartsAt)
	require.NotNil(t, breaks[0].EndsAt)
	assert.True(t, breakStart.Equal(*breaks[0].StartsAt))
	assert.True(t, breakEnd.Equal(*breaks[0].EndsAt))
	assert.Empty(t, buildShiftEntryBreaks(nil), "休憩なしは空配列（全休憩解除）")
}
