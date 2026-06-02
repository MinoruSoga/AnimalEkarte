package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
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
	repo        repository.RefundRepository
	accountRepo repository.AccountingRepository
	auditSvc    AuditService
}

// NewRefundService はRefundServiceを初期化して返す
func NewRefundService(repo repository.RefundRepository, accountRepo repository.AccountingRepository, auditSvc AuditService) RefundService {
	return &refundService{repo: repo, accountRepo: accountRepo, auditSvc: auditSvc}
}

func (s *refundService) Create(ctx context.Context, clinicID, billingID uint64, input CreateRefundInput) (*model.BillingRefund, error) {
	amount := input.Amount
	if amount <= 0 {
		return nil, apperrors.WrapInvalidInput("amount must be positive")
	}

	// 請求が存在するか確認（マルチテナント保護）
	billing, err := s.accountRepo.FindByID(ctx, clinicID, billingID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get billing", "error", err)
		return nil, apperrors.Wrap(err, "failed to get billing")
	}

	// 支払済みの請求のみ返金可能
	if billing.Status != model.BillingStatusCompleted {
		return nil, apperrors.WrapInvalidInput("支払済みの請求のみ返金できます")
	}

	// BUG-142: 返金可能残額チェック（過剰返金防止）— Payment 有無に関わらず常にチェック
	alreadyRefunded, sumErr := s.repo.SumByBillingID(ctx, clinicID, billingID)
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

	// #60 Phase 2: 支払方法を指定した返金は、その支払方法で受け取った額を上限とする。
	// 混在支払い(payment_splits)で「カード ¥3,000 で支払った会計にカードから ¥3,000 超の返金」を防ぐ。
	// 支払方法未指定(nil)は上の全体チェックのみ(後方互換)。
	if input.PaymentMethod != nil {
		var receivedByMethod int64
		for i := range billing.PaymentSplits {
			if billing.PaymentSplits[i].Method == *input.PaymentMethod {
				receivedByMethod += billing.PaymentSplits[i].Amount
			}
		}
		if receivedByMethod == 0 {
			return nil, apperrors.WrapInvalidInput("指定された支払方法での支払がありません")
		}
		refundedByMethod, sumErr := s.repo.SumByBillingIDAndPaymentMethod(ctx, clinicID, billingID, *input.PaymentMethod)
		if sumErr != nil {
			return nil, apperrors.Wrap(sumErr, "sum refunds by payment method")
		}
		if amount > receivedByMethod-refundedByMethod {
			return nil, apperrors.WrapInvalidInput("支払方法別の返金上限を超えています")
		}
	}

	refund := &model.BillingRefund{
		ClinicID:      clinicID,
		BillingID:     billingID,
		Amount:        amount,
		Reason:        input.Reason,
		RefundedBy:    input.StaffID,
		PaymentMethod: input.PaymentMethod,
	}
	if err := s.repo.Create(ctx, refund); err != nil {
		slog.ErrorContext(ctx, "failed to create refund", "error", err)
		return nil, apperrors.Wrap(err, "failed to create refund")
	}

	slog.InfoContext(ctx, "refund created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("billing_id", billingID),
		slog.Int64("amount", amount))

	// 監査ログ記録（best-effort）
	_ = s.auditSvc.LogEntry(ctx, &AuditLogInput{
		ClinicID:   &clinicID,
		ActorID:    input.StaffID,
		ActorType:  "staff",
		Action:     "create",
		Resource:   "billing_refund",
		ResourceID: &refund.ID,
		NewValue:   refund,
	})

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
