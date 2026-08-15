package reservation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestValidateReservationTypeAvailableTime(t *testing.T) {
	// 2026-05-04 is a Monday
	start := time.Date(2026, 5, 4, 10, 0, 0, 0, config.JST)
	end := time.Date(2026, 5, 4, 10, 30, 0, 0, config.JST)
	dayOfWeekMonday := int8(1)

	t.Run("reservationTypeID zero: noop", func(t *testing.T) {
		err := ValidateReservationTypeAvailableTime(context.Background(), &mockUnavailableTimeRepository{}, 1, 0, start, end)
		assert.NoError(t, err)
	})

	t.Run("invalid time range (end before start): error", func(t *testing.T) {
		err := ValidateReservationTypeAvailableTime(context.Background(), &mockUnavailableTimeRepository{}, 1, 5, end, start)
		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("nil unavailableRepo: skips the unavailable-time check", func(t *testing.T) {
		err := ValidateReservationTypeAvailableTime(context.Background(), nil, 1, 5, start, end)
		assert.NoError(t, err)
	})

	t.Run("repository lookup fails: wrapped error", func(t *testing.T) {
		repo := &mockUnavailableTimeRepository{
			findAllFn: func(_ context.Context, _, _ uint64) ([]model.ReservationTypeUnavailableTime, error) {
				return nil, errors.New("db error")
			},
		}
		err := ValidateReservationTypeAvailableTime(context.Background(), repo, 1, 5, start, end)
		assert.Error(t, err)
		assert.False(t, apperrors.IsInvalidInput(err))
	})

	t.Run("no unavailable times: success", func(t *testing.T) {
		repo := &mockUnavailableTimeRepository{
			findAllFn: func(_ context.Context, _, _ uint64) ([]model.ReservationTypeUnavailableTime, error) {
				return nil, nil
			},
		}
		err := ValidateReservationTypeAvailableTime(context.Background(), repo, 1, 5, start, end)
		assert.NoError(t, err)
	})

	t.Run("weekly unavailable time overlapping requested range: invalid input", func(t *testing.T) {
		repo := &mockUnavailableTimeRepository{
			findAllFn: func(_ context.Context, _, _ uint64) ([]model.ReservationTypeUnavailableTime, error) {
				return []model.ReservationTypeUnavailableTime{
					{
						UnavailableType: model.UnavailableTypeWeekly,
						DayOfWeek:       &dayOfWeekMonday,
						StartTime:       "09:00",
						EndTime:         "12:00",
					},
				}, nil
			},
		}
		err := ValidateReservationTypeAvailableTime(context.Background(), repo, 1, 5, start, end)
		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("weekly unavailable time not overlapping requested range: success", func(t *testing.T) {
		repo := &mockUnavailableTimeRepository{
			findAllFn: func(_ context.Context, _, _ uint64) ([]model.ReservationTypeUnavailableTime, error) {
				return []model.ReservationTypeUnavailableTime{
					{
						UnavailableType: model.UnavailableTypeWeekly,
						DayOfWeek:       &dayOfWeekMonday,
						StartTime:       "13:00",
						EndTime:         "14:00",
					},
				}, nil
			},
		}
		err := ValidateReservationTypeAvailableTime(context.Background(), repo, 1, 5, start, end)
		assert.NoError(t, err)
	})

	t.Run("specific-date unavailable time overlapping requested range: invalid input", func(t *testing.T) {
		specificDate := time.Date(2026, 5, 4, 0, 0, 0, 0, config.JST)
		repo := &mockUnavailableTimeRepository{
			findAllFn: func(_ context.Context, _, _ uint64) ([]model.ReservationTypeUnavailableTime, error) {
				return []model.ReservationTypeUnavailableTime{
					{
						UnavailableType: model.UnavailableTypeSpecific,
						SpecificDate:    &specificDate,
						StartTime:       "09:00",
						EndTime:         "12:00",
					},
				}, nil
			},
		}
		err := ValidateReservationTypeAvailableTime(context.Background(), repo, 1, 5, start, end)
		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})
}
