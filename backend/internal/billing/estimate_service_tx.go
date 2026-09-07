package billing

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *estimateService) createEstimateInTx(
	txCtx context.Context,
	clinicID uint64,
	input *CreateEstimateInput,
	estimate *model.Estimate,
	preparedItems []model.EstimateItem,
) (*model.Estimate, error) {
	// SD-2 + BE-refactor.md X-11: 親カルテが確定済みの場合は見積書追加を拒否。見積は
	// medical_record_id 任意（カルテに紐付かない独立見積も許容）のため、指定時のみガードする。
	if err := lockDraftMedicalRecordIfPresent(txCtx, s.medicalRecordRepo, clinicID, input.MedicalRecordID,
		"failed to find medical record", "確定済みカルテに見積書を追加できません"); err != nil {
		return nil, err
	}
	if err := s.validateEstimateRelatedFKs(txCtx, clinicID, input.MedicalRecordID, input.OwnerID, input.PetID); err != nil {
		return nil, err
	}
	if err := s.verifyCreatedByClinicMembership(txCtx, clinicID, *input.CreatedBy); err != nil {
		return nil, err
	}
	// TASK-012: clinic スコープの EST-{N} を原子採番してから INSERT する。
	estimateNo, err := s.repo.AllocateNextEstimateNo(txCtx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to allocate estimate number")
	}
	estimate.EstimateNo = estimateNo
	if err := s.repo.Create(txCtx, estimate); err != nil {
		return nil, apperrors.Wrap(err, "failed to create estimate")
	}
	if len(preparedItems) > 0 {
		if err := s.repo.ReplaceItems(txCtx, clinicID, estimate.ID, estimateItemsFromInput(estimate.ID, input.Items)); err != nil {
			return nil, apperrors.Wrap(err, "failed to save estimate items")
		}
	}
	// 再取得は commit 前。失敗したら INSERT ごと rollback し、成功を失敗応答へ反転させない。
	got, err := s.repo.FindByID(txCtx, clinicID, estimate.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get estimate after create")
	}
	return got, nil
}

type estimateUpdateTxResult struct {
	existing, updated                      *model.Estimate
	isBecomingApproved, isBecomingRejected bool
}

func (s *estimateService) updateEstimateInTx(
	txCtx context.Context,
	clinicID, id uint64,
	input *UpdateEstimateInput,
) (estimateUpdateTxResult, error) {
	var result estimateUpdateTxResult
	// All update paths first lock the editable parent. FindByID then loads active
	// items through this transaction, so header totals and item replacement share
	// one authoritative snapshot and serialization point.
	locked, err := s.repo.LockEditableByID(txCtx, clinicID, id)
	if err != nil && !apperrors.IsConflict(err) {
		return result, apperrors.Wrap(err, "failed to find estimate")
	}
	if err != nil {
		return result, err
	}
	result.existing = locked

	// Keep the prior error precedence: missing or locked estimates are rejected before
	// request validation. Validation still occurs before any write in this transaction.
	if input.Subtotal != nil && *input.Subtotal < 0 {
		return result, apperrors.WrapInvalidInput("subtotal must be 0 or greater")
	}
	if input.TaxTotal != nil && *input.TaxTotal < 0 {
		return result, apperrors.WrapInvalidInput("tax_total must be 0 or greater")
	}
	if input.TotalAmount != nil && *input.TotalAmount < 0 {
		return result, apperrors.WrapInvalidInput("total_amount must be 0 or greater")
	}
	if input.InsuranceAmount != nil && *input.InsuranceAmount < 0 {
		return result, apperrors.WrapInvalidInput("insurance_amount must be 0 or greater")
	}
	if input.DiscountAmount != nil && *input.DiscountAmount < 0 {
		return result, apperrors.WrapInvalidInput("discount_amount must be 0 or greater")
	}
	fields := buildEstimateUpdate(input)
	if len(fields) == 0 && input.Items == nil {
		return result, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	cmd := *input
	var preparedItems []model.EstimateItem
	if input.Items != nil {
		if err := validateEstimateItemInputs(*input.Items); err != nil {
			return result, err
		}
		preparedItems = estimateItemsFromInput(id, *input.Items)
		subtotal, taxTotal, totalAmount := calculateEstimateTotals(preparedItems)
		cmd.Subtotal = &subtotal
		cmd.TaxTotal = &taxTotal
		cmd.TotalAmount = &totalAmount
	}

	if input.Items == nil && len(result.existing.Items) > 0 {
		// Active persisted items are the source of truth for estimate totals. Header-only
		// PATCHes must not allow client-supplied totals to drift from those items.
		subtotal, taxTotal, totalAmount := calculateEstimateTotals(result.existing.Items)
		cmd.Subtotal = &subtotal
		cmd.TaxTotal = &taxTotal
		cmd.TotalAmount = &totalAmount
	}
	result.isBecomingApproved = input.Status != nil && *input.Status == model.EstimateStatusApproved &&
		result.existing.Status != model.EstimateStatusApproved
	result.isBecomingRejected = input.Status != nil && *input.Status == model.EstimateStatusRejected &&
		result.existing.Status != model.EstimateStatusRejected

	// SD-2 + BE-refactor.md X-11: 親カルテが確定済みの場合は見積書編集を拒否。
	if err := lockDraftMedicalRecordIfPresent(txCtx, s.medicalRecordRepo, clinicID, result.existing.MedicalRecordID,
		"failed to find medical record", "確定済みカルテの見積書は編集できません"); err != nil {
		return result, err
	}
	// UpdateIfNotLocked retains the status predicate as defense in depth. The parent
	// is already locked, so it cannot change between the authoritative read and write.
	got, err := s.repo.UpdateIfNotLocked(txCtx, clinicID, id, cmd)
	if err != nil {
		return result, apperrors.Wrap(err, "failed to update estimate")
	}
	if input.Items != nil {
		if err := s.repo.ReplaceItems(txCtx, clinicID, id, preparedItems); err != nil {
			return result, apperrors.Wrap(err, "failed to save estimate items")
		}
		got, err = s.repo.FindByID(txCtx, clinicID, id)
		if err != nil {
			return result, apperrors.Wrap(err, "failed to reload estimate after item replace")
		}
	}
	result.updated = got
	return result, nil
}
