package billing

import (
	"context"
	"log/slog"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// Estimate successor draft creation (TASK-012).
// Split from estimate_service.go (ARCH-A4-billing S3) — behavior unchanged.

// CreateSuccessor は承認済み/却下済み見積の後継ドラフトを新規作成する（TASK-012 FINAL B）。
// - 原見積は変更しない（unlock 経路は存在しない）
// - 確定カルテに紐付く原見積でも LockDraftMedicalRecord を呼ばず後継を許可する（明示訂正）
// - 監査は fail-closed（auditTx nil または LogEntryTx 失敗で TX 全体ロールバック）
func (s *estimateService) CreateSuccessor(
	ctx context.Context,
	clinicID, originalID uint64,
	input *CreateSuccessorInput,
) (*model.Estimate, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input is required")
	}
	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		return nil, apperrors.WrapInvalidInput("reason is required")
	}
	if len([]rune(reason)) > 500 {
		return nil, apperrors.WrapInvalidInput("reason must be at most 500 characters")
	}
	if input.ActorID == 0 {
		return nil, apperrors.WrapInvalidInput("actor is required")
	}
	if s.auditTx == nil {
		return nil, apperrors.WrapInternalServerError("estimate audit dependency is required")
	}

	original, err := s.repo.FindByID(ctx, clinicID, originalID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find estimate for successor", "error", err)
		return nil, apperrors.Wrap(err, "failed to find estimate")
	}
	if !isEstimateLocked(original.Status) {
		return nil, apperrors.WrapConflict("承認済みまたは却下済みの見積書のみ後継ドラフトを作成できます")
	}

	title := original.Title
	if input.Title != nil {
		title = *input.Title
	}
	comment := original.Comment
	if input.Comment != nil {
		comment = *input.Comment
	}
	notes := original.Notes
	if input.Notes != nil {
		notes = *input.Notes
	}
	actorID := input.ActorID
	originalIDCopy := original.ID

	var successor *model.Estimate
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// FINAL: 確定カルテでも LockDraftMedicalRecord を呼ばない（明示訂正パス）。
		// created_by は actor の clinic 所属を検証する。
		if err := s.verifyCreatedByClinicMembership(txCtx, clinicID, actorID); err != nil {
			return err
		}
		estimateNo, err := s.repo.AllocateNextEstimateNo(txCtx, clinicID)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to allocate estimate number for successor", "error", err)
			return apperrors.Wrap(err, "failed to allocate estimate number")
		}

		successor = &model.Estimate{
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
			slog.ErrorContext(txCtx, "failed to create successor estimate", "error", err)
			return apperrors.Wrap(err, "failed to create successor estimate")
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
			slog.ErrorContext(txCtx, "audit log failed for estimate supersede", "error", err, "successor_id", successor.ID)
			return apperrors.Wrap(err, "failed to write estimate supersede audit log")
		}
		return nil
	}); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "estimate successor created",
		slog.Uint64("original_id", original.ID),
		slog.Uint64("successor_id", successor.ID),
		slog.Uint64("clinic_id", clinicID))

	created, err := s.repo.FindByID(ctx, clinicID, successor.ID)
	if err != nil {
		// commit 済み成功を後段 read error で失敗応答へ反転させない: 最低限の successor を返す。
		slog.ErrorContext(ctx, "failed to get successor estimate after create", "error", err)
		return successor, nil
	}
	return created, nil
}
