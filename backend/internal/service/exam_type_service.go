// Package service provides business logic implementations for ExaminationType entity.
package service

import (
	"context"
	"fmt"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- ExamTypeService ----

type ExamTypeService interface {
	List(ctx context.Context) ([]model.ExaminationType, error)
	GetByID(ctx context.Context, id uint64) (*model.ExaminationType, error)
	Create(ctx context.Context, exType *model.ExaminationType) error
	Update(ctx context.Context, clinicID, id uint64, input UpdateExamTypeInput) (*model.ExaminationType, error)
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type examTypeService struct{ repo repository.ExamTypeRepository }

func NewExamTypeService(repo repository.ExamTypeRepository) ExamTypeService {
	return &examTypeService{repo: repo}
}

func (s *examTypeService) List(ctx context.Context) ([]model.ExaminationType, error) {
	return s.repo.FindAll(ctx)
}
func (s *examTypeService) GetByID(ctx context.Context, id uint64) (*model.ExaminationType, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *examTypeService) Create(ctx context.Context, exType *model.ExaminationType) error {
	return s.repo.Create(ctx, exType)
}
func (s *examTypeService) Update(ctx context.Context, clinicID, id uint64, input UpdateExamTypeInput) (*model.ExaminationType, error) {
	fields := buildExamTypeUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	exType, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
	if err != nil {
		return nil, fmt.Errorf("failed to update exam type: %w", err)
	}
	slog.InfoContext(ctx, "exam type updated", slog.Uint64("exam_type_id", id))
	return exType, nil
}
func (s *examTypeService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *examTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	return s.repo.Reorder(ctx, clinicID, ids)
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

func buildExamTypeUpdateFields(input UpdateExamTypeInput) map[string]any {
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
