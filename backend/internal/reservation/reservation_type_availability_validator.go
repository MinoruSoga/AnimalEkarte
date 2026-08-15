package reservation

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// ReservationTypeUnavailableTimeView is the read-only capability required by availability validation.
type ReservationTypeUnavailableTimeView interface {
	FindAll(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeUnavailableTime, error)
}

func ValidateReservationTypeAvailableTime(
	ctx context.Context,
	unavailableRepo ReservationTypeUnavailableTimeView,
	clinicID, reservationTypeID uint64,
	start, end time.Time,
) error {
	if reservationTypeID == 0 {
		return nil
	}
	if err := sharedkernel.ValidateTimeRange(start, end); err != nil {
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
