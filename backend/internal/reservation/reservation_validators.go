package reservation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	holiday "github.com/holiday-jp/holiday_jp-go"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// ReservationLimitError は予約制限違反エラー。
// フロントエンドが redirect_step を参照して画面遷移に使う。
type ReservationLimitError struct {
	Code         string // "SLOT_TAKEN" | "DAILY_LIMIT" | "MONTHLY_LIMIT" | "MAINTENANCE"
	Message      string
	RedirectStep int // 4=日付選択, 5=時間選択
}

func (e *ReservationLimitError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// IsReservationLimitError は err が ReservationLimitError かどうかを返す。
// エラーチェーンを辿るため errors.As を使用する。
func IsReservationLimitError(err error) (*ReservationLimitError, bool) {
	var e *ReservationLimitError
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

// ReservationValidators は予約制限チェックのインターフェース。
type ReservationValidators interface {
	ValidateAndCreate(ctx context.Context, input *CreateReservationInput) (*model.Reservation, error)
}

// CreateReservationInput は予約作成の入力。
type CreateReservationInput struct {
	ClinicID             uint64
	CustomerID           uint64
	ReservationTypeID    uint64
	StaffID              uint64 // 0 = 指名なし
	Date                 time.Time
	StartTime            string // "HHMM"
	EndTime              string // "HHMM"
	CustomerFields       []byte
	RequestText          string
	Settings             *model.LineReservationSetting
	TrimmingCourseID     *uint64  // BE-120: トリミングコース（category=trimming 時）
	TrimmingOptionIDs    []uint64 // BE-120: トリミングオプション
	TrimmingStyleRequest string   // BE-120: スタイルリクエスト
}

type reservationValidators struct {
	tx                 Transactor
	repo               ReservationRepository
	typeRepo           reservationTypeFinder
	staffRepo          ReservationStaffWriteGuard
	trimmingCourseRepo trimmingCourseFinder
	trimmingOptionRepo trimmingOptionFinder
	trimmingDetailRepo liffTrimmingDetailRepo
}

// NewReservationValidators はバリデーターを初期化して返す。
func NewReservationValidators(
	tx Transactor,
	repo ReservationRepository,
	typeRepo reservationTypeFinder,
	staffRepo ReservationStaffWriteGuard,
	trimmingCourseRepo trimmingCourseFinder,
	trimmingOptionRepo trimmingOptionFinder,
	trimmingDetailRepo liffTrimmingDetailRepo,
) ReservationValidators {
	return &reservationValidators{
		tx:                 tx,
		repo:               repo,
		typeRepo:           typeRepo,
		staffRepo:          staffRepo,
		trimmingCourseRepo: trimmingCourseRepo,
		trimmingOptionRepo: trimmingOptionRepo,
		trimmingDetailRepo: trimmingDetailRepo,
	}
}

func (v *reservationValidators) ValidateAndCreate(ctx context.Context, input *CreateReservationInput) (*model.Reservation, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("reservation input is required")
	}
	if input.Settings == nil {
		return nil, apperrors.WrapInternalServerError("LINE reservation settings are required")
	}
	if v.tx == nil {
		return nil, apperrors.WrapInternalServerError("LINE reservation transactor is required")
	}
	if v.repo == nil {
		return nil, apperrors.WrapInternalServerError("LINE reservation repository is required")
	}
	settings := input.Settings

	// 稼働状態チェック
	if settings.Status != "running" {
		return nil, &ReservationLimitError{
			Code:    "MAINTENANCE",
			Message: "只今、メンテナンス中です。暫くお待ちください。",
		}
	}

	// BUG-LINE-008: 業務時間・予約窓・休業日のサーバーサイド検証。
	// GET /available-times は正しく除外しているが、POST /reservations は素通りしていた。
	// 直接 API を叩かれても無効な予約を受け付けないようにする。
	if err := validateBusinessRules(ctx, settings, input.Date, input.StartTime, input.EndTime); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate business rules")
	}

	var result *model.Reservation
	if err := v.tx.WithTx(ctx, func(ctx context.Context) error {
		// BE-refactor.md X-9: LINE予約は不特定多数から到達するため、空き枠への同時アクセスが
		// 最も起こりやすい経路。AcquireBookingLock（clinic 単位 advisory xact lock）で
		// 競合チェック～INSERT を直列化する。
		if err := v.repo.AcquireBookingLock(ctx, input.ClinicID); err != nil {
			return err
		}
		if err := v.repo.AssertLineCustomerInClinic(ctx, input.ClinicID, input.CustomerID); err != nil {
			return apperrors.Wrap(err, "failed to verify LINE customer ownership")
		}
		// Request-derived masters and staff are revalidated inside the write transaction.
		// Their repository reads hold SHARE locks until appointment/detail/options commit.
		reservationType, err := v.validateReservationMasterOwnership(ctx, input)
		if err != nil {
			return err
		}
		if input.StaffID != 0 {
			if err := ValidateLineReservationStaffCapability(ctx, v.staffRepo, input.ClinicID, &input.StaffID, input.ReservationTypeID); err != nil {
				return err
			}
		}
		// 時間枠を SELECT FOR UPDATE でロック
		startDT, err := ToDateTime(input.Date, input.StartTime)
		if err != nil {
			return apperrors.WrapInvalidInput(err.Error())
		}
		endDT, err := ToDateTime(input.Date, input.EndTime)
		if err != nil {
			return apperrors.WrapInvalidInput(err.Error())
		}
		if err := validateTimeRange(startDT, endDT); err != nil {
			return err
		}

		// 時間枠の空きチェック（医師指定時は同一医師の重複、未指定時は出勤医師数を上限）
		var doctorIDPtr *uint64
		if input.StaffID != 0 {
			id := input.StaffID
			doctorIDPtr = &id
		}
		if err := CheckSlotConflict(ctx, v.repo, input.ClinicID, doctorIDPtr, startDT, endDT, nil); err != nil {
			// 出勤医師不在と通常の満員を区別する（RedirectStep: 4=日付選択, 5=時間選択）
			if errors.Is(err, errNoDoctorsOnDuty) {
				return &ReservationLimitError{
					Code:         "SLOT_TAKEN",
					Message:      "本日は医師が出勤していません。別の日をお選びください。",
					RedirectStep: 4,
				}
			}
			if errors.Is(err, apperrors.ErrConflict) {
				return &ReservationLimitError{
					Code:         "SLOT_TAKEN",
					Message:      "選択された時間枠は既に予約が入っています。別の時間をお選びください。",
					RedirectStep: 5,
				}
			}
			return err
		}
		if err := CheckReservationTypeCapacity(ctx, v.repo, v.typeRepo, input.ClinicID, input.ReservationTypeID, startDT, nil); err != nil {
			if errors.Is(err, apperrors.ErrConflict) {
				return &ReservationLimitError{
					Code:         "SLOT_TAKEN",
					Message:      "選択された時間枠は満員です。別の時間をお選びください。",
					RedirectStep: 5,
				}
			}
			return err
		}

		// 同日予約制限チェック（BE-refactor.md E-8: 日次・月次の構造的クローンを畳む）
		dayStart := time.Date(input.Date.Year(), input.Date.Month(), input.Date.Day(), 0, 0, 0, 0, input.Date.Location())
		dayEnd := dayStart.Add(24 * time.Hour)
		if err := checkCustomerReservationLimit(ctx, v.repo, input, dayStart, dayEnd,
			"DAILY_LIMIT", "1日内に予約できる件数を超えています。別の日をお選びください。",
			"failed to count daily reservations", settings.DailyLimit); err != nil {
			return err
		}

		// 同月予約制限チェック
		monthStart := time.Date(input.Date.Year(), input.Date.Month(), 1, 0, 0, 0, 0, input.Date.Location())
		monthEnd := monthStart.AddDate(0, 1, 0)
		if err := checkCustomerReservationLimit(ctx, v.repo, input, monthStart, monthEnd,
			"MONTHLY_LIMIT", "1ヶ月内に予約できる件数を超えています。別の月をお選びください。",
			"failed to count monthly reservations", settings.MonthlyLimit); err != nil {
			return err
		}

		// 確認番号生成
		confirmationNumber, err := generateConfirmationNumber(ctx, v.repo, input.ClinicID, input.Date)
		if err != nil {
			return apperrors.Wrap(err, "failed to generate confirmation number")
		}

		// 予約作成
		appt := buildLineReservation(input, startDT, endDT, confirmationNumber)
		if err := v.repo.Create(ctx, appt); err != nil {
			return apperrors.Wrap(err, "failed to create reservation")
		}
		if reservationType.Category == model.ReservationTypeCategoryTrimming && hasLineTrimmingDetailInput(input) {
			if v.trimmingDetailRepo == nil {
				return apperrors.WrapInternalServerError("trimming detail repository is required for a LINE trimming reservation")
			}
			detail := &model.AppointmentTrimmingDetail{
				ClinicID:      input.ClinicID,
				AppointmentID: appt.ID,
				CourseID:      input.TrimmingCourseID,
				StyleRequest:  input.TrimmingStyleRequest,
			}
			if err := v.trimmingDetailRepo.Create(ctx, detail); err != nil {
				return apperrors.Wrap(err, "failed to create LINE trimming detail")
			}
			if len(input.TrimmingOptionIDs) > 0 {
				if err := v.trimmingDetailRepo.SetOptions(ctx, input.ClinicID, appt.ID, input.TrimmingOptionIDs); err != nil {
					return apperrors.Wrap(err, "failed to set LINE trimming options")
				}
			}
		}
		result = appt
		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "failed to create line reservation", "error", err)
		return nil, apperrors.Wrap(err, "failed to create line reservation")
	}
	return result, nil
}

