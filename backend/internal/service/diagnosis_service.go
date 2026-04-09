// Package service provides business logic implementations for DiagnosisCategory and DiagnosisName entities.
package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- 列名定数 (#021) ----

const (
	colDiagnosisCategoryName        = "name"
	colDiagnosisCategoryIsActive    = "is_active"
	colDiagnosisCategoryDescription = "description"
	colDiagnosisCategorySortOrder   = "sort_order"

	colDiagnosisNameName                = "name"
	colDiagnosisNameIsActive            = "is_active"
	colDiagnosisNameDescription         = "description"
	colDiagnosisNameSortOrder           = "sort_order"
	colDiagnosisNameDiagnosisCategoryID = "diagnosis_category_id"
)

// ---- DiagnosisCategory Input DTOs ----

// CreateDiagnosisCategoryInput はカテゴリ作成の入力DTO
type CreateDiagnosisCategoryInput struct {
	Name        string
	IsActive    bool
	Description string
	SortOrder   int
}

// UpdateDiagnosisCategoryInput はカテゴリ更新の入力DTO（nil = 未指定 = 更新しない）
type UpdateDiagnosisCategoryInput struct {
	Name        *string
	IsActive    *bool
	Description *string
	SortOrder   *int
}

// ---- DiagnosisName Input DTOs ----

// CreateDiagnosisNameInput は診断名作成の入力DTO
type CreateDiagnosisNameInput struct {
	Name                string
	DiagnosisCategoryID uint64
	IsActive            bool
	Description         string
	SortOrder           int
}

// UpdateDiagnosisNameInput は診断名更新の入力DTO（nil = 未指定 = 更新しない）
type UpdateDiagnosisNameInput struct {
	Name                *string
	DiagnosisCategoryID *uint64
	IsActive            *bool
	Description         *string
	SortOrder           *int
}

// ---- DiagnosisCategoryService ----

type DiagnosisCategoryService interface {
	List(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisCategory, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisCategory, error)
	Create(ctx context.Context, clinicID uint64, input *CreateDiagnosisCategoryInput) (*model.DiagnosisCategory, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateDiagnosisCategoryInput) (*model.DiagnosisCategory, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

// diagnosisCategoryService
type diagnosisCategoryService struct {
	repo repository.DiagnosisCategoryRepository
}

func NewDiagnosisCategoryService(
	repo repository.DiagnosisCategoryRepository,
) DiagnosisCategoryService {
	return &diagnosisCategoryService{repo: repo}
}

func (s *diagnosisCategoryService) List(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisCategory, int64, error) {
	items, total, err := s.repo.FindAll(ctx, clinicID, page, limit)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list diagnosis categories")
	}
	return items, total, nil
}

func (s *diagnosisCategoryService) GetByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisCategory, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get diagnosis category")
	}
	return result, nil
}

func (s *diagnosisCategoryService) Create(ctx context.Context, clinicID uint64, input *CreateDiagnosisCategoryInput) (*model.DiagnosisCategory, error) {
	category := &model.DiagnosisCategory{
		ClinicID:    clinicID,
		Name:        input.Name,
		IsActive:    input.IsActive,
		Description: input.Description,
		SortOrder:   input.SortOrder,
	}
	if err := s.repo.Create(ctx, category); err != nil {
		return nil, apperrors.Wrap(err, "failed to create diagnosis category")
	}
	slog.InfoContext(ctx, "diagnosis category created",
		slog.Uint64("category_id", category.ID),
		slog.Uint64("clinic_id", clinicID))
	return category, nil
}

func (s *diagnosisCategoryService) Update(ctx context.Context, clinicID, id uint64, input *UpdateDiagnosisCategoryInput) (*model.DiagnosisCategory, error) {
	fields := buildDiagnosisCategoryUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update diagnosis category")
	}
	slog.InfoContext(ctx, "diagnosis category updated",
		slog.Uint64("category_id", id),
		slog.Uint64("clinic_id", clinicID))
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get diagnosis category after update")
	}
	return result, nil
}

func (s *diagnosisCategoryService) Delete(ctx context.Context, clinicID, id uint64) error {
	count, err := s.repo.CountNamesByCategoryID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check diagnosis category dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この診断カテゴリには診断名が登録されているため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete diagnosis category")
	}
	slog.InfoContext(ctx, "diagnosis category deleted",
		slog.Uint64("category_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}

func (s *diagnosisCategoryService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder diagnosis categories")
	}
	return nil
}

// buildDiagnosisCategoryUpdateFields はポインタが非 nil のフィールドのみ map に追加する (#021: 定数使用)
func buildDiagnosisCategoryUpdateFields(input *UpdateDiagnosisCategoryInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colDiagnosisCategoryName] = *input.Name
	}
	if input.IsActive != nil {
		fields[colDiagnosisCategoryIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colDiagnosisCategoryDescription] = *input.Description
	}
	if input.SortOrder != nil {
		fields[colDiagnosisCategorySortOrder] = *input.SortOrder
	}
	return fields
}

