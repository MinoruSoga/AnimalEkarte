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

type reservationCategoryGroupService struct {
	repo repository.ReservationTypeGroupRepository
}

func NewReservationTypeGroupService(repo repository.ReservationTypeGroupRepository) ReservationTypeGroupService {
	return &reservationCategoryGroupService{repo: repo}
}

func (s *reservationCategoryGroupService) List(ctx context.Context, clinicID uint64) ([]model.ReservationTypeGroup, error) {
	return s.repo.FindAll(ctx, clinicID)
}

func (s *reservationCategoryGroupService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeGroup, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *reservationCategoryGroupService) Create(ctx context.Context, clinicID uint64, input *CreateReservationTypeGroupInput) (*model.ReservationTypeGroup, error) {
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
	slog.InfoContext(ctx, "reservation_type_group created", "id", g.ID, "name", g.Name)
	return g, nil
}

func (s *reservationCategoryGroupService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationTypeGroupInput) (*model.ReservationTypeGroup, error) {
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
	if len(fields) == 0 {
		return s.repo.FindByID(ctx, clinicID, id)
	}
	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update reservation_type_group")
	}
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *reservationCategoryGroupService) Delete(ctx context.Context, clinicID, id uint64) error {
	count, err := s.repo.CountCategories(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to count categories in group")
	}
	if count > 0 {
		return apperrors.WrapConflict("このグループには予約区分が設定されています。先に予約区分のグループを変更してください。")
	}
	return s.repo.Delete(ctx, clinicID, id)
}

func (s *reservationCategoryGroupService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return s.repo.Reorder(ctx, clinicID, ids)
}
