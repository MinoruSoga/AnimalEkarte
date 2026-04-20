// Package service provides business logic implementations for DiagnosisType and DiagnosisName entities.
package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- DiagnosisType Input DTOs ----

// CreateDiagnosisTypeInput はカテゴリ作成の入力DTO
type CreateDiagnosisTypeInput struct {
	Name        string
	IsActive    bool
	Description string
	SortOrder   int
}

// UpdateDiagnosisTypeInput はカテゴリ更新の入力DTO（nil = 未指定 = 更新しない）
type UpdateDiagnosisTypeInput struct {
	Name        *string
	IsActive    *bool
	Description *string
	SortOrder   *int
}

// ---- DiagnosisName Input DTOs ----

// CreateDiagnosisNameInput は診断名作成の入力DTO
type CreateDiagnosisNameInput struct {
	Name            string
	DiagnosisTypeID uint64
	IsActive        bool
	Description     string
	SortOrder       int
}

// UpdateDiagnosisNameInput は診断名更新の入力DTO（nil = 未指定 = 更新しない）
type UpdateDiagnosisNameInput struct {
	Name            *string
	DiagnosisTypeID *uint64
	IsActive        *bool
	Description     *string
	SortOrder       *int
}

// ---- 列名定数 (#021) ----

// 共通列名定数（DiagnosisType / DiagnosisName で共有）
const (
	colDiagnosisName        = "name"
	colDiagnosisIsActive    = "is_active"
	colDiagnosisDescription = "description"
	colDiagnosisSortOrder   = "sort_order"
)

// DiagnosisName 固有列名定数
const (
	colDiagnosisNameDiagnosisTypeID = "diagnosis_type_id"
)

// ---- DiagnosisTypeService ----

type DiagnosisTypeService interface {
	List(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisType, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisType, error)
	Create(ctx context.Context, clinicID uint64, input *CreateDiagnosisTypeInput) (*model.DiagnosisType, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateDiagnosisTypeInput) (*model.DiagnosisType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

// diagnosisTypeService
type diagnosisTypeService struct {
	repo repository.DiagnosisTypeRepository
}

func NewDiagnosisTypeService(
	repo repository.DiagnosisTypeRepository,
) DiagnosisTypeService {
	return &diagnosisTypeService{repo: repo}
}

func (s *diagnosisTypeService) List(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisType, int64, error) {
	items, total, err := s.repo.FindAll(ctx, clinicID, page, limit)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list diagnosis categories")
	}
	return items, total, nil
}

func (s *diagnosisTypeService) GetByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisType, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get diagnosis type")
	}
	return result, nil
}

func (s *diagnosisTypeService) Create(ctx context.Context, clinicID uint64, input *CreateDiagnosisTypeInput) (*model.DiagnosisType, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, err
	}
	diagType := &model.DiagnosisType{
		ClinicID:    clinicID,
		Name:        input.Name,
		IsActive:    input.IsActive,
		Description: input.Description,
		SortOrder:   input.SortOrder,
	}
	if err := s.repo.Create(ctx, diagType); err != nil {
		return nil, apperrors.Wrap(err, "failed to create diagnosis type")
	}
	slog.InfoContext(ctx, "diagnosis type created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("diagnosis_type_id", diagType.ID))
	return diagType, nil
}

func (s *diagnosisTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateDiagnosisTypeInput) (*model.DiagnosisType, error) {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to get diagnosis type")
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, err
	}
	fields := buildDiagnosisTypeUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
	}
	result, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update diagnosis type")
	}
	slog.InfoContext(ctx, "diagnosis type updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("diagnosis_type_id", id))
	return result, nil
}

func (s *diagnosisTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
	count, err := s.repo.CountNamesByCategoryID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check diagnosis type dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この診断カテゴリには診断名が登録されているため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete diagnosis type")
	}
	slog.InfoContext(ctx, "diagnosis type deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("diagnosis_type_id", id))
	return nil
}

func (s *diagnosisTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder diagnosis categories")
	}
	slog.InfoContext(ctx, "diagnosis_types reordered",
		slog.Uint64("clinic_id", clinicID),
		slog.Int("count", len(ids)))
	return nil
}

// buildDiagnosisTypeUpdateFields はポインタが非 nil のフィールドのみ map に追加する (#021: 定数使用)
func buildDiagnosisTypeUpdateFields(input *UpdateDiagnosisTypeInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colDiagnosisName] = *input.Name
	}
	if input.IsActive != nil {
		fields[colDiagnosisIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colDiagnosisDescription] = *input.Description
	}
	if input.SortOrder != nil {
		fields[colDiagnosisSortOrder] = *input.SortOrder
	}
	return fields
}

