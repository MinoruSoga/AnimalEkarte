// Package service provides business logic implementations for ExaminationType entity.
package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- ExamTypeService ----

// CreateExamTypeInput は検査種別作成の入力DTO
type CreateExamTypeInput struct {
	Name        string
	Price       *int64
	IsActive    bool
	Description string
	ParentID    *uint64
	SortOrder   int
}

type ExamTypeService interface {
	List(ctx context.Context, clinicID uint64) ([]model.ExaminationType, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error)
	Create(ctx context.Context, clinicID uint64, input *CreateExamTypeInput) (*model.ExaminationType, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateExamTypeInput) (*model.ExaminationType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type examTypeService struct{ repo repository.ExamTypeRepository }

func NewExamTypeService(repo repository.ExamTypeRepository) ExamTypeService {
	return &examTypeService{repo: repo}
}

func (s *examTypeService) List(ctx context.Context, clinicID uint64) ([]model.ExaminationType, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list exam types")
	}
	return items, nil
}
func (s *examTypeService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get exam type")
	}
	return result, nil
}
func (s *examTypeService) Create(ctx context.Context, clinicID uint64, input *CreateExamTypeInput) (*model.ExaminationType, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, err
	}
	exType := &model.ExaminationType{
		ClinicID:    clinicID,
		Name:        input.Name,
		Price:       input.Price,
		IsActive:    input.IsActive,
		Description: input.Description,
		ParentID:    input.ParentID,
		SortOrder:   input.SortOrder,
	}
	if err := s.repo.Create(ctx, exType); err != nil {
		return nil, apperrors.Wrap(err, "failed to create exam type")
	}
	slog.InfoContext(ctx, "exam type created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("exam_type_id", exType.ID))
	return exType, nil
}
func (s *examTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateExamTypeInput) (*model.ExaminationType, error) {
	if err := validateOptionalName(input.Name); err != nil {
		return nil, err
	}
	fields := buildExamTypeUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("少なくとも1つのフィールドを指定してください")
	}
	exType, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update exam type")
	}
	slog.InfoContext(ctx, "exam type updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("exam_type_id", id))
	return exType, nil
}
func (s *examTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
	childCount, err := s.repo.CountChildrenByParentID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check exam type children")
	}
	if childCount > 0 {
		return apperrors.WrapConflict("この検査種別にはサブ種別が登録されているため削除できません")
	}
	count, err := s.repo.CountUsageByExamTypeID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check exam type dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この検査種別は検査記録で使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete exam type")
	}
	slog.InfoContext(ctx, "exam type deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("exam_type_id", id))
	return nil
}

func (s *examTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("並び順のIDリストが空です")
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder exam types")
	}
	slog.InfoContext(ctx, "exam_types reordered",
		slog.Uint64("clinic_id", clinicID),
		slog.Int("count", len(ids)))
	return nil
}

// UpdateExamTypeInput は検査種別更新のサービス入力 DTO
type UpdateExamTypeInput struct {
	Name          *string
	Price         *int64
	IsActive      *bool
	Description   *string
	ParentID      *uint64
	ClearParentID bool
	SortOrder     *int
}

const (
	colExamTypeName        = "name"
	colExamTypePrice       = "price"
	colExamTypeIsActive    = "is_active"
	colExamTypeDescription = "description"
	colExamTypeParentID    = "parent_id"
	colExamTypeSortOrder   = "sort_order"
)

func buildExamTypeUpdateFields(input *UpdateExamTypeInput) map[string]any {
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
	return fields
}
