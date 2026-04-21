// Package service provides business logic implementations for Procedure entity.
package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- ProcedureService ----

// CreateProcedureInput は診療項目作成のサービス入力 DTO
type CreateProcedureInput struct {
	Name        string
	Price       *int64
	IsActive    bool
	Description string
	Duration    *int
	Anesthesia  string
	ParentID    *uint64
	SortOrder   int
	TaxType     string   // "" = "excluded" (default), 変換はサービス層で行う (BUG-379)
	TaxRate     *float64 // nil = 0.10 (default)
}

// UpdateProcedureInput は処置更新のサービス入力 DTO
type UpdateProcedureInput struct {
	Name          *string
	Price         *int64
	IsActive      *bool
	Description   *string
	Duration      *int
	Anesthesia    *string
	ParentID      *uint64
	ClearParentID bool
	SortOrder     *int
	TaxType       *string
	TaxRate       *float64
}

const (
	colProcedureName        = "name"
	colProcedurePrice       = "price"
	colProcedureIsActive    = "is_active"
	colProcedureDescription = "description"
	colProcedureDuration    = "duration"
	colProcedureAnesthesia  = "anesthesia"
	colProcedureParentID    = "parent_id"
	colProcedureSortOrder   = "sort_order"
	colProcedureTaxType     = "tax_type"
	colProcedureTaxRate     = "tax_rate"
)

func buildProcedureUpdateFields(input *UpdateProcedureInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colProcedureName] = *input.Name
	}
	if input.Price != nil {
		fields[colProcedurePrice] = *input.Price
	}
	if input.IsActive != nil {
		fields[colProcedureIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colProcedureDescription] = *input.Description
	}
	if input.Duration != nil {
		fields[colProcedureDuration] = *input.Duration
	}
	if input.Anesthesia != nil {
		fields[colProcedureAnesthesia] = model.AnesthesiaType(*input.Anesthesia)
	}
	if input.ClearParentID {
		fields[colProcedureParentID] = nil
	} else if input.ParentID != nil {
		fields[colProcedureParentID] = *input.ParentID
	}
	if input.SortOrder != nil {
		fields[colProcedureSortOrder] = *input.SortOrder
	}
	if input.TaxType != nil {
		fields[colProcedureTaxType] = model.TaxType(*input.TaxType)
	}
	if input.TaxRate != nil {
		fields[colProcedureTaxRate] = *input.TaxRate
	}
	return fields
}

type ProcedureService interface {
	List(ctx context.Context, clinicID uint64) ([]model.Procedure, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Procedure, error)
	Create(ctx context.Context, clinicID uint64, input *CreateProcedureInput) (*model.Procedure, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateProcedureInput) (*model.Procedure, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type procedureService struct {
	repo repository.ProcedureRepository
}

func NewProcedureService(repo repository.ProcedureRepository) ProcedureService {
	return &procedureService{repo: repo}
}

func (s *procedureService) List(ctx context.Context, clinicID uint64) ([]model.Procedure, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list procedures")
	}
	return items, nil
}
func (s *procedureService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Procedure, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get procedure")
	}
	return result, nil
}
func (s *procedureService) Create(ctx context.Context, clinicID uint64, input *CreateProcedureInput) (*model.Procedure, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, err
	}
	if err := validateNonNegativePrice(input.Price); err != nil {
		return nil, err
	}
	if err := validateAnesthesiaType(input.Anesthesia); err != nil {
		return nil, err
	}
	if input.TaxType != "" {
		if err := validateTaxType(input.TaxType); err != nil {
			return nil, err
		}
	}
	// TaxType 変換: "" の場合はデフォルト "excluded" を使用 (BUG-379)
	taxType := model.TaxTypeExcluded
	if input.TaxType != "" {
		taxType = model.TaxType(input.TaxType)
	}
	taxRate := 0.10
	if input.TaxRate != nil {
		taxRate = *input.TaxRate
	}
	procedure := &model.Procedure{
		ClinicID:    clinicID,
		Name:        input.Name,
		Price:       input.Price,
		IsActive:    input.IsActive,
		Description: input.Description,
		Duration:    input.Duration,
		SortOrder:   input.SortOrder,
		TaxType:     taxType,
		TaxRate:     taxRate,
	}
	if input.Anesthesia != "" {
		procedure.Anesthesia = model.AnesthesiaType(input.Anesthesia)
	}
	if input.ParentID != nil {
		procedure.ParentID = input.ParentID
	}
	if err := s.repo.Create(ctx, procedure); err != nil {
		return nil, apperrors.Wrap(err, "failed to create procedure")
	}
	slog.InfoContext(ctx, "procedure created", slog.Uint64("clinic_id", clinicID), slog.Uint64("procedure_id", procedure.ID))
	return procedure, nil
}
func (s *procedureService) Update(ctx context.Context, clinicID, id uint64, input *UpdateProcedureInput) (*model.Procedure, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to get procedure")
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, err
	}
	if err := validateNonNegativePrice(input.Price); err != nil {
		return nil, err
	}
	if input.Anesthesia != nil {
		if err := validateAnesthesiaType(*input.Anesthesia); err != nil {
			return nil, err
		}
	}
	if input.TaxType != nil {
		if err := validateTaxType(*input.TaxType); err != nil {
			return nil, err
		}
	}
	fields := buildProcedureUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
	}
	procedure, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update procedure")
	}
	slog.InfoContext(ctx, "procedure updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("procedure_id", id))
	return procedure, nil
}
func (s *procedureService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to get procedure")
	}
	childCount, err := s.repo.CountChildrenByParentID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to count procedure children")
	}
	if childCount > 0 {
		return apperrors.WrapConflict("この処置は子処置が存在するため削除できません")
	}
	count, err := s.repo.CountUsageByProcedureID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check procedure dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この診療項目は診療記録で使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete procedure")
	}
	slog.InfoContext(ctx, "procedure deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("procedure_id", id))
	return nil
}

func (s *procedureService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder procedures")
	}
	slog.InfoContext(ctx, "procedures reordered", slog.Uint64("clinic_id", clinicID), slog.Int("count", len(ids)))
	return nil
}
