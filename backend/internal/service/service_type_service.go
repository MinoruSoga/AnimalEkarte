// Package service provides business logic implementations for ServiceType entity.
package service

import (
	"context"
	"fmt"

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
}

// UpdateServiceTypeInput はサービス種別更新のための入力データ（ポインタ型でゼロ値を区別する）
type UpdateServiceTypeInput struct {
	Name        *string
	Color       *string
	IsActive    *bool
	Description *string
	SortOrder   *int
}

// ---- DB column constants ----

const (
	colServiceTypeName        = "name"
	colServiceTypeColor       = "color"
	colServiceTypeIsActive    = "is_active"
	colServiceTypeDescription = "description"
	colServiceTypeSortOrder   = "sort_order"
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
	st := &model.ServiceType{
		ClinicID:    clinicID,
		Name:        input.Name,
		Color:       input.Color,
		IsActive:    input.IsActive,
		Description: input.Description,
		SortOrder:   input.SortOrder,
	}
	if err := s.repo.Create(ctx, st); err != nil {
		return nil, apperrors.Wrap(err, "failed to create service type")
	}
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
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *serviceTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
	exists, err := s.reservationRepo.ExistsByServiceTypeID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check reservation dependency: %w", err)
	}
	if exists {
		return apperrors.WrapAlreadyExists("service_type", "この項目は予約データで使用中のため削除できません")
	}
	return s.repo.Delete(ctx, clinicID, id)
}

func (s *serviceTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	return s.repo.Reorder(ctx, clinicID, ids)
}
