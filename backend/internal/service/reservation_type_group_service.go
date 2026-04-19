package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type CreateReservationTypeGroupInput struct {
	Name      string
	Color     string
	SortOrder int
	IsActive  bool
}

type UpdateReservationTypeGroupInput struct {
	Name      *string
	Color     *string
	SortOrder *int
	IsActive  *bool
}

type ReservationTypeGroupService interface {
	List(ctx context.Context, clinicID uint64) ([]model.ReservationTypeGroup, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeGroup, error)
	Create(ctx context.Context, clinicID uint64, input *CreateReservationTypeGroupInput) (*model.ReservationTypeGroup, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationTypeGroupInput) (*model.ReservationTypeGroup, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type reservationTypeGroupService struct {
	repo repository.ReservationTypeGroupRepository
}

func NewReservationTypeGroupService(repo repository.ReservationTypeGroupRepository) ReservationTypeGroupService {
	return &reservationTypeGroupService{repo: repo}
}

func (s *reservationTypeGroupService) List(ctx context.Context, clinicID uint64) ([]model.ReservationTypeGroup, error) {
	groups, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list reservation type groups")
	}
	return groups, nil
}

func (s *reservationTypeGroupService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeGroup, error) {
	group, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation type group")
	}
	return group, nil
}

func (s *reservationTypeGroupService) Create(ctx context.Context, clinicID uint64, input *CreateReservationTypeGroupInput) (*model.ReservationTypeGroup, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, err
	}
	color := input.Color
	if color == "" {
		color = "#3B82F6"
	}
	g := &model.ReservationTypeGroup{
		ClinicID:  clinicID,
		Name:      input.Name,
		Color:     color,
		SortOrder: input.SortOrder,
		IsActive:  true,
	}
	if err := s.repo.Create(ctx, g); err != nil {
		return nil, apperrors.Wrap(err, "failed to create reservation_type_group")
	}
	slog.InfoContext(ctx, "reservation_type_group created",
		slog.Uint64("id", g.ID),
		slog.String("name", g.Name))
	return g, nil
}

func (s *reservationTypeGroupService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationTypeGroupInput) (*model.ReservationTypeGroup, error) {
	if err := validateOptionalName(input.Name); err != nil {
		return nil, err
	}
	fields := buildReservationTypeGroupUpdateFields(input)
	if len(fields) == 0 {
		g, err := s.repo.FindByID(ctx, clinicID, id)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to get reservation type group")
		}
		return g, nil
	}
	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update reservation_type_group")
	}
	g, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get updated reservation type group")
	}
	slog.InfoContext(ctx, "reservation_type_group updated",
		slog.Uint64("id", g.ID),
		slog.Uint64("clinic_id", clinicID))
	return g, nil
}

func buildReservationTypeGroupUpdateFields(input *UpdateReservationTypeGroupInput) map[string]any {
	fields := map[string]any{}
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.Color != nil {
		fields["color"] = *input.Color
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	if input.IsActive != nil {
		fields["is_active"] = *input.IsActive
	}
	return fields
}

func (s *reservationTypeGroupService) Delete(ctx context.Context, clinicID, id uint64) error {
	count, err := s.repo.CountCategories(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to count categories in group")
	}
	if count > 0 {
		return apperrors.WrapConflict("このグループには予約区分が設定されています。先に予約区分のグループを変更してください。")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete reservation type group")
	}
	slog.InfoContext(ctx, "reservation_type_group deleted",
		slog.Uint64("id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}

func (s *reservationTypeGroupService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder reservation type groups")
	}
	return nil
}
