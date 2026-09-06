package billing

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

func (s *estimateService) createSuccessorInTx(
	txCtx context.Context,
	clinicID, actorID uint64,
	original *model.Estimate,
	title, comment, notes, reason string,
) (*model.Estimate, error) {
	// FINAL: 確定カルテでも LockDraftMedicalRecord を呼ばない（明示訂正パス）。
	// created_by は actor の clinic 所属を検証する。
	if err := s.verifyCreatedByClinicMembership(txCtx, clinicID, actorID); err != nil {
		return nil, err
	}
	estimateNo, err := s.repo.AllocateNextEstimateNo(txCtx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to allocate estimate number")
	}

	originalIDCopy := original.ID
	successor := &model.Estimate{
		ClinicID:             clinicID,
		EstimateNo:           estimateNo,
		MedicalRecordID:      original.MedicalRecordID,
		Title:                title,
		OwnerID:              original.OwnerID,
		PetID:                original.PetID,
		Status:               model.EstimateStatusDraft,
		Subtotal:             original.Subtotal,
		TaxTotal:             original.TaxTotal,
		TotalAmount:          original.TotalAmount,
		InsuranceAmount:      original.InsuranceAmount,
		DiscountAmount:       original.DiscountAmount,
		ValidUntil:           original.ValidUntil,
		Comment:              comment,
		Notes:                notes,
		CreatedBy:            &actorID,
		SupersedesEstimateID: &originalIDCopy,
	}
	if err := s.repo.Create(txCtx, successor); err != nil {
		return nil, apperrors.Wrap(err, "failed to create successor estimate")
	}
	if len(original.Items) > 0 {
		if err := s.repo.ReplaceItems(txCtx, clinicID, successor.ID, cloneEstimateItemsForSuccessor(successor.ID, original.Items)); err != nil {
			return nil, apperrors.Wrap(err, "failed to copy successor estimate items")
		}
	}

	// fail-closed: 監査失敗 → 後継 INSERT ごとロールバック。原行は未変更のまま。
	if err := s.auditTx.LogEntryTx(txCtx, &AuditEntry{
		ClinicID:   &clinicID,
		ActorID:    &actorID,
		ActorType:  sharedkernel.AuditActorTypeFor(&actorID),
		Action:     "supersede",
		Resource:   "estimate",
		ResourceID: &successor.ID,
		NewValue: map[string]any{
			"original_id":  original.ID,
			"successor_id": successor.ID,
			"reason":       reason,
			"estimate_no":  successor.EstimateNo,
		},
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to write estimate supersede audit log")
	}
	got, err := s.repo.FindByID(txCtx, clinicID, successor.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get successor estimate after create")
	}
	return got, nil
}
