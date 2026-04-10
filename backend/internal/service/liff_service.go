package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
	"gorm.io/gorm"
)

// LiffService はLIFF公開APIのビジネスロジックインターフェース
type LiffService interface {
	GetSettings(ctx context.Context, clinicID uint64) (*model.ReservationSetting, error)
	GetProfile(ctx context.Context, clinicID, customerID uint64) (*model.ReservationCustomer, error)
	GetCourses(ctx context.Context, clinicID uint64) ([]model.ServiceType, error)
	GetStaffs(ctx context.Context, clinicID, courseID uint64) ([]model.Staff, error)
	GetAvailableDates(ctx context.Context, clinicID, courseID, staffID uint64) ([]AvailableDateResult, BookingWindow, error)
	GetAvailableTimes(ctx context.Context, clinicID, courseID, staffID uint64, date time.Time) ([]TimeSlot, error)
	CreateReservation(ctx context.Context, clinicID, customerID uint64, input *CreateReservationInput) (*model.ReservationAppointment, error)
	GetMyReservations(ctx context.Context, clinicID, customerID uint64) ([]model.ReservationAppointment, error)
	CancelReservation(ctx context.Context, clinicID, customerID, reservationID uint64) error
}

type liffService struct {
	settingRepo  repository.ReservationSettingRepository
	courseRepo   repository.ReservationCourseRepository
	staffRepo    repository.ReservationStaffRepository
	scheduleRepo repository.ReservationScheduleRepository
	adminRepo    repository.ReservationAdminRepository
	customerRepo repository.ReservationCustomerRepository
	ownerRepo    repository.OwnerRepository
	validators   ReservationValidators
	notifier     ReservationNotifier
}

// NewLiffService はLIFFサービスを初期化して返す。
func NewLiffService(
	settingRepo repository.ReservationSettingRepository,
	courseRepo repository.ReservationCourseRepository,
	staffRepo repository.ReservationStaffRepository,
	scheduleRepo repository.ReservationScheduleRepository,
	adminRepo repository.ReservationAdminRepository,
	customerRepo repository.ReservationCustomerRepository,
	ownerRepo repository.OwnerRepository,
	db *gorm.DB,
	notifier ReservationNotifier,
) LiffService {
	return &liffService{
		settingRepo:  settingRepo,
		courseRepo:   courseRepo,
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
func (s *liffService) GetSettings(ctx context.Context, clinicID uint64) (*model.ReservationSetting, error) {
	setting, err := s.settingRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "get reservation setting")
	}
	return setting, nil
}

// GetProfile は顧客プロフィールを返す。
func (s *liffService) GetProfile(ctx context.Context, clinicID, customerID uint64) (*model.ReservationCustomer, error) {
	c, err := s.customerRepo.FindByID(ctx, clinicID, customerID)
	if err != nil {
		return nil, apperrors.Wrap(err, "get customer profile")
	}
	return c, nil
}

// GetCourses はLIFF向け公開コース一覧を返す（is_internal=false && reservation_visible=true）。
func (s *liffService) GetCourses(ctx context.Context, clinicID uint64) ([]model.ServiceType, error) {
	all, err := s.courseRepo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "get courses")
	}
	result := make([]model.ServiceType, 0, len(all))
	for _, c := range all {
		if !c.IsInternal && c.ReservationVisible {
			result = append(result, c)
		}
	}
	return result, nil
}

// GetStaffs はコース対応スタッフ一覧を返す（reservation_visible=true && courseIDを除外していない）。
func (s *liffService) GetStaffs(ctx context.Context, clinicID, courseID uint64) ([]model.Staff, error) {
	all, err := s.staffRepo.FindAllByClinicID(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "get staffs")
	}
	result := make([]model.Staff, 0, len(all))
	for _, st := range all {
		if !st.ReservationVisible {
			continue
		}
		excluded, err := s.staffRepo.FindExcludedServiceTypes(ctx, st.ID)
		if err != nil {
			return nil, apperrors.Wrap(err, "get excluded service types")
		}
		if isExcluded(excluded, courseID) {
			continue
		}
		result = append(result, st)
	}
	return result, nil
}

