// Package service provides business logic implementations for CheckupType entity.
package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- CheckupTypeService ----

type CheckupTypeService interface {
	List(ctx context.Context) ([]model.CheckupType, error)
	GetByID(ctx context.Context, id uint64) (*model.CheckupType, error)
	Create(ctx context.Context, checkupType *model.CheckupType) error
	Update(ctx context.Context, clinicID, id uint64, input UpdateCheckupTypeInput) (*model.CheckupType, error)
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type checkupTypeService struct {
	repo repository.CheckupTypeRepository
}

func NewCheckupTypeService(repo repository.CheckupTypeRepository) CheckupTypeService {
	return &checkupTypeService{repo: repo}
}

func (s *checkupTypeService) List(ctx context.Context) ([]model.CheckupType, error) {
	return s.repo.FindAll(ctx)
}
func (s *checkupTypeService) GetByID(ctx context.Context, id uint64) (*model.CheckupType, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *checkupTypeService) Create(ctx context.Context, checkupType *model.CheckupType) error {
	return s.repo.Create(ctx, checkupType)
}
func (s *checkupTypeService) Update(ctx context.Context, clinicID, id uint64, input UpdateCheckupTypeInput) (*model.CheckupType, error) {
	fields := buildCheckupTypeUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	checkupType, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update checkup type")
	}
	slog.InfoContext(ctx, "checkup type updated", slog.Uint64("checkup_type_id", id))
	return checkupType, nil
}
func (s *checkupTypeService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *checkupTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	return s.repo.Reorder(ctx, clinicID, ids)
}

// UpdateCheckupTypeInput はチェックアップ種別更新のサービス入力 DTO
type UpdateCheckupTypeInput struct {
	Name          *string
	Price         *int64
	IsActive      *bool
	Description   *string
	Interval      *string
	TargetAge     *string
	ParentID      *uint64
	ClearParentID bool
	SortOrder     *int
}

func buildCheckupTypeUpdateFields(input UpdateCheckupTypeInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.Price != nil {
		fields["price"] = input.Price
	}
	if input.IsActive != nil {
		fields["is_active"] = *input.IsActive
	}
	if input.Description != nil {
		fields["description"] = *input.Description
	}
	if input.Interval != nil {
		fields["interval"] = *input.Interval
	}
	if input.TargetAge != nil {
		fields["target_age"] = *input.TargetAge
	}
	if input.ClearParentID {
		fields["parent_id"] = nil
	} else if input.ParentID != nil {
		fields["parent_id"] = *input.ParentID
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	return fields
}
