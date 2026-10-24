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
	VisitType         *model.VisitType         // nil = tentative 行の既存値を維持。first/revisit のみ許可
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
	var prevOwnerID *uint64
	err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.repo.AcquireNextVisitIntentLock(ctx, clinicID, petID); err != nil {
			return err
		}
		if err := ValidateReservationPetNotDeceased(ctx, s.repo, clinicID, &petID); err != nil {
			return err
		}
		var writeErr error
		appointment, prevOwnerID, writeErr = s.repo.SetNextVisitIntent(ctx, clinicID, petID, input)
		return writeErr
	})
	if err != nil {
		return nil, err
	}
	// EMR-308: tentative 書込で飼い主の次回来院が変わるため post-commit で再同期する。
	// upsert で owner が差し替わった場合は旧飼い主側の stale タグも除去対象に含める。
	syncNextVisitTagsBestEffort(ctx, s.nextVisitTagSync, clinicID, appointment.OwnerID, prevOwnerID)
	slog.InfoContext(ctx, "next visit intent set",
		slog.Uint64("reservation_id", appointment.ID),
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("pet_id", petID))
	return appointment, nil
}

// ClearNextVisitIntent はペットの active tentative を除去する。他ステータスの予約や
// 他医院のデータには触れない。対象が存在しない場合も冪等に成功する。
func (s *reservationService) ClearNextVisitIntent(ctx context.Context, clinicID, petID uint64) error {
	return s.clearNextVisitIntent(ctx, clinicID, petID, nil)
}

func (s *reservationService) ClearNextVisitIntentIfMatches(ctx context.Context, clinicID, petID uint64, expected ClearNextVisitIntentExpectation) error {
	if expected.ReservationID == 0 || expected.UpdatedAt.IsZero() {
		return apperrors.WrapInvalidInput("intent identity and revision are required")
	}
	return s.clearNextVisitIntent(ctx, clinicID, petID, &expected)
}

func (s *reservationService) clearNextVisitIntent(ctx context.Context, clinicID, petID uint64, expected *ClearNextVisitIntentExpectation) error {
	if s.tx == nil {
		return apperrors.WrapInternalServerError("reservation transaction dependency is required")
	}
	// EMR-308: tentative は削除後に owner を引けないため、削除前に飼い主を解決しておく
	// （post-commit の next_visit タグ除去同期に使う）。
	var ownerID uint64
	var storedOwnerID *uint64
	err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.repo.AcquireNextVisitIntentLock(ctx, clinicID, petID); err != nil {
			return err
		}
		// pet の clinic ownership を確認し、cross-clinic 呼び出しは存在しない扱いにする。
		resolved, err := s.repo.FindPetOwnerInClinic(ctx, clinicID, petID)
		if err != nil {
			return err
		}
		ownerID = resolved
		_, storedOwnerID, err = s.repo.clearNextVisitIntentWithOwner(ctx, clinicID, petID, expected)
		return err
	})
	if err != nil {
		return err
	}
	syncNextVisitTagsBestEffort(ctx, s.nextVisitTagSync, clinicID, storedOwnerID, &ownerID)
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
	if input.VisitType != nil {
		switch *input.VisitType {
		case model.VisitTypeFirst, model.VisitTypeRevisit:
		default:
			return nil, apperrors.WrapInvalidInput("invalid visit_type")
		}
	}
	if s.tx == nil {
		return nil, apperrors.WrapInternalServerError("reservation transaction dependency is required")
	}
	var confirmed *model.Reservation
	var prevOwnerID *uint64
	err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		// 競合チェック→INSERT/UPDATE を直列化する clinic 単位 advisory lock（X-9 と同じ順序）。
		if err := s.repo.AcquireBookingLock(ctx, clinicID); err != nil {
			return err
		}
		// Codex #4: tentative mutation の lock 順序は intent advisory lock → 行ロック。
		// 非ロック read で pet を解決し（pet_id は不変列）、行ロック先取りによる
		// SetNextVisitIntent/ClearNextVisitIntent との AB-BA デッドロックを防ぐ。
		// Codex #16: pre-read は関連行ガードを通さない FindPetIDForIntentLock を使う。
		// 譲渡直後は行の owner_id が pet.owner_id と不整合で FindByID が行を除外するが、
		// ここで見失うと repo 側の owner 追従に辿り着かず not found で確定できない。
		petID, err := s.repo.FindPetIDForIntentLock(ctx, clinicID, id)
		if err != nil {
			return err
		}
		if petID != nil {
			if err := s.repo.AcquireNextVisitIntentLock(ctx, clinicID, *petID); err != nil {
				return err
			}
		}
		current, err := s.repo.LockAndFindByID(ctx, clinicID, id)
		if err != nil {
			return err
		}
		if !sharedkernel.PtrEquals(current.Status, model.ReservationStatusTentative) {
			return apperrors.WrapConflict("対象の予約は仮予約（次回予定）ではありません")
		}
		// Codex #16: post-commit の新旧 owner 双方をタグ再同期するため、昇格前の
		// 飼い主を保持する（repo 側がロック保持中に現飼い主へ owner_id を追従する）。
		prevOwnerID = current.OwnerID
		// EMR-308: tentative 行は死亡後も残るため、実予約化の境界でも他の予約書込経路と
		// 同じ死亡ペット拒否を適用する（行ロック取得後・booking 書込前）。
		if err := ValidateReservationPetNotDeceased(ctx, s.repo, clinicID, current.PetID); err != nil {
			return err
		}
		resolvedTypeID := current.ReservationTypeID
		if input.ReservationTypeID != nil {
			resolvedTypeID = *input.ReservationTypeID
		}
		if s.typeRepo != nil {
			reservationType, err := s.typeRepo.FindByID(ctx, clinicID, resolvedTypeID)
			if err != nil {
				return apperrors.Wrap(err, "failed to verify reservation type ownership")
			}
			// Codex #15: 無効化された予約区分では確定できない。
			if !reservationType.IsActive {
				return apperrors.WrapInvalidInput("無効化された予約区分は利用できません")
			}
			// Codex #6: migration fallback が general 以外の placeholder を残し得るため、
			// 区分未指定（= placeholder 維持）は stored 区分が general の場合のみ許可する。
			if input.ReservationTypeID == nil && reservationType.Category != model.ReservationTypeCategoryGeneral {
				return apperrors.WrapInvalidInput("次回予定のプレースホルダ区分では確定できません。予約区分を指定してください")
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
		// 休診日チェックは create 経路と同じく無条件で呼ぶ — validateClinicHoliday は
		// finder==nil で 500 fail-closed するため、組み立て漏れの静黙スキップを防ぐ。
		if err := s.validateCreateClinicHoliday(ctx, clinicID, input.StartTime); err != nil {
			return err
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
			VisitType:         input.VisitType,
		})
		return confirmErr
	})
	if err != nil {
		return nil, err
	}
	// EMR-308: tentative → 実予約確定で飼い主の次回来院の意味上の日付が変わるため
	// post-commit で next_visit タグを再同期する。Codex #16: repo が昇格時に owner を
	// 現飼い主へ追従するため、譲渡があった場合は旧飼い主側の stale タグも除去する。
	syncNextVisitTagsBestEffort(ctx, s.nextVisitTagSync, clinicID, confirmed.OwnerID, prevOwnerID)
	slog.InfoContext(ctx, "tentative appointment confirmed",
		slog.Uint64("reservation_id", confirmed.ID),
		slog.Uint64("clinic_id", clinicID))
	return confirmed, nil
}
