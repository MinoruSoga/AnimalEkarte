package medicalrecord

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// dischargeWithBillingInTx is the movement side of discharge (EMR-499 U1). It reads the stay's
// effective state under the row lock and refuses everything it cannot prove; it never writes the
// removed hospitalizations.status / end_date and never infers a discharge from planned dates.
//
// The discharge movement event itself (stay_movement_events.event_kind = 'discharge') is NOT written
// here: the INSERT needs an operation binding + success receipt (001 operation_success_required) which
// the discharge route does not carry, and no admit writer exists yet, so no stay can currently be
// proven "admitted". The 422 below is the stable fail-closed answer until EMR-499 U2 lands that writer.
func (s *hospitalizationService) dischargeWithBillingInTx(
	txCtx context.Context,
	clinicID, id uint64,
	input DischargeWithBillingInput,
	result *DischargeWithBillingResult,
) error {
	locked, err := s.hospRepo.LockByIDForUpdate(txCtx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to lock hospitalization for discharge")
	}
	if err := s.reservationRepo.AssertOwnerInClinic(txCtx, clinicID, locked.OwnerID); err != nil {
		return apperrors.Wrap(err, "failed to verify hospitalization owner ownership")
	}
	if _, err := s.reservationRepo.FindPetOwnerInClinic(txCtx, clinicID, locked.PetID); err != nil {
		return apperrors.Wrap(err, "failed to verify hospitalization pet ownership")
	}
	state, err := s.readDischargeStayState(txCtx, clinicID, locked)
	if err != nil {
		return err
	}
	switch state.State {
	case StayCageReleased:
		if state.ReleaseKind == StayCageReleasedByDischarge {
			return apperrors.WrapInvalidInput("hospitalization is already discharged")
		}
		return apperrors.WrapInvalidInput("hospitalization stay was cancelled and cannot be discharged")
	case StayCageAdmitted:
		if state.AdmittedAt == nil {
			return unknownDischargeStay()
		}
		// Calendar-day comparison in JST: discharge_date may carry no time of day.
		admittedDay := state.AdmittedAt.In(config.JST).Format(time.DateOnly)
		if input.DischargeDate.In(config.JST).Format(time.DateOnly) < admittedDay {
			return apperrors.WrapInvalidInput("discharge_date は入院開始日以降である必要があります")
		}
		return apperrors.WrapUnprocessableWithCode(codeStayDischargeWriterUnavailable,
			"退院の記録（滞在の移動履歴への書込）は、操作受付の仕組みの整備まで利用できません")
	case StayCageReserved:
		return apperrors.WrapInvalidInput("hospitalization has not been admitted")
	default:
		return unknownDischargeStay()
	}
}

func unknownDischargeStay() error {
	return apperrors.WrapUnprocessableWithCode(codeStayLifecycleUnknown,
		"この入院の滞在状態を確認できないため退院できません")
}

// readDischargeStayState reads the effective state inside the ambient transaction. A hospitalization
// without a stay case (legacy row) is unknown; there is no backfill and no guess.
func (s *hospitalizationService) readDischargeStayState(txCtx context.Context, clinicID uint64, locked *model.Hospitalization) (StayCageEffectiveState, error) {
	if locked.StayCaseID == nil || *locked.StayCaseID == 0 {
		return StayCageEffectiveState{}, unknownDischargeStay()
	}
	reader, ok := s.cageRepo.(StayCageEffectiveStateReader)
	if !ok {
		return StayCageEffectiveState{}, apperrors.WrapInternalServerError("stay effective-state reader dependency is required")
	}
	states, err := reader.ReadEffectiveStatesForCageAssignment(txCtx, clinicID, []uint64{*locked.StayCaseID})
	if err != nil {
		return StayCageEffectiveState{}, apperrors.Wrap(err, "failed to read stay effective state")
	}
	for _, st := range states {
		if st.ClinicID == clinicID && st.StayCaseID == *locked.StayCaseID {
			return st, nil
		}
	}
	return StayCageEffectiveState{}, unknownDischargeStay()
}

