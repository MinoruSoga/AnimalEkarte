// Package service provides business logic implementations for ReservationCategory entity.
package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- Input DTOs ----

// CreateReservationCategoryInput はサービス種別作成のための入力データ
type CreateReservationCategoryInput struct {
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

// UpdateReservationCategoryInput はサービス種別更新のための入力データ（ポインタ型でゼロ値を区別する）
type UpdateReservationCategoryInput struct {
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
	colReservationCategoryName                = "name"
	colReservationCategoryColor               = "color"
	colReservationCategoryIsActive            = "is_active"
	colReservationCategoryDescription         = "description"
	colReservationCategorySortOrder           = "sort_order"
	colReservationCategoryReservationDispName = "reservation_display_name"
	colReservationCategoryDurationMinutes     = "duration_minutes"
	colReservationCategoryShortName           = "short_name"
	colReservationCategoryShowShortName       = "show_short_name"
	colReservationCategoryReservationVisible  = "reservation_visible"
	colReservationCategoryReservationComment  = "reservation_comment"
	colReservationCategoryReservationDayOpt   = "reservation_day_option"
	colReservationCategoryIsInternal          = "is_internal"
)

// buildReservationCategoryUpdateFields は UpdateReservationCategoryInput から nil でないフィールドのみ map に変換する
func buildReservationCategoryUpdateFields(input *UpdateReservationCategoryInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colReservationCategoryName] = *input.Name
	}
	if input.Color != nil {
		fields[colReservationCategoryColor] = *input.Color
	}
	if input.IsActive != nil {
		fields[colReservationCategoryIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colReservationCategoryDescription] = *input.Description
	}
	if input.SortOrder != nil {
		fields[colReservationCategorySortOrder] = *input.SortOrder
	}
	if input.ReservationDisplayName != nil {
		fields[colReservationCategoryReservationDispName] = *input.ReservationDisplayName
	}
	if input.DurationMinutes != nil {
		fields[colReservationCategoryDurationMinutes] = *input.DurationMinutes
	}
	if input.ShortName != nil {
		fields[colReservationCategoryShortName] = *input.ShortName
	}
	if input.ShowShortName != nil {
		fields[colReservationCategoryShowShortName] = *input.ShowShortName
	}
	if input.ReservationVisible != nil {
		fields[colReservationCategoryReservationVisible] = *input.ReservationVisible
	}
	if input.ReservationComment != nil {
		fields[colReservationCategoryReservationComment] = *input.ReservationComment
	}
	if input.ReservationImageURL != nil {
		fields["reservation_image_url"] = *input.ReservationImageURL
	}
	if input.ReservationDayOption != nil {
		fields[colReservationCategoryReservationDayOpt] = *input.ReservationDayOption
	}
	if input.IsInternal != nil {
		fields[colReservationCategoryIsInternal] = *input.IsInternal
	}
	if input.GroupID != nil {
		fields["group_id"] = *input.GroupID
	}
	return fields
}

// ---- ReservationCategoryService ----

type ReservationCategoryService interface { //nolint:revive // ReservationCategory is a domain entity name, cannot avoid stutter
	List(ctx context.Context, clinicID uint64) ([]model.ReservationCategory, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.ReservationCategory, error)
	Create(ctx context.Context, clinicID uint64, input *CreateReservationCategoryInput) (*model.ReservationCategory, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationCategoryInput) (*model.ReservationCategory, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type reservationCategoryService struct {
	repo            repository.ReservationCategoryRepository
	reservationRepo repository.ReservationRepository
}

func NewReservationCategoryService(repo repository.ReservationCategoryRepository, reservationRepo repository.ReservationRepository) ReservationCategoryService {
	return &reservationCategoryService{repo: repo, reservationRepo: reservationRepo}
}

func (s *reservationCategoryService) List(ctx context.Context, clinicID uint64) ([]model.ReservationCategory, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list service types")
	}
	return items, nil
}

func (s *reservationCategoryService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ReservationCategory, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get service type")
	}
	return result, nil
}

func (s *reservationCategoryService) Create(ctx context.Context, clinicID uint64, input *CreateReservationCategoryInput) (*model.ReservationCategory, error) {
	reservationDayOption := model.ReservationDayOption(input.ReservationDayOption)
	if reservationDayOption == "" {
		reservationDayOption = model.DayOptionNone
	}

	st := &model.ReservationCategory{
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
	slog.InfoContext(ctx, "service type created", slog.Uint64("reservation_category_id", st.ID))
	return st, nil
}

func (s *reservationCategoryService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationCategoryInput) (*model.ReservationCategory, error) {
	fields := buildReservationCategoryUpdateFields(input)
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
	slog.InfoContext(ctx, "service type updated", slog.Uint64("reservation_category_id", id))
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get service type after update")
	}
	return result, nil
}

func (s *reservationCategoryService) Delete(ctx context.Context, clinicID, id uint64) error {
	exists, err := s.reservationRepo.ExistsByReservationCategoryID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check reservation dependency")
	}
	if exists {
		return apperrors.WrapConflict("この項目は予約データで使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete service type")
	}
	slog.InfoContext(ctx, "service type deleted", slog.Uint64("reservation_category_id", id), slog.Uint64("clinic_id", clinicID))
	return nil
}

func (s *reservationCategoryService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder service types")
	}
	return nil
}
