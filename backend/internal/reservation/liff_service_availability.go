package reservation

import (
	"context"
	"log/slog"
	"strings"
	"time"

	holiday "github.com/holiday-jp/holiday_jp-go"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// buildCapacityFilterFn は course.MaxConcurrent 制約下でのキャパシティフィルタ
// (warn-and-fallback付き)を構築する（BE-refactor.md E-10）。base のスロットに対しキャパシティ
// 超過分を除外し、フィルタ失敗時は warn ログを出して base をそのまま返す（fail-open・既存挙動維持）。
func (s *liffService) buildCapacityFilterFn(ctx context.Context, clinicID, typeID uint64, maxConcurrent int) func(date time.Time, base []TimeSlot) []TimeSlot {
	return func(date time.Time, base []TimeSlot) []TimeSlot {
		filtered, err := FilterSlotsByCapacity(ctx, base, s.reservationRepo, clinicID, typeID, date, maxConcurrent)
		if err != nil {
			slog.WarnContext(ctx, "failed to filter by capacity in dates, skipping", "error", err)
			return base
		}
		return filtered
	}
}

// GetAvailableDates は予約可能な日付一覧を返す。
func (s *liffService) GetAvailableDates(ctx context.Context, clinicID, typeID, staffID uint64) ([]AvailableDateResult, BookingWindow, error) {
	setting, err := s.settingRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get reservation setting", "error", err)
		return nil, BookingWindow{}, apperrors.Wrap(err, "failed to get reservation setting")
	}
	course, err := s.findActiveLiffCourse(ctx, clinicID, typeID)
	if err != nil {
		return nil, BookingWindow{}, err
	}

	// スタッフを事前取得（クロージャで再利用）
	visibleStaffs, err := s.resolveTargetStaffs(ctx, clinicID, typeID, staffID)
	if err != nil {
		return nil, BookingWindow{}, err
	}

	datesSettings, err := ParseAvailableDatesSettings(
		setting.ClosedWeekdays,
		setting.ClosedDates,
		setting.NationalHolidayClosed,
		setting.BookingWindowMinDays,
		setting.BookingWindowMaxDays,
		setting.CalendarMonths,
		string(course.ReservationDayOption),
	)
	if err != nil {
		return nil, BookingWindow{}, err
	}

	// G7-1: 日付ループN+1解消 — シフト/休憩/当日予約を予約受付期間の範囲でまとめてプリフェッチする。
	staffInputsFn, err := s.buildAvailableDatesStaffInputsFn(ctx, clinicID, visibleStaffs, datesSettings)
	if err != nil {
		return nil, BookingWindow{}, err
	}

	slotSettingsFn := func(date time.Time) TimeSlotsInput {
		// 一覧表示パスは既存挙動を維持し break_hours 破損時のエラーを無視する（意図的・スコープ外。
		// parseBusinessHoursForDate のコメント参照。書込パスの fail-closed 化のみ D10/F-2 対象）。
		bh, defaultBreaks, _ := ParseBusinessHoursForDate(ctx, setting, date)
		return TimeSlotsInput{
			BusinessHours:     bh,
			DefaultBreaks:     defaultBreaks,
			CourseDuration:    course.DurationMinutes,
			IntervalMinutes:   setting.TimeSlotIntervalMinutes,
			Mode:              setting.TimeSlotMode,
			MinCourseDuration: course.DurationMinutes,
		}
	}
	// BE-refactor.md E-10: capacity フィルタ(warn-and-fallback付き)を1回だけ構築し、両分岐で使う。
	// FilterSlotsByCapacity 内部で日付ごとの全スロットを1クエリにバッチ化済み
	// （reservationTypeCapacityBatchCounter、R2-4/D8）。日付間の反復は CalcAvailableDates の
	// 制御下にあり残るが、支配的だったスロット数分の N+1 は解消。
	var capacityFilter func(date time.Time, base []TimeSlot) []TimeSlot
	if course.MaxConcurrent != nil {
		capacityFilter = s.buildCapacityFilterFn(ctx, clinicID, typeID, *course.MaxConcurrent)
	}

	slotFilterFn, err := s.availableDateSlotFilter(ctx, clinicID, typeID, course, capacityFilter)
	if err != nil {
		return nil, BookingWindow{}, err
	}

	results, window, err := CalcAvailableDates(ctx, &AvailableDatesInput{
		Settings:       datesSettings,
		TypeID:         typeID,
		StaffID:        staffID,
		StaffInputsFn:  staffInputsFn,
		SlotSettingsFn: slotSettingsFn,
		SlotFilterFn:   slotFilterFn,
	})
	if err != nil {
		return nil, window, err
	}

	if err := s.applyOccupationGuard(ctx, clinicID, typeID, results); err != nil {
		return nil, window, err
	}

	return results, window, nil
}

