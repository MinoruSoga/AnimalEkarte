package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *reservationTypeService) List(ctx context.Context, clinicID uint64) ([]model.ReservationType, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list reservation types", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list reservation types")
	}
	return items, nil
}

func (s *reservationTypeService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get reservation type", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get reservation type")
	}
	return result, nil
}

func (s *reservationTypeService) Create(ctx context.Context, clinicID uint64, input *CreateReservationTypeInput) (*model.ReservationType, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	reservationDayOption := model.ReservationDayOption(input.ReservationDayOption)
	if reservationDayOption == "" {
		reservationDayOption = model.DayOptionNone
	}
	durationMinutes := 15
	if input.DurationMinutes != nil {
		durationMinutes = *input.DurationMinutes
	}
	reservationVisible := true
	if input.ReservationVisible != nil {
		reservationVisible = *input.ReservationVisible
	}
	category := model.ReservationTypeCategory(input.Category)
	if category == "" {
		category = model.ReservationTypeCategoryGeneral
	}

	st := &model.ReservationType{
		ClinicID:               clinicID,
		Name:                   input.Name,
		Color:                  input.Color,
		IsActive:               input.IsActive,
		Description:            input.Description,
		SortOrder:              input.SortOrder,
		Category:               category,
		ReservationDisplayName: input.ReservationDisplayName,
		DurationMinutes:        durationMinutes,
		ShortName:              input.ShortName,
		ShowShortName:          input.ShowShortName,
		ReservationVisible:     reservationVisible,
		ReservationComment:     input.ReservationComment,
		ReservationImageURL:    input.ReservationImageURL,
		ReservationDayOption:   reservationDayOption,
		IsInternal:             input.IsInternal,
		GroupID:                input.GroupID,
	}
	if err := s.repo.Create(ctx, st); err != nil {
		slog.ErrorContext(ctx, "failed to create reservation type", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to create reservation type")
	}
	slog.InfoContext(ctx, "reservation type created", slog.Uint64("clinic_id", clinicID), slog.Uint64("reservation_type_id", st.ID))
	return st, nil
}

func (s *reservationTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationTypeInput) (*model.ReservationType, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to get reservation type", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get reservation type")
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	fields := buildReservationTypeUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
	}
	result, err := s.repo.Update(ctx, clinicID, id, fields)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update reservation type", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to update reservation type")
	}
	slog.InfoContext(ctx, "reservation type updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("reservation_type_id", id))
	return result, nil
}

func (s *reservationTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to find reservation type")
	}
	count, err := s.repo.CountUsageByReservationTypeID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check reservation type usage", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to check reservation type usage")
	}
	if count > 0 {
		return apperrors.WrapConflict("この項目は予約データで使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to delete reservation type", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete reservation type")
	}
	slog.InfoContext(ctx, "reservation type deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("reservation_type_id", id))
	return nil
}

func (s *reservationTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		slog.ErrorContext(ctx, "failed to reorder reservation types", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to reorder reservation types")
	}
	slog.InfoContext(ctx, "reservation type reordered", slog.Uint64("clinic_id", clinicID), slog.Int("count", len(ids)))
	return nil
}
