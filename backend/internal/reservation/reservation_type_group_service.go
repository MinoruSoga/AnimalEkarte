package reservation

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

const (
	colReservationTypeGroupName      = "name"
	colReservationTypeGroupColor     = "color"
	colReservationTypeGroupSortOrder = "sort_order"
	colReservationTypeGroupIsActive  = "is_active"
)

func buildReservationTypeGroupUpdate(input *UpdateReservationTypeGroupInput) map[string]any {
	fields := map[string]any{}
	if input.Name != nil {
		fields[colReservationTypeGroupName] = *input.Name
	}
	if input.Color != nil {
		fields[colReservationTypeGroupColor] = *input.Color
	}
	if input.SortOrder != nil {
		fields[colReservationTypeGroupSortOrder] = *input.SortOrder
	}
	if input.IsActive != nil {
		fields[colReservationTypeGroupIsActive] = *input.IsActive
	}
	return fields
}

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

// defaultGroupColor はグループのデフォルトカラー（DB の default 値と一元管理）。
const defaultGroupColor = "#3B82F6"

// --- DB column constants ---

type ReservationTypeGroupService interface {
	List(ctx context.Context, clinicID uint64) ([]model.ReservationTypeGroup, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeGroup, error)
	Create(ctx context.Context, clinicID uint64, input *CreateReservationTypeGroupInput) (*model.ReservationTypeGroup, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationTypeGroupInput) (*model.ReservationTypeGroup, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type reservationTypeGroupService struct {
	repo ReservationTypeGroupRepository
}

func NewReservationTypeGroupService(repo ReservationTypeGroupRepository) ReservationTypeGroupService {
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
	if err := sharedkernel.ValidateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	color := input.Color
	if color == "" {
		color = defaultGroupColor
	}
	g := &model.ReservationTypeGroup{
		ClinicID:  clinicID,
		Name:      input.Name,
		Color:     color,
		SortOrder: input.SortOrder,
		IsActive:  input.IsActive,
	}
	if err := s.repo.Create(ctx, g); err != nil {
		return nil, apperrors.Wrap(err, "failed to create reservation_type_group")
	}
	slog.InfoContext(ctx, "reservation_type_group created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("reservation_type_group_id", g.ID),
		slog.String("name", g.Name))
	return g, nil
}

func (s *reservationTypeGroupService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationTypeGroupInput) (*model.ReservationTypeGroup, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgInputNotNil)
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation type group")
	}
	if err := sharedkernel.ValidateOptionalName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	fields := buildReservationTypeGroupUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgAtLeastOneField)
	}
	g, err := s.repo.Update(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update reservation_type_group")
	}
	slog.InfoContext(ctx, "reservation_type_group updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("reservation_type_group_id", id))
	return g, nil
}

func (s *reservationTypeGroupService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to get reservation type group")
	}
	count, err := s.repo.CountUsageByReservationTypeGroupID(ctx, clinicID, id)
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
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("reservation_type_group_id", id))
	return nil
}

func (s *reservationTypeGroupService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(sharedkernel.ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder reservation type groups")
	}
	slog.InfoContext(ctx, "reservation_type_groups reordered", slog.Uint64("clinic_id", clinicID), slog.Int("count", len(ids)))
	return nil
}
