package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ReservationCourseService は予約コース（service_types）のビジネスロジックインターフェース
type ReservationCourseService interface {
	List(ctx context.Context, clinicID uint64) ([]model.ServiceType, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.ServiceType, error)
	Create(ctx context.Context, clinicID uint64, input *CreateReservationCourseInput) (*model.ServiceType, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationCourseInput) (*model.ServiceType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	PatchStatus(ctx context.Context, clinicID, id uint64, isActive bool) (*model.ServiceType, error)
	PatchSortOrder(ctx context.Context, clinicID, id uint64, direction string) error
}

// CreateReservationCourseInput は予約コース作成の入力データ
type CreateReservationCourseInput struct {
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

// UpdateReservationCourseInput は予約コース更新の入力データ（ポインタ型でゼロ値を区別）
type UpdateReservationCourseInput struct {
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

type reservationCourseService struct {
	repo         repository.ReservationCourseRepository
	resAdminRepo repository.ReservationAdminRepository
	resRepo      repository.ReservationRepository
}

func NewReservationCourseService(repo repository.ReservationCourseRepository, resAdminRepo repository.ReservationAdminRepository, resRepo repository.ReservationRepository) ReservationCourseService {
	return &reservationCourseService{repo: repo, resAdminRepo: resAdminRepo, resRepo: resRepo}
}

func (s *reservationCourseService) List(ctx context.Context, clinicID uint64) ([]model.ServiceType, error) {
	result, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list reservation course")
	}
	return result, nil
}

func (s *reservationCourseService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ServiceType, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation course")
	}
	return result, nil
}

func (s *reservationCourseService) Create(ctx context.Context, clinicID uint64, input *CreateReservationCourseInput) (*model.ServiceType, error) {
	dayOption := model.ReservationDayOption(input.ReservationDayOption)
	if dayOption == "" {
		dayOption = model.DayOptionNone
	}
	st := &model.ServiceType{
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
		return nil, apperrors.Wrap(err, "failed to create reservation course")
	}
	slog.InfoContext(ctx, "reservation course created",
		slog.Uint64("service_type_id", st.ID),
		slog.Uint64("clinic_id", clinicID))
	created, err := s.repo.FindByID(ctx, clinicID, st.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation course after create")
	}
	return created, nil
}

func (s *reservationCourseService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationCourseInput) (*model.ServiceType, error) {
	fields := buildReservationCourseUpdateFields(input)
	if len(fields) == 0 {
		result, err := s.repo.FindByID(ctx, clinicID, id)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to get reservation course")
		}
		return result, nil
	}
	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update reservation course")
	}
	updated, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation course after update")
	}
	slog.InfoContext(ctx, "reservation course updated",
		slog.Uint64("service_type_id", id),
		slog.Uint64("clinic_id", clinicID))
	return updated, nil
}

func (s *reservationCourseService) Delete(ctx context.Context, clinicID, id uint64) error {
	exists, err := s.resRepo.ExistsByServiceTypeID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check reservation dependency")
	}
	if exists {
		return apperrors.WrapConflict("この予約コースは予約データで使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete reservation course")
	}
	slog.InfoContext(ctx, "reservation course deleted",
		slog.Uint64("service_type_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}

func (s *reservationCourseService) PatchStatus(ctx context.Context, clinicID, id uint64, isActive bool) (*model.ServiceType, error) {
	if err := s.repo.Update(ctx, clinicID, id, map[string]any{"is_active": isActive}); err != nil {
		return nil, apperrors.Wrap(err, "failed to patch status")
	}
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation course after patch status")
	}
	return result, nil
}

func (s *reservationCourseService) PatchSortOrder(ctx context.Context, clinicID, id uint64, direction string) error {
	if direction != "up" && direction != "down" {
		return apperrors.WrapInvalidInput("direction must be 'up' or 'down'")
	}
	if err := s.repo.SwapSortOrder(ctx, clinicID, id, direction); err != nil {
		return apperrors.Wrap(err, "failed to reorder reservation course")
	}
	return nil
}

func buildReservationCourseUpdateFields(input *UpdateReservationCourseInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.Color != nil {
		fields["color"] = *input.Color
	}
	if input.Description != nil {
		fields["description"] = *input.Description
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	if input.DurationMinutes != nil {
		fields["duration_minutes"] = *input.DurationMinutes
	}
	if input.ShortName != nil {
		fields["short_name"] = *input.ShortName
	}
	if input.ShowShortName != nil {
		fields["show_short_name"] = *input.ShowShortName
	}
	if input.ReservationVisible != nil {
		fields["reservation_visible"] = *input.ReservationVisible
	}
	if input.ReservationComment != nil {
		fields["reservation_comment"] = *input.ReservationComment
	}
	if input.ReservationDayOption != nil {
		fields["reservation_day_option"] = *input.ReservationDayOption
	}
	if input.IsInternal != nil {
		fields["is_internal"] = *input.IsInternal
	}
	return fields
}
