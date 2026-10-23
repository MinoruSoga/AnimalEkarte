package reservation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestReservationScheduleService_Save_ConvertsClocksToTimestampsOnWorkDate(t *testing.T) {
	// Arrange
	date := time.Date(2026, 10, 23, 0, 0, 0, 0, time.Local)
	start, end := "09:00", "18:00:00"
	var gotEntry *model.ShiftEntry
	var gotBreaks []model.ShiftEntryBreak
	repo := &mockReservationScheduleRepository{
		saveFn: func(_ context.Context, _ uint64, entry *model.ShiftEntry, breaks []model.ShiftEntryBreak) (*model.ShiftEntry, []model.ShiftEntryBreak, bool, error) {
			gotEntry, gotBreaks = entry, breaks
			return entry, breaks, true, nil
		},
	}
	svc := NewReservationScheduleService(repo)

	// Act
	_, _, err := svc.Save(context.Background(), 1, 2, date, &CreateReservationScheduleInput{
		ShiftType: string(model.ShiftTypeFull),
		WorkStart: &start,
		WorkEnd:   &end,
		Breaks:    []ReservationScheduleBreakInput{{Start: "12:00", End: "13:00"}},
	})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, gotEntry)
	assert.Equal(t, "09:00:00", rsvClock(gotEntry.StartsAt))
	assert.Equal(t, "18:00:00", rsvClock(gotEntry.EndsAt))
	assert.Equal(t, date.Format(time.DateOnly), gotEntry.StartsAt.In(time.Local).Format(time.DateOnly))
	require.Len(t, gotBreaks, 1)
	assert.Equal(t, "12:00:00", rsvClock(gotBreaks[0].StartsAt))
	assert.Equal(t, "13:00:00", rsvClock(gotBreaks[0].EndsAt))
}

func TestReservationScheduleService_Save_RejectsInvalidBreakTimeBeforeSave(t *testing.T) {
	// Arrange
	start, end := "09:00", "18:00"
	saves := 0
	repo := &mockReservationScheduleRepository{
		saveFn: func(_ context.Context, _ uint64, entry *model.ShiftEntry, breaks []model.ShiftEntryBreak) (*model.ShiftEntry, []model.ShiftEntryBreak, bool, error) {
			saves++
			return entry, breaks, true, nil
		},
	}
	svc := NewReservationScheduleService(repo)

	// Act
	_, _, err := svc.Save(context.Background(), 1, 2, time.Date(2026, 10, 23, 0, 0, 0, 0, time.Local), &CreateReservationScheduleInput{
		ShiftType: string(model.ShiftTypeFull),
		WorkStart: &start,
		WorkEnd:   &end,
		Breaks:    []ReservationScheduleBreakInput{{Start: "ab:cd", End: "13:00"}},
	})

	// Assert
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Zero(t, saves)
}
