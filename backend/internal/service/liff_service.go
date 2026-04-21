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
	"github.com/animal-ekarte/backend/internal/repository"
)

// LiffService はLIFF公開APIのビジネスロジックインターフェース
type LiffService interface {
	GetSettings(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error)
	GetProfile(ctx context.Context, clinicID, customerID uint64) (*model.LineCustomer, error)
	GetCourses(ctx context.Context, clinicID uint64) ([]model.ReservationType, error)
	GetTrimmingCourses(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error)
	GetTrimmingOptions(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error)
	GetStaffs(ctx context.Context, clinicID, typeID uint64) ([]model.Staff, error)
	GetAvailableDates(ctx context.Context, clinicID, typeID, staffID uint64) ([]AvailableDateResult, BookingWindow, error)
	GetAvailableTimes(ctx context.Context, clinicID, typeID, staffID uint64, date time.Time) ([]TimeSlot, error)
	CreateReservation(ctx context.Context, clinicID, customerID uint64, input *CreateReservationInput) (*model.Reservation, error)
	GetMyReservations(ctx context.Context, clinicID, customerID uint64) ([]model.Reservation, error)
	CancelReservation(ctx context.Context, clinicID, customerID, reservationID uint64) error
}

type liffService struct {
	settingRepo         repository.LineReservationSettingRepository
	typeLiffRepo        repository.ReservationTypeLiffRepository
	staffRepo           repository.ReservationStaffRepository
	scheduleRepo        repository.ReservationScheduleRepository
	adminRepo           repository.ReservationAdminRepository
	reservationRepo     repository.ReservationRepository
	customerRepo        repository.LineCustomerRepository
	ownerRepo           repository.OwnerRepository
	validators          ReservationValidators
	notifier            ReservationNotifier
	unavailableTimeRepo repository.ReservationTypeUnavailableTimeRepository // BE-117
	occupationRepo      repository.ReservationTypeOccupationRepository      // BE-117
	trimmingCourseRepo  repository.TrimmingCourseRepository                 // BE-120
	trimmingOptionRepo  repository.TrimmingOptionRepository                 // BE-120
	trimmingDetailRepo  repository.AppointmentTrimmingDetailRepository      // BE-120
}

// NewLiffService はLIFFサービスを初期化して返す。
func NewLiffService(
	settingRepo repository.LineReservationSettingRepository,
	typeLiffRepo repository.ReservationTypeLiffRepository,
	staffRepo repository.ReservationStaffRepository,
	scheduleRepo repository.ReservationScheduleRepository,
	adminRepo repository.ReservationAdminRepository,
	customerRepo repository.LineCustomerRepository,
	ownerRepo repository.OwnerRepository,
	tx repository.Transactor,
	reservationRepo repository.ReservationRepository,
	notifier ReservationNotifier,
	unavailableTimeRepo repository.ReservationTypeUnavailableTimeRepository,
	occupationRepo repository.ReservationTypeOccupationRepository,
	trimmingCourseRepo repository.TrimmingCourseRepository,
	trimmingOptionRepo repository.TrimmingOptionRepository,
	trimmingDetailRepo repository.AppointmentTrimmingDetailRepository,
) LiffService {
	return &liffService{
		settingRepo:         settingRepo,
		typeLiffRepo:        typeLiffRepo,
		staffRepo:           staffRepo,
		scheduleRepo:        scheduleRepo,
		adminRepo:           adminRepo,
		reservationRepo:     reservationRepo,
		customerRepo:        customerRepo,
		ownerRepo:           ownerRepo,
		validators:          NewReservationValidators(tx, reservationRepo),
		notifier:            notifier,
		unavailableTimeRepo: unavailableTimeRepo,
		occupationRepo:      occupationRepo,
		trimmingCourseRepo:  trimmingCourseRepo,
		trimmingOptionRepo:  trimmingOptionRepo,
		trimmingDetailRepo:  trimmingDetailRepo,
	}
}

// GetSettings はLIFF公開設定を返す（機密フィールドは除外）。
func (s *liffService) GetSettings(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
	setting, err := s.settingRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation setting")
	}
	return setting, nil
}

