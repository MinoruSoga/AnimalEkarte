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

type mockReservationTypeFinder struct {
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error)
}

func (m mockReservationTypeFinder) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func TestCheckReservationTypeCapacity(t *testing.T) {
	start := time.Date(2026, 6, 1, 9, 0, 0, 0, config.JST)
	maxConcurrent := 2

	t.Run("上限未満ならnilを返す", func(t *testing.T) {
		err := CheckReservationTypeCapacity(
			context.Background(),
			mockCapacityCounter{countFn: func(_ context.Context, _, _ uint64, _ time.Time, _ *uint64) (int64, error) {
				return 1, nil
			}},
			mockReservationTypeFinder{findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ReservationType, error) {
				return &model.ReservationType{ID: id, ClinicID: clinicID, MaxConcurrent: &maxConcurrent}, nil
			}},
			1,
			2,
			start,
			nil,
		)

		assert.NoError(t, err)
	})

	t.Run("上限以上ならconflictを返す", func(t *testing.T) {
		err := CheckReservationTypeCapacity(
			context.Background(),
			mockCapacityCounter{countFn: func(_ context.Context, _, _ uint64, _ time.Time, _ *uint64) (int64, error) {
				return 2, nil
			}},
			mockReservationTypeFinder{findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ReservationType, error) {
				return &model.ReservationType{ID: id, ClinicID: clinicID, MaxConcurrent: &maxConcurrent}, nil
			}},
			1,
			2,
			start,
			nil,
		)

		assert.ErrorIs(t, err, apperrors.ErrConflict)
	})

	t.Run("区分取得エラーはwrapする", func(t *testing.T) {
		baseErr := errors.New("db error")
		err := CheckReservationTypeCapacity(
			context.Background(),
			mockCapacityCounter{countFn: func(_ context.Context, _, _ uint64, _ time.Time, _ *uint64) (int64, error) {
				return 0, nil
			}},
			mockReservationTypeFinder{findByIDFn: func(_ context.Context, _ uint64, _ uint64) (*model.ReservationType, error) {
				return nil, baseErr
			}},
			1,
			2,
			start,
			nil,
		)

		assert.ErrorIs(t, err, baseErr)
	})
}
