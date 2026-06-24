package service

import (
	"context"
	"log/slog"
	"strings"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// isCardPaymentMethod は訂正フロー（#189）の対象となるカード系支払い手段か判定する。
// クレジットカードと電子マネーはレジ実機（カード端末）での金額入力を伴い、同種の打ち間違いリスクを持つ。
// 現金は #188（お釣り直接上書き）、銀行振込は端末入力を伴わないため本フロー対象外。
func isCardPaymentMethod(m model.PaymentMethod) bool {
	return m == model.PaymentMethodCreditCard || m == model.PaymentMethodElectronicMoney
}

// toValidationInputs は保存済み payment_splits を validatePaymentSplits 用の入力DTOへ変換する。
// 現金内訳は保存済みお釣りが派生値（received - amount）と異なる場合に #188 の上書きフラグを再導出し、
// 既存のお釣り上書きが訂正処理で誤って整合検証に弾かれる（#188 回帰）のを防ぐ。
func toValidationInputs(splits []model.PaymentSplit) []PaymentSplitInput {
	inputs := make([]PaymentSplitInput, 0, len(splits))
	for i := range splits {
		s := splits[i]
		in := PaymentSplitInput{
			Method:          s.Method,
			PaymentMethodID: s.PaymentMethodID,
			Amount:          s.Amount,
			ReceivedAmount:  s.ReceivedAmount,
			ChangeAmount:    s.ChangeAmount,
		}
		if s.Method == model.PaymentMethodCash && s.ChangeAmount != s.ReceivedAmount-s.Amount {
			in.ChangeOverride = true
		}
		inputs = append(inputs, in)
	}
	return inputs
}

// CorrectCreditPayment は確定済み会計のクレジット（カード）金額を確定後に訂正する（#189）。
// refund_service.Create と同じく WithTx + LockAndFindByID（FOR UPDATE）で TOCTOU を防ぎ、
// 訂正対象のカード内訳のみを書き換え、Payment.billing_amount を内訳合計に再計算する。
// 医療費（subtotal/tax/total_amount）・保険・他の支払い内訳は不変（最小スコープ）。
// 監査ログは before/after・理由・メモ・実行者を記録する（refund と同じく best-effort）。
func (s *accountingService) CorrectCreditPayment(ctx context.Context, input *CorrectCreditPaymentInput) (*model.Billing, error) {
	// 入力検証（トランザクション外・低コスト）
	if strings.TrimSpace(input.Reason) == "" {
		return nil, apperrors.WrapInvalidInput("訂正理由は必須です")
	}
	if !isCardPaymentMethod(input.Method) {
		return nil, apperrors.WrapInvalidInput("クレジットカードまたは電子マネーの訂正のみ対応しています")
	}
	if input.Amount <= 0 {
		return nil, apperrors.WrapInvalidInput("訂正金額は1円以上でなければなりません")
	}

	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// FOR UPDATE で確定済み会計を行ロック取得（多重訂正・返金との競合を防止）
		billing, err := s.repo.LockAndFindByID(txCtx, input.ClinicID, input.BillingID)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to lock billing for credit correction", "error", err, "billing_id", input.BillingID)
			return apperrors.Wrap(err, "failed to lock billing for credit correction")
		}

		// 確定済み（completed）以外は訂正不可。確定前は通常の編集フロー（PATCH）で対応する。
		if billing.Status != model.BillingStatusCompleted {
			return apperrors.WrapInvalidInput("確定済みの会計のみクレジット訂正できます")
		}

		// 訂正対象のカード内訳を特定
		targetIdx := -1
		for i := range billing.PaymentSplits {
			if billing.PaymentSplits[i].Method == input.Method {
				targetIdx = i
				break
			}
		}
		if targetIdx < 0 {
			return apperrors.WrapInvalidInput("対象のカード支払い内訳が存在しません")
		}
		target := billing.PaymentSplits[targetIdx]

		if len(billing.Payments) == 0 {
			return apperrors.WrapInvalidInput("支払い情報が存在しません")
		}

		// 訂正後の内訳を新規スライスとして構築（既存はミューテートしない）
		corrected := make([]model.PaymentSplit, len(billing.PaymentSplits))
		copy(corrected, billing.PaymentSplits)
		corrected[targetIdx].Amount = input.Amount
		corrected[targetIdx].ReceivedAmount = 0 // カード内訳は受領額・お釣りを持たない（既存の格納規約）
		corrected[targetIdx].ChangeAmount = 0
		corrected[targetIdx].PaidBy = input.StaffID

		var newBillingAmount int64
		for i := range corrected {
			newBillingAmount += corrected[i].Amount
		}

		// 整合検証（重複手段・金額>0・現金お釣り整合。#188 上書きは再導出済みで維持）
		if err := validatePaymentSplits(toValidationInputs(corrected), &newBillingAmount); err != nil {
			return apperrors.Wrap(err, "failed to validate corrected payment splits")
		}

		// 支払いヘッダは既存値を保全し billing_amount のみ再計算（医療費・保険・割引は不変）
		payment := billing.Payments[0]
		payment.BillingAmount = newBillingAmount
		if err := s.repo.SavePayment(txCtx, &payment); err != nil {
			slog.ErrorContext(txCtx, "failed to save payment during credit correction", "error", err, "billing_id", input.BillingID)
			return apperrors.Wrap(err, "failed to save payment during credit correction")
		}
		if err := s.repo.SavePaymentSplits(txCtx, corrected); err != nil {
			slog.ErrorContext(txCtx, "failed to save payment splits during credit correction", "error", err, "billing_id", input.BillingID)
			return apperrors.Wrap(err, "failed to save payment splits during credit correction")
		}

		slog.InfoContext(txCtx, "credit payment corrected",
			slog.Uint64("clinic_id", input.ClinicID),
			slog.Uint64("billing_id", input.BillingID),
			slog.String("method", string(input.Method)),
			slog.Int64("before_amount", target.Amount),
			slog.Int64("after_amount", input.Amount))

		// 監査ログ（best-effort。auditRepository は tx 非参加のため失敗しても訂正は確定する）
		s.logCreditCorrection(txCtx, input, &target, newBillingAmount)
		return nil
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to correct credit payment in transaction")
	}

	// 訂正後の最新レコードを返す
	updated, err := s.repo.FindByID(ctx, input.ClinicID, input.BillingID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to reload accounting after credit correction", "error", err, "billing_id", input.BillingID)
		return nil, apperrors.Wrap(err, "failed to reload accounting after credit correction")
	}
	return updated, nil
}

