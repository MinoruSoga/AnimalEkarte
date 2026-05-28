package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
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
	return result, nil
}

// filterApplicableUnavailableTimes は date に適用される不可時間帯を返す。
// 優先順位: specific > weekly（特定日設定が曜日設定を上書き）
func filterApplicableUnavailableTimes(times []model.ReservationTypeUnavailableTime, date time.Time) []model.ReservationTypeUnavailableTime {
	dateStr := date.In(jstLocation).Format("2006-01-02")
	var specific, weekly []model.ReservationTypeUnavailableTime
	for i := range times {
		switch times[i].UnavailableType {
		case model.UnavailableTypeSpecific:
			if times[i].SpecificDate != nil && times[i].SpecificDate.UTC().Format("2006-01-02") == dateStr {
				specific = append(specific, times[i])
			}
		case model.UnavailableTypeWeekly:
			if times[i].DayOfWeek != nil && int(*times[i].DayOfWeek) == int(date.In(jstLocation).Weekday()) {
				weekly = append(weekly, times[i])
			}
		}
	}
	if len(specific) > 0 {
		return specific
	}
	return weekly
}

// resolveTargetStaffs はtypeID・staffIDに基づいて対象スタッフを返す。
func (s *liffService) resolveTargetStaffs(ctx context.Context, clinicID, typeID, staffID uint64) ([]model.Staff, error) {
	if staffID != 0 {
		staff, err := s.staffRepo.FindByID(ctx, clinicID, staffID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get staff", "error", err)
			return nil, apperrors.Wrap(err, "failed to get staff")
		}
		if !staff.ReservationVisible {
			return nil, nil
		}
		return []model.Staff{*staff}, nil
	}

	all, err := s.staffRepo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get staffs", "error", err)
		return nil, apperrors.Wrap(err, "failed to get staffs")
	}
	return s.filterVisibleStaffsByTypeID(ctx, typeID, all)
}

// filterVisibleStaffsByTypeID は reservation_visible=true かつ typeID を除外していないスタッフを返す。
// FindAllExcludedReservationTypesByStaffIDs で一括取得して N+1 クエリを回避する。
func (s *liffService) filterVisibleStaffsByTypeID(ctx context.Context, typeID uint64, all []model.Staff) ([]model.Staff, error) {
	visibleIDs := make([]uint64, 0, len(all))
	for i := range all {
		if all[i].ReservationVisible {
			visibleIDs = append(visibleIDs, all[i].ID)
		}
	}
	if len(visibleIDs) == 0 {
		return nil, nil
	}

	allExclusions, err := s.staffRepo.FindAllExcludedReservationTypesByStaffIDs(ctx, visibleIDs)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get excluded service types", "error", err)
		return nil, apperrors.Wrap(err, "failed to get excluded service types")
	}

	// staffID → 除外予約区分 のマップを構築（O(N) で済む）
	excludeMap := make(map[uint64][]model.StaffReservationExclusion, len(allExclusions))
	for _, ex := range allExclusions {
		excludeMap[ex.StaffID] = append(excludeMap[ex.StaffID], ex)
	}

	result := make([]model.Staff, 0, len(visibleIDs))
	for i := range all {
		if !all[i].ReservationVisible {
			continue
		}
		if !isExcluded(excludeMap[all[i].ID], typeID) {
			result = append(result, all[i])
		}
	}
	return result, nil
}

// timeToHHMM は "HH:MM:SS" / "HH:MM" / "HHMM" を timeslot_engine が要求する "HHMM" 形式に変換する。
// PostgreSQL の time 型は GORM 経由で "HH:MM:SS" (8文字) として返るため正規化が必要。
func timeToHHMM(s string) string {
	clean := strings.ReplaceAll(s, ":", "")
	if len(clean) >= 4 {
		return clean[:4]
	}
	return clean
}

func ptrTimeToHHMM(s *string) *string {
	if s == nil {
		return nil
	}
	result := timeToHHMM(*s)
	return &result
}