func (s *liffService) availableDateSlotFilter(
	ctx context.Context,
	clinicID, typeID uint64,
	course *model.ReservationType,
	capacityFilter func(date time.Time, base []TimeSlot) []TimeSlot,
) (func(date time.Time, slots []TimeSlot) []TimeSlot, error) {
	if s.availableSlotRepo != nil {
		availableSlots, err := s.availableSlotRepo.FindAll(ctx, clinicID, typeID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get available slots", "error", err)
			return nil, apperrors.Wrap(err, "failed to get available slots")
		}
		if HasActiveAvailableSlots(availableSlots) || course.MaxConcurrent != nil {
			return func(date time.Time, slots []TimeSlot) []TimeSlot {
				merged := MergeAvailableTimeSlots(slots, availableSlots, date, course.DurationMinutes)
				if capacityFilter == nil {
					return merged
				}
				return capacityFilter(date, merged)
			}, nil
		}
		return nil, nil
	}
	if capacityFilter != nil {
		return capacityFilter, nil
	}
	return nil, nil
}

// applyOccupationGuard は職種紐付けが1件以上ある場合のみ、対応職種のスタッフが出勤している日かを
// チェックして results を in-place で更新する（BE-117）。0件（職種紐付けなし）は後方互換で素通り。
// G7-1: 日毎のカウント呼出をバッチ版1クエリに集約する。
func (s *liffService) applyOccupationGuard(ctx context.Context, clinicID, typeID uint64, results []AvailableDateResult) error {
	occupations, err := s.occupationRepo.FindAll(ctx, clinicID, typeID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get occupation guard", "error", err)
		return apperrors.Wrap(err, "failed to get occupation guard")
	}
	if len(occupations) == 0 {
		return nil
	}

	availableDates := make([]time.Time, 0, len(results))
	for _, r := range results {
		if !r.Available {
			continue
		}
		date, err := time.ParseInLocation(time.DateOnly, r.Date, config.JST)
		if err != nil {
			slog.ErrorContext(ctx, "failed to parse date", "error", err)
			return apperrors.Wrap(err, "failed to parse date")
		}
		availableDates = append(availableDates, date)
	}

	workingCounts, err := s.occupationRepo.CountWorkingStaffByReservationTypeIDs(ctx, clinicID, typeID, availableDates)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count working staff", "error", err)
		return apperrors.Wrap(err, "failed to count working staff")
	}
	for i, r := range results {
		if !r.Available {
			continue
		}
		if workingCounts[r.Date] == 0 {
			results[i].Available = false
			results[i].Reason = "staff_off"
		}
	}
	return nil
}

// isDateClosed は単日の休業判定(closed_weekdays/closed_dates/national_holiday_closed)を行う
// 線形走査の純関数（BE-refactor.md E-10）。CalcAvailableDates の prebuilt set 方式とは統合しない
// （ループ内 set は意図的設計 — CalcAvailableDates は複数日を走査するため事前構築が有効だが、
// 本関数は単日呼出のため都度構築で十分）。
func isDateClosed(settings AvailableDatesSettings, dateJST time.Time) bool {
	dateStr := dateJST.Format(time.DateOnly)
	wd := int(dateJST.Weekday())
	closedWeekdaySet := make(map[int]struct{}, len(settings.ClosedWeekdays))
	for _, w := range settings.ClosedWeekdays {
		closedWeekdaySet[w] = struct{}{}
	}
	closedDateSet := make(map[string]struct{}, len(settings.ClosedDates))
	for _, d := range settings.ClosedDates {
		closedDateSet[d] = struct{}{}
	}
	if _, closed := closedWeekdaySet[wd]; closed {
		return true
	}
	if _, closed := closedDateSet[dateStr]; closed {
		return true
	}
	if settings.NationalHolidayClosed && holiday.IsHoliday(dateJST) {
		return true
	}
	return false
}

// GetAvailableTimes は指定日の予約可能な時間枠一覧を返す（LIFF: 無効区分は拒否）。
func (s *liffService) GetAvailableTimes(ctx context.Context, clinicID, typeID, staffID uint64, date time.Time) ([]TimeSlot, error) {
	return s.getAvailableTimes(ctx, clinicID, typeID, staffID, date, true)
}

// GetStaffAvailableTimes は院内スタッフ向け空き枠一覧を返す（無効区分でも枠計算へ進む。BUG-015）。
func (s *liffService) GetStaffAvailableTimes(ctx context.Context, clinicID, typeID, staffID uint64, date time.Time) ([]TimeSlot, error) {
	return s.getAvailableTimes(ctx, clinicID, typeID, staffID, date, false)
}