// validateReservationMasterOwnership は master-FK（ReservationTypeID/TrimmingCourseID/
// TrimmingOptionIDs）が caller の clinic に属することを検証する（BE-refactor.md E-8:
// ValidateAndCreate の4責務分割・純粋抽出）。所有権失敗は best-effort に落とさず hard fail
// とし、orphan appointment を作らない。
func (v *reservationValidators) validateReservationMasterOwnership(ctx context.Context, input *CreateReservationInput) (*model.ReservationType, error) {
	if v.typeRepo == nil {
		return nil, apperrors.WrapInternalServerError("reservation type repository is required")
	}
	reservationType, err := v.typeRepo.FindByID(ctx, input.ClinicID, input.ReservationTypeID)
	if err != nil {
		slog.ErrorContext(ctx, "reservation type not found or belongs to different clinic", "error", err)
		return nil, apperrors.Wrap(err, "failed to verify reservation type ownership")
	}
	if !reservationType.IsActive || !reservationType.ReservationVisible || reservationType.IsInternal {
		return nil, apperrors.WrapInvalidInput("reservation type is not available for LINE reservation")
	}
	if reservationType.Category != model.ReservationTypeCategoryTrimming && hasLineTrimmingDetailInput(input) {
		return nil, apperrors.WrapInvalidInput("trimming fields require a trimming reservation type")
	}
	if v.trimmingCourseRepo != nil && input.TrimmingCourseID != nil {
		course, err := v.trimmingCourseRepo.FindByID(ctx, input.ClinicID, *input.TrimmingCourseID)
		if err != nil {
			slog.ErrorContext(ctx, "trimming course not found or belongs to different clinic", "error", err)
			return nil, apperrors.Wrap(err, "failed to verify trimming course ownership")
		}
		if !course.IsActive {
			return nil, apperrors.WrapInvalidInput("trimming course is inactive")
		}
	}
	if input.TrimmingCourseID != nil && v.trimmingCourseRepo == nil {
		return nil, apperrors.WrapInternalServerError("trimming course repository is required")
	}
	if v.trimmingOptionRepo != nil {
		for _, optionID := range input.TrimmingOptionIDs {
			option, err := v.trimmingOptionRepo.FindByID(ctx, input.ClinicID, optionID)
			if err != nil {
				slog.ErrorContext(ctx, "trimming option not found or belongs to different clinic", "error", err)
				return nil, apperrors.Wrap(err, "failed to verify trimming option ownership")
			}
			if !option.IsActive {
				return nil, apperrors.WrapInvalidInput("trimming option is inactive")
			}
		}
	}
	if len(input.TrimmingOptionIDs) > 0 && v.trimmingOptionRepo == nil {
		return nil, apperrors.WrapInternalServerError("trimming option repository is required")
	}
	return reservationType, nil
}