// buildStaffSlotInputs はスタッフ一覧と指定日からStaffSlotInputsを構築する。
func (s *liffService) buildStaffSlotInputs(ctx context.Context, clinicID uint64, staffs []model.Staff, date time.Time) ([]StaffSlotInput, error) {
	// 当日の全予約を一括取得（N+1回避）
	dayResv, err := s.adminRepo.FindAllByDay(ctx, clinicID, date)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get day reservations", "error", err)
		return nil, apperrors.Wrap(err, "failed to get day reservations")
	}

	inputs := make([]StaffSlotInput, 0, len(staffs))
	for i := range staffs {
		si := StaffSlotInput{StaffID: staffs[i].ID}

		// シフトエントリを取得
		entry, err := s.scheduleRepo.FindAllByDate(ctx, clinicID, staffs[i].ID, date)
		if err == nil && entry != nil {
			breaks, _ := s.scheduleRepo.FindAllBreaksByEntryID(ctx, entry.ID)
			override := &StaffScheduleOverride{
				ShiftType: string(entry.ShiftType),
				WorkStart: ptrTimeToHHMM(entry.StartTime), // PostgreSQL "HH:MM:SS" → "HHMM"
				WorkEnd:   ptrTimeToHHMM(entry.EndTime),   // PostgreSQL "HH:MM:SS" → "HHMM"
			}
			for _, b := range breaks {
				override.Breaks = append(override.Breaks, BreakPeriod{
					Start: timeToHHMM(b.BreakStart), // PostgreSQL "HH:MM:SS" → "HHMM"
					End:   timeToHHMM(b.BreakEnd),   // PostgreSQL "HH:MM:SS" → "HHMM"
				})
			}
			si.ScheduleOverride = override
		}

		// 当日の既存予約を絞り込み
		for j := range dayResv {
			if dayResv[j].Status == model.ReservationStatusCancelled {
				continue
			}
			if dayResv[j].DoctorID != nil && *dayResv[j].DoctorID == staffs[i].ID {
				si.ExistingResvs = append(si.ExistingResvs, ExistingReservation{
					StaffID:   staffs[i].ID,
					StartTime: dayResv[j].StartTime.Format("1504"),
					EndTime:   dayResv[j].EndTime.Format("1504"),
				})
			}
		}

		inputs = append(inputs, si)
	}
	return inputs, nil
}

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

// delegateStaff は指名なし時に no_staff_mode に従ってスタッフを自動割当する。
// 割当できない場合は 0 を返す（エラーではない）。
func (s *liffService) delegateStaff(ctx context.Context, clinicID, typeID uint64, mode string, date time.Time, startTime, endTime string) (uint64, error) {
	staffs, err := s.resolveTargetStaffs(ctx, clinicID, typeID, 0)
	if err != nil || len(staffs) == 0 {
		slog.ErrorContext(ctx, "failed to resolve target staffs liff", "error", err)
		return 0, apperrors.Wrap(err, "failed to resolve target staffs liff")
	}

	switch mode {
	case "top_priority":
		// 表示順1位（sort_order が最小）のスタッフに固定割当
		return staffs[0].ID, nil

	default: // "first_available"
		// 空き枠があるスタッフを表示順に探す
		dayResv, err := s.adminRepo.FindAllByDay(ctx, clinicID, date)
		if err != nil {
			return 0, nil //nolint:nilerr // 意図的フォールバック: 既存予約取得失敗時は空き確認をスキップして指名なしにする
		}
		startMin, err := minutesSinceMidnight(startTime)
		if err != nil {
			return 0, nil //nolint:nilerr // 意図的フォールバック: 時刻フォーマット不正時は空き確認をスキップして指名なしにする
		}
		endMin, err := minutesSinceMidnight(endTime)
		if err != nil {
			return 0, nil //nolint:nilerr // 意図的フォールバック: 時刻フォーマット不正時は空き確認をスキップして指名なしにする
		}
		for i := range staffs {
			if isStaffAvailable(staffs[i].ID, startMin, endMin, dayResv) {
				return staffs[i].ID, nil
			}
		}
		return 0, nil
	}
}

// isStaffAvailable はスタッフが指定時間枠で空いているか確認する。
func isStaffAvailable(staffID uint64, startMin, endMin int, dayResv []model.Reservation) bool {
	for i := range dayResv {
		if dayResv[i].Status == model.ReservationStatusCancelled {
			continue
		}
		if dayResv[i].DoctorID == nil || *dayResv[i].DoctorID != staffID {
			continue
		}
		rStart := dayResv[i].StartTime.Hour()*60 + dayResv[i].StartTime.Minute()
		rEnd := dayResv[i].EndTime.Hour()*60 + dayResv[i].EndTime.Minute()
		// 重複チェック: 新枠が既存予約と重なる場合は NG
		if startMin < rEnd && endMin > rStart {
			return false
		}
	}
	return true
}

// isExcluded は指定コースIDがスタッフの除外リストに含まれるか確認する。
func isExcluded(excluded []model.StaffReservationExclusion, typeID uint64) bool {
	for _, ex := range excluded {
		if ex.ReservationTypeID == typeID {
			return true
		}
	}
	return false
}
