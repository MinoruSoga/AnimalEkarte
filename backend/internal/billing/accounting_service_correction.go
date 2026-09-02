package billing

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
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
// 監査ログは before/after・理由・メモ・実行者を記録する（BE-refactor.md R1-2 で refund と同じく
// fail-closed 化済み。LogEntryTx が失敗すると訂正自体もロールバックされる）。
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
		return s.correctCreditPaymentInTx(txCtx, input)
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

func (s *accountingService) correctCreditPaymentInTx(txCtx context.Context, input *CorrectCreditPaymentInput) error {
	billing, err := s.repo.LockAndFindByID(txCtx, input.ClinicID, input.BillingID)
	if err != nil {
		slog.ErrorContext(txCtx, "failed to lock billing for credit correction", "error", err, "billing_id", input.BillingID)
		return apperrors.Wrap(err, "failed to lock billing for credit correction")
	}

	if billing.Status != model.BillingStatusCompleted {
		return apperrors.WrapInvalidInput("確定済みの会計のみクレジット訂正できます")
	}

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
	oldBillingAmount := billing.Payments[0].BillingAmount

	corrected := make([]model.PaymentSplit, len(billing.PaymentSplits))
	copy(corrected, billing.PaymentSplits)
	corrected[targetIdx].Amount = input.Amount
	corrected[targetIdx].ReceivedAmount = 0
	corrected[targetIdx].ChangeAmount = 0
	corrected[targetIdx].PaidBy = input.StaffID

	var newBillingAmount int64
	for i := range corrected {
		newBillingAmount += corrected[i].Amount
	}

	if err := validatePaymentSplits(toValidationInputs(corrected), nil); err != nil {
		return apperrors.Wrap(err, "failed to validate corrected payment splits")
	}

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

	if oldBillingAmount != newBillingAmount {
		slog.InfoContext(txCtx, "credit correction changed billing amount",
			slog.Uint64("clinic_id", input.ClinicID),
			slog.Uint64("billing_id", input.BillingID),
			slog.Int64("old_billing_amount", oldBillingAmount),
			slog.Int64("new_billing_amount", newBillingAmount))
	}

	postClose, err := s.resolvePostCloseInTx(txCtx, input.ClinicID, billing.ScheduledDate, input.IsPostClose)
	if err != nil {
		return err
	}
	if postClose {
		input.IsPostClose = true
		slog.WarnContext(txCtx, "credit correction on closed period",
			slog.Uint64("clinic_id", input.ClinicID),
			slog.Uint64("billing_id", input.BillingID),
			slog.String("scheduled_date", billing.ScheduledDate.Format(time.DateOnly)))
		if err := s.recordPostCloseAdjustment(
			txCtx,
			input.ClinicID,
			input.BillingID,
			billing.ScheduledDate,
			input.Reason,
			input.StaffID,
			newBillingAmount-oldBillingAmount,
		); err != nil {
			return err
		}
	}

	return s.logCreditCorrection(txCtx, input, &target, oldBillingAmount, newBillingAmount, billing.ScheduledDate)
}

// logCreditCorrection はクレジット訂正の監査ログを記録する（before/after・理由・メモ・実行者）。
// M-1: billing_amount の before/after と差額（delta）を明示し、売上への影響額を追跡可能にする。
// M-2: 締め済み期間への訂正（IsPostClose）は post_close フラグと対象締めの識別子（予定日）を記録する。
// BE-refactor.md R1-2: CorrectCreditPayment の ambient tx に参加する LogEntryTx を使う（fail-closed）。
// 呼び出し元は返されたエラーで tx をロールバックし、監査失敗時に訂正自体も無効にする。
func (s *accountingService) logCreditCorrection(ctx context.Context, input *CorrectCreditPaymentInput, before *model.PaymentSplit, oldBillingAmount, newBillingAmount int64, scheduledDate time.Time) error {
	// BIL-02: missing audit dependency must fail closed (mirror logPostCloseEdit).
	if s.auditTx == nil {
		return apperrors.WrapInternalServerError("billing audit dependency is required for credit correction")
	}
	actorType := sharedkernel.AuditActorTypeFor(input.StaffID)
	billingID := input.BillingID
	metadata := map[string]any{
		"reason": input.Reason,
		"memo":   input.Memo,
		// billing_amount の増減額を明示（売上への影響額の追跡用）。
		"billing_amount_delta": newBillingAmount - oldBillingAmount,
	}
	// M-2: 締め済み期間への訂正は監査エントリで可視化する（対象締めの識別子として予定日を記録）。
	if input.IsPostClose {
		metadata["post_close"] = true
		metadata["post_close_date"] = scheduledDate.Format(time.DateOnly)
	}
	if err := s.auditTx.LogEntryTx(ctx, &AuditEntry{
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
			"billing_amount":  oldBillingAmount,
		},
		NewValue: map[string]any{
			"method":          string(input.Method),
			"amount":          input.Amount,
			"received_amount": int64(0),
			"change_amount":   int64(0),
			"billing_amount":  newBillingAmount,
		},
		Metadata: metadata,
	}); err != nil {
		return apperrors.Wrap(err, "failed to write credit correction audit log")
	}
	return nil
}
