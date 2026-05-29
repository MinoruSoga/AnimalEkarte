package service

import (
	"context"
	"log/slog"
	"strings"
	"time"

	holiday "github.com/holiday-jp/holiday_jp-go"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// GetAvailableDates は予約可能な日付一覧を返す。
func (s *liffService) GetAvailableDates(ctx context.Context, clinicID, typeID, staffID uint64) ([]AvailableDateResult, BookingWindow, error) {
	setting, err := s.settingRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get reservation setting", "error", err)
		return nil, BookingWindow{}, apperrors.Wrap(err, "failed to get reservation setting")
	}
	course, err := s.typeLiffRepo.FindByID(ctx, clinicID, typeID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get course", "error", err)
		return nil, BookingWindow{}, apperrors.Wrap(err, "failed to get course")
	}

	// スタッフを事前取得（クロージャで再利用）
	visibleStaffs, err := s.resolveTargetStaffs(ctx, clinicID, typeID, staffID)
	if err != nil {
		return nil, BookingWindow{}, err
	}

	staffInputsFn := func(ctx context.Context, date time.Time, _ uint64, _ uint64) ([]StaffSlotInput, error) {
		return s.buildStaffSlotInputs(ctx, clinicID, visibleStaffs, date)
	}
	slotSettingsFn := func(date time.Time) TimeSlotsInput {
		bh, defaultBreaks := parseBusinessHoursForDate(setting, date)
		return TimeSlotsInput{
			BusinessHours:     bh,
			DefaultBreaks:     defaultBreaks,
			CourseDuration:    course.DurationMinutes,
			IntervalMinutes:   setting.TimeSlotIntervalMinutes,
			Mode:              setting.TimeSlotMode,
			MinCourseDuration: course.DurationMinutes,
		}
	}
	var slotFilterFn func(date time.Time, slots []TimeSlot) []TimeSlot
	if s.availableSlotRepo != nil {
		availableSlots, err := s.availableSlotRepo.FindAll(ctx, clinicID, typeID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get available slots", "error", err)
			return nil, BookingWindow{}, apperrors.Wrap(err, "failed to get available slots")
		}
		if hasActiveAvailableSlots(availableSlots) {
			slotFilterFn = func(date time.Time, slots []TimeSlot) []TimeSlot {
				return filterTimeSlotsByAvailableSlots(slots, availableSlots, date)
			}
		}
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
		slog.WarnContext(ctx, "failed to parse available dates settings, using defaults", "error", err)
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

	// BE-117: 職種ガード — 職種紐付けが1件以上ある場合のみチェック（0件は後方互換で素通り）
	occupations, err := s.occupationRepo.FindAll(ctx, clinicID, typeID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get occupation guard", "error", err)
		return nil, window, apperrors.Wrap(err, "failed to get occupation guard")
	}
	if len(occupations) > 0 {
		for i, r := range results {
			if !r.Available {
				continue
			}
			date, err := time.ParseInLocation("2006-01-02", r.Date, jstLocation)
			if err != nil {
				slog.ErrorContext(ctx, "failed to parse date", "error", err)
				return nil, window, apperrors.Wrap(err, "failed to parse date")
			}
			count, err := s.occupationRepo.CountWorkingStaffByReservationTypeID(ctx, clinicID, typeID, date)
			if err != nil {
				slog.ErrorContext(ctx, "failed to count working staff", "error", err)
				return nil, window, apperrors.Wrap(err, "failed to count working staff")
			}
			if count == 0 {
				results[i].Available = false
				results[i].Reason = "staff_off"
			}
		}
	}

	return results, window, nil
}

// GetAvailableTimes は指定日の予約可能な時間枠一覧を返す。
func (s *liffService) GetAvailableTimes(ctx context.Context, clinicID, typeID, staffID uint64, date time.Time) ([]TimeSlot, error) {
	setting, err := s.settingRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get reservation setting", "error", err)
		return nil, apperrors.Wrap(err, "failed to get reservation setting")
	}
	course, err := s.typeLiffRepo.FindByID(ctx, clinicID, typeID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get course", "error", err)
		return nil, apperrors.Wrap(err, "failed to get course")
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
		slog.WarnContext(ctx, "failed to parse available dates settings, using defaults", "error", err)
	}
	dateJST := date.In(jstLocation)
	dateStr := dateJST.Format("2006-01-02")
	wd := int(dateJST.Weekday())
	closedWeekdaySet := make(map[int]struct{}, len(datesSettings.ClosedWeekdays))
	for _, w := range datesSettings.ClosedWeekdays {
		closedWeekdaySet[w] = struct{}{}
	}
	closedDateSet := make(map[string]struct{}, len(datesSettings.ClosedDates))
	for _, d := range datesSettings.ClosedDates {
		closedDateSet[d] = struct{}{}
	}
	if _, closed := closedWeekdaySet[wd]; closed {
		return []TimeSlot{}, nil
	}
	if _, closed := closedDateSet[dateStr]; closed {
		return []TimeSlot{}, nil
	}
	if datesSettings.NationalHolidayClosed && holiday.IsHoliday(dateJST) {
		return []TimeSlot{}, nil
	}

	visibleStaffs, err := s.resolveTargetStaffs(ctx, clinicID, typeID, staffID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to resolve target staffs", "error", err)
		return nil, apperrors.Wrap(err, "failed to resolve target staffs")
	}

	staffInputs, err := s.buildStaffSlotInputs(ctx, clinicID, visibleStaffs, date)
	if err != nil {
		slog.ErrorContext(ctx, "failed to build staff slot inputs", "error", err)
		return nil, apperrors.Wrap(err, "failed to build staff slot inputs")
	}

	bh, defaultBreaks := parseBusinessHoursForDate(setting, date)
	input := &TimeSlotsInput{
		BusinessHours:     bh,
		DefaultBreaks:     defaultBreaks,
		CourseDuration:    course.DurationMinutes,
		IntervalMinutes:   setting.TimeSlotIntervalMinutes,
		Mode:              setting.TimeSlotMode,
		MinCourseDuration: course.DurationMinutes,
		Staffs:            staffInputs,
	}
	// BE-117: 予約不可時間を DefaultBreaks に追加
	unavailableTimes, err := s.unavailableTimeRepo.FindAll(ctx, clinicID, typeID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get unavailable times", "error", err)
		return nil, apperrors.Wrap(err, "failed to get unavailable times")
	}
	applicable := filterApplicableUnavailableTimes(unavailableTimes, date)
	for i := range applicable {
		// モデルの "HH:MM" → timeslot_engine の "HHMM"（コロン除去）
		input.DefaultBreaks = append(input.DefaultBreaks, BreakPeriod{
			Start: strings.ReplaceAll(applicable[i].StartTime, ":", ""),
			End:   strings.ReplaceAll(applicable[i].EndTime, ":", ""),
		})
	}

	result, err := GenerateTimeSlots(input)
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate time slots", "error", err)
		return nil, apperrors.Wrap(err, "failed to generate time slots")
	}
	if s.availableSlotRepo == nil {
		return result, nil
	}
	availableSlots, err := s.availableSlotRepo.FindAll(ctx, clinicID, typeID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get available slots", "error", err)
		return nil, apperrors.Wrap(err, "failed to get available slots")
	}
	if !hasActiveAvailableSlots(availableSlots) {
		return result, nil
	}
	return filterTimeSlotsByAvailableSlots(result, availableSlots, date), nil
}

func filterTimeSlotsByAvailableSlots(slots []TimeSlot, availableSlots []model.ReservationTypeAvailableSlot, date time.Time) []TimeSlot {
	applicableSlots := filterApplicableAvailableSlots(availableSlots, date)
	allowedStarts := make(map[string]struct{}, len(applicableSlots))
	for i := range applicableSlots {
		allowedStarts[strings.ReplaceAll(applicableSlots[i].StartTime, ":", "")] = struct{}{}
	}
	filtered := make([]TimeSlot, 0, len(slots))
	for i := range slots {
		if _, ok := allowedStarts[slots[i].StartTime]; ok {
			filtered = append(filtered, slots[i])
		}
	}
	return filtered
}
