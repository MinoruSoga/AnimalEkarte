// Package medicalrecord provides procedure use cases.
package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
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

func buildProcedureUpdate(input *UpdateProcedureInput) map[string]any {
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
	setNullableUint64Field(fields, colProcedureParentID, input.ClearParentID, input.ParentID)
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
	repo       ProcedureRepository
	transactor Transactor
}

// NewProcedureService constructs ProcedureService with a Transactor for atomic delete (MRC-07).
func NewProcedureService(repo ProcedureRepository, transactor Transactor) ProcedureService {
	return &procedureService{repo: repo, transactor: transactor}
}

func (s *procedureService) List(ctx context.Context, clinicID uint64) ([]model.Procedure, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list procedures", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list procedures")
	}
	return items, nil
}
func (s *procedureService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Procedure, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get procedure", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get procedure")
	}
	return result, nil
}
func (s *procedureService) Create(ctx context.Context, clinicID uint64, input *CreateProcedureInput) (*model.Procedure, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	if err := validateNonNegativePrice(input.Price); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate non negative price")
	}
	if err := validateAnesthesiaType(input.Anesthesia); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate anesthesia type")
	}
	if input.TaxType != "" {
		if err := validateTaxType(input.TaxType); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate tax type")
		}
	}
	if err := s.validateParentOwnership(ctx, clinicID, input.ParentID); err != nil {
		return nil, err
	}
	// TaxType 変換: "" の場合はデフォルト "excluded" を使用 (BUG-379)
	taxType := model.TaxTypeExcluded
	if input.TaxType != "" {
		taxType = model.TaxType(input.TaxType)
	}
	taxRate := DefaultTaxRate
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
		if conflict := apperrors.AsNameUniqueConflict(
			err,
			input.Name,
			apperrors.ConstraintProcedureName,
			apperrors.CodeProcedureNameConflict,
		); conflict != nil {
			return nil, conflict
		}
		slog.ErrorContext(ctx, "failed to create procedure", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to create procedure")
	}
	slog.InfoContext(ctx, "procedure created", slog.Uint64("clinic_id", clinicID), slog.Uint64("procedure_id", procedure.ID))
	return procedure, nil
}
func (s *procedureService) Update(ctx context.Context, clinicID, id uint64, input *UpdateProcedureInput) (*model.Procedure, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(errMsgInputNotNil)
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to get procedure", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get procedure")
	}
	if err := s.validateParentOwnership(ctx, clinicID, input.ParentID); err != nil {
		return nil, err
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	if err := validateNonNegativePrice(input.Price); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate non negative price")
	}
	if input.Anesthesia != nil {
		if err := validateAnesthesiaType(*input.Anesthesia); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate anesthesia type")
		}
	}
	if input.TaxType != nil {
		if err := validateTaxType(*input.TaxType); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate tax type")
		}
	}
	fields := buildProcedureUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(errMsgAtLeastOneField)
	}
	procedure, err := s.repo.Update(ctx, clinicID, id, fields)
	if err != nil {
		nameForConflict := ""
		if input.Name != nil {
			nameForConflict = *input.Name
		}
		if conflict := apperrors.AsNameUniqueConflict(
			err,
			nameForConflict,
			apperrors.ConstraintProcedureName,
			apperrors.CodeProcedureNameConflict,
		); conflict != nil {
			return nil, conflict
		}
		slog.ErrorContext(ctx, "failed to update procedure", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to update procedure")
	}
	slog.InfoContext(ctx, "procedure updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("procedure_id", id))
	return procedure, nil
}
func (s *procedureService) Delete(ctx context.Context, clinicID, id uint64) error {
	// MRC-07: usage/children checks and soft-delete share one ambient transaction.
	if s.transactor == nil {
		return apperrors.WrapInternalServerError("procedure write transaction dependency is required")
	}
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if _, err := s.repo.FindByID(txCtx, clinicID, id); err != nil {
			return apperrors.Wrap(err, "failed to get procedure")
		}
		childCount, err := s.repo.CountChildrenByParentID(txCtx, clinicID, id)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to count procedure children", "error", err, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to count procedure children")
		}
		if childCount > 0 {
			return apperrors.WrapConflict("この処置は子処置が存在するため削除できません")
		}
		count, err := s.repo.CountUsageByProcedureID(txCtx, clinicID, id)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to check procedure dependencies", "error", err, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to check procedure dependencies")
		}
		if count > 0 {
			return apperrors.WrapConflict("この診療項目は診療記録で使用中のため削除できません")
		}
		if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
			slog.ErrorContext(txCtx, "failed to delete procedure", "error", err, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to delete procedure")
		}
		return nil
	}); err != nil {
		return err
	}
	slog.InfoContext(ctx, "procedure deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("procedure_id", id))
	return nil
}

func (s *procedureService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(errMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		slog.ErrorContext(ctx, "failed to reorder procedures", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to reorder procedures")
	}
	slog.InfoContext(ctx, "procedures reordered", slog.Uint64("clinic_id", clinicID), slog.Int("count", len(ids)))
	return nil
}

// validateParentOwnership verifies a request-supplied parent_id belongs to the caller's
// clinic before it is persisted (X-14 self-ref master FK guard).
func (s *procedureService) validateParentOwnership(ctx context.Context, clinicID uint64, parentID *uint64) error {
	return validateOwnedMasterFK(ctx, "parent procedure", clinicID, parentID,
		func(actx context.Context, cid, mid uint64) error {
			_, err := s.repo.FindByID(actx, cid, mid)
			return err
		})
}