func hasLineTrimmingDetailInput(input *CreateReservationInput) bool {
	return input != nil && (input.TrimmingCourseID != nil || len(input.TrimmingOptionIDs) > 0 || input.TrimmingStyleRequest != "")
}

// checkCustomerReservationLimit は指定期間内の顧客の予約件数が limit 以上かどうかを検証する
// （BE-refactor.md E-8: 同日・同月の構造的クローンを畳む）。limit が nil または 0 以下の場合は
// チェックをスキップする。countErrMsg はカウントクエリ失敗時の apperrors.Wrap メッセージ、
// code/msg は上限超過時に返す ReservationLimitError の内容（呼び出し元ごとの既存文言を再現）。
func checkCustomerReservationLimit(ctx context.Context, repo ReservationRepository, input *CreateReservationInput, from, to time.Time, code, msg, countErrMsg string, limit *int) error {
	if limit == nil || *limit <= 0 {
		return nil
	}
	count, err := repo.CountByCustomerAndDateRange(ctx, input.ClinicID, input.CustomerID, from, to)
	if err != nil {
		return apperrors.Wrap(err, countErrMsg)
	}
	if int(count) >= *limit {
		return &ReservationLimitError{
			Code:         code,
			Message:      msg,
			RedirectStep: 4,
		}
	}
	return nil
}

