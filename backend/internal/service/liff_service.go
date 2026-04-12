package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// LiffService はLIFF公開APIのビジネスロジックインターフェース
type LiffService interface {
	GetSettings(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error)
	GetProfile(ctx context.Context, clinicID, customerID uint64) (*model.LineCustomer, error)
	GetCourses(ctx context.Context, clinicID uint64) ([]model.ReservationType, error)
	GetStaffs(ctx context.Context, clinicID, typeID uint64) ([]model.Staff, error)
	GetAvailableDates(ctx context.Context, clinicID, typeID, staffID uint64) ([]AvailableDateResult, BookingWindow, error)
	GetAvailableTimes(ctx context.Context, clinicID, typeID, staffID uint64, date time.Time) ([]TimeSlot, error)
	CreateReservation(ctx context.Context, clinicID, customerID uint64, input *CreateReservationInput) (*model.Appointment, error)
	GetMyReservations(ctx context.Context, clinicID, customerID uint64) ([]model.Appointment, error)
	CancelReservation(ctx context.Context, clinicID, customerID, reservationID uint64) error
}

type liffService struct {
	settingRepo  repository.LineReservationSettingRepository
	typeLiffRepo repository.ReservationTypeLiffRepository
	staffRepo    repository.ReservationStaffRepository
	scheduleRepo repository.ReservationScheduleRepository
	adminRepo    repository.ReservationAdminRepository
	customerRepo repository.LineCustomerRepository
	ownerRepo    repository.OwnerRepository
	validators   ReservationValidators
	notifier     ReservationNotifier
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
	db *gorm.DB,
	notifier ReservationNotifier,
) LiffService {
	return &liffService{
		settingRepo:  settingRepo,
		typeLiffRepo: typeLiffRepo,
		staffRepo:    staffRepo,
		scheduleRepo: scheduleRepo,
		adminRepo:    adminRepo,
		customerRepo: customerRepo,
		ownerRepo:    ownerRepo,
		validators:   NewReservationValidators(db),
		notifier:     notifier,
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
	all, err := s.staffRepo.FindAllByClinicID(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get staffs")
	}
	result := make([]model.Staff, 0, len(all))
	for i := range all {
		if !all[i].ReservationVisible {
			continue
		}
		excluded, err := s.staffRepo.FindExcludedReservationTypes(ctx, all[i].ID)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to get excluded service types")
		}
		if isExcluded(excluded, typeID) {
			continue
		}
		result = append(result, all[i])
	}
	return result, nil
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
		bh, defaultBreaks := s.parseBusinessHoursForDate(setting, date)
		return TimeSlotsInput{
			BusinessHours:     bh,
			DefaultBreaks:     defaultBreaks,
			CourseDuration:    course.DurationMinutes,
			IntervalMinutes:   setting.TimeSlotIntervalMinutes,
			Mode:              setting.TimeSlotMode,
			MinCourseDuration: course.DurationMinutes,
		}
	}

	datesSettings, _ := ParseAvailableDatesSettings(
		setting.ClosedWeekdays,
		setting.ClosedDates,
		setting.NationalHolidayClosed,
		setting.BookingWindowMinDays,
		setting.BookingWindowMaxDays,
		setting.CalendarMonths,
		string(course.ReservationDayOption),
	)

	return CalcAvailableDates(ctx, &AvailableDatesInput{
		Settings:       datesSettings,
		TypeID:         typeID,
		StaffID:        staffID,
		StaffInputsFn:  staffInputsFn,
		SlotSettingsFn: slotSettingsFn,
	})
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
	datesSettings, _ := ParseAvailableDatesSettings(
		setting.ClosedWeekdays,
		setting.ClosedDates,
		setting.NationalHolidayClosed,
		setting.BookingWindowMinDays,
		setting.BookingWindowMaxDays,
		setting.CalendarMonths,
		string(course.ReservationDayOption),
	)
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
	if datesSettings.NationalHolidayClosed && isJapaneseHoliday(dateJST) {
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

	bh, defaultBreaks := s.parseBusinessHoursForDate(setting, date)
	input := &TimeSlotsInput{
		BusinessHours:     bh,
		DefaultBreaks:     defaultBreaks,
		CourseDuration:    course.DurationMinutes,
		IntervalMinutes:   setting.TimeSlotIntervalMinutes,
		Mode:              setting.TimeSlotMode,
		MinCourseDuration: course.DurationMinutes,
		Staffs:            staffInputs,
	}
	result, err := GenerateTimeSlots(input)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to generate time slots")
	}
	return result, nil
}

