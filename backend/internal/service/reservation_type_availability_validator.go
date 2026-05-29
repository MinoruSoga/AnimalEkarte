package service

import (
	"context"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/repository"
)

func validateReservationTypeAvailableTime(ctx context.Context, repo repository.ReservationTypeUnavailableTimeRepository, clinicID, reservationTypeID uint64, start, end time.Time) error {
	if repo == nil || reservationTypeID == 0 {
		return nil
	}
	if err := validateTimeRange(start, end); err != nil {
		return err
	}
	unavailableTimes, err := repo.FindAll(ctx, clinicID, reservationTypeID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get unavailable reservation type times")
	}
	applicable := filterApplicableUnavailableTimes(unavailableTimes, start)
	startTime := start.In(jstLocation).Format("15:04")
	endTime := end.In(jstLocation).Format("15:04")
	for i := range applicable {
		if startTime < applicable[i].EndTime && endTime > applicable[i].StartTime {
			return apperrors.WrapInvalidInput("選択した時間はこの予約区分の予約不可時間に含まれています")
		}
	}
	return nil
}