// buildLineReservation は LINE 予約の model.Reservation エンティティを組み立てる純関数
// （BE-refactor.md E-8）。
func buildLineReservation(input *CreateReservationInput, startDT, endDT time.Time, confirmationNumber string) *model.Reservation {
	doctorID := &input.StaffID
	if input.StaffID == 0 {
		doctorID = nil
	}
	customerFields := json.RawMessage("{}")
	if len(input.CustomerFields) > 0 {
		customerFields = input.CustomerFields
	}
	notes := input.RequestText
	if confirmationNumber != "" {
		if notes != "" {
			notes = confirmationNumber + " " + notes
		} else {
			notes = confirmationNumber
		}
	}
	return &model.Reservation{
		ClinicID:          input.ClinicID,
		StartTime:         startDT,
		EndTime:           endDT,
		ReservationTypeID: input.ReservationTypeID,
		DoctorID:          doctorID,
		Status:            model.ReservationStatusConfirmed,
		Source:            model.ReservationSourceLine,
		LineCustomerID:    &input.CustomerID,
		IsStaffDelegated:  input.StaffID == 0,
		CustomerFields:    customerFields,
		Notes:             notes,
		VisitType:         model.VisitTypeRevisit,
	}
}

// validateBusinessRules は予約可能日・営業時間・休憩時間などの業務ルールを検証する（BUG-LINE-008）。
// - 過去日・予約窓（min/max days）範囲外
// - 休業曜日（closed_weekdays）
// - 休業日（closed_dates）
// - 祝日（national_holiday_closed=true 時のみ）
// - 営業時間外（business_hours / business_hours_by_weekday）
// - 休憩時間と重複（break_hours）
// エラーは apperrors.WrapInvalidInput で 400 を返す。
func validateBusinessRules(ctx context.Context, settings *model.LineReservationSetting, date time.Time, startTime, endTime string) error {
	loc := config.JST
	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	dateStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc)

	// 1. 予約窓チェック
	minDate := today.AddDate(0, 0, settings.BookingWindowMinDays)
	maxDate := today.AddDate(0, 0, settings.BookingWindowMaxDays)
	if dateStart.Before(minDate) {
		return apperrors.WrapInvalidInput(fmt.Sprintf("予約可能日は %s 以降です", minDate.Format(time.DateOnly)))
	}
	if dateStart.After(maxDate) {
		return apperrors.WrapInvalidInput(fmt.Sprintf("予約可能日は %s 以前です", maxDate.Format(time.DateOnly)))
	}

	// 2. 休業曜日
	// A-3: JSON 破損時は fail-closed（予約拒否）。休業曜日チェックを検証できない状態で
	// 予約を通すと、休業曜日に予約が誤って確定しうるため（break_hours と対称）。
	if len(settings.ClosedWeekdays) > 0 {
		var closedWeekdays []int
		if err := json.Unmarshal(settings.ClosedWeekdays, &closedWeekdays); err != nil {
			return apperrors.Wrap(err, "invalid closed_weekdays")
		}
		wd := int(dateStart.Weekday())
		for _, cwd := range closedWeekdays {
			if cwd == wd {
				return apperrors.WrapInvalidInput("指定日は休業曜日のため予約できません")
			}
		}
	}

	// 3. 休業日
	// A-3: JSON 破損時は fail-closed（予約拒否）。休業日チェックを検証できない状態で
	// 予約を通すと、休業日に予約が誤って確定しうるため（break_hours と対称）。
	if len(settings.ClosedDates) > 0 {
		var closedDates []string
		if err := json.Unmarshal(settings.ClosedDates, &closedDates); err != nil {
			return apperrors.Wrap(err, "invalid closed_dates")
		}
		dateStr := dateStart.Format(time.DateOnly)
		for _, cd := range closedDates {
			if cd == dateStr {
				return apperrors.WrapInvalidInput("指定日は休業日のため予約できません")
			}
		}
	}

	// 4. 祝日
	if settings.NationalHolidayClosed && holiday.IsHoliday(dateStart) {
		return apperrors.WrapInvalidInput("指定日は祝日休業のため予約できません")
	}

	// 5. 営業時間
	// D10/F-2: break_hours の unmarshal 失敗は fail-closed（予約拒否）。休憩時間との重複を
	// 検証できない状態で予約を通すと、休憩時間帯の予約が誤って確定しうるため。
	bh, breaks, err := ParseBusinessHoursForDate(ctx, settings, dateStart)
	if err != nil {
		return apperrors.Wrap(err, "invalid break_hours")
	}
	bsStart, err := MinutesSinceMidnight(bh.Start)
	if err != nil {
		return apperrors.Wrap(err, "invalid business_hours.start")
	}
	bsEnd, err := MinutesSinceMidnight(bh.End)
	if err != nil {
		return apperrors.Wrap(err, "invalid business_hours.end")
	}
	reqStart, err := MinutesSinceMidnight(startTime)
	if err != nil {
		return apperrors.WrapInvalidInput(err.Error())
	}
	reqEnd, err := MinutesSinceMidnight(endTime)
	if err != nil {
		return apperrors.WrapInvalidInput(err.Error())
	}
	if reqStart < bsStart || reqEnd > bsEnd {
		return apperrors.WrapInvalidInput(fmt.Sprintf("営業時間外の予約はできません (営業時間 %s-%s)", bh.Start, bh.End))
	}

	// 6. 休憩時間との重複（半開区間 [start, end) で重複判定）
	// D10/F-2: 個別エントリの形式不正も fail-closed。continue でスキップすると、その
	// エントリだけ重複判定を回避でき休憩時間帯の予約が誤って確定しうる。
	for _, b := range breaks {
		brStart, err := MinutesSinceMidnight(b.Start)
		if err != nil {
			return apperrors.Wrap(err, "invalid break_hours entry")
		}
		brEnd, err := MinutesSinceMidnight(b.End)
		if err != nil {
			return apperrors.Wrap(err, "invalid break_hours entry")
		}
		if reqStart < brEnd && reqEnd > brStart {
			return apperrors.WrapInvalidInput(fmt.Sprintf("休憩時間(%s-%s)と重複しています", b.Start, b.End))
		}
	}

	return nil
}

// ToDateTime は日付と "HHMM" 文字列を time.Time に変換する。
func ToDateTime(date time.Time, hhmm string) (time.Time, error) {
	mins, err := MinutesSinceMidnight(hhmm)
	if err != nil {
		return time.Time{}, err
	}
	h := mins / 60
	m := mins % 60
	return time.Date(date.Year(), date.Month(), date.Day(), h, m, 0, 0, config.JST), nil
}

// generateConfirmationNumber は "R-YYYYMMDD-XXXX" 形式の確認番号を生成する。
// 同日の予約件数+1をシーケンス番号として使用する。
func generateConfirmationNumber(ctx context.Context, repo ReservationRepository, clinicID uint64, date time.Time) (string, error) {
	count, err := repo.CountByDateAndSource(ctx, clinicID, date, model.ReservationSourceLine)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count reservations by date and source", "error", err)
		return "", apperrors.Wrap(err, "failed to count reservations by date and source")
	}
	seq := int(count) + 1
	return fmt.Sprintf("R-%s-%04d", date.Format("20060102"), seq), nil
}