// ---- DiagnosisNameService ----

type DiagnosisNameService interface {
	List(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisName, int64, error)
	ListByCategoryID(ctx context.Context, clinicID, categoryID uint64, page, limit int) ([]model.DiagnosisName, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisName, error)
	Create(ctx context.Context, clinicID uint64, input *CreateDiagnosisNameInput) (*model.DiagnosisName, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateDiagnosisNameInput) (*model.DiagnosisName, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

// diagnosisNameService (#020: categoryRepo FK validation)
type diagnosisNameService struct {
	repo         repository.DiagnosisNameRepository
	categoryRepo repository.DiagnosisCategoryRepository
}

func NewDiagnosisNameService(
	repo repository.DiagnosisNameRepository,
	categoryRepo repository.DiagnosisCategoryRepository,
) DiagnosisNameService {
	return &diagnosisNameService{repo: repo, categoryRepo: categoryRepo}
}

func (s *diagnosisNameService) List(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisName, int64, error) {
	items, total, err := s.repo.FindAll(ctx, clinicID, page, limit)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list diagnosis names")
	}
	return items, total, nil
}

func (s *diagnosisNameService) ListByCategoryID(ctx context.Context, clinicID, categoryID uint64, page, limit int) ([]model.DiagnosisName, int64, error) {
	items, total, err := s.repo.FindByCategoryID(ctx, clinicID, categoryID, page, limit)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list diagnosis names by category")
	}
	return items, total, nil
}

func (s *diagnosisNameService) GetByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisName, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get diagnosis name")
	}
	return result, nil
}

func (s *diagnosisNameService) Create(ctx context.Context, clinicID uint64, input *CreateDiagnosisNameInput) (*model.DiagnosisName, error) {
	// #020: FK validation — diagnosis_category_id の存在確認
	if _, err := s.categoryRepo.FindByID(ctx, clinicID, input.DiagnosisCategoryID); err != nil {
		return nil, apperrors.WrapInvalidInput("診断カテゴリが見つかりません")
	}
	name := &model.DiagnosisName{
		ClinicID:            clinicID,
		Name:                input.Name,
		DiagnosisCategoryID: input.DiagnosisCategoryID,
		IsActive:            input.IsActive,
		Description:         input.Description,
		SortOrder:           input.SortOrder,
	}
	if err := s.repo.Create(ctx, name); err != nil {
		return nil, apperrors.Wrap(err, "failed to create diagnosis name")
	}
	slog.InfoContext(ctx, "diagnosis name created",
		slog.Uint64("name_id", name.ID),
		slog.Uint64("clinic_id", clinicID))
	return name, nil
}

func (s *diagnosisNameService) Update(ctx context.Context, clinicID, id uint64, input *UpdateDiagnosisNameInput) (*model.DiagnosisName, error) {
	// #020: FK validation — diagnosis_category_id が変更される場合のみ確認
	if input.DiagnosisCategoryID != nil {
		if _, err := s.categoryRepo.FindByID(ctx, clinicID, *input.DiagnosisCategoryID); err != nil {
			return nil, apperrors.WrapInvalidInput("診断カテゴリが見つかりません")
		}
	}
	fields := buildDiagnosisNameUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update diagnosis name")
	}
	slog.InfoContext(ctx, "diagnosis name updated",
		slog.Uint64("name_id", id),
		slog.Uint64("clinic_id", clinicID))
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get diagnosis name after update")
	}
	return result, nil
}

func (s *diagnosisNameService) Delete(ctx context.Context, clinicID, id uint64) error {
	count, err := s.repo.CountClinicalPlansByDiagnosisNameID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check diagnosis name dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この診断名は診療記録で使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete diagnosis name")
	}
	slog.InfoContext(ctx, "diagnosis name deleted",
		slog.Uint64("name_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}

func (s *diagnosisNameService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder diagnosis names")
	}
	return nil
}

// buildDiagnosisNameUpdateFields はポインタが非 nil のフィールドのみ map に追加する (#021: 定数使用)
func buildDiagnosisNameUpdateFields(input *UpdateDiagnosisNameInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colDiagnosisNameName] = *input.Name
	}
	if input.DiagnosisCategoryID != nil {
		fields[colDiagnosisNameDiagnosisCategoryID] = *input.DiagnosisCategoryID
	}
	if input.IsActive != nil {
		fields[colDiagnosisNameIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colDiagnosisNameDescription] = *input.Description
	}
	if input.SortOrder != nil {
		fields[colDiagnosisNameSortOrder] = *input.SortOrder
	}
	return fields
}
