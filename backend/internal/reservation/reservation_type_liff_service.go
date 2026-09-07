package reservation

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	colReservationTypeLiffName                 = "name"
	colReservationTypeLiffColor                = "color"
	colReservationTypeLiffDescription          = "description"
	colReservationTypeLiffSortOrder            = "sort_order"
	colReservationTypeLiffDurationMinutes      = "duration_minutes"
	colReservationTypeLiffMaxConcurrent        = "max_concurrent"
	colReservationTypeLiffShortName            = "short_name"
	colReservationTypeLiffShowShortName        = "show_short_name"
	colReservationTypeLiffReservationVisible   = "reservation_visible"
	colReservationTypeLiffReservationComment   = "reservation_comment"
	colReservationTypeLiffReservationDayOption = "reservation_day_option"
	colReservationTypeLiffIsInternal           = "is_internal"
	colReservationTypeLiffIsActive             = "is_active"
)

func buildReservationTypeLiffUpdate(input *UpdateReservationTypeLiffInput) map[string]any {
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
	if input.ClearMaxConcurrent {
		fields[colReservationTypeLiffMaxConcurrent] = nil
	} else if input.MaxConcurrent != nil {
		fields[colReservationTypeLiffMaxConcurrent] = *input.MaxConcurrent
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
	if input.IsActive != nil {
		fields[colReservationTypeLiffIsActive] = *input.IsActive
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
	MaxConcurrent        *int
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
	MaxConcurrent        *int
	ClearMaxConcurrent   bool
	ShortName            *string
	ShowShortName        *bool
	ReservationVisible   *bool
	ReservationComment   *string
	ReservationDayOption *string
	IsInternal           *bool
	IsActive             *bool
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
	repo    ReservationTypeLiffRepository
	resRepo reservationTypeUsageChecker
}

// NewReservationTypeLiffService は LIFF 予約コース service を構築する。
// （BE9-2C R①: 旧 resAdminRepo 引数は代入のみで全 file 使用ゼロの dead 依存のため除去）
func NewReservationTypeLiffService(repo ReservationTypeLiffRepository, resRepo reservationTypeUsageChecker) ReservationTypeLiffService {
	return &reservationTypeLiffService{repo: repo, resRepo: resRepo}
}

func (s *reservationTypeLiffService) List(ctx context.Context, clinicID uint64) ([]model.ReservationType, error) {
	result, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list reservation course")
	}
	return result, nil
}

func (s *reservationTypeLiffService) Create(ctx context.Context, clinicID uint64, input *CreateReservationTypeLiffInput) (*model.ReservationType, error) {
	if err := validateMaxConcurrent(input.MaxConcurrent); err != nil {
		return nil, err
	}
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
		MaxConcurrent:        input.MaxConcurrent,
		ShortName:            input.ShortName,
		ShowShortName:        input.ShowShortName,
		ReservationVisible:   input.ReservationVisible,
		ReservationComment:   input.ReservationComment,
		ReservationDayOption: dayOption,
		IsInternal:           input.IsInternal,
	}
	if err := s.repo.Create(ctx, st); err != nil {
		return nil, apperrors.Wrap(err, "failed to create reservation course")
	}
	// RSV-03: return the write result; Create populates ID and compensated flags.
	slog.InfoContext(ctx, "reservation course created",
		slog.Uint64("reservation_type_id", st.ID),
		slog.Uint64("clinic_id", clinicID))
	return st, nil
}

func (s *reservationTypeLiffService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationTypeLiffInput) (*model.ReservationType, error) {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation course")
	}
	if err := validateMaxConcurrent(input.MaxConcurrent); err != nil {
		return nil, err
	}
	fields := buildReservationTypeLiffUpdate(input)
	if len(fields) == 0 {
		result, err := s.repo.FindByID(ctx, clinicID, id)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to get reservation course")
		}
		return result, nil
	}
	updated, err := s.repo.Update(ctx, clinicID, id, *input)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update reservation course")
	}
	slog.InfoContext(ctx, "reservation course updated",
		slog.Uint64("reservation_type_id", id),
		slog.Uint64("clinic_id", clinicID))
	return updated, nil
}

func (s *reservationTypeLiffService) Delete(ctx context.Context, clinicID, id uint64) error {
	// RSV-07: lock master + re-check deps + soft delete in one transaction.
	if err := s.repo.DeleteWithDependencyChecks(ctx, clinicID, id, s.resRepo); err != nil {
		return err
	}
	slog.InfoContext(ctx, "reservation course deleted",
		slog.Uint64("reservation_type_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}

func (s *reservationTypeLiffService) PatchStatus(ctx context.Context, clinicID, id uint64, isActive bool) (*model.ReservationType, error) {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation type")
	}
	result, err := s.repo.Update(ctx, clinicID, id, UpdateReservationTypeLiffInput{IsActive: &isActive})
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to patch status")
	}
	return result, nil
}

func (s *reservationTypeLiffService) PatchSortOrder(ctx context.Context, clinicID, id uint64, direction string) error {
	if direction != "up" && direction != "down" {
		return apperrors.WrapInvalidInput("direction must be 'up' or 'down'")
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to get reservation type")
	}
	if err := s.repo.UpdateSortOrder(ctx, clinicID, id, direction); err != nil {
		return apperrors.Wrap(err, "failed to reorder reservation course")
	}
	return nil
}