// CreateReservation は予約を確定する。staffID=0 の場合は no_staff_mode に従って自動割当する。
func (s *liffService) CreateReservation(ctx context.Context, clinicID, customerID uint64, input *CreateReservationInput) (*model.Appointment, error) {
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

	// 顧客の追加フィールドを更新（プロフィール自動保存）
	if len(input.CustomerFields) > 0 && string(input.CustomerFields) != "{}" {
		if err := s.customerRepo.UpdateAdditionalFields(ctx, clinicID, customerID, input.CustomerFields); err != nil {
			slog.WarnContext(ctx, "failed to update customer additional fields (best-effort)", "error", err)
		}
	}

	// 自動オーナー紐付け: customer_fields の氏名+電話番号で owners を検索し、1件一致で自動リンク
	s.tryAutoLinkOwner(ctx, clinicID, customerID, input.CustomerFields)

	// Phase 6: 予約確定通知（LINE + メール）fire-and-forget
	if s.notifier != nil {
		// 通知メッセージ用にリレーションをロード（enriched が nil の場合は元の appt を使う）
		notifyAppt := appt
		if enriched, err := s.adminRepo.FindByIDForNotify(ctx, clinicID, appt.ID); err == nil && enriched != nil {
			notifyAppt = enriched
		}
		customer, _ := s.customerRepo.FindByID(ctx, clinicID, customerID)
		s.notifier.NotifyCreated(ctx, notifyAppt, customer)
	}

	return appt, nil
}

// GetMyReservations は顧客自身の予約一覧を返す。
func (s *liffService) GetMyReservations(ctx context.Context, clinicID, customerID uint64) ([]model.Appointment, error) {
	items, err := s.adminRepo.FindByCustomerID(ctx, clinicID, customerID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get my reservations")
	}
	return items, nil
}

// CancelReservation は予約をキャンセルする。
func (s *liffService) CancelReservation(ctx context.Context, clinicID, customerID, reservationID uint64) error {
	// Phase 6: キャンセル通知のために事前にアポを取得する
	var apptForNotify *model.Appointment
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
		customer, _ := s.customerRepo.FindByID(ctx, clinicID, customerID)
		s.notifier.NotifyCancelled(ctx, apptForNotify, customer)
	}

	return nil
}

// ---- 内部ヘルパー ----

// resolveTargetStaffs はtypeID・staffIDに基づいて対象スタッフを返す。
func (s *liffService) resolveTargetStaffs(ctx context.Context, clinicID, typeID, staffID uint64) ([]model.Staff, error) {
	if staffID != 0 {
		staff, err := s.staffRepo.FindByID(ctx, staffID)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to get staff")
		}
		if !staff.ReservationVisible {
			return nil, nil
		}
		return []model.Staff{*staff}, nil
	}

	all, err := s.staffRepo.FindAllByClinicID(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get staffs")
	}
	result := make([]model.Staff, 0, len(all))
	for i := range all {
		if !all[i].ReservationVisible {
			continue
		}
		excluded, err := s.staffRepo.FindExcludedReservationTypes(ctx, all[i].ID)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to get excluded service types")
		}
		if !isExcluded(excluded, typeID) {
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

// parseBusinessHoursForDate は指定日の営業時間・休憩時間を解析する。
// BusinessHoursByWeekday に該当曜日の設定があればそれを優先する。
func (s *liffService) parseBusinessHoursForDate(setting *model.LineReservationSetting, date time.Time) (BusinessHours, []BreakPeriod) {
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
func isStaffAvailable(staffID uint64, startMin, endMin int, dayResv []model.Appointment) bool {
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
