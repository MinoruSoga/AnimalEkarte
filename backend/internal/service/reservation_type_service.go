// Package service provides business logic implementations for ReservationType entity.
package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- Input DTOs ----

// CreateReservationTypeInput はサービス種別作成のための入力データ
type CreateReservationTypeInput struct {
	Name        string
	Color       string
	IsActive    bool
	Description string
	SortOrder   int
	GroupID     *uint64

	// LINE予約用フィールド
	ReservationDisplayName string
	DurationMinutes        int
	ShortName              string
	ShowShortName          bool
	ReservationVisible     bool
	ReservationComment     string
	ReservationImageURL    string
	ReservationDayOption   string
	IsInternal             bool
}

// UpdateReservationTypeInput はサービス種別更新のための入力データ（ポインタ型でゼロ値を区別する）
type UpdateReservationTypeInput struct {
	Name        *string
	Color       *string
	IsActive    *bool
	Description *string
	SortOrder   *int
	GroupID     *uint64

	// LINE予約用フィールド
	ReservationDisplayName *string
	DurationMinutes        *int
	ShortName              *string
	ShowShortName          *bool
	ReservationVisible     *bool
	ReservationComment     *string
	ReservationImageURL    *string
	ReservationDayOption   *string
	IsInternal             *bool
}

// ---- DB column constants ----

const (
	colReservationTypeName                = "name"
	colReservationTypeColor               = "color"
	colReservationTypeIsActive            = "is_active"
	colReservationTypeDescription         = "description"
	colReservationTypeSortOrder           = "sort_order"
	colReservationTypeReservationDispName = "reservation_display_name"
	colReservationTypeDurationMinutes     = "duration_minutes"
	colReservationTypeShortName           = "short_name"
	colReservationTypeShowShortName       = "show_short_name"
	colReservationTypeReservationVisible  = "reservation_visible"
	colReservationTypeReservationComment  = "reservation_comment"
	colReservationTypeReservationDayOpt   = "reservation_day_option"
	colReservationTypeIsInternal          = "is_internal"
	colReservationTypeReservationImageURL = "reservation_image_url"
)

// buildReservationTypeUpdateFields は UpdateReservationTypeInput から nil でないフィールドのみ map に変換する
func buildReservationTypeUpdateFields(input *UpdateReservationTypeInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colReservationTypeName] = *input.Name
	}
	if input.Color != nil {
		fields[colReservationTypeColor] = *input.Color
	}
	if input.IsActive != nil {
		fields[colReservationTypeIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colReservationTypeDescription] = *input.Description
	}
	if input.SortOrder != nil {
		fields[colReservationTypeSortOrder] = *input.SortOrder
	}
	if input.ReservationDisplayName != nil {
		fields[colReservationTypeReservationDispName] = *input.ReservationDisplayName
	}
	if input.DurationMinutes != nil {
		fields[colReservationTypeDurationMinutes] = *input.DurationMinutes
	}
	if input.ShortName != nil {
		fields[colReservationTypeShortName] = *input.ShortName
	}
	if input.ShowShortName != nil {
		fields[colReservationTypeShowShortName] = *input.ShowShortName
	}
	if input.ReservationVisible != nil {
		fields[colReservationTypeReservationVisible] = *input.ReservationVisible
	}
	if input.ReservationComment != nil {
		fields[colReservationTypeReservationComment] = *input.ReservationComment
	}
	if input.ReservationImageURL != nil {
		fields[colReservationTypeReservationImageURL] = *input.ReservationImageURL
	}
	if input.ReservationDayOption != nil {
		fields[colReservationTypeReservationDayOpt] = *input.ReservationDayOption
	}
	if input.IsInternal != nil {
		fields[colReservationTypeIsInternal] = *input.IsInternal
	}
	if input.GroupID != nil {
		fields["group_id"] = *input.GroupID
	}
	return fields
}

// ---- ReservationTypeService ----

type ReservationTypeService interface { //nolint:revive // ReservationType is a domain entity name, cannot avoid stutter
	List(ctx context.Context, clinicID uint64) ([]model.ReservationType, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error)
	Create(ctx context.Context, clinicID uint64, input *CreateReservationTypeInput) (*model.ReservationType, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationTypeInput) (*model.ReservationType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type reservationTypeService struct {
	repo            repository.ReservationTypeRepository
	reservationRepo repository.ReservationRepository
}

func NewReservationTypeService(repo repository.ReservationTypeRepository, reservationRepo repository.ReservationRepository) ReservationTypeService {
	return &reservationTypeService{repo: repo, reservationRepo: reservationRepo}
}

func (s *reservationTypeService) List(ctx context.Context, clinicID uint64) ([]model.ReservationType, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list service types")
	}
	return items, nil
}

func (s *reservationTypeService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get service type")
	}
	return result, nil
}

func (s *reservationTypeService) Create(ctx context.Context, clinicID uint64, input *CreateReservationTypeInput) (*model.ReservationType, error) {
	if err := validateMasterName(input.Name); err != nil {
		return nil, err
	}
	reservationDayOption := model.ReservationDayOption(input.ReservationDayOption)
	if reservationDayOption == "" {
		reservationDayOption = model.DayOptionNone
	}

	st := &model.ReservationType{
		ClinicID:               clinicID,
		Name:                   input.Name,
		Color:                  input.Color,
		IsActive:               input.IsActive,
		Description:            input.Description,
		SortOrder:              input.SortOrder,
		ReservationDisplayName: input.ReservationDisplayName,
		DurationMinutes:        input.DurationMinutes,
		ShortName:              input.ShortName,
		ShowShortName:          input.ShowShortName,
		ReservationVisible:     input.ReservationVisible,
		ReservationComment:     input.ReservationComment,
		ReservationImageURL:    input.ReservationImageURL,
		ReservationDayOption:   reservationDayOption,
		IsInternal:             input.IsInternal,
		GroupID:                input.GroupID,
	}
	if err := s.repo.Create(ctx, st); err != nil {
		return nil, apperrors.Wrap(err, "failed to create service type")
	}
	slog.InfoContext(ctx, "service type created", slog.Uint64("reservation_type_id", st.ID))
	return st, nil
}

func (s *reservationTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationTypeInput) (*model.ReservationType, error) {
	if err := validateOptionalMasterName(input.Name); err != nil {
		return nil, err
	}
	fields := buildReservationTypeUpdateFields(input)
	if len(fields) == 0 {
		result, err := s.repo.FindByID(ctx, clinicID, id)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to get service type")
		}
		return result, nil
	}
	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update service type")
	}
	slog.InfoContext(ctx, "service type updated", slog.Uint64("reservation_type_id", id))
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get service type after update")
	}
	return result, nil
}

func (s *reservationTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
	exists, err := s.reservationRepo.ExistsByReservationTypeID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check reservation dependency")
	}
	if exists {
		return apperrors.WrapConflict("この項目は予約データで使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete service type")
	}
	slog.InfoContext(ctx, "service type deleted", slog.Uint64("reservation_type_id", id), slog.Uint64("clinic_id", clinicID))
	return nil
}

func (s *reservationTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder service types")
	}
	return nil
}
