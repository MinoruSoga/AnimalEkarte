// Package service provides business logic implementations for ExaminationType entity.
package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

const (
	colExamTypeName           = "name"
	colExamTypePrice          = "price"
	colExamTypeIsActive       = "is_active"
	colExamTypeDescription    = "description"
	colExamTypeParentID       = "parent_id"
	colExamTypeSortOrder      = "sort_order"
	colExamTypeIsNonInsurance = "is_non_insurance"
)

func buildExamTypeUpdate(input *UpdateExamTypeInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colExamTypeName] = *input.Name
	}
	if input.Price != nil {
		fields[colExamTypePrice] = *input.Price
	}
	if input.IsActive != nil {
		fields[colExamTypeIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colExamTypeDescription] = *input.Description
	}
	if input.ClearParentID {
		fields[colExamTypeParentID] = nil
	} else if input.ParentID != nil {
		fields[colExamTypeParentID] = *input.ParentID
	}
	if input.SortOrder != nil {
		fields[colExamTypeSortOrder] = *input.SortOrder
	}
	if input.IsNonInsurance != nil {
		fields[colExamTypeIsNonInsurance] = *input.IsNonInsurance
	}
	return fields
}

// ---- ExaminationTypeService ----

// CreateExamTypeInput は検査種別作成の入力DTO
type CreateExamTypeInput struct {
	Name           string
	Price          *int64
	IsActive       bool
	Description    string
	ParentID       *uint64
	SortOrder      int
	IsNonInsurance bool
}

// UpdateExamTypeInput は検査種別更新のサービス入力 DTO
type UpdateExamTypeInput struct {
	Name           *string
	Price          *int64
	IsActive       *bool
	Description    *string
	ParentID       *uint64
	ClearParentID  bool
	SortOrder      *int
	IsNonInsurance *bool
}

type ExaminationTypeService interface {
	List(ctx context.Context, clinicID uint64) ([]model.ExaminationType, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error)
	Create(ctx context.Context, clinicID uint64, input *CreateExamTypeInput) (*model.ExaminationType, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateExamTypeInput) (*model.ExaminationType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type examTypeService struct{ repo repository.ExamTypeRepository }

func NewExamTypeService(repo repository.ExamTypeRepository) ExaminationTypeService {
	return &examTypeService{repo: repo}
}

func (s *examTypeService) List(ctx context.Context, clinicID uint64) ([]model.ExaminationType, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list exam types", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list exam types")
	}
	return items, nil
}
func (s *examTypeService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get exam type", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get exam type")
	}
	return result, nil
}
func (s *examTypeService) Create(ctx context.Context, clinicID uint64, input *CreateExamTypeInput) (*model.ExaminationType, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	exType := &model.ExaminationType{
		ClinicID:       clinicID,
		Name:           input.Name,
		Price:          input.Price,
		IsActive:       input.IsActive,
		Description:    input.Description,
		ParentID:       input.ParentID,
		SortOrder:      input.SortOrder,
		IsNonInsurance: input.IsNonInsurance,
	}
	if err := s.repo.Create(ctx, exType); err != nil {
		slog.ErrorContext(ctx, "failed to create exam type", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to create exam type")
	}
	slog.InfoContext(ctx, "exam type created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("exam_type_id", exType.ID))
	return exType, nil
}
func (s *examTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateExamTypeInput) (*model.ExaminationType, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to get exam type", "error", err)
		return nil, apperrors.Wrap(err, "failed to get exam type")
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	fields := buildExamTypeUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
	}
	exType, err := s.repo.Update(ctx, clinicID, id, fields)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update exam type", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to update exam type")
	}
	slog.InfoContext(ctx, "exam type updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("exam_type_id", id))
	return exType, nil
}
func (s *examTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to find exam type")
	}
	childCount, err := s.repo.CountChildrenByParentID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check exam type children", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to check exam type children")
	}
	if childCount > 0 {
		return apperrors.WrapConflict("この検査種別にはサブ種別が登録されているため削除できません")
	}
	count, err := s.repo.CountUsageByExamTypeID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check exam type dependencies", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to check exam type dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この検査種別は検査記録で使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to delete exam type", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete exam type")
	}
	slog.InfoContext(ctx, "exam type deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("exam_type_id", id))
	return nil
}

func (s *examTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		slog.ErrorContext(ctx, "failed to reorder exam types", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to reorder exam types")
	}
	slog.InfoContext(ctx, "exam type reordered",
		slog.Uint64("clinic_id", clinicID),
		slog.Int("count", len(ids)))
	return nil
}
