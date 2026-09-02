package reservation

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

func (s *reservationService) validateCreateClosedDaysIfConfigured(ctx context.Context, clinicID uint64, startTime time.Time) error {
	if s.settingFinder == nil {
		return nil
	}
	return s.validateCreateClosedDays(ctx, clinicID, startTime)
}

func validateReservationPetNotDeceasedOnRelink(
	ctx context.Context,
	repo petByIDInClinicFinder,
	clinicID uint64,
	petID *uint64,
) error {
	if petID == nil || *petID == 0 {
		return nil
	}
	return ValidateReservationPetNotDeceased(ctx, repo, clinicID, petID)
}

func (s *reservationService) validateCreateClosedDays(ctx context.Context, clinicID uint64, startTime time.Time) error {
	if s.settingFinder == nil {
		return apperrors.WrapInternalServerError("LINE reservation settings are required")
	}
	settings, err := s.settingFinder.FindByClinicID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load LINE reservation settings for closed-day check", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to load LINE reservation settings")
	}
	if settings == nil {
		return apperrors.WrapInternalServerError("LINE reservation settings are required")
	}
	return validateClosedDays(settings, startTime)
}

func validateClinicHoliday(ctx context.Context, finder clinicHolidayFinder, clinicID uint64, startTime time.Time) error {
	if finder == nil {
		return apperrors.WrapInternalServerError("clinic holiday lookup is required")
	}
	dateJST := startTime.In(config.JST)
	dateStart := time.Date(dateJST.Year(), dateJST.Month(), dateJST.Day(), 0, 0, 0, 0, config.JST)
	holiday, err := finder.FindByDate(ctx, clinicID, dateStart)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if holiday != nil {
		return apperrors.WrapInvalidInput("指定日は休診日のため予約できません")
	}
	return nil
}

func (s *reservationService) validateCreateClinicHoliday(ctx context.Context, clinicID uint64, startTime time.Time) error {
	return validateClinicHoliday(ctx, s.holidayFinder, clinicID, startTime)
}

func ShouldEnforceReservationBookingConstraints(status model.ReservationStatus, route *string) bool {
	if route != nil {
		switch *route {
		case "reception", "exam_room", "record_shortcut":
			return false
		}
	}
	return shouldEnforceReservationBookingConstraintsForStatus(status)
}

// shouldEnforceClosedDayConstraintOnUpdate is deliberately route-independent:
// shortcut routes can skip create-time availability checks, but a reschedule
// must never move a pending booking to a configured closed day or holiday.
func shouldEnforceClosedDayConstraintOnUpdate(status model.ReservationStatus, _ *string) bool {
	return shouldEnforceReservationBookingConstraintsForStatus(status)
}

func shouldEnforceReservationBookingConstraintsForStatus(status model.ReservationStatus) bool {
	switch status {
	case model.ReservationStatusCheckedIn,
		model.ReservationStatusInConsultation,
		model.ReservationStatusAccounting,
		model.ReservationStatusCompleted:
		return false
	default:
		return true
	}
}