// logCreditCorrection はクレジット訂正の監査ログを記録する（before/after・理由・メモ・実行者）。
func (s *accountingService) logCreditCorrection(ctx context.Context, input *CorrectCreditPaymentInput, before *model.PaymentSplit, newBillingAmount int64) {
	if s.auditSvc == nil {
		return
	}
	actorType := model.AuditActorTypeSystem
	if input.StaffID != nil {
		actorType = model.AuditActorTypeStaff
	}
	billingID := input.BillingID
	if err := s.auditSvc.LogEntry(ctx, &AuditLogInput{
		ClinicID:   &input.ClinicID,
		ActorID:    input.StaffID,
		ActorType:  actorType,
		Action:     model.AuditActionBillingCreditCorrection,
		Resource:   "billing",
		ResourceID: &billingID,
		OldValue: map[string]any{
			"method":          string(before.Method),
			"amount":          before.Amount,
			"received_amount": before.ReceivedAmount,
			"change_amount":   before.ChangeAmount,
		},
		NewValue: map[string]any{
			"method":          string(input.Method),
			"amount":          input.Amount,
			"received_amount": int64(0),
			"change_amount":   int64(0),
			"billing_amount":  newBillingAmount,
		},
		Metadata: map[string]any{
			"reason": input.Reason,
			"memo":   input.Memo,
		},
	}); err != nil {
		slog.WarnContext(ctx, "audit log failed for credit correction", "error", err, "billing_id", input.BillingID)
	}
}
