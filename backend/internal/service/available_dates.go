package service

import (
	"context"
	"encoding/json"
	"time"

	holiday "github.com/holiday-jp/holiday_jp-go"
)

// AvailableDateResult は1日分の空き状況を表す。
type AvailableDateResult struct {
	Date      string // "2006-01-02"
	Weekday   int    // 0=Sun, 1=Mon, ..., 6=Sat
	Available bool
	Reason    string // "closed" | "holiday" | "staff_off" | "no_slots" | ""
}

// BookingWindow は予約受付期間を表す。
type BookingWindow struct {
	Start string // "2006-01-02"
	End   string // "2006-01-02"
}

// AvailableDatesInput は空き日付計算の入力。
type AvailableDatesInput struct {
	Settings       AvailableDatesSettings
	CourseID       uint64
	StaffID        uint64 // 0 = 指名なし
	StaffInputsFn  func(ctx context.Context, date time.Time, courseID, staffID uint64) ([]StaffSlotInput, error)
	SlotSettingsFn func() TimeSlotsInput
}

// AvailableDatesSettings は空き日付計算に必要な設定項目。
type AvailableDatesSettings struct {
	ClosedWeekdays        []int    // 0=Sun,...,6=Sat
	ClosedDates           []string // "2006-01-02"
	NationalHolidayClosed bool
	BookingWindowMinDays  int
	BookingWindowMaxDays  int
	CalendarMonths        int
	ReservationDayOption  string // "none" | "saturday" | "weekday" | "anyday"
}

// ParseAvailableDatesSettings は ReservationSetting の JSONB フィールドから設定を解析する。
func ParseAvailableDatesSettings(
	closedWeekdaysJSON []byte,
	closedDatesJSON []byte,
	nationalHolidayClosed bool,
	bookingWindowMinDays, bookingWindowMaxDays, calendarMonths int,
	reservationDayOption string,
) (AvailableDatesSettings, error) {
	var closedWeekdays []int
	if err := json.Unmarshal(closedWeekdays0(closedWeekdaysJSON), &closedWeekdays); err != nil {
		closedWeekdays = nil
	}
	var closedDates []string
	if err := json.Unmarshal(closedDates0(closedDatesJSON), &closedDates); err != nil {
		closedDates = nil
	}
	return AvailableDatesSettings{
		ClosedWeekdays:        closedWeekdays,
		ClosedDates:           closedDates,
		NationalHolidayClosed: nationalHolidayClosed,
		BookingWindowMinDays:  bookingWindowMinDays,
		BookingWindowMaxDays:  bookingWindowMaxDays,
		CalendarMonths:        calendarMonths,
		ReservationDayOption:  reservationDayOption,
	}, nil
}

func closedWeekdays0(b []byte) []byte {
	if len(b) == 0 {
		return []byte("[]")
	}
	return b
}

func closedDates0(b []byte) []byte {
	if len(b) == 0 {
		return []byte("[]")
	}
	return b
}

// CalcAvailableDates は予約可能な日付一覧を計算して返す。
func CalcAvailableDates(ctx context.Context, input AvailableDatesInput) ([]AvailableDateResult, BookingWindow, error) {
	now := time.Now().In(jstLocation())
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, jstLocation())

	minDate := today.AddDate(0, 0, input.Settings.BookingWindowMinDays)
	maxDate := today.AddDate(0, 0, input.Settings.BookingWindowMaxDays)

	window := BookingWindow{
		Start: minDate.Format("2006-01-02"),
		End:   maxDate.Format("2006-01-02"),
	}

	// 休業日セット
	closedDateSet := make(map[string]struct{}, len(input.Settings.ClosedDates))
	for _, d := range input.Settings.ClosedDates {
		closedDateSet[d] = struct{}{}
	}
	// 休業曜日セット
	closedWeekdaySet := make(map[int]struct{}, len(input.Settings.ClosedWeekdays))
	for _, w := range input.Settings.ClosedWeekdays {
		closedWeekdaySet[w] = struct{}{}
	}

	results := make([]AvailableDateResult, 0, input.Settings.BookingWindowMaxDays)

	for d := minDate; !d.After(maxDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		wd := int(d.Weekday()) // 0=Sun,...,6=Sat

		result := AvailableDateResult{
			Date:    dateStr,
			Weekday: wd,
		}

		// 休業曜日チェック
		if _, closed := closedWeekdaySet[wd]; closed {
			result.Available = false
			result.Reason = "closed"
			results = append(results, result)
			continue
		}

		// 休業日チェック
		if _, closed := closedDateSet[dateStr]; closed {
			result.Available = false
			result.Reason = "closed"
			results = append(results, result)
			continue
		}

		// 祝日チェック
		if input.Settings.NationalHolidayClosed && isJapaneseHoliday(d) {
			result.Available = false
			result.Reason = "holiday"
			results = append(results, result)
			continue
		}

		// コースの曜日オプションチェック
		if !checkDayOption(input.Settings.ReservationDayOption, wd) {
			result.Available = false
			result.Reason = "closed"
			results = append(results, result)
			continue
		}

		// スタッフ個人設定・時間枠チェック
		if input.StaffInputsFn != nil && input.SlotSettingsFn != nil {
			staffInputs, err := input.StaffInputsFn(ctx, d, input.CourseID, input.StaffID)
			if err != nil {
				return nil, window, err
			}

			// 全スタッフが休日かチェック
			allOff := true
			for _, si := range staffInputs {
				if si.ScheduleOverride == nil {
					allOff = false
					break
				}
				if si.ScheduleOverride.ShiftType != "off" && si.ScheduleOverride.ShiftType != "paid_leave" {
					allOff = false
					break
				}
			}
			if allOff && len(staffInputs) > 0 {
				result.Available = false
				result.Reason = "staff_off"
				results = append(results, result)
				continue
			}

			// 時間枠が1つ以上あるかチェック
			slotInput := input.SlotSettingsFn()
			slotInput.Staffs = staffInputs
			slots, err := GenerateTimeSlots(slotInput)
			if err != nil {
				return nil, window, err
			}
			if len(slots) == 0 {
				result.Available = false
				result.Reason = "no_slots"
				results = append(results, result)
				continue
			}
		}

		result.Available = true
		results = append(results, result)
	}

	return results, window, nil
}

// isJapaneseHoliday は指定日が日本の祝日かどうかを返す。
func isJapaneseHoliday(t time.Time) bool {
	return holiday.IsHoliday(t)
}

// checkDayOption はコースの曜日オプションに対して予約可能かチェックする。
func checkDayOption(option string, weekday int) bool {
	switch option {
	case "saturday":
		return weekday == 6 // 土曜のみ
	case "weekday":
		return weekday >= 1 && weekday <= 5 // 月〜金
	case "anyday":
		return true
	default: // "none"
		return true
	}
}

// jstLocation は JST タイムゾーンを返す。
func jstLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.FixedZone("JST", 9*60*60)
	}
	return loc
}
