// Package billing provides billing item update use cases.
package billing

import (
	"context"
	"log/slog"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
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
		return nil, apperrors.Wrap(err, "failed to get billing item")
	}
	if err := validateUpdateBillingItemDecimals(input); err != nil {
		return nil, err
	}
	// EMR-229: name 更新を許可するが空文字・空白のみは create と同じく拒否する
	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		return nil, apperrors.WrapInvalidInput("商品名は必須です")
	}

	fields := buildBillingItemUpdate(input)
	if len(fields) == 0 {
		return item, nil
	}

	var updated *model.BillingItem
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// BUG-463: 親請求をロックし、draft のときだけ更新する。確定済みの訂正（correct）は未接続のため拒否する。
		billing, err := s.billingRepo.LockAndFindByID(txCtx, clinicID, item.BillingID)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock billing before updating item")
		}
		if err := rejectItemUpdateUnlessDraft(billing, "更新"); err != nil {
			return err
		}

		if err := s.repo.Update(txCtx, clinicID, id, *input); err != nil {
			return apperrors.Wrap(err, "failed to update billing item")
		}

		itemID := id
		if err := s.recordBillingItemPostClose(txCtx, input.IsPostClose, clinicID, item.BillingID, billing, input.StaffID, input.PostCloseReason, "update", &itemID); err != nil {
			return err
		}

		slog.InfoContext(txCtx, "billing item updated",
			slog.Uint64("clinic_id", clinicID),
			slog.Uint64("item_id", id),
			slog.Uint64("billing_id", item.BillingID),
		)
		updated, err = s.repo.FindByID(txCtx, clinicID, id)
		if err != nil {
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

	return s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		billing, err := s.billingRepo.LockAndFindByID(txCtx, clinicID, billingID)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock billing before deleting item")
		}
		if err := rejectIfBillingFinalized(billing, "削除"); err != nil {
			return err
		}

		if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
			return apperrors.Wrap(err, "failed to delete billing item")
		}

		// BUG-463 residual / W-013: 締め済み日の削除は、締め後の変更記録が未接続のため固定エラーで拒否する（held for EMR-509）。
		// 予防接種 claim の解放監査（BUG-440）は、明細側の予防接種列が廃止され請求元が billing_source_bindings へ移ったため、対象行が作られない。
		itemID := id
		if err := s.recordBillingItemPostClose(txCtx, input.IsPostClose, clinicID, billingID, billing, input.StaffID, input.PostCloseReason, "delete", &itemID); err != nil {
			return err
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
	itemSubtotal, err := discountBaseAmount(item.UnitPrice, item.Quantity)
	if err != nil {
		return nil, err
	}
	return BuildDiscountSuggestions(itemSubtotal, campaigns, ownerRate), nil
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
		if err := s.closeRepo.LockCloseBoundary(ctx, clinicID, billing.ScheduledDate); err != nil {
			return err
		}
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
	// held for EMR-509: 締め後の台帳追記（cash_register_close_adjustments）は退役し、後継の遅延イベント・実査訂正は未接続のため、
	// 締め済み期間の明細変更は書き込み前に固定エラー（422）で拒否する。reason・staffID・operation・itemID は 509 の writer 接続まで未使用。
	return apperrors.WrapFinanceConsumerPending("レジ締め済み期間の会計明細の変更は、締め後の変更記録（遅延イベント・実査訂正）が未接続のため利用できません")
}