// GetProfile は顧客プロフィールを返す。
func (s *liffService) GetProfile(ctx context.Context, clinicID, customerID uint64) (*model.LineCustomer, error) {
	c, err := s.customerRepo.FindByID(ctx, clinicID, customerID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get customer profile")
	}
	return c, nil
}

// GetTrimmingCourses はLIFF向けトリミングコース一覧を返す（is_active=true のみ）。
func (s *liffService) GetTrimmingCourses(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error) {
	all, err := s.trimmingCourseRepo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get trimming courses")
	}
	result := make([]model.TrimmingCourse, 0, len(all))
	for i := range all {
		if all[i].IsActive {
			result = append(result, all[i])
		}
	}
	return result, nil
}

// GetTrimmingOptions はLIFF向けトリミングオプション一覧を返す（is_active=true のみ）。
func (s *liffService) GetTrimmingOptions(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error) {
	all, err := s.trimmingOptionRepo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get trimming options")
	}
	result := make([]model.TrimmingOption, 0, len(all))
	for i := range all {
		if all[i].IsActive {
			result = append(result, all[i])
		}
	}
	return result, nil
}

// GetCourses はLIFF向け公開コース一覧を返す（is_internal=false && reservation_visible=true）。
func (s *liffService) GetCourses(ctx context.Context, clinicID uint64) ([]model.ReservationType, error) {
	all, err := s.typeLiffRepo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get courses")
	}
	result := make([]model.ReservationType, 0, len(all))
	for i := range all {
		if !all[i].IsInternal && all[i].ReservationVisible {
			result = append(result, all[i])
		}
	}
	return result, nil
}

// GetStaffs は予約区分対応スタッフ一覧を返す（reservation_visible=true && typeIDを除外していない）。
func (s *liffService) GetStaffs(ctx context.Context, clinicID, typeID uint64) ([]model.Staff, error) {
	all, err := s.staffRepo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get staffs")
	}
	return s.filterVisibleStaffsByTypeID(ctx, typeID, all)
}

