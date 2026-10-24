package billing

import (
	"context"

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

// RefundService は返金の入口。
// held for EMR-506: 旧実装は billing_refunds（請求ごとの返金行。上限は請求合計・支払方法別の受領額）を読み書きしていたが、
// 置換先は受領ごとの返金イベント（fund_consumption_events の refund）で、上限は請求ではなく受領の残額になる。
// 旧の上限規則は新仕様へそのまま写さない。reader / writer が未接続のため、2 操作とも保存へ触れる前に固定エラー（422）で拒否する。
type RefundService interface {
	Create(ctx context.Context, clinicID, billingID uint64, input CreateRefundInput) error
	ListByBillingID(ctx context.Context, clinicID, billingID uint64) error
}

type refundService struct{}

// NewRefundService は quarantine 中の RefundService を返す。held for EMR-506。
func NewRefundService() RefundService {
	return &refundService{}
}

func (*refundService) Create(_ context.Context, _, _ uint64, input CreateRefundInput) error {
	if input.Amount <= 0 {
		return apperrors.WrapInvalidInput("amount must be positive")
	}
	return apperrors.WrapFinanceConsumerPending("返金の登録は、実返金の記録（受領ごとの返金イベント）の保存先が未接続のため利用できません")
}

func (*refundService) ListByBillingID(_ context.Context, _, _ uint64) error {
	return apperrors.WrapFinanceConsumerPending("返金一覧の参照は、実返金の記録の読取が未接続のため利用できません")
}
