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

// CreateCheckupTypeInput は定期健診種別作成の入力DTO
type CreateCheckupTypeInput struct {
	Name        string
	Price       *int64
	IsActive    bool
	Description string
	Interval    string
	TargetAge   string
	ParentID    *uint64
	SortOrder   int
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

const (
	colCheckupTypeName        = "name"
	colCheckupTypePrice       = "price"
	colCheckupTypeIsActive    = "is_active"
	colCheckupTypeDescription = "description"
	colCheckupTypeInterval    = "interval"
	colCheckupTypeTargetAge   = "target_age"
	colCheckupTypeParentID    = "parent_id"
	colCheckupTypeSortOrder   = "sort_order"
)

func buildCheckupTypeUpdate(input *UpdateCheckupTypeInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colCheckupTypeName] = *input.Name
	}
	if input.Price != nil {
		fields[colCheckupTypePrice] = *input.Price
	}
	if input.IsActive != nil {
		fields[colCheckupTypeIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colCheckupTypeDescription] = *input.Description
	}
	if input.Interval != nil {
		fields[colCheckupTypeInterval] = *input.Interval
	}
	if input.TargetAge != nil {
		fields[colCheckupTypeTargetAge] = *input.TargetAge
	}
	if input.ClearParentID {
		fields[colCheckupTypeParentID] = nil
	} else if input.ParentID != nil {
		fields[colCheckupTypeParentID] = *input.ParentID
	}
	if input.SortOrder != nil {
		fields[colCheckupTypeSortOrder] = *input.SortOrder
	}
	return fields
}

type CheckupTypeService interface {
	List(ctx context.Context, clinicID uint64) ([]model.CheckupType, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.CheckupType, error)
	Create(ctx context.Context, clinicID uint64, input *CreateCheckupTypeInput) (*model.CheckupType, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateCheckupTypeInput) (*model.CheckupType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type checkupTypeService struct {
	repo repository.CheckupTypeRepository
}

func NewCheckupTypeService(repo repository.CheckupTypeRepository) CheckupTypeService {
	return &checkupTypeService{repo: repo}
}

func (s *checkupTypeService) List(ctx context.Context, clinicID uint64) ([]model.CheckupType, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list checkup types")
	}
	return items, nil
}
func (s *checkupTypeService) GetByID(ctx context.Context, clinicID, id uint64) (*model.CheckupType, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get checkup type")
	}
	return result, nil
}
func (s *checkupTypeService) Create(ctx context.Context, clinicID uint64, input *CreateCheckupTypeInput) (*model.CheckupType, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, err
	}
	checkupType := &model.CheckupType{
		ClinicID:    clinicID,
		Name:        input.Name,
		Price:       input.Price,
		IsActive:    input.IsActive,
		Description: input.Description,
		Interval:    input.Interval,
		TargetAge:   input.TargetAge,
		ParentID:    input.ParentID,
		SortOrder:   input.SortOrder,
	}
	if err := s.repo.Create(ctx, checkupType); err != nil {
		return nil, apperrors.Wrap(err, "failed to create checkup type")
	}
	slog.InfoContext(ctx, "checkup type created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("checkup_type_id", checkupType.ID))
	return checkupType, nil
}
func (s *checkupTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateCheckupTypeInput) (*model.CheckupType, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to get checkup type")
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, err
	}
	fields := buildCheckupTypeUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
	}
	checkupType, err := s.repo.Update(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update checkup type")
	}
	slog.InfoContext(ctx, "checkup type updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("checkup_type_id", id))
	return checkupType, nil
}
func (s *checkupTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to get checkup type")
	}
	childCount, err := s.repo.CountChildrenByParentID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check checkup type children")
	}
	if childCount > 0 {
		return apperrors.WrapConflict("この定期健診種別にはサブ種別が登録されているため削除できません")
	}
	count, err := s.repo.CountUsageByCheckupTypeID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check checkup type dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この定期健診種別は健診記録で使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete checkup type")
	}
	slog.InfoContext(ctx, "checkup type deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("checkup_type_id", id))
	return nil
}

func (s *checkupTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder checkup types")
	}
	slog.InfoContext(ctx, "checkup type reordered",
		slog.Uint64("clinic_id", clinicID),
		slog.Int("count", len(ids)))
	return nil
}