// createDischargeBilling is the PRE-EXISTING billing builder, moved verbatim and intentionally NOT
// called: create_accounting=true is refused with stay_discharge_billing_unavailable before any state
// is touched. It still targets the removed Billing/BillingItem shape; EMR-505/506/508 (unit 4) own its
// replacement and the call site. It is kept so that hunk has a single owner and no history is lost.
func (s *hospitalizationService) createDischargeBilling(
	txCtx context.Context,
	clinicID, id uint64,
	locked *model.Hospitalization,
	hospType model.HospitalizationType, // from stay_cases.stay_kind via stayKindForWireType (EMR-499 U1)
	input DischargeWithBillingInput,
	result *DischargeWithBillingResult,
) error {
	// EMR-253: 同じ入院に既存の非削除 billing がある場合は二重作成しない。
	// cancelled 行は部分 UNIQUE（hospitalization_id, deleted_at IS NULL）を占有したまま残るため
	// 同一 tx で解放してから新規作成する。waiting/pending/completed の既存行はそのまま会計として
	// 再利用し、確定は POST /accountings/complete の takeover 経路に委ねる。
	occupant, err := s.accountingRepo.FindByHospitalizationID(txCtx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to lookup existing hospitalization billing")
	}
	if occupant != nil && occupant.Status == model.BillingStatusCancelled {
		if err := s.accountingRepo.SoftDeleteCancelled(txCtx, clinicID, occupant.ID); err != nil {
			return apperrors.Wrap(err, "failed to release cancelled hospitalization billing")
		}
		occupant = nil
	}
	if occupant != nil {
		result.AccountingID = &occupant.ID
		resourceID := id
		if err := s.auditTx.LogEntryTx(txCtx, &AuditEntry{
			ClinicID:   &clinicID,
			ActorID:    input.ActorID,
			ActorType:  model.AuditActorTypeStaff,
			Action:     model.AuditActionHospitalizationDischargeWithBilling,
			Resource:   model.AuditResourceHospitalization,
			ResourceID: &resourceID,
			NewValue: map[string]any{
				"billing_id":      occupant.ID,
				"subtotal_amount": occupant.Subtotal,
				"tax_amount":      occupant.TaxTotal,
				"total_amount":    occupant.TotalAmount,
				"reused":          true,
			},
		}); err != nil {
			return apperrors.Wrap(err, "failed to audit hospitalization discharge billing")
		}
		return nil
	}

	carePlanItems, err := s.carePlanItemRepo.FindByHospitalizationID(txCtx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to get care plan items")
	}
	billing := &model.Billing{
		ClinicID:          clinicID,
		HospitalizationID: &id,
		PetID:             &locked.PetID,
		OwnerID:           &locked.OwnerID,
		Status:            model.BillingStatusWaiting,
		ScheduledDate:     input.DischargeDate,
	}
	if err := s.accountingRepo.Create(txCtx, clinicID, billing); err != nil {
		return apperrors.Wrap(err, "failed to create billing")
	}

	var subtotalAmount int64
	for i := range carePlanItems {
		item := &carePlanItems[i]
		billingItem := &model.BillingItem{
			BillingID: billing.ID,
			Category: sharedkernel.ResolveItemCategory(sharedkernel.ItemCategoryResolverInput{
				Source:              model.ItemSourceHospitalization,
				CarePlanType:        item.Type,
				IsSurgery:           item.Procedure != nil && item.Procedure.IsSurgery,
				HospitalizationType: hospType,
			}),
			Name:      item.Name,
			UnitPrice: item.UnitPrice,
			Quantity:  1.0,
			TaxType:   model.TaxTypeExcluded,
			TaxRate:   sharedkernel.DefaultTaxRate,
			Source:    model.ItemSourceHospitalization,
			SortOrder: i,
		}
		if err := s.billingItemRepo.Create(txCtx, billingItem); err != nil {
			return apperrors.Wrap(err, "failed to create billing item")
		}
		subtotalAmount += item.UnitPrice
	}

	taxAmount := int64(float64(subtotalAmount) * sharedkernel.DefaultTaxRate)
	totalAmount := subtotalAmount + taxAmount
	if len(carePlanItems) > 0 {
		if err := s.billingItemRepo.UpdateBillingTotals(txCtx, clinicID, billing.ID, subtotalAmount, taxAmount, totalAmount); err != nil {
			return apperrors.Wrap(err, "failed to update billing totals")
		}
	}

	resourceID := id
	if err := s.auditTx.LogEntryTx(txCtx, &AuditEntry{
		ClinicID:   &clinicID,
		ActorID:    input.ActorID,
		ActorType:  model.AuditActorTypeStaff,
		Action:     model.AuditActionHospitalizationDischargeWithBilling,
		Resource:   model.AuditResourceHospitalization,
		ResourceID: &resourceID,
		NewValue: map[string]any{
			"billing_id":      billing.ID,
			"subtotal_amount": subtotalAmount,
			"tax_amount":      taxAmount,
			"total_amount":    totalAmount,
		},
	}); err != nil {
		return apperrors.Wrap(err, "failed to audit hospitalization discharge billing")
	}

	result.AccountingID = &billing.ID
	return nil
}
