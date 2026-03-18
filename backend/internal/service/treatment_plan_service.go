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
	Insurance        bool
	UnitPrice        float64
	Quantity         float64
	DiscountRate     float64
	DiscountAmount   float64
	Subtotal         float64
	SortOrder        int
}

type UpdateTreatmentPlanInput struct {
	TreatmentContent *string
	Memo             *string
	Insurance        *bool
	UnitPrice        *float64
	Quantity         *float64
	DiscountRate     *float64
	DiscountAmount   *float64
	Subtotal         *float64
	SortOrder        *int
}

type TreatmentPlanService interface {
	ListByMedicalRecord(ctx context.Context, medicalRecordID uint64) ([]model.TreatmentPlan, error)
	ListByHospitalization(ctx context.Context, hospitalizationID uint64) ([]model.TreatmentPlan, error)
	Create(ctx context.Context, medicalRecordID, hospitalizationID *uint64, input *CreateTreatmentPlanInput) (*model.TreatmentPlan, error)
	Update(ctx context.Context, id uint64, input *UpdateTreatmentPlanInput) (*model.TreatmentPlan, error)
	Delete(ctx context.Context, id uint64) error
}

type treatmentPlanService struct {
	repo repository.TreatmentPlanRepository
}

func NewTreatmentPlanService(repo repository.TreatmentPlanRepository) TreatmentPlanService {
	return &treatmentPlanService{repo: repo}
}

func (s *treatmentPlanService) ListByMedicalRecord(ctx context.Context, medicalRecordID uint64) ([]model.TreatmentPlan, error) {
	return s.repo.ListByMedicalRecordID(ctx, medicalRecordID)
}

func (s *treatmentPlanService) ListByHospitalization(ctx context.Context, hospitalizationID uint64) ([]model.TreatmentPlan, error) {
	return s.repo.ListByHospitalizationID(ctx, hospitalizationID)
}

func (s *treatmentPlanService) Create(ctx context.Context, medicalRecordID, hospitalizationID *uint64, input *CreateTreatmentPlanInput) (*model.TreatmentPlan, error) {
	subtotal := input.Subtotal
	if subtotal == 0 && input.UnitPrice > 0 {
		qty := input.Quantity
		if qty == 0 {
			qty = 1
		}
		subtotal = input.UnitPrice*qty*(1-input.DiscountRate/100) - input.DiscountAmount
	}

	plan := &model.TreatmentPlan{
		MedicalRecordID:   medicalRecordID,
		HospitalizationID: hospitalizationID,
		TreatmentContent:  input.TreatmentContent,
		Memo:              input.Memo,
		Insurance:         input.Insurance,
		UnitPrice:         input.UnitPrice,
		Quantity:          input.Quantity,
		DiscountRate:      input.DiscountRate,
		DiscountAmount:    input.DiscountAmount,
		Subtotal:          subtotal,
		SortOrder:         input.SortOrder,
	}
	if err := s.repo.Create(ctx, plan); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "treatment plan created", slog.Uint64("treatment_plan_id", plan.ID))
	return s.repo.FindByID(ctx, plan.ID)
}

func (s *treatmentPlanService) Update(ctx context.Context, id uint64, input *UpdateTreatmentPlanInput) (*model.TreatmentPlan, error) {
	fields := buildTreatmentPlanUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if err := s.repo.Update(ctx, id, fields); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "treatment plan updated", slog.Uint64("treatment_plan_id", id))
	return s.repo.FindByID(ctx, id)
}

func (s *treatmentPlanService) Delete(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	slog.InfoContext(ctx, "treatment plan deleted", slog.Uint64("treatment_plan_id", id))
	return nil
}

func buildTreatmentPlanUpdateFields(input *UpdateTreatmentPlanInput) map[string]any {
	fields := map[string]any{}
	if input.TreatmentContent != nil {
		fields["treatment_content"] = *input.TreatmentContent
	}
	if input.Memo != nil {
		fields["memo"] = *input.Memo
	}
	if input.Insurance != nil {
		fields["insurance"] = *input.Insurance
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

// compile-time check
var _ TreatmentPlanService = (*treatmentPlanService)(nil)
