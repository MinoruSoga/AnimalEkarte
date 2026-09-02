package reservation

import (
	"context"
	"encoding/json"
	"time"

	holiday "github.com/holiday-jp/holiday_jp-go"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// AvailableDateResult は1日分の空き状況を表す。
type AvailableDateResult struct {
	Date      string // time.DateOnly
	Weekday   int    // 0=Sun, 1=Mon, ..., 6=Sat
	Available bool
	Reason    string // "closed" | "holiday" | "staff_off" | "no_slots" | ""
}

// BookingWindow は予約受付期間を表す。
type BookingWindow struct {
	Start string // time.DateOnly
	End   string // time.DateOnly
}

// AvailableDatesInput は空き日付計算の入力。
type AvailableDatesInput struct {
	Settings       AvailableDatesSettings
	TypeID         uint64
	StaffID        uint64 // 0 = 指名なし
	StaffInputsFn  func(ctx context.Context, date time.Time, typeID, staffID uint64) ([]StaffSlotInput, error)
	SlotSettingsFn func(date time.Time) TimeSlotsInput
	SlotFilterFn   func(date time.Time, slots []TimeSlot) []TimeSlot
}

// AvailableDatesSettings は空き日付計算に必要な設定項目。
type AvailableDatesSettings struct {
	ClosedWeekdays        []int    // 0=Sun,...,6=Sat
	ClosedDates           []string // time.DateOnly
	NationalHolidayClosed bool
	BookingWindowMinDays  int
	BookingWindowMaxDays  int
	CalendarMonths        int
	ReservationDayOption  string // "none" | "saturday" | "weekday" | "anyday"
}

// ParseAvailableDatesSettings は LineReservationSetting の JSONB フィールドから設定を解析する。
func ParseAvailableDatesSettings(
	closedWeekdaysJSON []byte,
	closedDatesJSON []byte,
	nationalHolidayClosed bool,
	bookingWindowMinDays, bookingWindowMaxDays, calendarMonths int,
	reservationDayOption string,
) (AvailableDatesSettings, error) {
	var closedWeekdays []int
	if err := json.Unmarshal(orEmptyJSONArray(closedWeekdaysJSON), &closedWeekdays); err != nil {
		return AvailableDatesSettings{}, apperrors.WrapInvalidInput("休診設定の解析に失敗しました: " + err.Error())
	}
	var closedDates []string
	if err := json.Unmarshal(orEmptyJSONArray(closedDatesJSON), &closedDates); err != nil {
		return AvailableDatesSettings{}, apperrors.WrapInvalidInput("休診設定の解析に失敗しました: " + err.Error())
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

// orEmptyJSONArray は nil または空スライスを JSON の空配列 "[]" に変換する。
func orEmptyJSONArray(b []byte) []byte {
	if len(b) == 0 {
		return []byte("[]")
	}
	return b
}

// maxAvailableDatesResultCap is the upper bound for the results slice capacity.
// Matches binding max on booking_window_max_days (line_reservation_setting_request).
const maxAvailableDatesResultCap = 366

// availableDatesResultCap clamps BookingWindowMaxDays for make([]T, 0, cap).
// Negative values become 0 (avoids makeslice panic); values above the bind max are capped.
func availableDatesResultCap(bookingWindowMaxDays int) int {
	if bookingWindowMaxDays < 0 {
		return 0
	}
	if bookingWindowMaxDays > maxAvailableDatesResultCap {
		return maxAvailableDatesResultCap
	}
	return bookingWindowMaxDays
}

// BookingWindowDates は BookingWindowMinDays/MaxDays から予約受付期間 [minDate, maxDate]（JST日付・時刻0時）を計算する。
// CalcAvailableDates と GetAvailableDates のプリフェッチ経路が同一の窓計算を共有するために抽出した（G7-1）。
func BookingWindowDates(settings AvailableDatesSettings) (minDate, maxDate time.Time) {
	now := time.Now().In(config.JST)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, config.JST)
	minDays := settings.BookingWindowMinDays
	maxDays := settings.BookingWindowMaxDays
	if minDays < 0 {
		minDays = 0
	}
	if maxDays < 0 {
		maxDays = 0
	}
	if maxDays > maxAvailableDatesResultCap {
		maxDays = maxAvailableDatesResultCap
	}
	if minDays > maxDays {
		minDays = maxDays
	}
	minDate = today.AddDate(0, 0, minDays)
	maxDate = today.AddDate(0, 0, maxDays)
	return minDate, maxDate
}

// CalcAvailableDates は予約可能な日付一覧を計算して返す。
func CalcAvailableDates(ctx context.Context, input *AvailableDatesInput) ([]AvailableDateResult, BookingWindow, error) {
	minDate, maxDate := BookingWindowDates(input.Settings)

	window := BookingWindow{
		Start: minDate.Format(time.DateOnly),
		End:   maxDate.Format(time.DateOnly),
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

	// Defensive cap: binding rejects negative/huge values on write (RSV-04), but
	// previously-persisted bad settings must not panic makeslice or OOM a request.
	results := make([]AvailableDateResult, 0, availableDatesResultCap(input.Settings.BookingWindowMaxDays))

	for d := minDate; !d.After(maxDate); d = d.AddDate(0, 0, 1) {
		result, err := evaluateAvailableDate(ctx, d, input, closedDateSet, closedWeekdaySet)
		if err != nil {
			return nil, window, err
		}
		results = append(results, result)
	}

	return results, window, nil
}

func evaluateAvailableDate(
	ctx context.Context,
	d time.Time,
	input *AvailableDatesInput,
	closedDateSet map[string]struct{},
	closedWeekdaySet map[int]struct{},
) (AvailableDateResult, error) {
	dateStr := d.Format(time.DateOnly)
	wd := int(d.Weekday()) // 0=Sun,...,6=Sat

	result := AvailableDateResult{
		Date:    dateStr,
		Weekday: wd,
	}

	// 休業曜日チェック
	if _, closed := closedWeekdaySet[wd]; closed {
		result.Available = false
		result.Reason = "closed"
		return result, nil
	}

	// 休業日チェック
	if _, closed := closedDateSet[dateStr]; closed {
		result.Available = false
		result.Reason = "closed"
		return result, nil
	}

	// 祝日チェック
	if input.Settings.NationalHolidayClosed && holiday.IsHoliday(d) {
		result.Available = false
		result.Reason = "holiday"
		return result, nil
	}

	// コースの曜日オプションチェック
	if !checkDayOption(input.Settings.ReservationDayOption, wd) {
		result.Available = false
		result.Reason = "closed"
		return result, nil
	}

	// スタッフ個人設定・時間枠チェック
	if input.StaffInputsFn != nil && input.SlotSettingsFn != nil {
		staffInputs, err := input.StaffInputsFn(ctx, d, input.TypeID, input.StaffID)
		if err != nil {
			return result, err
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
			return result, nil
		}

		// 時間枠が1つ以上あるかチェック
		slotInput := input.SlotSettingsFn(d)
		slotInput.Staffs = staffInputs
		slots, err := GenerateTimeSlots(&slotInput)
		if err != nil {
			return result, err
		}
		if input.SlotFilterFn != nil {
			slots = input.SlotFilterFn(d, slots)
		}
		if len(slots) == 0 {
			result.Available = false
			result.Reason = "no_slots"
			return result, nil
		}
	}

	result.Available = true
	return result, nil
}

// checkDayOption はコースの曜日オプションに対して予約可能かチェックする。
func checkDayOption(option string, weekday int) bool {
	switch model.ReservationDayOption(option) {
	case model.DayOptionSaturday:
		return weekday == 6 // 土曜のみ
	case model.DayOptionWeekday:
		return weekday >= 1 && weekday <= 5 // 月〜金
	case model.DayOptionAnyday:
		return true
	default: // "none"
		return true
	}
}
