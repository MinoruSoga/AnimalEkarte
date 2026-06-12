package service

import (
	"context"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type reservationTypeFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error)
}

type reservationTypeCapacityCounter interface {
	CountByTypeAndStartTime(ctx context.Context, clinicID, reservationTypeID uint64, startTime time.Time, excludeID *uint64) (int64, error)
}

func checkReservationTypeCapacity(
	ctx context.Context,
	reservationRepo reservationTypeCapacityCounter,
	typeFinder reservationTypeFinder,
	clinicID, reservationTypeID uint64,
	startTime time.Time,
	excludeID *uint64,
) error {
	if reservationRepo == nil || typeFinder == nil {
		return nil
	}
	reservationType, err := typeFinder.FindByID(ctx, clinicID, reservationTypeID)
	if err != nil {
		return apperrors.Wrap(err, "failed to find reservation type for capacity")
	}
	if reservationType.MaxConcurrent == nil {
		return nil
	}
	count, err := reservationRepo.CountByTypeAndStartTime(ctx, clinicID, reservationTypeID, startTime, excludeID)
	if err != nil {
		return apperrors.Wrap(err, "failed to count reservations by type and start time")
	}
	if count >= int64(*reservationType.MaxConcurrent) {
		return apperrors.WrapConflict("この時間帯の予約受入枠が満員です")
	}
	return nil
}
