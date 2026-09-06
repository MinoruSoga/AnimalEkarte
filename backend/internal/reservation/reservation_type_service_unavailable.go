package reservation

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- 予約不可時間 ----

func (s *reservationTypeService) ListUnavailableTimes(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeUnavailableTime, error) {
	// 予約区分の存在確認
	if _, err := s.repo.FindByID(ctx, clinicID, reservationTypeID); err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation type")
	}
	items, err := s.unavailableTimeRepo.FindAll(ctx, clinicID, reservationTypeID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list unavailable times")
	}
	return items, nil
}

func (s *reservationTypeService) CreateUnavailableTime(ctx context.Context, clinicID, reservationTypeID uint64, input CreateUnavailableTimeInput) (*model.ReservationTypeUnavailableTime, error) {
	// 予約区分の存在確認
	if _, err := s.repo.FindByID(ctx, clinicID, reservationTypeID); err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation type")
	}
	// 種別バリデーション
	if err := validateUnavailableTimeInput(input); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate unavailable time input")
	}
	// 重複チェック
	existing, err := s.unavailableTimeRepo.FindAll(ctx, clinicID, reservationTypeID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to check existing unavailable times")
	}
	if err := validateUnavailableTimeNotOverlaps(existing, input); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate unavailable time not overlaps")
	}

	t := &model.ReservationTypeUnavailableTime{
		ClinicID:          clinicID,
		ReservationTypeID: reservationTypeID,
		UnavailableType:   model.UnavailableType(input.UnavailableType),
		DayOfWeek:         input.DayOfWeek,
		SpecificDate:      input.SpecificDate,
		StartTime:         input.StartTime,
		EndTime:           input.EndTime,
	}
	if err := s.unavailableTimeRepo.Create(ctx, t); err != nil {
		return nil, apperrors.Wrap(err, "failed to create unavailable time")
	}
	slog.InfoContext(ctx, "unavailable time created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("reservation_type_id", reservationTypeID))
	return t, nil
}

func (s *reservationTypeService) DeleteUnavailableTime(ctx context.Context, clinicID, reservationTypeID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, reservationTypeID); err != nil {
		return apperrors.Wrap(err, "failed to get reservation type")
	}
	if _, err := s.unavailableTimeRepo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "unavailable time not found")
	}
	if err := s.unavailableTimeRepo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete unavailable time")
	}
	slog.InfoContext(ctx, "unavailable time deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("reservation_type_id", reservationTypeID),
		slog.Uint64("unavailable_time_id", id))
	return nil
}

// ---- バリデーションヘルパー ----

// validateUnavailableTimeInput は CreateUnavailableTimeInput の整合性を検証する
func validateUnavailableTimeInput(input CreateUnavailableTimeInput) error {
	switch input.UnavailableType {
	case string(model.UnavailableTypeWeekly):
		if input.DayOfWeek == nil {
			return apperrors.WrapInvalidInput("曜日の指定は必須です")
		}
		if *input.DayOfWeek < 0 || *input.DayOfWeek > 6 {
			return apperrors.WrapInvalidInput("day_of_week は 0（日曜）から 6（土曜）の範囲で指定してください")
		}
	case string(model.UnavailableTypeSpecific):
		if input.SpecificDate == nil {
			return apperrors.WrapInvalidInput("specific タイプでは specific_date の指定は必須です")
		}
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("不正な予約不可時間種別です: %s", input.UnavailableType))
	}
	// StartTime < EndTime（VARCHAR(5) "HH:MM" は辞書順比較で正しく機能する）
	if input.StartTime >= input.EndTime {
		return apperrors.WrapInvalidInput("開始時刻は終了時刻より前に指定してください")
	}
	return nil
}

// validateUnavailableTimeNotOverlaps は既存設定との時間帯重複を検証する
// weekly と specific の混在は許可（LIFF で specific が weekly より優先される）
func validateUnavailableTimeNotOverlaps(existing []model.ReservationTypeUnavailableTime, input CreateUnavailableTimeInput) error {
	for i := range existing {
		if string(existing[i].UnavailableType) != input.UnavailableType {
			continue
		}
		switch existing[i].UnavailableType {
		case model.UnavailableTypeWeekly:
			if input.DayOfWeek == nil || existing[i].DayOfWeek == nil || *existing[i].DayOfWeek != *input.DayOfWeek {
				continue
			}
		case model.UnavailableTypeSpecific:
			if input.SpecificDate == nil || existing[i].SpecificDate == nil {
				continue
			}
			if existing[i].SpecificDate.In(time.Local).Format(time.DateOnly) != input.SpecificDate.In(time.Local).Format(time.DateOnly) {
				continue
			}
		}
		// 時間帯が交差するか（start < other.end && end > other.start）
		if input.StartTime < existing[i].EndTime && input.EndTime > existing[i].StartTime {
			return apperrors.WrapConflict("指定した時間帯は既存の予約不可時間と重複しています")
		}
	}
	return nil
}