// GetAvailableDates は予約可能な日付一覧を返す。
func (s *liffService) GetAvailableDates(ctx context.Context, clinicID, typeID, staffID uint64) ([]AvailableDateResult, BookingWindow, error) {
	setting, err := s.settingRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		return nil, BookingWindow{}, apperrors.Wrap(err, "failed to get reservation setting")
	}
	course, err := s.typeLiffRepo.FindByID(ctx, clinicID, typeID)
	if err != nil {
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
		return nil, window, apperrors.Wrap(err, "failed to get occupation guard")
	}
	if len(occupations) > 0 {
		for i, r := range results {
			if !r.Available {
				continue
			}
			date, err := time.ParseInLocation("2006-01-02", r.Date, jstLocation())
			if err != nil {
				return nil, window, apperrors.Wrap(err, "failed to parse date")
			}
			count, err := s.occupationRepo.CountByStaffID(ctx, clinicID, typeID, date)
			if err != nil {
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
		return nil, apperrors.Wrap(err, "failed to get reservation setting")
	}
	course, err := s.typeLiffRepo.FindByID(ctx, clinicID, typeID)
	if err != nil {
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
	dateJST := date.In(jstLocation())
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
		return nil, apperrors.Wrap(err, "failed to resolve target staffs")
	}

	staffInputs, err := s.buildStaffSlotInputs(ctx, clinicID, visibleStaffs, date)
	if err != nil {
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
		return nil, apperrors.Wrap(err, "failed to generate time slots")
	}
	return result, nil
}

// filterApplicableUnavailableTimes は date に適用される不可時間帯を返す。
// 優先順位: specific > weekly（特定日設定が曜日設定を上書き）
func filterApplicableUnavailableTimes(times []model.ReservationTypeUnavailableTime, date time.Time) []model.ReservationTypeUnavailableTime {
	dateStr := date.In(jstLocation()).Format("2006-01-02")
	var specific, weekly []model.ReservationTypeUnavailableTime
	for i := range times {
		switch times[i].UnavailableType {
		case model.UnavailableTypeSpecific:
			if times[i].SpecificDate != nil && times[i].SpecificDate.UTC().Format("2006-01-02") == dateStr {
				specific = append(specific, times[i])
			}
		case model.UnavailableTypeWeekly:
			if times[i].DayOfWeek != nil && int(*times[i].DayOfWeek) == int(date.In(jstLocation()).Weekday()) {
				weekly = append(weekly, times[i])
			}
		}
	}
	if len(specific) > 0 {
		return specific
	}
	return weekly
}

// CreateReservation は予約を確定する。staffID=0 の場合は no_staff_mode に従って自動割当する。
func (s *liffService) CreateReservation(ctx context.Context, clinicID, customerID uint64, input *CreateReservationInput) (*model.Reservation, error) {
	setting, err := s.settingRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation setting")
	}
	input.ClinicID = clinicID
	input.CustomerID = customerID
	input.Settings = setting

	// TASK-RES-025: 指名なし委譲ロジック
	if input.StaffID == 0 {
		date, err := toDateTime(input.Date, input.StartTime)
		if err == nil {
			assignedID, err := s.delegateStaff(ctx, clinicID, input.ReservationTypeID, setting.NoStaffMode, date, input.StartTime, input.EndTime)
			if err == nil && assignedID != 0 {
				input.StaffID = assignedID
			}
		}
	}

	appt, err := s.validators.ValidateAndCreate(ctx, input)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to validate and create appointment")
	}

	// BE-120: category=trimming の場合、トリミング詳細を作成（best-effort）
	if input.TrimmingCourseID != nil {
		detail := &model.AppointmentTrimmingDetail{
			ClinicID:      clinicID,
			AppointmentID: appt.ID,
			CourseID:      input.TrimmingCourseID,
			StyleRequest:  input.TrimmingStyleRequest,
		}
		if err := s.trimmingDetailRepo.Create(ctx, detail); err != nil {
			slog.WarnContext(ctx, "failed to create trimming detail (best-effort)", "error", err, "appointment_id", appt.ID)
		} else if len(input.TrimmingOptionIDs) > 0 {
			if err := s.trimmingDetailRepo.SetOptions(ctx, appt.ID, input.TrimmingOptionIDs); err != nil {
				slog.WarnContext(ctx, "failed to set trimming options (best-effort)", "error", err, "appointment_id", appt.ID)
			}
		}
	}

	// 顧客の追加フィールドを更新（プロフィール自動保存）
	if len(input.CustomerFields) > 0 && string(input.CustomerFields) != "{}" {
		if err := s.customerRepo.UpdateAdditionalFields(ctx, clinicID, customerID, input.CustomerFields); err != nil {
			slog.WarnContext(ctx, "failed to update customer additional fields (best-effort)", "error", err)
		}
	}

	// 自動オーナー紐付け: customer_fields の氏名+電話番号で owners を検索し、1件一致で自動リンク
	s.tryAutoLinkOwner(ctx, clinicID, customerID, input.CustomerFields)
	s.tryAttachReservationOwnerPet(ctx, clinicID, customerID, appt, input.CustomerFields)

	// Phase 6: 予約確定通知（LINE + メール）fire-and-forget
	if s.notifier != nil {
		// 通知メッセージ用にリレーションをロード（enriched が nil の場合は元の appt を使う）
		notifyAppt := appt
		if enriched, err := s.adminRepo.FindByIDForNotify(ctx, clinicID, appt.ID); err == nil && enriched != nil {
			notifyAppt = enriched
		}
		// best-effort: 通知失敗は予約成否に影響させない
		customer, custErr := s.customerRepo.FindByID(ctx, clinicID, customerID)
		if custErr != nil {
			slog.WarnContext(ctx, "failed to get customer for notification (best-effort)", "error", custErr)
		}
		s.notifier.NotifyCreated(ctx, notifyAppt, customer)
	}

	return appt, nil
}

