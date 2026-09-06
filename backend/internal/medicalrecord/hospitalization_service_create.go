package medicalrecord

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

func (s *hospitalizationService) createHospitalizationInTx(
	txCtx context.Context,
	clinicID uint64,
	input *CreateHospitalizationInput,
	hospitalization *model.Hospitalization,
) error {
	ownerID, petID := input.OwnerID, input.PetID
	if err := sharedkernel.ValidateReservationOwnerPetLinks(txCtx, s.reservationRepo, clinicID, &ownerID, &petID); err != nil {
		return err
	}
	if err := validatePetNotDeceased(txCtx, s.petRepo, clinicID, petID, "死亡したペットは入院登録できません"); err != nil {
		return err
	}
	if err := validateOwnedMasterFK(txCtx, "cage", clinicID, input.CageID,
		func(actx context.Context, cid, mid uint64) error {
			_, err := s.cageRepo.FindByID(actx, cid, mid)
			return err
		}); err != nil {
		return err
	}
	if err := s.validateDoctor(txCtx, clinicID, input.DoctorID); err != nil {
		return err
	}
	if err := s.hospRepo.Create(txCtx, hospitalization); err != nil {
		return apperrors.Wrap(err, "failed to create hospitalization")
	}
	return s.createNestedTreatmentPlansInTx(txCtx, clinicID, input, hospitalization.ID)
}

func (s *hospitalizationService) createNestedTreatmentPlansInTx(
	txCtx context.Context,
	clinicID uint64,
	input *CreateHospitalizationInput,
	hospID uint64,
) error {
	if len(input.TreatmentPlans) == 0 {
		return nil
	}
	if s.treatmentPlanRepo == nil {
		return apperrors.WrapInternalServerError("hospitalization treatment plan repository is required for nested create")
	}
	if s.carePlanItemRepo == nil {
		return apperrors.WrapInternalServerError("hospitalization care plan item repository is required for nested create")
	}
	for i := range input.TreatmentPlans {
		planInput := &input.TreatmentPlans[i]
		if err := validateTreatmentPlanMoney(planInput.UnitPrice, planInput.Quantity, planInput.DiscountRate, planInput.DiscountAmount); err != nil {
			return err
		}
		if planInput.TreatmentContent == "" {
			return apperrors.WrapInvalidInput("treatment_content is required")
		}
		subtotal := computeTreatmentPlanSubtotal(planInput.UnitPrice, planInput.Quantity, planInput.DiscountRate, planInput.DiscountAmount)
		plan := &model.TreatmentPlan{
			ClinicID:          clinicID,
			HospitalizationID: &hospID,
			TreatmentContent:  planInput.TreatmentContent,
			Memo:              planInput.Memo,
			IsInsurance:       planInput.IsInsurance,
			UnitPrice:         planInput.UnitPrice,
			Quantity:          planInput.Quantity,
			DiscountRate:      planInput.DiscountRate,
			DiscountAmount:    planInput.DiscountAmount,
			Subtotal:          subtotal,
			SortOrder:         planInput.SortOrder,
		}
		if err := s.treatmentPlanRepo.Create(txCtx, plan); err != nil {
			return apperrors.Wrap(err, "failed to create nested treatment plan")
		}
		careItem := &model.CarePlanItem{
			HospitalizationID: hospID,
			Type:              model.CarePlanTypeInstruction,
			Name:              planInput.TreatmentContent,
			Notes:             planInput.Memo,
			Status:            model.CarePlanStatusActive,
			SortOrder:         planInput.SortOrder,
		}
		if err := s.carePlanItemRepo.Create(txCtx, careItem); err != nil {
			return apperrors.Wrap(err, "failed to create nested care plan item")
		}
	}
	return nil
}
