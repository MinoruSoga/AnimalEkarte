package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type CreateTreatmentPlanInput struct {
	TreatmentContent string
	Memo             string
	IsInsurance      bool
	UnitPrice        int64
	Quantity         float64
	DiscountRate     float64
	DiscountAmount   int64
	Subtotal         int64
	SortOrder        int
}

type UpdateTreatmentPlanInput struct {
	TreatmentContent *string
	Memo             *string
	IsInsurance      *bool
	UnitPrice        *int64
	Quantity         *float64
	DiscountRate     *float64
	DiscountAmount   *int64
	Subtotal         *int64
	SortOrder        *int
}

func buildTreatmentPlanUpdate(input *UpdateTreatmentPlanInput) map[string]any {
	fields := map[string]any{}
	if input.TreatmentContent != nil {
		fields["treatment_content"] = *input.TreatmentContent
	}
	if input.Memo != nil {
		fields["memo"] = *input.Memo
	}
	if input.IsInsurance != nil {
		fields["is_insurance"] = *input.IsInsurance
	}
	if input.UnitPrice != nil {
		fields["unit_price"] = *input.UnitPrice
	}
	if input.Quantity != nil {
		fields["quantity"] = *input.Quantity
	}
	if input.DiscountRate != nil {
		fields["discount_rate"] = *input.DiscountRate
	}
	if input.DiscountAmount != nil {
		fields["discount_amount"] = *input.DiscountAmount
	}
	if input.Subtotal != nil {
		fields["subtotal"] = *input.Subtotal
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	return fields
}

type TreatmentPlanService interface {
	ListByMedicalRecord(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.TreatmentPlan, error)
	ListByHospitalization(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.TreatmentPlan, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.TreatmentPlan, error)
	Create(ctx context.Context, clinicID uint64, medicalRecordID, hospitalizationID *uint64, input *CreateTreatmentPlanInput) (*model.TreatmentPlan, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateTreatmentPlanInput) (*model.TreatmentPlan, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

type treatmentPlanService struct {
	repo repository.TreatmentPlanRepository
}

func NewTreatmentPlanService(repo repository.TreatmentPlanRepository) TreatmentPlanService {
	return &treatmentPlanService{repo: repo}
}

func (s *treatmentPlanService) GetByID(ctx context.Context, clinicID, id uint64) (*model.TreatmentPlan, error) {
	plan, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get treatment plan")
	}
	return plan, nil
}

func (s *treatmentPlanService) ListByMedicalRecord(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.TreatmentPlan, error) {
	plans, err := s.repo.FindByMedicalRecordID(ctx, clinicID, medicalRecordID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list treatment plans by medical record")
	}
	return plans, nil
}

func (s *treatmentPlanService) ListByHospitalization(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.TreatmentPlan, error) {
	plans, err := s.repo.FindByHospitalizationID(ctx, clinicID, hospitalizationID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list treatment plans by hospitalization")
	}
	return plans, nil
}

func (s *treatmentPlanService) Create(ctx context.Context, clinicID uint64, medicalRecordID, hospitalizationID *uint64, input *CreateTreatmentPlanInput) (*model.TreatmentPlan, error) {
	subtotal := input.Subtotal
	if subtotal == 0 && input.UnitPrice > 0 {
		qty := input.Quantity
		if qty == 0 {
			qty = 1
		}
		subtotal = int64(float64(input.UnitPrice)*qty*(1-input.DiscountRate/100)) - input.DiscountAmount
	}

	plan := &model.TreatmentPlan{
		MedicalRecordID:   medicalRecordID,
		HospitalizationID: hospitalizationID,
		TreatmentContent:  input.TreatmentContent,
		Memo:              input.Memo,
		IsInsurance:       input.IsInsurance,
		UnitPrice:         input.UnitPrice,
		Quantity:          input.Quantity,
		DiscountRate:      input.DiscountRate,
		DiscountAmount:    input.DiscountAmount,
		Subtotal:          subtotal,
		SortOrder:         input.SortOrder,
	}
	if err := s.repo.Create(ctx, plan); err != nil {
		return nil, apperrors.Wrap(err, "failed to create treatment plan")
	}
	slog.InfoContext(ctx, "treatment plan created", slog.Uint64("clinic_id", clinicID), slog.Uint64("treatment_plan_id", plan.ID))

	result, err := s.repo.FindByID(ctx, clinicID, plan.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to reload treatment plan after create")
	}
	return result, nil
}

func (s *treatmentPlanService) Update(ctx context.Context, clinicID, id uint64, input *UpdateTreatmentPlanInput) (*model.TreatmentPlan, error) {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to find treatment plan")
	}
	fields := buildTreatmentPlanUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update treatment plan")
	}
	slog.InfoContext(ctx, "treatment plan updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("treatment_plan_id", id))

	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to reload treatment plan after update")
	}
	return result, nil
}

func (s *treatmentPlanService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to find treatment plan")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete treatment plan")
	}
	slog.InfoContext(ctx, "treatment plan deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("treatment_plan_id", id))
	return nil
}

// compile-time check
var _ TreatmentPlanService = (*treatmentPlanService)(nil)
