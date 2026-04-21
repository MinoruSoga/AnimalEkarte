package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

const (
	colReservationTypeLiffName                 = "name"
	colReservationTypeLiffColor                = "color"
	colReservationTypeLiffDescription          = "description"
	colReservationTypeLiffSortOrder            = "sort_order"
	colReservationTypeLiffDurationMinutes      = "duration_minutes"
	colReservationTypeLiffShortName            = "short_name"
	colReservationTypeLiffShowShortName        = "show_short_name"
	colReservationTypeLiffReservationVisible   = "reservation_visible"
	colReservationTypeLiffReservationComment   = "reservation_comment"
	colReservationTypeLiffReservationDayOption = "reservation_day_option"
	colReservationTypeLiffIsInternal           = "is_internal"
)

func buildReservationTypeLiffUpdateFields(input *UpdateReservationTypeLiffInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colReservationTypeLiffName] = *input.Name
	}
	if input.Color != nil {
		fields[colReservationTypeLiffColor] = *input.Color
	}
	if input.Description != nil {
		fields[colReservationTypeLiffDescription] = *input.Description
	}
	if input.SortOrder != nil {
		fields[colReservationTypeLiffSortOrder] = *input.SortOrder
	}
	if input.DurationMinutes != nil {
		fields[colReservationTypeLiffDurationMinutes] = *input.DurationMinutes
	}
	if input.ShortName != nil {
		fields[colReservationTypeLiffShortName] = *input.ShortName
	}
	if input.ShowShortName != nil {
		fields[colReservationTypeLiffShowShortName] = *input.ShowShortName
	}
	if input.ReservationVisible != nil {
		fields[colReservationTypeLiffReservationVisible] = *input.ReservationVisible
	}
	if input.ReservationComment != nil {
		fields[colReservationTypeLiffReservationComment] = *input.ReservationComment
	}
	if input.ReservationDayOption != nil {
		fields[colReservationTypeLiffReservationDayOption] = *input.ReservationDayOption
	}
	if input.IsInternal != nil {
		fields[colReservationTypeLiffIsInternal] = *input.IsInternal
	}
	return fields
}

// CreateReservationTypeLiffInput は予約コース作成の入力データ
type CreateReservationTypeLiffInput struct {
	Name                 string
	Color                string
	Description          string
	SortOrder            int
	DurationMinutes      int
	ShortName            string
	ShowShortName        bool
	ReservationVisible   bool
	ReservationComment   string
	ReservationDayOption string
	IsInternal           bool
}

// UpdateReservationTypeLiffInput は予約コース更新の入力データ（ポインタ型でゼロ値を区別）
type UpdateReservationTypeLiffInput struct {
	Name                 *string
	Color                *string
	Description          *string
	SortOrder            *int
	DurationMinutes      *int
	ShortName            *string
	ShowShortName        *bool
	ReservationVisible   *bool
	ReservationComment   *string
	ReservationDayOption *string
	IsInternal           *bool
}

// ReservationTypeLiffService は予約コース（reservation_types）のビジネスロジックインターフェース
type ReservationTypeLiffService interface {
	List(ctx context.Context, clinicID uint64) ([]model.ReservationType, error)
	Create(ctx context.Context, clinicID uint64, input *CreateReservationTypeLiffInput) (*model.ReservationType, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationTypeLiffInput) (*model.ReservationType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	PatchStatus(ctx context.Context, clinicID, id uint64, isActive bool) (*model.ReservationType, error)
	PatchSortOrder(ctx context.Context, clinicID, id uint64, direction string) error
}

type reservationTypeLiffService struct {
	repo         repository.ReservationTypeLiffRepository
	resAdminRepo repository.ReservationAdminRepository
	resRepo      repository.ReservationQueryRepository
}

func NewReservationTypeLiffService(repo repository.ReservationTypeLiffRepository, resAdminRepo repository.ReservationAdminRepository, resRepo repository.ReservationQueryRepository) ReservationTypeLiffService {
	return &reservationTypeLiffService{repo: repo, resAdminRepo: resAdminRepo, resRepo: resRepo}
}

func (s *reservationTypeLiffService) List(ctx context.Context, clinicID uint64) ([]model.ReservationType, error) {
	result, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list reservation course", "error", err)
		return nil, apperrors.Wrap(err, "failed to list reservation course")
	}
	return result, nil
}

