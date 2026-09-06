package reservation

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- 予約可能開始時刻 ----

func (s *reservationTypeService) ListAvailableSlots(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeAvailableSlot, error) {
	if _, err := s.repo.FindByID(ctx, clinicID, reservationTypeID); err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation type")
	}
	items, err := s.availableSlotRepo.FindAll(ctx, clinicID, reservationTypeID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list available slots")
	}
	return items, nil
}

func (s *reservationTypeService) CreateAvailableSlot(ctx context.Context, clinicID, reservationTypeID uint64, input CreateAvailableSlotInput) (*model.ReservationTypeAvailableSlot, error) {
	if _, err := s.repo.FindByID(ctx, clinicID, reservationTypeID); err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation type")
	}
	if err := validateAvailableSlotInput(input); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate available slot input")
	}
	existing, err := s.availableSlotRepo.FindAll(ctx, clinicID, reservationTypeID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to check existing available slots")
	}
	if err := validateAvailableSlotNotDuplicated(existing, input); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate available slot not duplicated")
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}
	slot := &model.ReservationTypeAvailableSlot{
		ClinicID:          clinicID,
		ReservationTypeID: reservationTypeID,
		AvailableType:     model.AvailableSlotType(input.AvailableType),
		DayOfWeek:         input.DayOfWeek,
		SpecificDate:      input.SpecificDate,
		StartTime:         input.StartTime,
		IsActive:          isActive,
	}
	if err := s.availableSlotRepo.Create(ctx, slot); err != nil {
		return nil, apperrors.Wrap(err, "failed to create available slot")
	}
	slog.InfoContext(ctx, "available slot created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("reservation_type_id", reservationTypeID))
	return slot, nil
}

func (s *reservationTypeService) DeleteAvailableSlot(ctx context.Context, clinicID, reservationTypeID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, reservationTypeID); err != nil {
		return apperrors.Wrap(err, "failed to get reservation type")
	}
	if _, err := s.availableSlotRepo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "available slot not found")
	}
	if err := s.availableSlotRepo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete available slot")
	}
	slog.InfoContext(ctx, "available slot deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("reservation_type_id", reservationTypeID),
		slog.Uint64("available_slot_id", id))
	return nil
}

func validateAvailableSlotInput(input CreateAvailableSlotInput) error {
	switch input.AvailableType {
	case string(model.AvailableSlotTypeWeekly):
		if input.DayOfWeek == nil {
			return apperrors.WrapInvalidInput("曜日の指定は必須です")
		}
		if *input.DayOfWeek < 0 || *input.DayOfWeek > 6 {
			return apperrors.WrapInvalidInput("day_of_week は 0（日曜）から 6（土曜）の範囲で指定してください")
		}
	case string(model.AvailableSlotTypeSpecific):
		if input.SpecificDate == nil {
			return apperrors.WrapInvalidInput("specific タイプでは specific_date の指定は必須です")
		}
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("不正な予約可能枠種別です: %s", input.AvailableType))
	}
	if _, err := time.ParseInLocation("15:04", input.StartTime, time.Local); err != nil {
		return apperrors.WrapInvalidInput("start_time は HH:MM 形式で指定してください")
	}
	return nil
}

func validateAvailableSlotNotDuplicated(existing []model.ReservationTypeAvailableSlot, input CreateAvailableSlotInput) error {
	for i := range existing {
		if string(existing[i].AvailableType) != input.AvailableType || existing[i].StartTime != input.StartTime {
			continue
		}
		switch existing[i].AvailableType {
		case model.AvailableSlotTypeWeekly:
			if input.DayOfWeek != nil && existing[i].DayOfWeek != nil && *existing[i].DayOfWeek == *input.DayOfWeek {
				return apperrors.WrapConflict("指定した予約可能枠は既に登録されています")
			}
		case model.AvailableSlotTypeSpecific:
			if input.SpecificDate != nil && existing[i].SpecificDate != nil &&
				existing[i].SpecificDate.In(time.Local).Format(time.DateOnly) == input.SpecificDate.In(time.Local).Format(time.DateOnly) {
				return apperrors.WrapConflict("指定した予約可能枠は既に登録されています")
			}
		}
	}
	return nil
}

func HasActiveAvailableSlots(slots []model.ReservationTypeAvailableSlot) bool {
	for i := range slots {
		if slots[i].IsActive {
			return true
		}
	}
	return false
}

func FilterApplicableAvailableSlots(slots []model.ReservationTypeAvailableSlot, date time.Time) []model.ReservationTypeAvailableSlot {
	dateJST := date.In(config.JST)
	dateStr := dateJST.Format(time.DateOnly)
	dayOfWeek := int8(dateJST.Weekday()) //nolint:gosec // Weekday() は 0-6 を返すため int8 に安全に収まる
	result := make([]model.ReservationTypeAvailableSlot, 0, len(slots))
	for i := range slots {
		if !slots[i].IsActive {
			continue
		}
		switch slots[i].AvailableType {
		case model.AvailableSlotTypeWeekly:
			if slots[i].DayOfWeek != nil && *slots[i].DayOfWeek == dayOfWeek {
				result = append(result, slots[i])
			}
		case model.AvailableSlotTypeSpecific:
			if slots[i].SpecificDate != nil && slots[i].SpecificDate.In(time.Local).Format(time.DateOnly) == dateStr {
				result = append(result, slots[i])
			}
		}
	}
	return result
}