// GetAvailableDates は予約可能な日付一覧を返す。
func (s *liffService) GetAvailableDates(ctx context.Context, clinicID, courseID, staffID uint64) ([]AvailableDateResult, BookingWindow, error) {
	setting, err := s.settingRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		return nil, BookingWindow{}, apperrors.Wrap(err, "get reservation setting")
	}
	course, err := s.courseRepo.FindByID(ctx, clinicID, courseID)
	if err != nil {
		return nil, BookingWindow{}, apperrors.Wrap(err, "get course")
	}

	// スタッフを事前取得（クロージャで再利用）
	visibleStaffs, err := s.resolveTargetStaffs(ctx, clinicID, courseID, staffID)
	if err != nil {
		return nil, BookingWindow{}, err
	}

	bh, defaultBreaks := s.parseBusinessHours(setting)

	staffInputsFn := func(ctx context.Context, date time.Time, _ uint64, _ uint64) ([]StaffSlotInput, error) {
		return s.buildStaffSlotInputs(ctx, clinicID, visibleStaffs, date)
	}
	slotSettingsFn := func() TimeSlotsInput {
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

	return CalcAvailableDates(ctx, AvailableDatesInput{
		Settings:       datesSettings,
		CourseID:       courseID,
		StaffID:        staffID,
		StaffInputsFn:  staffInputsFn,
		SlotSettingsFn: slotSettingsFn,
	})
}

// GetAvailableTimes は指定日の予約可能な時間枠一覧を返す。
func (s *liffService) GetAvailableTimes(ctx context.Context, clinicID, courseID, staffID uint64, date time.Time) ([]TimeSlot, error) {
	setting, err := s.settingRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "get reservation setting")
	}
	course, err := s.courseRepo.FindByID(ctx, clinicID, courseID)
	if err != nil {
		return nil, apperrors.Wrap(err, "get course")
	}

	visibleStaffs, err := s.resolveTargetStaffs(ctx, clinicID, courseID, staffID)
	if err != nil {
		return nil, err
	}

	staffInputs, err := s.buildStaffSlotInputs(ctx, clinicID, visibleStaffs, date)
	if err != nil {
		return nil, err
	}

	bh, defaultBreaks := s.parseBusinessHours(setting)
	input := TimeSlotsInput{
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
func (s *liffService) CreateReservation(ctx context.Context, clinicID, customerID uint64, input *CreateReservationInput) (*model.ReservationAppointment, error) {
	setting, err := s.settingRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "get reservation setting")
	}
	input.ClinicID = clinicID
	input.CustomerID = customerID
	input.Settings = setting

	// TASK-RES-025: 指名なし委譲ロジック
	if input.StaffID == 0 {
		date, err := toDateTime(input.Date, input.StartTime)
		if err == nil {
			assignedID, err := s.delegateStaff(ctx, clinicID, input.ServiceTypeID, setting.NoStaffMode, date, input.StartTime, input.EndTime)
			if err == nil && assignedID != 0 {
				input.StaffID = assignedID
			}
		}
	}

	appt, err := s.validators.ValidateAndCreate(ctx, input)
	if err != nil {
		return nil, err
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
func (s *liffService) GetMyReservations(ctx context.Context, clinicID, customerID uint64) ([]model.ReservationAppointment, error) {
	items, err := s.adminRepo.FindByCustomerID(ctx, clinicID, customerID)
	if err != nil {
		return nil, apperrors.Wrap(err, "get my reservations")
	}
	return items, nil
}

// CancelReservation は予約をキャンセルする。
func (s *liffService) CancelReservation(ctx context.Context, clinicID, customerID, reservationID uint64) error {
	// Phase 6: キャンセル通知のために事前にアポを取得する
	var apptForNotify *model.ReservationAppointment
	if s.notifier != nil {
		var err error
		apptForNotify, err = s.adminRepo.FindByIDForNotify(ctx, clinicID, reservationID)
		if err != nil {
			slog.WarnContext(ctx, "failed to find appointment for notification (best-effort)", "error", err)
		}
	}

	if err := s.adminRepo.CancelByID(ctx, clinicID, customerID, reservationID); err != nil {
		return apperrors.Wrap(err, "cancel reservation")
	}

	// Phase 6: キャンセル通知（LINE + メール）fire-and-forget
	if s.notifier != nil && apptForNotify != nil {
		customer, _ := s.customerRepo.FindByID(ctx, clinicID, customerID)
		s.notifier.NotifyCancelled(ctx, apptForNotify, customer)
	}

	return nil
}

// ---- 内部ヘルパー ----

// resolveTargetStaffs はcourseID・staffIDに基づいて対象スタッフを返す。
func (s *liffService) resolveTargetStaffs(ctx context.Context, clinicID, courseID, staffID uint64) ([]model.Staff, error) {
	if staffID != 0 {
		staff, err := s.staffRepo.FindByID(ctx, staffID)
		if err != nil {
			return nil, apperrors.Wrap(err, "get staff")
		}
		if !staff.ReservationVisible {
			return nil, nil
		}
		return []model.Staff{*staff}, nil
	}

	all, err := s.staffRepo.FindAllByClinicID(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "get staffs")
	}
	result := make([]model.Staff, 0, len(all))
	for _, st := range all {
		if !st.ReservationVisible {
			continue
		}
		excluded, err := s.staffRepo.FindExcludedServiceTypes(ctx, st.ID)
		if err != nil {
			return nil, apperrors.Wrap(err, "get excluded service types")
		}
		if !isExcluded(excluded, courseID) {
			result = append(result, st)
		}
	}
	return result, nil
}