// ---- DiagnosisNameService ----

type DiagnosisNameService interface {
	List(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisName, int64, error)
	ListByCategoryID(ctx context.Context, clinicID, categoryID uint64, page, limit int) ([]model.DiagnosisName, int64, error)
	ListNames(ctx context.Context, clinicID uint64, typeID *uint64) ([]model.DiagnosisName, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisName, error)
	Create(ctx context.Context, clinicID uint64, input *CreateDiagnosisNameInput) (*model.DiagnosisName, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateDiagnosisNameInput) (*model.DiagnosisName, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

// diagnosisNameService (#020: typeRepo FK validation)
type diagnosisNameService struct {
	repo     repository.DiagnosisNameRepository
	typeRepo repository.DiagnosisTypeRepository
}

func NewDiagnosisNameService(
	repo repository.DiagnosisNameRepository,
	typeRepo repository.DiagnosisTypeRepository,
) DiagnosisNameService {
	return &diagnosisNameService{repo: repo, typeRepo: typeRepo}
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
		return nil, 0, apperrors.Wrap(err, "failed to list diagnosis names by type")
	}
	return items, total, nil
}

// ListNames はページネーションなしで診断名の一覧を返す (#418)。
// typeID が非 nil の場合は該当カテゴリのみ、nil の場合は全件を返す。
func (s *diagnosisNameService) ListNames(ctx context.Context, clinicID uint64, typeID *uint64) ([]model.DiagnosisName, error) {
	items, err := s.repo.FindAllActive(ctx, clinicID, typeID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list diagnosis names")
	}
	return items, nil
}

func (s *diagnosisNameService) GetByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisName, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get diagnosis name")
	}
	return result, nil
}

func (s *diagnosisNameService) Create(ctx context.Context, clinicID uint64, input *CreateDiagnosisNameInput) (*model.DiagnosisName, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, err
	}
	// #020: FK validation — diagnosis_type_id の存在確認
	if _, err := s.typeRepo.FindByID(ctx, clinicID, input.DiagnosisTypeID); err != nil {
		return nil, apperrors.WrapInvalidInput("診断カテゴリが見つかりません")
	}
	name := &model.DiagnosisName{
		ClinicID:        clinicID,
		Name:            input.Name,
		DiagnosisTypeID: input.DiagnosisTypeID,
		IsActive:        input.IsActive,
		Description:     input.Description,
		SortOrder:       input.SortOrder,
	}
	if err := s.repo.Create(ctx, name); err != nil {
		return nil, apperrors.Wrap(err, "failed to create diagnosis name")
	}
	slog.InfoContext(ctx, "diagnosis name created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("diagnosis_name_id", name.ID))
	return name, nil
}

func (s *diagnosisNameService) Update(ctx context.Context, clinicID, id uint64, input *UpdateDiagnosisNameInput) (*model.DiagnosisName, error) {
	// #424: 存在確認
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to get diagnosis name")
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, err
	}
	// #020: FK validation — diagnosis_type_id が変更される場合のみ確認
	if input.DiagnosisTypeID != nil {
		if _, err := s.typeRepo.FindByID(ctx, clinicID, *input.DiagnosisTypeID); err != nil {
			return nil, apperrors.WrapInvalidInput("診断カテゴリが見つかりません")
		}
	}
	fields := buildDiagnosisNameUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
	}
	result, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update diagnosis name")
	}
	slog.InfoContext(ctx, "diagnosis name updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("diagnosis_name_id", id))
	return result, nil
}

func (s *diagnosisNameService) Delete(ctx context.Context, clinicID, id uint64) error {
	count, err := s.repo.CountClinicalPlansByDiagnosisNameID(ctx, clinicID, id)
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
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("diagnosis_name_id", id))
	return nil
}

func (s *diagnosisNameService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder diagnosis names")
	}
	slog.InfoContext(ctx, "diagnosis_names reordered",
		slog.Uint64("clinic_id", clinicID),
		slog.Int("count", len(ids)))
	return nil
}

// buildDiagnosisNameUpdateFields はポインタが非 nil のフィールドのみ map に追加する (#426: DiagnosisName 専用定数使用)
func buildDiagnosisNameUpdateFields(input *UpdateDiagnosisNameInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colDiagnosisName] = *input.Name
	}
	if input.DiagnosisTypeID != nil {
		fields[colDiagnosisNameDiagnosisTypeID] = *input.DiagnosisTypeID
	}
	if input.IsActive != nil {
		fields[colDiagnosisIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colDiagnosisDescription] = *input.Description
	}
	if input.SortOrder != nil {
		fields[colDiagnosisSortOrder] = *input.SortOrder
	}
	return fields
}
