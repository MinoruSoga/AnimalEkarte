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

type ProcedureService interface {
	List(ctx context.Context) ([]model.Procedure, error)
	GetByID(ctx context.Context, id uint64) (*model.Procedure, error)
	Create(ctx context.Context, procedure *model.Procedure) error
	Update(ctx context.Context, clinicID, id uint64, input *UpdateProcedureInput) (*model.Procedure, error)
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type procedureService struct {
	repo repository.ProcedureRepository
}

func NewProcedureService(repo repository.ProcedureRepository) ProcedureService {
	return &procedureService{repo: repo}
}

func (s *procedureService) List(ctx context.Context) ([]model.Procedure, error) {
	return s.repo.FindAll(ctx)
}
func (s *procedureService) GetByID(ctx context.Context, id uint64) (*model.Procedure, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *procedureService) Create(ctx context.Context, procedure *model.Procedure) error {
	return s.repo.Create(ctx, procedure)
}
func (s *procedureService) Update(ctx context.Context, clinicID, id uint64, input *UpdateProcedureInput) (*model.Procedure, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	fields := buildProcedureUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	procedure, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update procedure")
	}
	slog.InfoContext(ctx, "procedure updated", slog.Uint64("procedure_id", id))
	return procedure, nil
}
func (s *procedureService) Delete(ctx context.Context, id uint64) error {
	count, err := s.repo.CountUsageByProcedureID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check procedure dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この診療項目は診療記録で使用中のため削除できません")
	}
	return s.repo.Delete(ctx, id)
}

func (s *procedureService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	return s.repo.Reorder(ctx, clinicID, ids)
}

// UpdateProcedureInput は処置更新のサービス入力 DTO
type UpdateProcedureInput struct {
	Name          *string
	Price         *int64
	IsActive      *bool
	Description   *string
	Duration      *int
	Anesthesia    *model.AnesthesiaType
	ParentID      *uint64
	ClearParentID bool
	SortOrder     *int
	TaxType       *model.TaxType
	TaxRate       *float64
}

func buildProcedureUpdateFields(input *UpdateProcedureInput) map[string]any {
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
	if input.Duration != nil {
		fields["duration"] = *input.Duration
	}
	if input.Anesthesia != nil {
		fields["anesthesia"] = *input.Anesthesia
	}
	if input.ClearParentID {
		fields["parent_id"] = nil
	} else if input.ParentID != nil {
		fields["parent_id"] = *input.ParentID
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	if input.TaxType != nil {
		fields["tax_type"] = *input.TaxType
	}
	if input.TaxRate != nil {
		fields["tax_rate"] = *input.TaxRate
	}
	return fields
}