// GetMyReservations は顧客自身の予約一覧を返す。
func (s *liffService) GetMyReservations(ctx context.Context, clinicID, customerID uint64) ([]model.Reservation, error) {
	items, err := s.adminRepo.FindByCustomerID(ctx, clinicID, customerID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get my reservations")
	}
	return items, nil
}

// CancelReservation は予約をキャンセルする。
func (s *liffService) CancelReservation(ctx context.Context, clinicID, customerID, reservationID uint64) error {
	// Phase 6: キャンセル通知のために事前にアポを取得する
	var apptForNotify *model.Reservation
	if s.notifier != nil {
		var err error
		apptForNotify, err = s.adminRepo.FindByIDForNotify(ctx, clinicID, reservationID)
		if err != nil {
			slog.WarnContext(ctx, "failed to find appointment for notification (best-effort)", "error", err)
		}
	}

	if err := s.adminRepo.CancelByID(ctx, clinicID, customerID, reservationID); err != nil {
		return apperrors.Wrap(err, "failed to cancel reservation")
	}

	// Phase 6: キャンセル通知（LINE + メール）fire-and-forget
	if s.notifier != nil && apptForNotify != nil {
		// best-effort: 通知失敗は予約キャンセル成否に影響させない
		customer, custErr := s.customerRepo.FindByID(ctx, clinicID, customerID)
		if custErr != nil {
			slog.WarnContext(ctx, "failed to get customer for cancel notification (best-effort)", "error", custErr)
		}
		s.notifier.NotifyCancelled(ctx, apptForNotify, customer)
	}

	return nil
}

// tryAttachReservationOwnerPet は line_customer の owner 紐付け状態に応じて、
// 予約へ owner_id / pet_id を best-effort で反映する。
func (s *liffService) tryAttachReservationOwnerPet(
	ctx context.Context,
	clinicID, customerID uint64,
	appt *model.Reservation,
	customerFields []byte,
) {
	if s.reservationRepo == nil || appt == nil {
		return
	}

	customer, err := s.customerRepo.FindByID(ctx, clinicID, customerID)
	if err != nil || customer == nil || customer.OwnerID == nil {
		return
	}

	fields := map[string]any{
		"owner_id": *customer.OwnerID,
	}
	if petID := resolveReservationPetID(customer, customerFields); petID != nil {
		fields["pet_id"] = *petID
	}

	updated, err := s.reservationRepo.UpdateFields(ctx, clinicID, appt.ID, fields)
	if err != nil {
		slog.WarnContext(ctx, "failed to attach owner/pet to line reservation (best-effort)", "error", err)
		return
	}
	*appt = *updated
}

// ---- 内部ヘルパー ----

// resolveTargetStaffs はtypeID・staffIDに基づいて対象スタッフを返す。
func (s *liffService) resolveTargetStaffs(ctx context.Context, clinicID, typeID, staffID uint64) ([]model.Staff, error) {
	if staffID != 0 {
		staff, err := s.staffRepo.FindByID(ctx, clinicID, staffID)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to get staff")
		}
		if !staff.ReservationVisible {
			return nil, nil
		}
		return []model.Staff{*staff}, nil
	}

	all, err := s.staffRepo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get staffs")
	}
	return s.filterVisibleStaffsByTypeID(ctx, typeID, all)
}

