package medicalrecord

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

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
	if locked.Status == model.HospitalizationStatusDischarged {
		return apperrors.WrapInvalidInput("hospitalization is already discharged")
	}
	if err := validateHospitalizationDateRange(locked.StartDate, input.DischargeDate); err != nil {
		return err
	}
	if err := s.reservationRepo.AssertOwnerInClinic(txCtx, clinicID, locked.OwnerID); err != nil {
		return apperrors.Wrap(err, "failed to verify hospitalization owner ownership")
	}
	if _, err := s.reservationRepo.FindPetOwnerInClinic(txCtx, clinicID, locked.PetID); err != nil {
		return apperrors.Wrap(err, "failed to verify hospitalization pet ownership")
	}
	if input.CreateAccounting {
		if input.ActorID == nil || *input.ActorID == 0 {
			return apperrors.WrapInvalidInput("staff actor is required to create discharge billing")
		}
		if s.auditTx == nil {
			return apperrors.WrapInternalServerError("hospitalization discharge billing audit dependency is required")
		}
		// EMR-253: no-duplicate ガード（FindByHospitalizationID/SoftDeleteCancelled）を含め required。
		if s.accountingRepo == nil {
			return apperrors.WrapInternalServerError("hospitalization discharge billing repository dependency is required")
		}
	}

	dischargedStatus := model.HospitalizationStatusDischarged
	if _, err := s.hospRepo.UpdateIfNotDischarged(txCtx, clinicID, id, UpdateHospitalizationInput{
		Status:  &dischargedStatus,
		EndDate: &input.DischargeDate,
	}); err != nil {
		return apperrors.Wrap(err, "failed to discharge hospitalization")
	}
	if !input.CreateAccounting {
		return nil
	}

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
				HospitalizationType: locked.HospitalizationType,
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
