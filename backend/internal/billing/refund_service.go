package billing

import (
	"context"
	"log/slog"

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
	closeRepo     CashRegisterCloseRepository
}

type refundServiceOption func(*refundService)

// WithRefundCloseRepository serializes refunds with cash-register close on the
// refund occurrence date and the original billing date.
func WithRefundCloseRepository(repo CashRegisterCloseRepository) refundServiceOption {
	return func(s *refundService) {
		s.closeRepo = repo
	}
}

// NewRefundService はRefundServiceを初期化して返す
func NewRefundService(
	repo RefundRepository,
	accountRepo accountingBillingView,
	auditTxLogger billingAuditTxLogger,
	transactor Transactor,
	opts ...refundServiceOption,
) RefundService {
	s := &refundService{repo: repo, accountRepo: accountRepo, auditTxLogger: auditTxLogger, transactor: transactor}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *refundService) Create(ctx context.Context, clinicID, billingID uint64, input CreateRefundInput) (*model.BillingRefund, error) {
	amount := input.Amount
	if amount <= 0 {
		return nil, apperrors.WrapInvalidInput("amount must be positive")
	}

	var refund *model.BillingRefund
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		created, err := s.createRefundInTx(txCtx, clinicID, billingID, amount, input)
		refund = created
		return err
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
