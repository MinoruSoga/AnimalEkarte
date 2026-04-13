// Package service provides business logic implementations for HospitalizationPlan entity.
package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- HospitalizationPlanService ----

type HospitalizationPlanService interface {
	List(ctx context.Context, clinicID uint64) ([]model.HospitalizationPlan, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.HospitalizationPlan, error)
	Create(ctx context.Context, plan *model.HospitalizationPlan) error
	Update(ctx context.Context, clinicID, id uint64, input UpdateHospitalizationPlanInput) (*model.HospitalizationPlan, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type hospitalizationPlanService struct {
	repo repository.HospitalizationPlanRepository
}

func NewHospitalizationPlanService(repo repository.HospitalizationPlanRepository) HospitalizationPlanService {
	return &hospitalizationPlanService{repo: repo}
}

func (s *hospitalizationPlanService) List(ctx context.Context, clinicID uint64) ([]model.HospitalizationPlan, error) {
	result, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list hospitalization plan")
	}
	return result, nil
}
func (s *hospitalizationPlanService) GetByID(ctx context.Context, clinicID, id uint64) (*model.HospitalizationPlan, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get hospitalization plan")
	}
	return result, nil
}
func (s *hospitalizationPlanService) Create(ctx context.Context, plan *model.HospitalizationPlan) error {
	if err := s.repo.Create(ctx, plan); err != nil {
		return apperrors.Wrap(err, "failed to create hospitalization plan")
	}
	slog.InfoContext(ctx, "hospitalization plan created",
		slog.Uint64("hospitalization_plan_id", plan.ID),
		slog.Uint64("clinic_id", plan.ClinicID))
	return nil
}
func (s *hospitalizationPlanService) Update(ctx context.Context, clinicID, id uint64, input UpdateHospitalizationPlanInput) (*model.HospitalizationPlan, error) {
	fields := buildHospitalizationPlanUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	plan, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update hospitalization plan")
	}
	slog.InfoContext(ctx, "hospitalization plan updated", slog.Uint64("hospitalization_plan_id", id))
	return plan, nil
}
func (s *hospitalizationPlanService) Delete(ctx context.Context, clinicID, id uint64) error {
	count, err := s.repo.CountCarePlanItemsByPlanID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check hospitalization plan dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この入院プランはケアプランで使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete hospitalization plan")
	}
	slog.InfoContext(ctx, "hospitalization plan deleted",
		slog.Uint64("hospitalization_plan_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}

func (s *hospitalizationPlanService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder hospitalization plan")
	}
	return nil
}

// UpdateHospitalizationPlanInput は入院プラン更新のサービス入力 DTO
type UpdateHospitalizationPlanInput struct {
	Name        *string
	Price       *int64
	IsActive    *bool
	Description *string
	BodySize    *model.BodySize
	BillingUnit *model.BillingUnit
	SortOrder   *int
	TaxType     *model.TaxType
	TaxRate     *float64
}

func buildHospitalizationPlanUpdateFields(input UpdateHospitalizationPlanInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.Price != nil {
		fields["price"] = *input.Price
	}
	if input.IsActive != nil {
		fields["is_active"] = *input.IsActive
	}
	if input.Description != nil {
		fields["description"] = *input.Description
	}
	if input.BodySize != nil {
		fields["body_size"] = *input.BodySize
	}
	if input.BillingUnit != nil {
		fields["billing_unit"] = *input.BillingUnit
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