func (s *liffService) getAvailableTimes(
	ctx context.Context,
	clinicID, typeID, staffID uint64,
	date time.Time,
	requireActive bool,
) ([]TimeSlot, error) {
	setting, err := s.settingRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get reservation setting", "error", err)
		return nil, apperrors.Wrap(err, "failed to get reservation setting")
	}
	var course *model.ReservationType
	if requireActive {
		course, err = s.findActiveLiffCourse(ctx, clinicID, typeID)
	} else {
		course, err = s.findLiffCourse(ctx, clinicID, typeID)
	}
	if err != nil {
		return nil, err
	}

	// 定休日チェック（GetAvailableDates で除外済みのはずだが多重防御）
	datesSettings, err := ParseAvailableDatesSettings(
		setting.ClosedWeekdays,
		setting.ClosedDates,
		setting.NationalHolidayClosed,
		setting.BookingWindowMinDays,
		setting.BookingWindowMaxDays,
		setting.CalendarMonths,
		string(course.ReservationDayOption),
	)
	if err != nil {
		return nil, err
	}
	dateJST := date.In(config.JST)
	if isDateClosed(datesSettings, dateJST) {
		return []TimeSlot{}, nil
	}

	visibleStaffs, err := s.resolveTargetStaffs(ctx, clinicID, typeID, staffID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to resolve target staffs", "error", err)
		return nil, apperrors.Wrap(err, "failed to resolve target staffs")
	}

	staffInputs, err := s.buildStaffSlotInputsForDate(ctx, clinicID, visibleStaffs, date)
	if err != nil {
		slog.ErrorContext(ctx, "failed to build staff slot inputs", "error", err)
		return nil, apperrors.Wrap(err, "failed to build staff slot inputs")
	}

	// 一覧表示パスは既存挙動を維持し break_hours 破損時のエラーを無視する（意図的・スコープ外。
	// parseBusinessHoursForDate のコメント参照。書込パスの fail-closed 化のみ D10/F-2 対象）。
	bh, defaultBreaks, _ := ParseBusinessHoursForDate(ctx, setting, date)
	input := &TimeSlotsInput{
		BusinessHours:     bh,
		DefaultBreaks:     defaultBreaks,
		CourseDuration:    course.DurationMinutes,
		IntervalMinutes:   setting.TimeSlotIntervalMinutes,
		Mode:              setting.TimeSlotMode,
		MinCourseDuration: course.DurationMinutes,
		Staffs:            staffInputs,
	}
	if err := s.appendDateUnavailableBreaks(ctx, clinicID, typeID, date, input); err != nil {
		return nil, err
	}

	result, err := GenerateTimeSlots(input)
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate time slots", "error", err)
		return nil, apperrors.Wrap(err, "failed to generate time slots")
	}
	return s.mergeAndFilterGeneratedSlots(ctx, clinicID, typeID, date, course, result)
}

func (s *liffService) appendDateUnavailableBreaks(
	ctx context.Context,
	clinicID, typeID uint64,
	date time.Time,
	input *TimeSlotsInput,
) error {
	// BE-117: 予約不可時間を DefaultBreaks に追加
	unavailableTimes, err := s.unavailableTimeRepo.FindAll(ctx, clinicID, typeID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get unavailable times", "error", err)
		return apperrors.Wrap(err, "failed to get unavailable times")
	}
	applicable := filterApplicableUnavailableTimes(unavailableTimes, date)
	for i := range applicable {
		// モデルの "HH:MM" → timeslot_engine の "HHMM"（コロン除去）
		input.DefaultBreaks = append(input.DefaultBreaks, BreakPeriod{
			Start: strings.ReplaceAll(applicable[i].StartTime, ":", ""),
			End:   strings.ReplaceAll(applicable[i].EndTime, ":", ""),
		})
	}
	return nil
}

func (s *liffService) mergeAndFilterGeneratedSlots(
	ctx context.Context,
	clinicID, typeID uint64,
	date time.Time,
	course *model.ReservationType,
	result []TimeSlot,
) ([]TimeSlot, error) {
	if s.availableSlotRepo != nil {
		availableSlots, err := s.availableSlotRepo.FindAll(ctx, clinicID, typeID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get available slots", "error", err)
			return nil, apperrors.Wrap(err, "failed to get available slots")
		}
		if HasActiveAvailableSlots(availableSlots) {
			result = MergeAvailableTimeSlots(result, availableSlots, date, course.DurationMinutes)
		}
	}
	if course.MaxConcurrent != nil {
		filtered, err := FilterSlotsByCapacity(ctx, result, s.reservationRepo, clinicID, typeID, date, *course.MaxConcurrent)
		if err != nil {
			slog.ErrorContext(ctx, "failed to filter slots by capacity", "error", err)
			return nil, apperrors.Wrap(err, "failed to filter slots by capacity")
		}
		result = filtered
	}
	return result, nil
}
