package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
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
	ValidateAndCreate(ctx context.Context, input *CreateReservationInput) (*model.Appointment, error)
}

// CreateReservationInput は予約作成の入力。
type CreateReservationInput struct {
	ClinicID          uint64
	CustomerID        uint64
	ReservationTypeID uint64
	StaffID           uint64 // 0 = 指名なし
	Date              time.Time
	StartTime         string // "HHMM"
	EndTime           string // "HHMM"
	CustomerFields    []byte
	RequestText       string
	Settings          *model.LineReservationSetting
}

type reservationValidators struct {
	tx   repository.Transactor
	repo repository.ReservationRepository
}

// NewReservationValidators はバリデーターを初期化して返す。
func NewReservationValidators(tx repository.Transactor, repo repository.ReservationRepository) ReservationValidators {
	return &reservationValidators{tx: tx, repo: repo}
}

func (v *reservationValidators) ValidateAndCreate(ctx context.Context, input *CreateReservationInput) (*model.Appointment, error) {
	settings := input.Settings

	// 稼働状態チェック
	if settings.Status != "running" {
		return nil, &ReservationLimitError{
			Code:    "MAINTENANCE",
			Message: "只今、メンテナンス中です。暫くお待ちください。",
		}
	}

	var result *model.Appointment
	if err := v.tx.WithTx(ctx, func(ctx context.Context) error {
		// 時間枠を SELECT FOR UPDATE でロック
		startDT, err := toDateTime(input.Date, input.StartTime)
		if err != nil {
			return apperrors.WrapInvalidInput(err.Error())
		}
		endDT, err := toDateTime(input.Date, input.EndTime)
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
		if err := checkSlotConflict(ctx, v.repo, input.ClinicID, doctorIDPtr, startDT, endDT, nil); err != nil {
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

		// 同日予約制限チェック
		if settings.DailyLimit != nil && *settings.DailyLimit > 0 {
			dayStart := time.Date(input.Date.Year(), input.Date.Month(), input.Date.Day(), 0, 0, 0, 0, input.Date.Location())
			dayEnd := dayStart.Add(24 * time.Hour)
			dailyCount, err := v.repo.CountByCustomerAndDateRange(ctx, input.ClinicID, input.CustomerID, dayStart, dayEnd)
			if err != nil {
				return apperrors.Wrap(err, "failed to count daily reservations")
			}
			if int(dailyCount) >= *settings.DailyLimit {
				return &ReservationLimitError{
					Code:         "DAILY_LIMIT",
					Message:      "1日内に予約できる件数を超えています。別の日をお選びください。",
					RedirectStep: 4,
				}
			}
		}

		// 同月予約制限チェック
		if settings.MonthlyLimit != nil && *settings.MonthlyLimit > 0 {
			monthStart := time.Date(input.Date.Year(), input.Date.Month(), 1, 0, 0, 0, 0, input.Date.Location())
			monthEnd := monthStart.AddDate(0, 1, 0)
			monthlyCount, err := v.repo.CountByCustomerAndDateRange(ctx, input.ClinicID, input.CustomerID, monthStart, monthEnd)
			if err != nil {
				return apperrors.Wrap(err, "failed to count monthly reservations")
			}
			if int(monthlyCount) >= *settings.MonthlyLimit {
				return &ReservationLimitError{
					Code:         "MONTHLY_LIMIT",
					Message:      "1ヶ月内に予約できる件数を超えています。別の月をお選びください。",
					RedirectStep: 4,
				}
			}
		}

		// 確認番号生成
		confirmationNumber, err := generateConfirmationNumber(ctx, v.repo, input.ClinicID, input.Date)
		if err != nil {
			return apperrors.Wrap(err, "failed to generate confirmation number")
		}

		// 予約作成
		doctorID := &input.StaffID
		if input.StaffID == 0 {
			doctorID = nil
		}
		customerFields := []byte("{}")
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

		appt := &model.Appointment{
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
		if err := v.repo.Create(ctx, appt); err != nil {
			return apperrors.Wrap(err, "failed to create reservation")
		}
		result = appt
		return nil
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to create line reservation")
	}
	return result, nil
}

// toDateTime は日付と "HHMM" 文字列を time.Time に変換する。
func toDateTime(date time.Time, hhmm string) (time.Time, error) {
	mins, err := minutesSinceMidnight(hhmm)
	if err != nil {
		return time.Time{}, err
	}
	h := mins / 60
	m := mins % 60
	return time.Date(date.Year(), date.Month(), date.Day(), h, m, 0, 0, jstLocation()), nil
}

// generateConfirmationNumber は "R-YYYYMMDD-XXXX" 形式の確認番号を生成する。
// 同日の予約件数+1をシーケンス番号として使用する。
func generateConfirmationNumber(ctx context.Context, repo repository.ReservationRepository, clinicID uint64, date time.Time) (string, error) {
	count, err := repo.CountByDateAndSource(ctx, clinicID, date, model.ReservationSourceLine)
	if err != nil {
		return "", err
	}
	seq := int(count) + 1
	return fmt.Sprintf("R-%s-%04d", date.Format("20060102"), seq), nil
}
