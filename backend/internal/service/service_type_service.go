// Package service provides business logic implementations for ServiceType entity.
package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- Input DTOs ----

// CreateServiceTypeInput はサービス種別作成のための入力データ
type CreateServiceTypeInput struct {
	Name        string
	Color       string
	IsActive    bool
	Description string
	SortOrder   int

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

// UpdateServiceTypeInput はサービス種別更新のための入力データ（ポインタ型でゼロ値を区別する）
type UpdateServiceTypeInput struct {
	Name        *string
	Color       *string
	IsActive    *bool
	Description *string
	SortOrder   *int

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
	colServiceTypeName                = "name"
	colServiceTypeColor               = "color"
	colServiceTypeIsActive            = "is_active"
	colServiceTypeDescription         = "description"
	colServiceTypeSortOrder           = "sort_order"
	colServiceTypeReservationDispName = "reservation_display_name"
	colServiceTypeDurationMinutes     = "duration_minutes"
	colServiceTypeShortName           = "short_name"
	colServiceTypeShowShortName       = "show_short_name"
	colServiceTypeReservationVisible  = "reservation_visible"
	colServiceTypeReservationComment  = "reservation_comment"
	colServiceTypeReservationDayOpt   = "reservation_day_option"
	colServiceTypeIsInternal          = "is_internal"
)

// buildServiceTypeUpdateFields は UpdateServiceTypeInput から nil でないフィールドのみ map に変換する
func buildServiceTypeUpdateFields(input *UpdateServiceTypeInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colServiceTypeName] = *input.Name
	}
	if input.Color != nil {
		fields[colServiceTypeColor] = *input.Color
	}
	if input.IsActive != nil {
		fields[colServiceTypeIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colServiceTypeDescription] = *input.Description
	}
	if input.SortOrder != nil {
		fields[colServiceTypeSortOrder] = *input.SortOrder
	}
	if input.ReservationDisplayName != nil {
		fields[colServiceTypeReservationDispName] = *input.ReservationDisplayName
	}
	if input.DurationMinutes != nil {
		fields[colServiceTypeDurationMinutes] = *input.DurationMinutes
	}
	if input.ShortName != nil {
		fields[colServiceTypeShortName] = *input.ShortName
	}
	if input.ShowShortName != nil {
		fields[colServiceTypeShowShortName] = *input.ShowShortName
	}
	if input.ReservationVisible != nil {
		fields[colServiceTypeReservationVisible] = *input.ReservationVisible
	}
	if input.ReservationComment != nil {
		fields[colServiceTypeReservationComment] = *input.ReservationComment
	}
	if input.ReservationImageURL != nil {
		fields["reservation_image_url"] = *input.ReservationImageURL
	}
	if input.ReservationDayOption != nil {
		fields[colServiceTypeReservationDayOpt] = *input.ReservationDayOption
	}
	if input.IsInternal != nil {
		fields[colServiceTypeIsInternal] = *input.IsInternal
	}
	return fields
}

// ---- ServiceTypeService ----

type ServiceTypeService interface { //nolint:revive // ServiceType is a domain entity name, cannot avoid stutter
	List(ctx context.Context, clinicID uint64) ([]model.ServiceType, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.ServiceType, error)
	Create(ctx context.Context, clinicID uint64, input *CreateServiceTypeInput) (*model.ServiceType, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateServiceTypeInput) (*model.ServiceType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type serviceTypeService struct {
	repo            repository.ServiceTypeRepository
	reservationRepo repository.ReservationRepository
}

func NewServiceTypeService(repo repository.ServiceTypeRepository, reservationRepo repository.ReservationRepository) ServiceTypeService {
	return &serviceTypeService{repo: repo, reservationRepo: reservationRepo}
}

func (s *serviceTypeService) List(ctx context.Context, clinicID uint64) ([]model.ServiceType, error) {
	return s.repo.FindAll(ctx, clinicID)
}

func (s *serviceTypeService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ServiceType, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *serviceTypeService) Create(ctx context.Context, clinicID uint64, input *CreateServiceTypeInput) (*model.ServiceType, error) {
	reservationDayOption := model.ReservationDayOption(input.ReservationDayOption)
	if reservationDayOption == "" {
		reservationDayOption = model.DayOptionNone
	}

	st := &model.ServiceType{
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
	}
	if err := s.repo.Create(ctx, st); err != nil {
		return nil, apperrors.Wrap(err, "failed to create service type")
	}
	slog.InfoContext(ctx, "service type created", slog.Uint64("service_type_id", st.ID))
	return st, nil
}

func (s *serviceTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateServiceTypeInput) (*model.ServiceType, error) {
	fields := buildServiceTypeUpdateFields(input)
	if len(fields) == 0 {
		return s.repo.FindByID(ctx, clinicID, id)
	}
	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update service type")
	}
	slog.InfoContext(ctx, "service type updated", slog.Uint64("service_type_id", id))
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *serviceTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
	exists, err := s.reservationRepo.ExistsByServiceTypeID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check reservation dependency")
	}
	if exists {
		return apperrors.WrapConflict("この項目は予約データで使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete service type")
	}
	slog.InfoContext(ctx, "service type deleted", slog.Uint64("service_type_id", id), slog.Uint64("clinic_id", clinicID))
	return nil
}

func (s *serviceTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	return s.repo.Reorder(ctx, clinicID, ids)
}
