// Package medicalrecord provides hospitalization plan use cases.
package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

const (
	colHospitalizationPlanName        = "name"
	colHospitalizationPlanPrice       = "price"
	colHospitalizationPlanIsActive    = "is_active"
	colHospitalizationPlanDescription = "description"
	colHospitalizationPlanBodySize    = "body_size"
	colHospitalizationPlanBillingUnit = "billing_unit"
	colHospitalizationPlanSortOrder   = "sort_order"
	colHospitalizationPlanTaxType     = "tax_type"
	colHospitalizationPlanTaxRate     = "tax_rate"
)

func buildHospitalizationPlanUpdate(input UpdateHospitalizationPlanInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colHospitalizationPlanName] = *input.Name
	}
	if input.Price != nil {
		fields[colHospitalizationPlanPrice] = *input.Price
	}
	if input.IsActive != nil {
		fields[colHospitalizationPlanIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colHospitalizationPlanDescription] = *input.Description
	}
	if input.BodySize != nil {
		fields[colHospitalizationPlanBodySize] = model.BodySize(*input.BodySize)
	}
	if input.BillingUnit != nil {
		fields[colHospitalizationPlanBillingUnit] = model.BillingUnit(*input.BillingUnit)
	}
	if input.SortOrder != nil {
		fields[colHospitalizationPlanSortOrder] = *input.SortOrder
	}
	if input.TaxType != nil {
		fields[colHospitalizationPlanTaxType] = model.TaxType(*input.TaxType)
	}
	if input.TaxRate != nil {
		fields[colHospitalizationPlanTaxRate] = *input.TaxRate
	}
	return fields
}

// ---- HospitalizationPlanService ----

// CreateHospitalizationPlanInput は入院プラン作成の入力DTO
type CreateHospitalizationPlanInput struct {
	Name        string
	Price       *int64
	IsActive    bool
	Description string
	SortOrder   int
	TaxType     string
	TaxRate     *float64
	BodySize    string
	BillingUnit string
}

// UpdateHospitalizationPlanInput は入院プラン更新のサービス入力 DTO
type UpdateHospitalizationPlanInput struct {
	Name        *string
	Price       *int64
	IsActive    *bool
	Description *string
	BodySize    *string
	BillingUnit *string
	SortOrder   *int
	TaxType     *string
	TaxRate     *float64
}

// --- DB column constants ---

type HospitalizationPlanService interface {
	List(ctx context.Context, clinicID uint64) ([]model.HospitalizationPlan, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.HospitalizationPlan, error)
	Create(ctx context.Context, clinicID uint64, input *CreateHospitalizationPlanInput) (*model.HospitalizationPlan, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateHospitalizationPlanInput) (*model.HospitalizationPlan, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type hospitalizationPlanService struct {
	repo HospitalizationPlanRepository
}

func NewHospitalizationPlanService(repo HospitalizationPlanRepository) HospitalizationPlanService {
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
func (s *hospitalizationPlanService) Create(ctx context.Context, clinicID uint64, input *CreateHospitalizationPlanInput) (*model.HospitalizationPlan, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	taxType := model.TaxTypeExcluded
	if input.TaxType != "" {
		taxType = model.TaxType(input.TaxType)
	}
	taxRate := sharedkernel.DefaultTaxRate
	if input.TaxRate != nil {
		taxRate = *input.TaxRate
	}
	plan := &model.HospitalizationPlan{
		ClinicID:    clinicID,
		Name:        input.Name,
		Price:       input.Price,
		IsActive:    input.IsActive,
		Description: input.Description,
		SortOrder:   input.SortOrder,
		TaxType:     taxType,
		TaxRate:     taxRate,
	}
	if input.BodySize != "" {
		bs := model.BodySize(input.BodySize)
		plan.BodySize = &bs
	}
	if input.BillingUnit != "" {
		bu := model.BillingUnit(input.BillingUnit)
		plan.BillingUnit = &bu
	}
	if err := s.repo.Create(ctx, plan); err != nil {
		return nil, apperrors.Wrap(err, "failed to create hospitalization plan")
	}
	slog.InfoContext(ctx, "hospitalization plan created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("hospitalization_plan_id", plan.ID))
	return plan, nil
}
func (s *hospitalizationPlanService) Update(ctx context.Context, clinicID, id uint64, input *UpdateHospitalizationPlanInput) (*model.HospitalizationPlan, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(errMsgInputNotNil)
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to get hospitalization plan")
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	fields := buildHospitalizationPlanUpdate(*input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(errMsgAtLeastOneField)
	}
	plan, err := s.repo.Update(ctx, clinicID, id, *input)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update hospitalization plan")
	}
	slog.InfoContext(ctx, "hospitalization plan updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("hospitalization_plan_id", id))
	return plan, nil
}
func (s *hospitalizationPlanService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to find hospitalization plan")
	}
	count, err := s.repo.CountUsageByHospitalizationPlanID(ctx, clinicID, id)
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
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("hospitalization_plan_id", id))
	return nil
}

func (s *hospitalizationPlanService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(errMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder hospitalization plan")
	}
	slog.InfoContext(ctx, "hospitalization plans reordered",
		slog.Uint64("clinic_id", clinicID),
		slog.Int("count", len(ids)))
	return nil
}
