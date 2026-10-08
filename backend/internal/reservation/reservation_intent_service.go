package reservation

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// nextVisitIntentMaxYearsAhead bounds how far ahead a next-visit intent date may be.
const nextVisitIntentMaxYearsAhead = 2

// ConfirmTentativeInput は tentative 予約を本予約へ確定する service 層コマンド。
// ReservationTypeID/DoctorID/IsDesignated/Notes は未指定なら現在値を維持する。
type ConfirmTentativeInput struct {
	StartTime         time.Time // 必須: 確定後の開始時刻
	EndTime           time.Time // 必須: 確定後の終了時刻
	ReservationTypeID *uint64   // nil = tentative の診療区分（general placeholder）を維持
	DoctorID          *uint64   // nil/0 = 担当医未指定のまま
	IsDesignated      *bool
	Status            *model.ReservationStatus // nil = confirmed。pending または confirmed のみ許可
	Notes             *string
}

// validateNextVisitIntentDate は次回予定日を JST 暦日で検証する。
// 過去日と約2年先を超える日付は拒否する（server-side enforcement）。
func validateNextVisitIntentDate(intendedDate, now time.Time) error {
	if intendedDate.IsZero() {
		return apperrors.WrapInvalidInput("intended_date is required")
	}
	todayJST := now.In(config.JST)
	today := time.Date(todayJST.Year(), todayJST.Month(), todayJST.Day(), 0, 0, 0, 0, config.JST)
	targetJST := intendedDate.In(config.JST)
	target := time.Date(targetJST.Year(), targetJST.Month(), targetJST.Day(), 0, 0, 0, 0, config.JST)
	if target.Before(today) {
		return apperrors.WrapInvalidInput("次回予定日は本日以降の日付を指定してください")
	}
	if target.After(today.AddDate(nextVisitIntentMaxYearsAhead, 0, 0)) {
		return apperrors.WrapInvalidInput("次回予定日は2年以内の日付を指定してください")
	}
	return nil
}

// SetNextVisitIntent はペットの次回予定（tentative appointment）を作成または更新する。
// tentative は予約枠容量・時間枠競合・通知の対象外であり、ここではそれらを評価しない。
func (s *reservationService) SetNextVisitIntent(
	ctx context.Context,
	clinicID, petID uint64,
	input SetNextVisitIntentInput,
) (*model.Reservation, error) {
	if err := validateNextVisitIntentDate(input.IntendedDate, time.Now()); err != nil {
		return nil, err
	}
	if s.tx == nil {
		return nil, apperrors.WrapInternalServerError("reservation transaction dependency is required")
	}
	var appointment *model.Reservation
	err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := ValidateReservationPetNotDeceased(ctx, s.repo, clinicID, &petID); err != nil {
			return err
		}
		var writeErr error
		appointment, writeErr = s.repo.SetNextVisitIntent(ctx, clinicID, petID, input)
		return writeErr
	})
	if err != nil {
		return nil, err
	}
	// EMR-308: tentative 書込で飼い主の次回来院が変わるため post-commit で再同期する。
	syncNextVisitTagBestEffort(ctx, s.nextVisitTagSync, clinicID, appointment.OwnerID)
	slog.InfoContext(ctx, "next visit intent set",
		slog.Uint64("reservation_id", appointment.ID),
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("pet_id", petID))
	return appointment, nil
}

// ClearNextVisitIntent はペットの active tentative を除去する。他ステータスの予約や
// 他医院のデータには触れない。対象が存在しない場合も冪等に成功する。
func (s *reservationService) ClearNextVisitIntent(ctx context.Context, clinicID, petID uint64) error {
	if s.tx == nil {
		return apperrors.WrapInternalServerError("reservation transaction dependency is required")
	}
	// EMR-308: tentative は削除後に owner を引けないため、削除前に飼い主を解決しておく
	// （post-commit の next_visit タグ除去同期に使う）。
	var ownerID uint64
	err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		// pet の clinic ownership を確認し、cross-clinic 呼び出しは存在しない扱いにする。
		resolved, err := s.repo.FindPetOwnerInClinic(ctx, clinicID, petID)
		if err != nil {
			return err
		}
		ownerID = resolved
		_, err = s.repo.ClearNextVisitIntent(ctx, clinicID, petID)
		return err
	})
	if err != nil {
		return err
	}
	syncNextVisitTagBestEffort(ctx, s.nextVisitTagSync, clinicID, &ownerID)
	return nil
}

