package service

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/config"
	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/repository"
)

func validateReservationTypeAvailableTime(
	ctx context.Context,
	unavailableRepo repository.ReservationTypeUnavailableTimeRepository,
	clinicID, reservationTypeID uint64,
	start, end time.Time,
) error {
	if reservationTypeID == 0 {
		return nil
	}
	if err := validateTimeRange(start, end); err != nil {
		return err
	}
	startTime := start.In(config.JST).Format("15:04")
	endTime := end.In(config.JST).Format("15:04")

	if unavailableRepo != nil {
		unavailableTimes, err := unavailableRepo.FindAll(ctx, clinicID, reservationTypeID)
		if err != nil {
			return apperrors.Wrap(err, "failed to get unavailable reservation type times")
		}
		applicable := filterApplicableUnavailableTimes(unavailableTimes, start)
		for i := range applicable {
			if startTime < applicable[i].EndTime && endTime > applicable[i].StartTime {
				return apperrors.WrapInvalidInput("選択した時間はこの予約区分の予約不可時間に含まれています")
			}
		}
	}
	return nil
}
