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
	// This legacy Create does not write stay acceptance proof. Do not create
	// a physical assignment whose occupancy the effective-state reader cannot
	// resolve, regardless of the requested/default legacy status.
	if input.CageID != nil || hospitalization.CageID != nil {
		return hospitalizationCageConflict(codeOccupancyUnknown, "滞在の有効状態を確認できないためケージを割り当てできません")
	}
	ownerID, petID := input.OwnerID, input.PetID
	if err := sharedkernel.ValidateReservationOwnerPetLinks(txCtx, s.reservationRepo, clinicID, &ownerID, &petID); err != nil {
		return err
	}
	if err := validatePetNotDeceased(txCtx, s.petLife, clinicID, petID, "死亡したペットは入院登録できません"); err != nil {
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
	// EMR-499 U1: the stay case carries kind + PLANNED interval. No movement event is written here,
	// so the effective state stays unknown (reserved/admitted need acceptance proof, EMR-499 Q1/Q2)
	// and a non-NULL cage stays refused above. Nothing is inferred from the planned dates.
	stay := &model.StayCase{
		ClinicID:       clinicID,
		OwnerID:        input.OwnerID,
		PetID:          input.PetID,
		StayKind:       input.StayKind,
		PlannedEntryAt: input.PlannedEntryAt,
		PlannedExitAt:  input.PlannedExitAt,
	}
	if err := s.hospRepo.CreateStayCase(txCtx, stay); err != nil {
		return apperrors.Wrap(err, "failed to create stay case")
	}
	hospitalization.StayCaseID = &stay.ID
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
		// quantity / discount_rate は exact 列。検証済みの float64 request 値を最短10進表現で橋渡しする。
		quantity, err := exactNumericFromFloat(planInput.Quantity, "quantity")
		if err != nil {
			return err
		}
		discountRate, err := exactNumericFromFloat(planInput.DiscountRate, "discount_rate")
		if err != nil {
			return err
		}
		plan := &model.TreatmentPlan{
			ClinicID:          clinicID,
			HospitalizationID: &hospID,
			TreatmentContent:  planInput.TreatmentContent,
			Memo:              planInput.Memo,
			IsInsurance:       sharedkernel.Ptr(planInput.IsInsurance),
			UnitPrice:         sharedkernel.Ptr(planInput.UnitPrice),
			Quantity:          &quantity,
			DiscountRate:      &discountRate,
			DiscountAmount:    sharedkernel.Ptr(planInput.DiscountAmount),
			SortOrder:         sharedkernel.Ptr(planInput.SortOrder),
		}
		if err := s.treatmentPlanRepo.Create(txCtx, plan); err != nil {
			return apperrors.Wrap(err, "failed to create nested treatment plan")
		}
		careItem := &model.CarePlanItem{
			HospitalizationID: hospID,
			Type:              model.CarePlanTypeInstruction,
			Name:              planInput.TreatmentContent,
			Notes:             planInput.Memo,
			Status:            sharedkernel.Ptr(model.CarePlanStatusActive),
			SortOrder:         sharedkernel.Ptr(planInput.SortOrder),
		}
		if err := s.carePlanItemRepo.Create(txCtx, careItem); err != nil {
			return apperrors.Wrap(err, "failed to create nested care plan item")
		}
	}
	return nil
}
