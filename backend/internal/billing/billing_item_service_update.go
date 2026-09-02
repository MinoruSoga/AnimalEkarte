// Package service provides business logic implementations for BillingItem entity.
package billing

import (
	"context"
	"log/slog"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

func (s *billingItemService) UpdateItem(ctx context.Context, clinicID, id uint64, input *UpdateBillingItemInput) (*model.BillingItem, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("請求明細の更新内容は必須です")
	}
	// #115 / BUG-463: 締め後編集は理由必須（handler 迂回経路にも強制）
	if err := requirePostCloseReason(input.IsPostClose, input.PostCloseReason); err != nil {
		return nil, err
	}
	item, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get billing item", "error", err)
		return nil, apperrors.Wrap(err, "failed to get billing item")
	}
	if input.UnitPrice != nil && *input.UnitPrice < 0 {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgPriceZeroOrMore)
	}
	if input.Quantity != nil && *input.Quantity <= 0 {
		return nil, apperrors.WrapInvalidInput("数量は正の値である必要があります")
	}

	fields := buildBillingItemUpdate(input)
	if len(fields) == 0 {
		return item, nil
	}

	var updated *model.BillingItem
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// BUG-463 / BUG-009: lock parent; cancelled 拒否、completed は理由付き訂正のみ許可
		billing, err := s.billingRepo.LockAndFindByID(txCtx, clinicID, item.BillingID)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock billing before updating item")
		}
		if err := rejectUnlessCompletedCorrectionAllowed(billing, "更新", input.PostCloseReason); err != nil {
			return err
		}

		if err := s.repo.Update(txCtx, clinicID, id, fields); err != nil {
			slog.ErrorContext(txCtx, "failed to update billing item", "error", err)
			return apperrors.Wrap(err, "failed to update billing item")
		}

		completedCorrection := billing.Status == model.BillingStatusCompleted
		if err := s.recalculateTotals(txCtx, clinicID, item.BillingID, completedCorrection); err != nil {
			slog.ErrorContext(txCtx, "failed to recalculate billing totals after update",
				slog.Uint64("billing_id", item.BillingID),
				slog.String("error", err.Error()),
			)
			return apperrors.Wrap(err, "failed to recalculate billing totals")
		}

		itemID := id
		if completedCorrection {
			// BUG-009: 確定済み訂正は常に監査。締め済み日なら adjustment 台帳も同 tx。
			if err := s.recordCompletedItemCorrection(txCtx, clinicID, item.BillingID, billing, input.StaffID, input.PostCloseReason, "update", &itemID); err != nil {
				return err
			}
		} else if err := s.recordBillingItemPostClose(txCtx, input.IsPostClose, clinicID, item.BillingID, billing, input.StaffID, input.PostCloseReason, "update", &itemID); err != nil {
			return err
		}

		slog.InfoContext(txCtx, "billing item updated",
			slog.Uint64("clinic_id", clinicID),
			slog.Uint64("item_id", id),
			slog.Uint64("billing_id", item.BillingID),
		)
		updated, err = s.repo.FindByID(txCtx, clinicID, id)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get billing item after update", "error", err)
			return apperrors.Wrap(err, "failed to get billing item after update")
		}
		return nil
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to update billing item in transaction")
	}

	return updated, nil
}

func (s *billingItemService) DeleteItem(ctx context.Context, clinicID, id uint64, input *DeleteBillingItemInput) error {
	if input == nil {
		input = &DeleteBillingItemInput{}
	}
	// #115 / BUG-463 residual: 締め後削除は理由必須（handler 迂回経路にも強制）
	if err := requirePostCloseReason(input.IsPostClose, input.PostCloseReason); err != nil {
		return err
	}

	item, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to get billing item")
	}
	billingID := item.BillingID
	// Delete が vaccination_id を NULL 化する前に claim 解放対象を保持する（BUG-440）。
	releasedVaccinationID := item.VaccinationID

	return s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		billing, err := s.billingRepo.LockAndFindByID(txCtx, clinicID, billingID)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock billing before deleting item")
		}
		if err := rejectIfBillingFinalized(billing, "削除"); err != nil {
			return err
		}

		if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
			slog.ErrorContext(txCtx, "failed to delete billing item", "error", err, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to delete billing item")
		}

		if err := s.recalculateTotals(txCtx, clinicID, billingID, false); err != nil {
			slog.ErrorContext(txCtx, "failed to recalculate billing totals after delete",
				slog.Uint64("billing_id", billingID),
				slog.String("error", err.Error()),
			)
			return apperrors.Wrap(err, "failed to recalculate billing totals")
		}

		// BUG-463 residual / W-013: 締め後削除は adjustment 台帳 + 監査を同 tx fail-closed で記録する。
		// vaccination claim 解放監査（BUG-440）と両立し、両方発火し得る。
		itemID := id
		if err := s.recordBillingItemPostClose(txCtx, input.IsPostClose, clinicID, billingID, billing, input.StaffID, input.PostCloseReason, "delete", &itemID); err != nil {
			return err
		}

		// BUG-440: vaccination claim 解放時のみ immutable actor 監査（同 tx fail-closed）。
		// 非 vaccination 明細削除は claim-release 監査しない（ledger 方針: claim 解放の追跡が目的）。
		if releasedVaccinationID != nil {
			if err := s.logVaccinationClaimRelease(txCtx, clinicID, billingID, id, *releasedVaccinationID, input.StaffID); err != nil {
				return err
			}
		}

		slog.InfoContext(txCtx, "billing item deleted",
			slog.Uint64("clinic_id", clinicID),
			slog.Uint64("item_id", id),
			slog.Uint64("billing_id", billingID),
		)
		return nil
	})
}

