// Package refund owns billing_refunds data access (BE8-4 batch8 — leaf domain).
package refund

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Repository is the data access interface for billing refunds.
type Repository interface {
	Create(ctx context.Context, refund *model.BillingRefund) error
	FindByBillingID(ctx context.Context, clinicID, billingID uint64) ([]model.BillingRefund, error)
	SumByBillingID(ctx context.Context, clinicID, billingID uint64) (int64, error)
	// SumByBillingIDAndPaymentMethod は指定支払方法(ENUM)への返金合計を返す(#60 Phase 2 方法別上限)。
	SumByBillingIDAndPaymentMethod(ctx context.Context, clinicID, billingID uint64, method model.PaymentMethod) (int64, error)
}

type repository struct {
	db *gorm.DB
}

// New constructs a Repository.
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, refund *model.BillingRefund) error {
	if err := repohelpers.DBOrTx(ctx, r.db).Create(refund).Error; err != nil {
		return apperrors.FromGORM(err, "billing_refund", "")
	}
	return nil
}

func (r *repository) FindByBillingID(ctx context.Context, clinicID, billingID uint64) ([]model.BillingRefund, error) {
	refunds := make([]model.BillingRefund, 0)
	if err := r.db.WithContext(ctx).
		Preload("RefundedByStaff", "deleted_at IS NULL").
		Scopes(repohelpers.ClinicScope(clinicID)).Where("billing_id = ?", billingID).
		Order("created_at DESC").
		Find(&refunds).Error; err != nil {
		return nil, apperrors.FromGORM(err, "billing_refund", fmt.Sprintf("billing_id=%d", billingID))
	}
	return refunds, nil
}

// BE-refactor.md R1-1 (D2): refund_service.Create が本メソッドを WithTx 内で txCtx 付きで呼び
// 「返金可能残額チェック（トランザクション内で再計算）」する。dbOrTx で ambient tx に参加しないと
// 同一 tx 内で直前に作成した返金の未コミット行を合計に含められず、TOCTOU チェックが機能しない。
func (r *repository) SumByBillingID(ctx context.Context, clinicID, billingID uint64) (int64, error) {
	var total int64
	if err := repohelpers.DBOrTx(ctx, r.db).
		Model(&model.BillingRefund{}).
		Scopes(repohelpers.ClinicScope(clinicID)).Where("billing_id = ?", billingID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error; err != nil {
		return 0, apperrors.FromGORM(err, "billing_refund", fmt.Sprintf("billing_id=%d", billingID))
	}
	return total, nil
}

func (r *repository) SumByBillingIDAndPaymentMethod(ctx context.Context, clinicID, billingID uint64, method model.PaymentMethod) (int64, error) {
	var total int64
	if err := repohelpers.DBOrTx(ctx, r.db).
		Model(&model.BillingRefund{}).
		Scopes(repohelpers.ClinicScope(clinicID)).Where("billing_id = ? AND payment_method = ?", billingID, method).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error; err != nil {
		return 0, apperrors.FromGORM(err, "billing_refund", fmt.Sprintf("billing_id=%d method=%s", billingID, method))
	}
	return total, nil
}
