package reservation

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

func ValidateReservationTypeAvailableTime(
	ctx context.Context,
	unavailableRepo ReservationTypeUnavailableTimeRepository,
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

// filterApplicableUnavailableTimes は date に適用される不可時間帯を返す。
// 優先順位: specific > weekly（特定日設定が曜日設定を上書き）
// （BE9-2C R①: service/liff_service_availability_filters.go の意図的複製。
// liff 系(R⑤)移動時に単一実装へ自己解消する）
func filterApplicableUnavailableTimes(times []model.ReservationTypeUnavailableTime, date time.Time) []model.ReservationTypeUnavailableTime {
	dateStr := date.In(config.JST).Format(time.DateOnly)
	var specific, weekly []model.ReservationTypeUnavailableTime
	for i := range times {
		switch times[i].UnavailableType {
		case model.UnavailableTypeSpecific:
			if times[i].SpecificDate != nil && times[i].SpecificDate.In(config.JST).Format(time.DateOnly) == dateStr {
				specific = append(specific, times[i])
			}
		case model.UnavailableTypeWeekly:
			if times[i].DayOfWeek != nil && int(*times[i].DayOfWeek) == int(date.In(config.JST).Weekday()) {
				weekly = append(weekly, times[i])
			}
		}
	}
	if len(specific) > 0 {
		return specific
	}
	return weekly
}