// ConfirmTentative は tentative 予約を本予約へ遷移する。tentative には時刻・担当医が
// 未確定のため、ここで通常の booking 検証（時間帯不可・休診日・時間枠競合・区分容量・
// 担当医 capability）をまとめて適用する。確定後は通常予約と同じ挙動になる。
func (s *reservationService) ConfirmTentative(
	ctx context.Context,
	clinicID, id uint64,
	input *ConfirmTentativeInput,
) (*model.Reservation, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	if err := validateTimeRange(input.StartTime, input.EndTime); err != nil {
		return nil, err
	}
	targetStatus := model.ReservationStatusConfirmed
	if input.Status != nil {
		switch *input.Status {
		case model.ReservationStatusPending, model.ReservationStatusConfirmed:
			targetStatus = *input.Status
		default:
			return nil, apperrors.WrapInvalidInput("仮予約は pending または confirmed のみ確定できます")
		}
	}
	if s.tx == nil {
		return nil, apperrors.WrapInternalServerError("reservation transaction dependency is required")
	}
	var confirmed *model.Reservation
	err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		// 競合チェック→INSERT/UPDATE を直列化する clinic 単位 advisory lock（X-9 と同じ順序）。
		if err := s.repo.AcquireBookingLock(ctx, clinicID); err != nil {
			return err
		}
		current, err := s.repo.LockAndFindByID(ctx, clinicID, id)
		if err != nil {
			return err
		}
		if current.Status != model.ReservationStatusTentative {
			return apperrors.WrapConflict("対象の予約は仮予約（次回予定）ではありません")
		}
		resolvedTypeID := current.ReservationTypeID
		if input.ReservationTypeID != nil {
			resolvedTypeID = *input.ReservationTypeID
		}
		if s.typeRepo != nil {
			if _, err := s.typeRepo.FindByID(ctx, clinicID, resolvedTypeID); err != nil {
				return apperrors.Wrap(err, "failed to verify reservation type ownership")
			}
		}
		doctorID := normalizeCreateDoctorID(input.DoctorID)
		if err := ValidateReservationStaffCapability(ctx, s.reservationStaffRepo, clinicID, doctorID, resolvedTypeID); err != nil {
			return err
		}
		if err := ValidateReservationTypeAvailableTime(ctx, s.unavailableTimeRepo, clinicID, resolvedTypeID, input.StartTime, input.EndTime); err != nil {
			return err
		}
		if err := s.validateCreateClosedDaysIfConfigured(ctx, clinicID, input.StartTime); err != nil {
			return err
		}
		if s.holidayFinder != nil {
			if err := s.validateCreateClinicHoliday(ctx, clinicID, input.StartTime); err != nil {
				return err
			}
		}
		if err := CheckSlotConflict(ctx, s.repo, clinicID, doctorID, input.StartTime, input.EndTime, &id); err != nil {
			return err
		}
		if err := CheckReservationTypeCapacity(ctx, s.repo, s.typeRepo, clinicID, resolvedTypeID, input.StartTime, &id); err != nil {
			return err
		}
		var confirmErr error
		confirmed, confirmErr = s.repo.ConfirmTentative(ctx, clinicID, id, ConfirmTentativeUpdate{
			StartTime:         input.StartTime,
			EndTime:           input.EndTime,
			ReservationTypeID: resolvedTypeID,
			DoctorID:          doctorID,
			IsDesignated:      input.IsDesignated,
			Notes:             input.Notes,
			Status:            targetStatus,
		})
		return confirmErr
	})
	if err != nil {
		return nil, err
	}
	// EMR-308: tentative → 実予約確定で飼い主の次回来院の意味上の日付が変わるため
	// post-commit で next_visit タグを再同期する。
	syncNextVisitTagBestEffort(ctx, s.nextVisitTagSync, clinicID, confirmed.OwnerID)
	slog.InfoContext(ctx, "tentative appointment confirmed",
		slog.Uint64("reservation_id", confirmed.ID),
		slog.Uint64("clinic_id", clinicID))
	return confirmed, nil
}
