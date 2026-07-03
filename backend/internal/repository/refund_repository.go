package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// RefundRepository は返金レコードのデータアクセスインターフェース
type RefundRepository interface {
	Create(ctx context.Context, refund *model.BillingRefund) error
	FindByBillingID(ctx context.Context, clinicID, billingID uint64) ([]model.BillingRefund, error)
	SumByBillingID(ctx context.Context, clinicID, billingID uint64) (int64, error)
	// SumByBillingIDAndPaymentMethod は指定支払方法(ENUM)への返金合計を返す(#60 Phase 2 方法別上限)。
	SumByBillingIDAndPaymentMethod(ctx context.Context, clinicID, billingID uint64, method model.PaymentMethod) (int64, error)
}

type refundRepository struct {
	db *gorm.DB
}

// NewRefundRepository はRefundRepositoryを初期化して返す
func NewRefundRepository(db *gorm.DB) RefundRepository {
	return &refundRepository{db: db}
}

func (r *refundRepository) Create(ctx context.Context, refund *model.BillingRefund) error {
	if err := dbOrTx(ctx, r.db).Create(refund).Error; err != nil {
		return apperrors.FromGORM(err, "billing_refund", "")
	}
	return nil
}

func (r *refundRepository) FindByBillingID(ctx context.Context, clinicID, billingID uint64) ([]model.BillingRefund, error) {
	refunds := make([]model.BillingRefund, 0)
	if err := r.db.WithContext(ctx).
		Preload("RefundedByStaff", "deleted_at IS NULL").
		Scopes(clinicScope(clinicID)).Where("billing_id = ?", billingID).
		Order("created_at DESC").
		Find(&refunds).Error; err != nil {
		return nil, apperrors.FromGORM(err, "billing_refund", fmt.Sprintf("billing_id=%d", billingID))
	}
	return refunds, nil
}

// BE-refactor.md R1-1 (D2): refund_service.Create が本メソッドを WithTx 内で txCtx 付きで呼び
// 「返金可能残額チェック（トランザクション内で再計算）」する。dbOrTx で ambient tx に参加しないと
// 同一 tx 内で直前に作成した返金の未コミット行を合計に含められず、TOCTOU チェックが機能しない。
func (r *refundRepository) SumByBillingID(ctx context.Context, clinicID, billingID uint64) (int64, error) {
	var total int64
	if err := dbOrTx(ctx, r.db).
		Model(&model.BillingRefund{}).
		Scopes(clinicScope(clinicID)).Where("billing_id = ?", billingID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error; err != nil {
		return 0, apperrors.FromGORM(err, "billing_refund", fmt.Sprintf("billing_id=%d", billingID))
	}
	return total, nil
}

func (r *refundRepository) SumByBillingIDAndPaymentMethod(ctx context.Context, clinicID, billingID uint64, method model.PaymentMethod) (int64, error) {
	var total int64
	if err := dbOrTx(ctx, r.db).
		Model(&model.BillingRefund{}).
		Scopes(clinicScope(clinicID)).Where("billing_id = ? AND payment_method = ?", billingID, method).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error; err != nil {
		return 0, apperrors.FromGORM(err, "billing_refund", fmt.Sprintf("billing_id=%d method=%s", billingID, method))
	}
	return total, nil
}
