package billing

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// CreateRefundInput は返金作成の入力DTO。
type CreateRefundInput struct {
	StaffID       *uint64
	Amount        int64
	Reason        string
	PaymentMethod *model.PaymentMethod // 返金先の支払手段（nullable・ENUM。会計 payment_splits.method と同体系）
}

// RefundService は返金ビジネスロジックのインターフェース
type RefundService interface {
	Create(ctx context.Context, clinicID, billingID uint64, input CreateRefundInput) (*model.BillingRefund, error)
	ListByBillingID(ctx context.Context, clinicID, billingID uint64) ([]model.BillingRefund, error)
}

type refundService struct {
	repo          RefundRepository
	accountRepo   accountingBillingView
	auditTxLogger billingAuditTxLogger // fail-closed: ambient tx に参加して監査を書く（#211）
	transactor    Transactor
}

// NewRefundService はRefundServiceを初期化して返す
func NewRefundService(repo RefundRepository, accountRepo accountingBillingView, auditTxLogger billingAuditTxLogger, transactor Transactor) RefundService {
	return &refundService{repo: repo, accountRepo: accountRepo, auditTxLogger: auditTxLogger, transactor: transactor}
}

func (s *refundService) Create(ctx context.Context, clinicID, billingID uint64, input CreateRefundInput) (*model.BillingRefund, error) {
	amount := input.Amount
	if amount <= 0 {
		return nil, apperrors.WrapInvalidInput("amount must be positive")
	}

	var refund *model.BillingRefund
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// 請求が存在するか確認（マルチテナント保護）— FOR UPDATE でロック
		billing, err := s.accountRepo.LockAndFindByID(txCtx, clinicID, billingID)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get billing", "error", err)
			return apperrors.Wrap(err, "failed to get billing")
		}

		// 支払済みの請求のみ返金可能
		if billing.Status != model.BillingStatusCompleted {
			return apperrors.WrapInvalidInput("支払済みの請求のみ返金できます")
		}

		// 返金可能残額チェック（トランザクション内で再計算）
		alreadyRefunded, sumErr := s.repo.SumByBillingID(txCtx, clinicID, billingID)
		if sumErr != nil {
			return apperrors.Wrap(sumErr, "sum refunds")
		}
		totalAmount := billing.TotalAmount
		if len(billing.Payments) > 0 {
			totalAmount = billing.Payments[0].TotalAmount
		}
		available := totalAmount - alreadyRefunded
		if amount > available {
			return apperrors.WrapInvalidInput("リファンド額が請求残高を超えています")
		}

		// #60 Phase 2: 支払方法を指定した返金は、その支払方法で受け取った額を上限とする。
		if input.PaymentMethod != nil {
			var receivedByMethod int64
			for i := range billing.PaymentSplits {
				if billing.PaymentSplits[i].Method == *input.PaymentMethod {
					receivedByMethod += billing.PaymentSplits[i].Amount
				}
			}
			if receivedByMethod == 0 {
				return apperrors.WrapInvalidInput("指定された支払方法での支払がありません")
			}
			refundedByMethod, sumErr := s.repo.SumByBillingIDAndPaymentMethod(txCtx, clinicID, billingID, *input.PaymentMethod)
			if sumErr != nil {
				return apperrors.Wrap(sumErr, "sum refunds by payment method")
			}
			if amount > receivedByMethod-refundedByMethod {
				return apperrors.WrapInvalidInput("支払方法別の返金上限を超えています")
			}
		}

		refund = &model.BillingRefund{
			ClinicID:      clinicID,
			BillingID:     billingID,
			Amount:        amount,
			Reason:        input.Reason,
			RefundedBy:    input.StaffID,
			PaymentMethod: input.PaymentMethod,
			RefundedAt:    time.Now(),
		}
		if err := s.repo.Create(txCtx, refund); err != nil {
			slog.ErrorContext(txCtx, "failed to create refund", "error", err)
			return apperrors.Wrap(err, "failed to create refund")
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
			return apperrors.Wrap(err, "failed to write refund audit log")
		}

		return nil
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to create refund in transaction")
	}

	return refund, nil
}

func (s *refundService) ListByBillingID(ctx context.Context, clinicID, billingID uint64) ([]model.BillingRefund, error) {
	// マルチテナント保護
	if _, err := s.accountRepo.FindByID(ctx, clinicID, billingID); err != nil {
		slog.ErrorContext(ctx, "failed to get billing", "error", err)
		return nil, apperrors.Wrap(err, "failed to get billing")
	}
	items, err := s.repo.FindByBillingID(ctx, clinicID, billingID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list refunds", "error", err)
		return nil, apperrors.Wrap(err, "failed to list refunds")
	}
	return items, nil
}