// buildStaffSlotInputs はスタッフ一覧と指定日からStaffSlotInputsを構築する。
func (s *liffService) buildStaffSlotInputs(ctx context.Context, clinicID uint64, staffs []model.Staff, date time.Time) ([]StaffSlotInput, error) {
	// 当日の全予約を一括取得（N+1回避）
	dayResv, err := s.adminRepo.FindByDay(ctx, clinicID, date)
	if err != nil {
		return nil, apperrors.Wrap(err, "get day reservations")
	}

	inputs := make([]StaffSlotInput, 0, len(staffs))
	for _, staff := range staffs {
		si := StaffSlotInput{StaffID: staff.ID}

		// シフトエントリを取得
		entry, err := s.scheduleRepo.FindByDate(ctx, clinicID, staff.ID, date)
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
		for _, r := range dayResv {
			if r.Status == model.ReservationStatusCancelled {
				continue
			}
			if r.DoctorID != nil && *r.DoctorID == staff.ID {
				si.ExistingResvs = append(si.ExistingResvs, ExistingReservation{
					StaffID:   staff.ID,
					StartTime: r.StartTime.Format("1504"),
					EndTime:   r.EndTime.Format("1504"),
				})
			}
		}

		inputs = append(inputs, si)
	}
	return inputs, nil
}

// parseBusinessHours は設定JSONから営業時間・休憩時間を解析する。
func (s *liffService) parseBusinessHours(setting *model.ReservationSetting) (BusinessHours, []BreakPeriod) {
	var bh BusinessHours
	if err := json.Unmarshal(setting.BusinessHours, &bh); err != nil {
		bh = BusinessHours{Start: "0900", End: "1900"}
	}
	var breaks []BreakPeriod
	if err := json.Unmarshal(setting.BreakHours, &breaks); err != nil {
		breaks = nil
	}
	return bh, breaks
}

// delegateStaff は指名なし時に no_staff_mode に従ってスタッフを自動割当する。
// 割当できない場合は 0 を返す（エラーではない）。
func (s *liffService) delegateStaff(ctx context.Context, clinicID, courseID uint64, mode string, date time.Time, startTime, endTime string) (uint64, error) {
	staffs, err := s.resolveTargetStaffs(ctx, clinicID, courseID, 0)
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
			return 0, nil // 失敗しても指名なしにフォールバック
		}
		startMin, err := minutesSinceMidnight(startTime)
		if err != nil {
			return 0, nil
		}
		endMin, err := minutesSinceMidnight(endTime)
		if err != nil {
			return 0, nil
		}
		for _, st := range staffs {
			if isStaffAvailable(st.ID, startMin, endMin, dayResv) {
				return st.ID, nil
			}
		}
		return 0, nil
	}
}

// isStaffAvailable はスタッフが指定時間枠で空いているか確認する。
func isStaffAvailable(staffID uint64, startMin, endMin int, dayResv []model.ReservationAppointment) bool {
	for _, r := range dayResv {
		if r.Status == model.ReservationStatusCancelled {
			continue
		}
		if r.DoctorID == nil || *r.DoctorID != staffID {
			continue
		}
		rStart := r.StartTime.Hour()*60 + r.StartTime.Minute()
		rEnd := r.EndTime.Hour()*60 + r.EndTime.Minute()
		// 重複チェック: 新枠が既存予約と重なる場合は NG
		if startMin < rEnd && endMin > rStart {
			return false
		}
	}
	return true
}

// isExcluded は指定コースIDがスタッフの除外リストに含まれるか確認する。
func isExcluded(excluded []model.StaffExcludedServiceType, courseID uint64) bool {
	for _, ex := range excluded {
		if ex.ServiceTypeID == courseID {
			return true
		}
	}
	return false
}

// tryAutoLinkOwner は予約顧客の氏名+電話番号で owners テーブルを検索し、
// 1件だけ一致した場合に reservation_customers.owner_id を自動紐付けする。
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
