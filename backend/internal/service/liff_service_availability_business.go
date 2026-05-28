package service

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// parseBusinessHoursForDate はパッケージ内共通のヘルパー（liffService・validator から呼ばれる）。
func parseBusinessHoursForDate(setting *model.LineReservationSetting, date time.Time) (BusinessHours, []BreakPeriod) {
	var bh BusinessHours
	if err := json.Unmarshal(setting.BusinessHours, &bh); err != nil {
		bh = BusinessHours{Start: "0900", End: "1900"}
	}
	var breaks []BreakPeriod
	if err := json.Unmarshal(setting.BreakHours, &breaks); err != nil {
		breaks = nil
	}

	// 曜日別営業時間があれば上書き（例: 土曜だけ短縮営業）
	if len(setting.BusinessHoursByWeekday) > 0 {
		var byWeekday map[string]BusinessHours
		if err := json.Unmarshal(setting.BusinessHoursByWeekday, &byWeekday); err == nil {
			key := strconv.Itoa(int(date.In(jstLocation).Weekday()))
			if wdBH, ok := byWeekday[key]; ok {
				bh = wdBH
			}
		}
	}

	return bh, breaks
}