func (s *reservationTypeLiffService) Create(ctx context.Context, clinicID uint64, input *CreateReservationTypeLiffInput) (*model.ReservationType, error) {
	dayOption := model.ReservationDayOption(input.ReservationDayOption)
	if dayOption == "" {
		dayOption = model.DayOptionNone
	}
	st := &model.ReservationType{
		ClinicID:             clinicID,
		Name:                 input.Name,
		Color:                input.Color,
		IsActive:             true,
		Description:          input.Description,
		SortOrder:            input.SortOrder,
		DurationMinutes:      input.DurationMinutes,
		ShortName:            input.ShortName,
		ShowShortName:        input.ShowShortName,
		ReservationVisible:   input.ReservationVisible,
		ReservationComment:   input.ReservationComment,
		ReservationDayOption: dayOption,
		IsInternal:           input.IsInternal,
	}
	if err := s.repo.Create(ctx, st); err != nil {
		slog.ErrorContext(ctx, "failed to create reservation course", "error", err)
		return nil, apperrors.Wrap(err, "failed to create reservation course")
	}
	slog.InfoContext(ctx, "reservation course created",
		slog.Uint64("reservation_type_id", st.ID),
		slog.Uint64("clinic_id", clinicID))
	created, err := s.repo.FindByID(ctx, clinicID, st.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get reservation course after create", "error", err)
		return nil, apperrors.Wrap(err, "failed to get reservation course after create")
	}
	return created, nil
}

func (s *reservationTypeLiffService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationTypeLiffInput) (*model.ReservationType, error) {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to get reservation course", "error", err)
		return nil, apperrors.Wrap(err, "failed to get reservation course")
	}
	fields := buildReservationTypeLiffUpdateFields(input)
	if len(fields) == 0 {
		result, err := s.repo.FindByID(ctx, clinicID, id)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get reservation course", "error", err)
			return nil, apperrors.Wrap(err, "failed to get reservation course")
		}
		return result, nil
	}
	updated, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update reservation course", "error", err)
		return nil, apperrors.Wrap(err, "failed to update reservation course")
	}
	slog.InfoContext(ctx, "reservation course updated",
		slog.Uint64("reservation_type_id", id),
		slog.Uint64("clinic_id", clinicID))
	return updated, nil
}

func (s *reservationTypeLiffService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to get reservation course", "error", err)
		return apperrors.Wrap(err, "failed to get reservation course")
	}
	exists, err := s.resRepo.ExistsByReservationTypeID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check reservation dependency", "error", err)
		return apperrors.Wrap(err, "failed to check reservation dependency")
	}
	if exists {
		return apperrors.WrapConflict("この予約コースは予約データで使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to delete reservation course", "error", err)
		return apperrors.Wrap(err, "failed to delete reservation course")
	}
	slog.InfoContext(ctx, "reservation course deleted",
		slog.Uint64("reservation_type_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}

func (s *reservationTypeLiffService) PatchStatus(ctx context.Context, clinicID, id uint64, isActive bool) (*model.ReservationType, error) {
	result, err := s.repo.UpdateFields(ctx, clinicID, id, map[string]any{"is_active": isActive})
	if err != nil {
		slog.ErrorContext(ctx, "failed to patch status", "error", err)
		return nil, apperrors.Wrap(err, "failed to patch status")
	}
	return result, nil
}

func (s *reservationTypeLiffService) PatchSortOrder(ctx context.Context, clinicID, id uint64, direction string) error {
	if direction != "up" && direction != "down" {
		return apperrors.WrapInvalidInput("direction must be 'up' or 'down'")
	}
	if err := s.repo.SwapSortOrder(ctx, clinicID, id, direction); err != nil {
		slog.ErrorContext(ctx, "failed to reorder reservation course", "error", err)
		return apperrors.Wrap(err, "failed to reorder reservation course")
	}
	return nil
}
