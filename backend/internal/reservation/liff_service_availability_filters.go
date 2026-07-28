package reservation

import (
	"time"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// filterApplicableUnavailableTimes は date に適用される不可時間帯を返す。
// 優先順位: specific > weekly（特定日設定が曜日設定を上書き）
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