func reservationUint64PtrEqual(a, b *uint64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func resolveUpdateReservationTypeID(current *model.Reservation, input *UpdateReservationInput) uint64 {
	if input != nil && input.ReservationTypeID != nil {
		return *input.ReservationTypeID
	}
	if current == nil {
		return 0
	}
	return current.ReservationTypeID
}

func reservationScheduleFieldsChanged(current *model.Reservation, input *UpdateReservationInput) bool {
	if current == nil || input == nil {
		return false
	}
	resolvedStart, resolvedEnd, resolvedDoctorID := resolveUpdateParams(current, input)
	resolvedTypeID := resolveUpdateReservationTypeID(current, input)
	return !resolvedStart.Equal(current.StartTime) ||
		!resolvedEnd.Equal(current.EndTime) ||
		!reservationUint64PtrEqual(resolvedDoctorID, current.DoctorID) ||
		resolvedTypeID != current.ReservationTypeID
}

// reservationStaffCapabilityRequiresRevalidation reports whether Update must
// re-run staff capability against the resolved doctor/type. Independent of
// BUG-006 slot-conflict skip. Payload presence is not enough: memo-only PATCH
// resends unchanged DoctorID/ReservationTypeID.
func reservationStaffCapabilityRequiresRevalidation(current *model.Reservation, input *UpdateReservationInput) bool {
	if current == nil || input == nil {
		return false
	}
	_, _, resolvedDoctorID := resolveUpdateParams(current, input)
	resolvedTypeID := resolveUpdateReservationTypeID(current, input)
	return !reservationUint64PtrEqual(resolvedDoctorID, current.DoctorID) ||
		resolvedTypeID != current.ReservationTypeID
}

// shouldReevaluateReservationBookingConstraintsOnUpdate reports whether Update
// should re-run slot conflict / on-duty absence checks.
//
// Decision (BUG-006):
//   - Skip when start/end/doctor/type are unchanged. Frontend memo-only PATCH
//     sends the full schedule payload; presence is not an actual change.
//   - Skip when the current status is already checked_in or later. The visit
//     already occupies the slot; re-checking on-duty doctors caused false 409s.
//   - Reservation route is not a skip signal on Update. Rescheduling a still
//     pending appointment must still conflict-check even if it was created via
//     reception / exam_room / record_shortcut.
func shouldReevaluateReservationBookingConstraintsOnUpdate(current *model.Reservation, input *UpdateReservationInput) bool {
	if !reservationScheduleFieldsChanged(current, input) {
		return false
	}
	return shouldEnforceReservationBookingConstraintsForStatus(current.Status)
}

// validateTimeRange は end_time > start_time を確認する共通バリデーション
// （BE9-2C R①: 実装は sharedkernel.ValidateTimeRange へ昇格・本関数は既存呼び出し面互換の delegate）。
func validateTimeRange(startTime, endTime time.Time) error {
	return sharedkernel.ValidateTimeRange(startTime, endTime)
}

// errNoDoctorsOnDuty は当日の出勤医師が 0 人のためスロット予約不可を示すセンチネルエラー。
// *apperrors.AppError なので RespondError が errors.As で Message を抽出し日本語メッセージを返す。
// LINE パスでは reservation_validators.go が errors.Is でこれを識別し RedirectStep: 4 を返す。
// ※ WrapConflict はパッケージレベル変数として固定ポインタを保持するため errors.Is が機能する。
var errNoDoctorsOnDuty = apperrors.WrapConflict("本日は医師が出勤していないため予約できません")

type slotConflictChecker interface {
	HasDoctorConflict(ctx context.Context, clinicID, doctorID uint64, start, end time.Time, excludeID *uint64) (bool, error)
	CountOnDutyDoctors(ctx context.Context, clinicID uint64, date time.Time) (int64, error)
	CountConflicts(ctx context.Context, clinicID uint64, start, end time.Time, excludeID *uint64) (int64, error)
}

// checkDoctorSlotConflict は特定医師の時間枠重複をチェックする（SELECT FOR UPDATE）。
func checkDoctorSlotConflict(ctx context.Context, repo slotConflictChecker, clinicID, doctorID uint64, start, end time.Time, excludeID *uint64) error {
	conflict, err := repo.HasDoctorConflict(ctx, clinicID, doctorID, start, end, excludeID)
	if err != nil {
		return apperrors.Wrap(err, "check doctor slot conflict")
	}
	if conflict {
		return apperrors.WrapConflict("この時間枠は既に予約が入っています")
	}
	return nil
}

// checkCapacitySlotConflict は出勤医師数を上限として時間枠の空き確認をする（SELECT FOR UPDATE）。
// 出勤医師が 0 人の場合は errNoDoctorsOnDuty を返す（LINE パスで RedirectStep を分岐するため）。
func checkCapacitySlotConflict(ctx context.Context, repo slotConflictChecker, clinicID uint64, start, end time.Time, excludeID *uint64) error {
	doctorCount, err := repo.CountOnDutyDoctors(ctx, clinicID, start)
	if err != nil {
		return apperrors.Wrap(err, "count on-duty doctors")
	}
	if doctorCount == 0 {
		return errNoDoctorsOnDuty
	}

	conflictCount, err := repo.CountConflicts(ctx, clinicID, start, end, excludeID)
	if err != nil {
		return apperrors.Wrap(err, "count conflicts")
	}
	if conflictCount >= doctorCount {
		return apperrors.WrapConflict("この時間枠は満員です（出勤医師数に達しています）")
	}
	return nil
}

// CheckSlotConflict は時間枠の空き・重複をチェックする（SELECT FOR UPDATE）。
//
//   - doctor_id 指定時 → 同一医師の重複のみチェック（別医師は許可）
//   - doctor_id nil 時 → その日の出勤医師数を上限として全予約件数をチェック
//
// excludeID が非 nil の場合、その予約 ID を競合対象から除外する（Update 時の自己競合防止）。
// 競合がある場合は apperrors.ErrConflict ラップエラーを返す。
func CheckSlotConflict(ctx context.Context, repo slotConflictChecker, clinicID uint64, doctorID *uint64, startTime, endTime time.Time, excludeID *uint64) error {
	if doctorID != nil {
		return checkDoctorSlotConflict(ctx, repo, clinicID, *doctorID, startTime, endTime, excludeID)
	}
	return checkCapacitySlotConflict(ctx, repo, clinicID, startTime, endTime, excludeID)
}

// resolveUpdateParams は現在の予約と更新入力から、競合チェックに使用する時刻・医師 ID を確定する。
// 未指定フィールドは現在値を維持する。DoctorID=0 は NULL（医師未指定）として扱う。
func resolveUpdateParams(current *model.Reservation, input *UpdateReservationInput) (start, end time.Time, doctorID *uint64) {
	start = current.StartTime
	if input.StartTime != nil {
		start = *input.StartTime
	}
	end = current.EndTime
	if input.EndTime != nil {
		end = *input.EndTime
	}
	doctorID = current.DoctorID
	if input.DoctorID != nil {
		if *input.DoctorID == 0 {
			doctorID = nil // 0 は「医師未指定」として NULL 扱い
		} else {
			doctorID = input.DoctorID
		}
	}
	return start, end, doctorID
}

func validateLineReservationCheckedInLink(current *model.Reservation, input *UpdateReservationInput) error {
	if input.Status == nil || *input.Status != model.ReservationStatusCheckedIn {
		return nil
	}
	if current.Source != model.ReservationSourceLine || current.LineCustomerID == nil {
		return nil
	}
	// AUD-001: 0 クリアの意味と揃える（*uint64(0) を non-nil のまま通さない）
	ownerID, petID := resolveFinalOwnerPet(current, input)
	if ownerID == nil || petID == nil {
		return apperrors.WrapInvalidInput("LINE予約を受付済みにする前に飼主とペットの紐付けが必要です")
	}
	return nil
}

func isTransitioningToInConsultation(current *model.Reservation, input *UpdateReservationInput) bool {
	if input == nil || input.Status == nil || *input.Status != model.ReservationStatusInConsultation {
		return false
	}
	if current != nil && current.Status == model.ReservationStatusInConsultation {
		return false
	}
	return true
}

func validateInConsultationHasMedicalRecord(
	ctx context.Context,
	repo ReservationRepository,
	clinicID, id uint64,
	current *model.Reservation,
	input *UpdateReservationInput,
) error {
	if !isTransitioningToInConsultation(current, input) {
		return nil
	}
	if repo == nil {
		return apperrors.WrapInternalServerError("reservation repository is required")
	}
	count, err := repo.CountMedicalRecordsByReservationID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count medical records for in_consultation transition", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to check medical records for consultation")
	}
	if count <= 0 {
		return apperrors.WrapConflict("診療を開始するにはカルテが必要です")
	}
	return nil
}

// updateWithConflictCheck は SELECT FOR UPDATE + トランザクション内で競合チェック + 予約更新を実行する。
// 時刻・医師変更がある場合にのみ呼び出す。
