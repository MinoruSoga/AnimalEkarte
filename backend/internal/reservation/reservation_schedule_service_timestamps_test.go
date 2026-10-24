package reservation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestReservationScheduleService_Save_PassesExplicitTimestampsToRepository(t *testing.T) {
	// Arrange
	date := time.Date(2026, 10, 23, 0, 0, 0, 0, config.JST)
	startsAt := rsvTS(t, date, "09:00")
	endsAt := rsvTS(t, date, "18:00:00")
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
		StartsAt:  startsAt,
		EndsAt:    endsAt,
		Breaks:    []ReservationScheduleBreakInput{{StartsAt: *rsvTS(t, date, "12:00"), EndsAt: *rsvTS(t, date, "13:00")}},
	})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, gotEntry)
	assert.True(t, startsAt.Equal(*gotEntry.StartsAt))
	assert.True(t, endsAt.Equal(*gotEntry.EndsAt))
	assert.Equal(t, "09:00:00", rsvClock(gotEntry.StartsAt))
	assert.Equal(t, "18:00:00", rsvClock(gotEntry.EndsAt))
	assert.Equal(t, date.Format(time.DateOnly), gotEntry.StartsAt.In(config.JST).Format(time.DateOnly))
	require.Len(t, gotBreaks, 1)
	assert.Equal(t, "12:00:00", rsvClock(gotBreaks[0].StartsAt))
	assert.Equal(t, "13:00:00", rsvClock(gotBreaks[0].EndsAt))
}

func TestReservationScheduleService_Save_AcceptsOvernightShift(t *testing.T) {
	// Arrange: 22:00 から翌日 07:00。翌日は明示日時で表し、時刻の大小から推測しない。
	date := time.Date(2026, 10, 23, 0, 0, 0, 0, config.JST)
	nextDay := date.AddDate(0, 0, 1)
	var gotEntry *model.ShiftEntry
	repo := &mockReservationScheduleRepository{
		saveFn: func(_ context.Context, _ uint64, entry *model.ShiftEntry, breaks []model.ShiftEntryBreak) (*model.ShiftEntry, []model.ShiftEntryBreak, bool, error) {
			gotEntry = entry
			return entry, breaks, true, nil
		},
	}
	svc := NewReservationScheduleService(repo)

	// Act
	_, _, err := svc.Save(context.Background(), 1, 2, date, &CreateReservationScheduleInput{
		ShiftType: string(model.ShiftTypeFull),
		StartsAt:  rsvTS(t, date, "22:00"),
		EndsAt:    rsvTS(t, nextDay, "07:00"),
		Breaks:    []ReservationScheduleBreakInput{{StartsAt: *rsvTS(t, nextDay, "01:00"), EndsAt: *rsvTS(t, nextDay, "02:00")}},
	})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, gotEntry)
	assert.Equal(t, "2026-10-24", gotEntry.EndsAt.In(config.JST).Format(time.DateOnly), "ends_at は翌日の明示日時のまま保存する")
}

func TestReservationScheduleService_Save_RejectsInvalidShapeBeforeSave(t *testing.T) {
	date := time.Date(2026, 10, 23, 0, 0, 0, 0, config.JST)
	nextDay := date.AddDate(0, 0, 1)
	tests := []struct {
		name  string
		input *CreateReservationScheduleInput
	}{
		{name: "ends_at before starts_at on the same date", input: &CreateReservationScheduleInput{ShiftType: string(model.ShiftTypeFull), StartsAt: rsvTS(t, date, "22:00"), EndsAt: rsvTS(t, date, "07:00")}},
		{name: "off shift with times", input: &CreateReservationScheduleInput{ShiftType: string(model.ShiftTypeOff), StartsAt: rsvTS(t, date, "09:00"), EndsAt: rsvTS(t, date, "18:00")}},
		{name: "off shift with a break", input: &CreateReservationScheduleInput{ShiftType: string(model.ShiftTypeOff), Breaks: []ReservationScheduleBreakInput{{StartsAt: *rsvTS(t, date, "12:00"), EndsAt: *rsvTS(t, date, "13:00")}}}},
		{name: "break outside the shift", input: &CreateReservationScheduleInput{ShiftType: string(model.ShiftTypeFull), StartsAt: rsvTS(t, date, "09:00"), EndsAt: rsvTS(t, date, "18:00"), Breaks: []ReservationScheduleBreakInput{{StartsAt: *rsvTS(t, date, "17:30"), EndsAt: *rsvTS(t, date, "18:30")}}}},
		{name: "overlapping breaks", input: &CreateReservationScheduleInput{ShiftType: string(model.ShiftTypeFull), StartsAt: rsvTS(t, date, "09:00"), EndsAt: rsvTS(t, date, "18:00"), Breaks: []ReservationScheduleBreakInput{{StartsAt: *rsvTS(t, date, "12:00"), EndsAt: *rsvTS(t, date, "13:00")}, {StartsAt: *rsvTS(t, date, "12:30"), EndsAt: *rsvTS(t, date, "13:30")}}}},
		{name: "starts_at on another date than the schedule date", input: &CreateReservationScheduleInput{ShiftType: string(model.ShiftTypeFull), StartsAt: rsvTS(t, nextDay, "09:00"), EndsAt: rsvTS(t, nextDay, "18:00")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saves := 0
			repo := &mockReservationScheduleRepository{
				saveFn: func(_ context.Context, _ uint64, entry *model.ShiftEntry, breaks []model.ShiftEntryBreak) (*model.ShiftEntry, []model.ShiftEntryBreak, bool, error) {
					saves++
					return entry, breaks, true, nil
				},
			}
			svc := NewReservationScheduleService(repo)

			_, _, err := svc.Save(context.Background(), 1, 2, date, tt.input)

			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err))
			assert.Zero(t, saves)
		})
	}
}