// GetDiscountSuggestions は指定明細に適用可能な割引候補を返す (#81 Q-I スタッフ選択)。
func (s *billingItemService) GetDiscountSuggestions(ctx context.Context, clinicID, itemID uint64) ([]DiscountSuggestion, error) {
	item, err := s.repo.FindByID(ctx, clinicID, itemID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find billing item")
	}
	billing, err := s.billingRepo.FindByID(ctx, clinicID, item.BillingID)
	if err != nil || billing == nil {
		return nil, apperrors.Wrap(err, "failed to find billing")
	}
	ownerRate := s.resolveOwnerDiscountRate(ctx, clinicID, billing.OwnerID)
	var campaigns []*model.Campaign
	if s.campaignRepo != nil {
		campaigns, err = s.campaignRepo.FindAllApplicableForItem(ctx, clinicID, billing.ScheduledDate, item.Category, item.MerchandiseItemID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to find applicable campaigns for suggestions", "error", err)
			campaigns = nil // best-effort
		}
	}
	itemSubtotal := int64(float64(item.UnitPrice) * item.Quantity)
	return BuildDiscountSuggestions(itemSubtotal, campaigns, ownerRate), nil
}

// recalculateTotals は billing の全明細から subtotal/tax_total/total_amount を再計算して保存する。
// allowCompleted は BUG-009 の確定済み明細訂正経路でのみ true。
func (s *billingItemService) recalculateTotals(ctx context.Context, clinicID, billingID uint64, allowCompleted bool) error {
	items, err := s.repo.FindByBillingID(ctx, clinicID, billingID)
	if err != nil {
		return apperrors.Wrap(err, "failed to find billing items")
	}
	subtotal, taxTotal, totalAmount := CalculateBillingTotals(items)
	var writeErr error
	if allowCompleted {
		writeErr = s.repo.UpdateBillingTotalsForCompletedCorrection(ctx, clinicID, billingID, subtotal, taxTotal, totalAmount)
	} else {
		writeErr = s.repo.UpdateBillingTotals(ctx, clinicID, billingID, subtotal, taxTotal, totalAmount)
	}
	if writeErr != nil {
		return apperrors.Wrap(writeErr, "failed to update billing totals")
	}
	return nil
}

// recordCompletedItemCorrection は確定済み明細の理由付き訂正を監査し、締め済み日なら adjustment も書く（BUG-009）。
func (s *billingItemService) recordCompletedItemCorrection(
	ctx context.Context,
	clinicID, billingID uint64,
	billing *model.Billing,
	staffID *uint64,
	reason *string,
	operation string,
	itemID *uint64,
) error {
	if reason == nil || strings.TrimSpace(*reason) == "" {
		return apperrors.WrapInvalidInput("確定済み会計の明細修正には修正理由（post_close_reason）が必要です")
	}
	if billing == nil {
		return apperrors.WrapInternalServerError("billing is required for completed item correction")
	}
	// 締め済み日のみ台帳追記（close 無しで createPostCloseAdjustment すると Conflict になる）
	if s.closeRepo != nil {
		if err := s.createClosedDayAdjustmentIfNeeded(ctx, clinicID, billingID, billing, staffID, reason); err != nil {
			return err
		}
	}
	return s.logBillingItemPostCloseEdit(ctx, clinicID, billingID, itemID, staffID, reason, operation)
}

func (s *billingItemService) createClosedDayAdjustmentIfNeeded(
	ctx context.Context,
	clinicID, billingID uint64,
	billing *model.Billing,
	staffID *uint64,
	reason *string,
) error {
	closed, err := s.closeRepo.HasCloseOnDate(ctx, clinicID, billing.ScheduledDate)
	if err != nil {
		return apperrors.Wrap(err, "failed to re-check cash register close state for completed item correction")
	}
	if !closed {
		return nil
	}
	return createPostCloseAdjustment(ctx, s.closeRepo, clinicID, billingID, billing.ScheduledDate, strings.TrimSpace(*reason), staffID, 0)
}

