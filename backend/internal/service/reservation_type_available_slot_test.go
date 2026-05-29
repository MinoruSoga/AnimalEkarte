package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type mockAvailableSlotRepository struct {
	findAllFn  func(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeAvailableSlot, error)
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeAvailableSlot, error)
	createFn   func(ctx context.Context, slot *model.ReservationTypeAvailableSlot) error
	deleteFn   func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockAvailableSlotRepository) FindAll(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeAvailableSlot, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID, reservationTypeID)
	}
	return nil, nil
}

func (m *mockAvailableSlotRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeAvailableSlot, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.ReservationTypeAvailableSlot{ID: id}, nil
}

func (m *mockAvailableSlotRepository) Create(ctx context.Context, slot *model.ReservationTypeAvailableSlot) error {
	if m.createFn != nil {
		return m.createFn(ctx, slot)
	}
	return nil
}

func (m *mockAvailableSlotRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func TestReservationService_Create_RejectsUnavailableStartSlot(t *testing.T) {
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, jstLocation) // Monday
	end := start.Add(60 * time.Minute)
	repo := &mockReservationRepository{
		createFn: func(_ context.Context, _ *model.Reservation) error {
			t.Fatal("reservation must not be created outside available slots")
			return nil
		},
	}
	availableSlotRepo := &mockAvailableSlotRepository{
		findAllFn: func(_ context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeAvailableSlot, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(5), reservationTypeID)
			dayOfWeek := int8(1)
			return []model.ReservationTypeAvailableSlot{
				{ReservationTypeID: 5, AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &dayOfWeek, StartTime: "09:45", IsActive: true},
				{ReservationTypeID: 5, AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &dayOfWeek, StartTime: "12:30", IsActive: true},
			}, nil
		},
	}
	svc := NewReservationServiceWithAvailability(repo, nil, nil, nil, availableSlotRepo)

	result, err := svc.Create(context.Background(), &CreateManualReservationInput{
		ClinicID:          1,
		StartTime:         start,
		EndTime:           end,
		ReservationTypeID: 5,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsInvalidInput(err), "expected ErrInvalidInput but got: %v", err)
}

func TestFilterApplicableAvailableSlots_WeeklyAndSpecific(t *testing.T) {
	monday := int8(1)
	tuesday := int8(2)
	specificDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, jstLocation)

	result := filterApplicableAvailableSlots([]model.ReservationTypeAvailableSlot{
		{AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &monday, StartTime: "09:45", IsActive: true},
		{AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &tuesday, StartTime: "12:30", IsActive: true},
		{AvailableType: model.AvailableSlotTypeSpecific, SpecificDate: &specificDate, StartTime: "14:00", IsActive: true},
		{AvailableType: model.AvailableSlotTypeSpecific, SpecificDate: &specificDate, StartTime: "15:00", IsActive: false},
	}, date)

	assert.Len(t, result, 2)
	assert.Equal(t, "09:45", result[0].StartTime)
	assert.Equal(t, "14:00", result[1].StartTime)
}

func TestFilterTimeSlotsByAvailableSlots(t *testing.T) {
	monday := int8(1)
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, jstLocation)

	result := filterTimeSlotsByAvailableSlots(
		[]TimeSlot{
			{StartTime: "0900", EndTime: "1000"},
			{StartTime: "0945", EndTime: "1045"},
			{StartTime: "1230", EndTime: "1330"},
		},
		[]model.ReservationTypeAvailableSlot{
			{AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &monday, StartTime: "09:45", IsActive: true},
			{AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &monday, StartTime: "12:30", IsActive: true},
		},
		date,
	)

	assert.Equal(t, []TimeSlot{
		{StartTime: "0945", EndTime: "1045"},
		{StartTime: "1230", EndTime: "1330"},
	}, result)
}
