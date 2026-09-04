package billing

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *refundService) createRefundInTx(
	txCtx context.Context,
	clinicID, billingID uint64,
	amount int64,
	input CreateRefundInput,
) (*model.BillingRefund, error) {
	// 請求が存在するか確認（マルチテナント保護）— FOR UPDATE でロック
	billing, err := s.accountRepo.LockAndFindByID(txCtx, clinicID, billingID)
	if err != nil {
		slog.ErrorContext(txCtx, "failed to get billing", "error", err)
		return nil, apperrors.Wrap(err, "failed to get billing")
	}

	// 支払済みの請求のみ返金可能
	if billing.Status != model.BillingStatusCompleted {
		return nil, apperrors.WrapInvalidInput("支払済みの請求のみ返金できます")
	}

	refundedAt := time.Now()
	if s.closeRepo != nil {
		for _, date := range uniqueCloseBoundaryDates(refundedAt, billing.ScheduledDate) {
			if err := s.closeRepo.LockCloseBoundary(txCtx, clinicID, date); err != nil {
				return nil, err
			}
		}
	}

	// 返金可能残額チェック（トランザクション内で再計算）
	alreadyRefunded, sumErr := s.repo.SumByBillingID(txCtx, clinicID, billingID)
	if sumErr != nil {
		return nil, apperrors.Wrap(sumErr, "sum refunds")
	}
	totalAmount := billing.TotalAmount
	if len(billing.Payments) > 0 {
		totalAmount = billing.Payments[0].TotalAmount
	}
	available := totalAmount - alreadyRefunded
	if amount > available {
		return nil, apperrors.WrapInvalidInput("リファンド額が請求残高を超えています")
	}

	if err := s.assertRefundPaymentMethodCap(txCtx, clinicID, billingID, amount, input.PaymentMethod, billing); err != nil {
		return nil, err
	}

	refund := &model.BillingRefund{
		ClinicID:      clinicID,
		BillingID:     billingID,
		Amount:        amount,
		Reason:        input.Reason,
		RefundedBy:    input.StaffID,
		PaymentMethod: input.PaymentMethod,
		RefundedAt:    refundedAt,
	}
	if err := s.repo.Create(txCtx, refund); err != nil {
		slog.ErrorContext(txCtx, "failed to create refund", "error", err)
		return nil, apperrors.Wrap(err, "failed to create refund")
	}

	slog.InfoContext(txCtx, "refund created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("billing_id", billingID),
		slog.Int64("amount", amount))

	// 監査ログ記録（fail-closed: 失敗→tx ロールバック→返金無効）
	if err := s.auditTxLogger.LogEntryTx(txCtx, &AuditEntry{
		ClinicID:   &clinicID,
		ActorID:    input.StaffID,
		ActorType:  model.AuditActorTypeStaff,
		Action:     model.AuditActionBillingRefundCreate,
		Resource:   "billing_refund",
		ResourceID: &refund.ID,
		NewValue: map[string]any{
			"amount":         refund.Amount,
			"reason":         refund.Reason,
			"payment_method": refund.PaymentMethod,
		},
	}); err != nil {
		slog.ErrorContext(txCtx, "audit log failed for refund create", "error", err, "refund_id", refund.ID)
		return nil, apperrors.Wrap(err, "failed to write refund audit log")
	}

	return refund, nil
}

func (s *refundService) assertRefundPaymentMethodCap(
	txCtx context.Context,
	clinicID, billingID uint64,
	amount int64,
	paymentMethod *model.PaymentMethod,
	billing *model.Billing,
) error {
	// #60 Phase 2: 支払方法を指定した返金は、その支払方法で受け取った額を上限とする。
	if paymentMethod == nil {
		return nil
	}
	var receivedByMethod int64
	for i := range billing.PaymentSplits {
		if billing.PaymentSplits[i].Method == *paymentMethod {
			receivedByMethod += billing.PaymentSplits[i].Amount
		}
	}
	if receivedByMethod == 0 {
		return apperrors.WrapInvalidInput("指定された支払方法での支払がありません")
	}
	refundedByMethod, sumErr := s.repo.SumByBillingIDAndPaymentMethod(txCtx, clinicID, billingID, *paymentMethod)
	if sumErr != nil {
		return apperrors.Wrap(sumErr, "sum refunds by payment method")
	}
	if amount > receivedByMethod-refundedByMethod {
		return apperrors.WrapInvalidInput("支払方法別の返金上限を超えています")
	}
	return nil
}