// GetBilling は締め後編集判定用に請求ヘッダを返す。
func (s *billingItemService) GetBilling(ctx context.Context, clinicID, billingID uint64) (*model.Billing, error) {
	billing, err := s.billingRepo.FindByID(ctx, clinicID, billingID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find billing")
	}
	return billing, nil
}

// GetBillingForItem は明細 ID から親請求ヘッダを返す。
func (s *billingItemService) GetBillingForItem(ctx context.Context, clinicID, itemID uint64) (*model.Billing, error) {
	item, err := s.repo.FindByID(ctx, clinicID, itemID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find billing item")
	}
	billing, err := s.billingRepo.FindByID(ctx, clinicID, item.BillingID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find billing")
	}
	return billing, nil
}

// recordBillingItemPostClose は write 時に締め状態を再評価し、adjustment 台帳 + 監査を同一 tx で残す（W-013）。
// handler の IsPostClose フラグは候補 read に過ぎないため HasCloseOnDate で再評価する。
func (s *billingItemService) recordBillingItemPostClose(
	ctx context.Context,
	handlerFlag bool,
	clinicID, billingID uint64,
	billing *model.Billing,
	staffID *uint64,
	reason *string,
	operation string,
	itemID *uint64,
) error {
	postClose := handlerFlag
	if s.closeRepo != nil && billing != nil {
		closed, err := s.closeRepo.HasCloseOnDate(ctx, clinicID, billing.ScheduledDate)
		if err != nil {
			return apperrors.Wrap(err, "failed to re-check cash register close state for billing item")
		}
		if closed {
			postClose = true
		}
	}
	if !postClose {
		return nil
	}
	if reason == nil || strings.TrimSpace(*reason) == "" {
		return apperrors.WrapInvalidInput("レジ締め済み期間の会計編集には post_close_reason の入力が必要です")
	}
	// 明細経路は total 再計算後でも delta を行単位で確定しづらいため 0 とし、reason で追跡する。
	if err := createPostCloseAdjustment(ctx, s.closeRepo, clinicID, billingID, billing.ScheduledDate, *reason, staffID, 0); err != nil {
		return err
	}
	return s.logBillingItemPostCloseEdit(ctx, clinicID, billingID, itemID, staffID, reason, operation)
}

// logBillingItemPostCloseEdit はレジ締め済み期間の明細編集監査ログを記録する（#115 / BUG-463）。
// ambient tx に参加する LogEntryTx を使い、監査失敗時は明細変更ごとロールバックする（fail-closed）。
func (s *billingItemService) logBillingItemPostCloseEdit(
	ctx context.Context,
	clinicID, billingID uint64,
	itemID *uint64,
	staffID *uint64,
	reason *string,
	operation string,
) error {
	if s.auditTx == nil {
		return apperrors.WrapInternalServerError("billing audit dependency is required for post-close edits")
	}
	meta := map[string]any{
		"billing_id": billingID,
		"operation":  operation,
	}
	if itemID != nil {
		meta["item_id"] = *itemID
	}
	if reason != nil {
		meta["reason"] = *reason
	}
	if err := s.auditTx.LogEntryTx(ctx, &AuditEntry{
		ClinicID:   &clinicID,
		ActorID:    staffID,
		ActorType:  sharedkernel.AuditActorTypeFor(staffID),
		Action:     model.AuditActionBillingPostCloseEdit,
		Resource:   "billing_item",
		ResourceID: itemID,
		Metadata:   meta,
	}); err != nil {
		return apperrors.Wrap(err, "failed to write post_close_edit audit log")
	}
	return nil
}

// logVaccinationClaimRelease は明細削除による予防接種 claim 解放を監査する（BUG-440 / DEC-28 A）。
// ambient tx の LogEntryTx で fail-closed: 監査失敗時は delete + totals 再計算ごとロールバックする。
// metadata は ID のみ（PII 非格納）。
func (s *billingItemService) logVaccinationClaimRelease(
	ctx context.Context,
	clinicID, billingID, itemID, vaccinationID uint64,
	staffID *uint64,
) error {
	if s.auditTx == nil {
		return apperrors.WrapInternalServerError("billing audit dependency is required for vaccination claim release")
	}
	meta := map[string]any{
		"billing_id":     billingID,
		"item_id":        itemID,
		"vaccination_id": vaccinationID,
		"reason":         "billing_item_delete",
	}
	resourceID := itemID
	if err := s.auditTx.LogEntryTx(ctx, &AuditEntry{
		ClinicID:   &clinicID,
		ActorID:    staffID,
		ActorType:  sharedkernel.AuditActorTypeFor(staffID),
		Action:     model.AuditActionBillingVaccinationClaimRelease,
		Resource:   "billing_item",
		ResourceID: &resourceID,
		Metadata:   meta,
	}); err != nil {
		return apperrors.Wrap(err, "failed to write vaccination claim release audit log")
	}
	return nil
}