// filterVisibleStaffsByTypeID は reservation_visible=true かつ typeID を除外していないスタッフを返す。
// FindExcludedReservationTypesByStaffIDs で一括取得して N+1 クエリを回避する。
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

	allExclusions, err := s.staffRepo.FindExcludedReservationTypesByStaffIDs(ctx, visibleIDs)
	if err != nil {
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

// buildStaffSlotInputs はスタッフ一覧と指定日からStaffSlotInputsを構築する。
func (s *liffService) buildStaffSlotInputs(ctx context.Context, clinicID uint64, staffs []model.Staff, date time.Time) ([]StaffSlotInput, error) {
	// 当日の全予約を一括取得（N+1回避）
	dayResv, err := s.adminRepo.FindByDay(ctx, clinicID, date)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get day reservations")
	}

	inputs := make([]StaffSlotInput, 0, len(staffs))
	for i := range staffs {
		si := StaffSlotInput{StaffID: staffs[i].ID}

		// シフトエントリを取得
		entry, err := s.scheduleRepo.FindByDate(ctx, clinicID, staffs[i].ID, date)
		if err == nil && entry != nil {
			breaks, _ := s.scheduleRepo.FindBreaksByEntryID(ctx, entry.ID)
			override := &StaffScheduleOverride{
				ShiftType: string(entry.ShiftType),
				WorkStart: entry.StartTime,
				WorkEnd:   entry.EndTime,
			}
			for _, b := range breaks {
				override.Breaks = append(override.Breaks, BreakPeriod{
					Start: b.BreakStart,
					End:   b.BreakEnd,
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
			key := strconv.Itoa(int(date.In(jstLocation()).Weekday()))
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
		return 0, err
	}

	switch mode {
	case "top_priority":
		// 表示順1位（sort_order が最小）のスタッフに固定割当
		return staffs[0].ID, nil

	default: // "first_available"
		// 空き枠があるスタッフを表示順に探す
		dayResv, err := s.adminRepo.FindByDay(ctx, clinicID, date)
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

func resolveReservationPetID(customer *model.LineCustomer, customerFields []byte) *uint64 {
	if customer == nil || customer.Owner == nil || len(customer.Owner.Pets) == 0 {
		return nil
	}
	if len(customer.Owner.Pets) == 1 {
		id := customer.Owner.Pets[0].ID
		return &id
	}

	var fields struct {
		Pets []struct {
			Name string `json:"name"`
		} `json:"pets"`
	}
	if len(customerFields) == 0 || json.Unmarshal(customerFields, &fields) != nil || len(fields.Pets) == 0 {
		return nil
	}

	wantName := strings.TrimSpace(fields.Pets[0].Name)
	if wantName == "" {
		return nil
	}
	for i := range customer.Owner.Pets {
		if strings.TrimSpace(customer.Owner.Pets[i].Name) == wantName {
			id := customer.Owner.Pets[i].ID
			return &id
		}
	}
	return nil
}

// tryAutoLinkOwner は予約顧客の氏名+電話番号で owners テーブルを検索し、
// 1件だけ一致した場合に line_customers.owner_id を自動紐付けする。
// best-effort: 失敗しても予約処理は中断しない。
func (s *liffService) tryAutoLinkOwner(ctx context.Context, clinicID, customerID uint64, customerFields []byte) {
	if s.ownerRepo == nil {
		return
	}

	// 既に紐付け済みならスキップ
	customer, err := s.customerRepo.FindByID(ctx, clinicID, customerID)
	if err != nil || customer == nil || customer.OwnerID != nil {
		return
	}

	// customer_fields から氏名と電話番号を抽出
	var fields struct {
		CustomerName string `json:"customer_name"`
		Phone        string `json:"phone"`
		OwnerName    string `json:"owner_name"`
	}
	if len(customerFields) == 0 {
		return
	}
	if err := json.Unmarshal(customerFields, &fields); err != nil {
		return
	}

	// owner_name を優先、なければ customer_name を使用
	name := fields.OwnerName
	if name == "" {
		name = fields.CustomerName
	}
	phone := fields.Phone
	if name == "" || phone == "" {
		return
	}

	// owners テーブルで氏名+電話番号の完全一致検索（1件のみ返す）
	owner, err := s.ownerRepo.FindByNameAndPhone(ctx, clinicID, name, phone)
	if err != nil {
		slog.WarnContext(ctx, "auto-link owner lookup failed (best-effort)", "error", err)
		return
	}
	if owner == nil {
		return // 0件 or 複数件 → 紐付けしない
	}

	// 自動紐付け実行
	if err := s.customerRepo.UpdateOwnerLink(ctx, clinicID, customerID, &owner.ID); err != nil {
		slog.WarnContext(ctx, "auto-link owner update failed (best-effort)", "error", err)
		return
	}
	slog.InfoContext(ctx, "auto-linked LINE customer to owner",
		"customer_id", customerID, "owner_id", owner.ID, "name", name)
}
