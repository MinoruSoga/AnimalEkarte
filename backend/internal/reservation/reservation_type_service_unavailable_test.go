package reservation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- ListUnavailableTimes ----

func TestReservationTypeService_ListUnavailableTimes(t *testing.T) {
	t.Run("returns unavailable times for reservation type", func(t *testing.T) {
		const (
			clinicID          uint64 = 1
			reservationTypeID uint64 = 10
		)
		want := []model.ReservationTypeUnavailableTime{
			{ID: 1, ClinicID: clinicID, ReservationTypeID: reservationTypeID, UnavailableType: model.UnavailableTypeWeekly},
		}
		unavailableRepo := &mockUnavailableTimeRepository{
			findAllFn: func(_ context.Context, cid, rtID uint64) ([]model.ReservationTypeUnavailableTime, error) {
				assert.Equal(t, clinicID, cid)
				assert.Equal(t, reservationTypeID, rtID)
				return want, nil
			},
		}
		svc := NewReservationTypeService(&mockReservationTypeRepository{}, unavailableRepo, &mockReservationTypeOccupationRepository{}, &mockOccupationRepository{}, nil)

		got, err := svc.ListUnavailableTimes(context.Background(), clinicID, reservationTypeID)

		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("returns error when reservation type not found", func(t *testing.T) {
		repo := &mockReservationTypeRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
				return nil, apperrors.WrapNotFound("reservation_type", "10")
			},
		}
		svc := NewReservationTypeService(repo, &mockUnavailableTimeRepository{}, &mockReservationTypeOccupationRepository{}, &mockOccupationRepository{}, nil)

		_, err := svc.ListUnavailableTimes(context.Background(), 1, 10)

		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("propagates repository error listing unavailable times", func(t *testing.T) {
		unavailableRepo := &mockUnavailableTimeRepository{
			findAllFn: func(_ context.Context, _, _ uint64) ([]model.ReservationTypeUnavailableTime, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewReservationTypeService(&mockReservationTypeRepository{}, unavailableRepo, &mockReservationTypeOccupationRepository{}, &mockOccupationRepository{}, nil)

		_, err := svc.ListUnavailableTimes(context.Background(), 1, 10)

		assert.Error(t, err)
	})
}

// ---- CreateUnavailableTime ----

func TestReservationTypeService_CreateUnavailableTime(t *testing.T) {
	dayOfWeek := int8(1)
	specificDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	t.Run("creates weekly unavailable time successfully", func(t *testing.T) {
		var captured *model.ReservationTypeUnavailableTime
		unavailableRepo := &mockUnavailableTimeRepository{
			createFn: func(_ context.Context, tt *model.ReservationTypeUnavailableTime) error {
				captured = tt
				return nil
			},
		}
		svc := NewReservationTypeService(&mockReservationTypeRepository{}, unavailableRepo, &mockReservationTypeOccupationRepository{}, &mockOccupationRepository{}, nil)

		got, err := svc.CreateUnavailableTime(context.Background(), 1, 10, CreateUnavailableTimeInput{
			UnavailableType: string(model.UnavailableTypeWeekly),
			DayOfWeek:       &dayOfWeek,
			StartTime:       "09:00",
			EndTime:         "10:00",
		})

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, uint64(1), got.ClinicID)
		assert.Equal(t, uint64(10), got.ReservationTypeID)
		assert.Equal(t, model.UnavailableTypeWeekly, got.UnavailableType)
		assert.Same(t, captured, got)
	})

	t.Run("creates specific unavailable time successfully", func(t *testing.T) {
		unavailableRepo := &mockUnavailableTimeRepository{}
		svc := NewReservationTypeService(&mockReservationTypeRepository{}, unavailableRepo, &mockReservationTypeOccupationRepository{}, &mockOccupationRepository{}, nil)

		got, err := svc.CreateUnavailableTime(context.Background(), 1, 10, CreateUnavailableTimeInput{
			UnavailableType: string(model.UnavailableTypeSpecific),
			SpecificDate:    &specificDate,
			StartTime:       "09:00",
			EndTime:         "10:00",
		})

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, model.UnavailableTypeSpecific, got.UnavailableType)
	})

	t.Run("returns error when reservation type not found", func(t *testing.T) {
		repo := &mockReservationTypeRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
				return nil, apperrors.WrapNotFound("reservation_type", "10")
			},
		}
		svc := NewReservationTypeService(repo, &mockUnavailableTimeRepository{}, &mockReservationTypeOccupationRepository{}, &mockOccupationRepository{}, nil)

		_, err := svc.CreateUnavailableTime(context.Background(), 1, 10, CreateUnavailableTimeInput{
			UnavailableType: string(model.UnavailableTypeWeekly),
			DayOfWeek:       &dayOfWeek,
			StartTime:       "09:00",
			EndTime:         "10:00",
		})

		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("returns invalid input for validation failure", func(t *testing.T) {
		svc := NewReservationTypeService(&mockReservationTypeRepository{}, &mockUnavailableTimeRepository{}, &mockReservationTypeOccupationRepository{}, &mockOccupationRepository{}, nil)

		_, err := svc.CreateUnavailableTime(context.Background(), 1, 10, CreateUnavailableTimeInput{
			UnavailableType: string(model.UnavailableTypeWeekly),
			DayOfWeek:       nil,
			StartTime:       "09:00",
			EndTime:         "10:00",
		})

		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("propagates error checking existing unavailable times", func(t *testing.T) {
		unavailableRepo := &mockUnavailableTimeRepository{
			findAllFn: func(_ context.Context, _, _ uint64) ([]model.ReservationTypeUnavailableTime, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewReservationTypeService(&mockReservationTypeRepository{}, unavailableRepo, &mockReservationTypeOccupationRepository{}, &mockOccupationRepository{}, nil)

		_, err := svc.CreateUnavailableTime(context.Background(), 1, 10, CreateUnavailableTimeInput{
			UnavailableType: string(model.UnavailableTypeWeekly),
			DayOfWeek:       &dayOfWeek,
			StartTime:       "09:00",
			EndTime:         "10:00",
		})

		assert.Error(t, err)
	})

	t.Run("returns conflict when overlapping with existing unavailable time", func(t *testing.T) {
		unavailableRepo := &mockUnavailableTimeRepository{
			findAllFn: func(_ context.Context, _, _ uint64) ([]model.ReservationTypeUnavailableTime, error) {
				return []model.ReservationTypeUnavailableTime{
					{UnavailableType: model.UnavailableTypeWeekly, DayOfWeek: &dayOfWeek, StartTime: "09:30", EndTime: "10:30"},
				}, nil
			},
		}
		svc := NewReservationTypeService(&mockReservationTypeRepository{}, unavailableRepo, &mockReservationTypeOccupationRepository{}, &mockOccupationRepository{}, nil)

		_, err := svc.CreateUnavailableTime(context.Background(), 1, 10, CreateUnavailableTimeInput{
			UnavailableType: string(model.UnavailableTypeWeekly),
			DayOfWeek:       &dayOfWeek,
			StartTime:       "09:00",
			EndTime:         "10:00",
		})

		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
	})

	t.Run("propagates repository create error", func(t *testing.T) {
		unavailableRepo := &mockUnavailableTimeRepository{
			createFn: func(_ context.Context, _ *model.ReservationTypeUnavailableTime) error {
				return errors.New("db error")
			},
		}
		svc := NewReservationTypeService(&mockReservationTypeRepository{}, unavailableRepo, &mockReservationTypeOccupationRepository{}, &mockOccupationRepository{}, nil)

		_, err := svc.CreateUnavailableTime(context.Background(), 1, 10, CreateUnavailableTimeInput{
			UnavailableType: string(model.UnavailableTypeWeekly),
			DayOfWeek:       &dayOfWeek,
			StartTime:       "09:00",
			EndTime:         "10:00",
		})

		assert.Error(t, err)
	})
}

// ---- validateUnavailableTimeInput ----

func TestValidateUnavailableTimeInput(t *testing.T) {
	dayOfWeek := int8(3)
	invalidDayOfWeek := int8(7)
	negativeDayOfWeek := int8(-1)
	specificDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		input   CreateUnavailableTimeInput
		wantErr bool
	}{
		{
			name: "valid weekly input",
			input: CreateUnavailableTimeInput{
				UnavailableType: string(model.UnavailableTypeWeekly),
				DayOfWeek:       &dayOfWeek,
				StartTime:       "09:00",
				EndTime:         "10:00",
			},
			wantErr: false,
		},
		{
			name: "weekly missing day_of_week",
			input: CreateUnavailableTimeInput{
				UnavailableType: string(model.UnavailableTypeWeekly),
				DayOfWeek:       nil,
				StartTime:       "09:00",
				EndTime:         "10:00",
			},
			wantErr: true,
		},
		{
			name: "weekly day_of_week above range",
			input: CreateUnavailableTimeInput{
				UnavailableType: string(model.UnavailableTypeWeekly),
				DayOfWeek:       &invalidDayOfWeek,
				StartTime:       "09:00",
				EndTime:         "10:00",
			},
			wantErr: true,
		},
		{
			name: "weekly day_of_week below range",
			input: CreateUnavailableTimeInput{
				UnavailableType: string(model.UnavailableTypeWeekly),
				DayOfWeek:       &negativeDayOfWeek,
				StartTime:       "09:00",
				EndTime:         "10:00",
			},
			wantErr: true,
		},
		{
			name: "valid specific input",
			input: CreateUnavailableTimeInput{
				UnavailableType: string(model.UnavailableTypeSpecific),
				SpecificDate:    &specificDate,
				StartTime:       "09:00",
				EndTime:         "10:00",
			},
			wantErr: false,
		},
		{
			name: "specific missing specific_date",
			input: CreateUnavailableTimeInput{
				UnavailableType: string(model.UnavailableTypeSpecific),
				SpecificDate:    nil,
				StartTime:       "09:00",
				EndTime:         "10:00",
			},
			wantErr: true,
		},
		{
			name: "unknown unavailable type",
			input: CreateUnavailableTimeInput{
				UnavailableType: "bogus",
				StartTime:       "09:00",
				EndTime:         "10:00",
			},
			wantErr: true,
		},
		{
			name: "empty unavailable type (zero value)",
			input: CreateUnavailableTimeInput{
				StartTime: "09:00",
				EndTime:   "10:00",
			},
			wantErr: true,
		},
		{
			name: "start_time equal to end_time",
			input: CreateUnavailableTimeInput{
				UnavailableType: string(model.UnavailableTypeWeekly),
				DayOfWeek:       &dayOfWeek,
				StartTime:       "09:00",
				EndTime:         "09:00",
			},
			wantErr: true,
		},
		{
			name: "start_time after end_time",
			input: CreateUnavailableTimeInput{
				UnavailableType: string(model.UnavailableTypeWeekly),
				DayOfWeek:       &dayOfWeek,
				StartTime:       "10:00",
				EndTime:         "09:00",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUnavailableTimeInput(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, apperrors.IsInvalidInput(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---- validateUnavailableTimeNotOverlaps ----

func TestValidateUnavailableTimeNotOverlaps(t *testing.T) {
	day1 := int8(1)
	day2 := int8(2)
	date1 := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2026, 1, 16, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		existing []model.ReservationTypeUnavailableTime
		input    CreateUnavailableTimeInput
		wantErr  bool
	}{
		{
			name:     "no existing entries",
			existing: nil,
			input: CreateUnavailableTimeInput{
				UnavailableType: string(model.UnavailableTypeWeekly),
				DayOfWeek:       &day1,
				StartTime:       "09:00",
				EndTime:         "10:00",
			},
			wantErr: false,
		},
		{
			name: "different unavailable type is skipped",
			existing: []model.ReservationTypeUnavailableTime{
				{UnavailableType: model.UnavailableTypeSpecific, SpecificDate: &date1, StartTime: "09:00", EndTime: "10:00"},
			},
			input: CreateUnavailableTimeInput{
				UnavailableType: string(model.UnavailableTypeWeekly),
				DayOfWeek:       &day1,
				StartTime:       "09:00",
				EndTime:         "10:00",
			},
			wantErr: false,
		},
		{
			name: "weekly different day of week does not overlap",
			existing: []model.ReservationTypeUnavailableTime{
				{UnavailableType: model.UnavailableTypeWeekly, DayOfWeek: &day2, StartTime: "09:00", EndTime: "10:00"},
			},
			input: CreateUnavailableTimeInput{
				UnavailableType: string(model.UnavailableTypeWeekly),
				DayOfWeek:       &day1,
				StartTime:       "09:00",
				EndTime:         "10:00",
			},
			wantErr: false,
		},
		{
			name: "weekly same day non-overlapping time ranges",
			existing: []model.ReservationTypeUnavailableTime{
				{UnavailableType: model.UnavailableTypeWeekly, DayOfWeek: &day1, StartTime: "11:00", EndTime: "12:00"},
			},
			input: CreateUnavailableTimeInput{
				UnavailableType: string(model.UnavailableTypeWeekly),
				DayOfWeek:       &day1,
				StartTime:       "09:00",
				EndTime:         "10:00",
			},
			wantErr: false,
		},
		{
			name: "weekly same day overlapping time ranges conflict",
			existing: []model.ReservationTypeUnavailableTime{
				{UnavailableType: model.UnavailableTypeWeekly, DayOfWeek: &day1, StartTime: "09:30", EndTime: "10:30"},
			},
			input: CreateUnavailableTimeInput{
				UnavailableType: string(model.UnavailableTypeWeekly),
				DayOfWeek:       &day1,
				StartTime:       "09:00",
				EndTime:         "10:00",
			},
			wantErr: true,
		},
		{
			name: "weekly nil day_of_week on input is skipped",
			existing: []model.ReservationTypeUnavailableTime{
				{UnavailableType: model.UnavailableTypeWeekly, DayOfWeek: &day1, StartTime: "09:00", EndTime: "10:00"},
			},
			input: CreateUnavailableTimeInput{
				UnavailableType: string(model.UnavailableTypeWeekly),
				DayOfWeek:       nil,
				StartTime:       "09:00",
				EndTime:         "10:00",
			},
			wantErr: false,
		},
		{
			name: "weekly nil day_of_week on existing is skipped",
			existing: []model.ReservationTypeUnavailableTime{
				{UnavailableType: model.UnavailableTypeWeekly, DayOfWeek: nil, StartTime: "09:00", EndTime: "10:00"},
			},
			input: CreateUnavailableTimeInput{
				UnavailableType: string(model.UnavailableTypeWeekly),
				DayOfWeek:       &day1,
				StartTime:       "09:00",
				EndTime:         "10:00",
			},
			wantErr: false,
		},
		{
			name: "specific different date does not overlap",
			existing: []model.ReservationTypeUnavailableTime{
				{UnavailableType: model.UnavailableTypeSpecific, SpecificDate: &date2, StartTime: "09:00", EndTime: "10:00"},
			},
			input: CreateUnavailableTimeInput{
				UnavailableType: string(model.UnavailableTypeSpecific),
				SpecificDate:    &date1,
				StartTime:       "09:00",
				EndTime:         "10:00",
			},
			wantErr: false,
		},
		{
			name: "specific same date overlapping time ranges conflict",
			existing: []model.ReservationTypeUnavailableTime{
				{UnavailableType: model.UnavailableTypeSpecific, SpecificDate: &date1, StartTime: "09:30", EndTime: "10:30"},
			},
			input: CreateUnavailableTimeInput{
				UnavailableType: string(model.UnavailableTypeSpecific),
				SpecificDate:    &date1,
				StartTime:       "09:00",
				EndTime:         "10:00",
			},
			wantErr: true,
		},
		{
			name: "specific nil specific_date on input is skipped",
			existing: []model.ReservationTypeUnavailableTime{
				{UnavailableType: model.UnavailableTypeSpecific, SpecificDate: &date1, StartTime: "09:00", EndTime: "10:00"},
			},
			input: CreateUnavailableTimeInput{
				UnavailableType: string(model.UnavailableTypeSpecific),
				SpecificDate:    nil,
				StartTime:       "09:00",
				EndTime:         "10:00",
			},
			wantErr: false,
		},
		{
			name: "specific nil specific_date on existing is skipped",
			existing: []model.ReservationTypeUnavailableTime{
				{UnavailableType: model.UnavailableTypeSpecific, SpecificDate: nil, StartTime: "09:00", EndTime: "10:00"},
			},
			input: CreateUnavailableTimeInput{
				UnavailableType: string(model.UnavailableTypeSpecific),
				SpecificDate:    &date1,
				StartTime:       "09:00",
				EndTime:         "10:00",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUnavailableTimeNotOverlaps(tt.existing, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, apperrors.IsConflict(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
