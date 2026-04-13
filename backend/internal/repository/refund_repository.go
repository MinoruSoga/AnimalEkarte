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
}

type refundRepository struct {
	db *gorm.DB
}

// NewRefundRepository はRefundRepositoryを初期化して返す
func NewRefundRepository(db *gorm.DB) RefundRepository {
	return &refundRepository{db: db}
}

func (r *refundRepository) Create(ctx context.Context, refund *model.BillingRefund) error {
	if err := r.db.WithContext(ctx).Create(refund).Error; err != nil {
		return apperrors.FromGORM(err, "billing_refund", "")
	}
	return nil
}

func (r *refundRepository) FindByBillingID(ctx context.Context, clinicID, billingID uint64) ([]model.BillingRefund, error) {
	refunds := make([]model.BillingRefund, 0)
	if err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).Where("billing_id = ?", billingID).
		Order("created_at DESC").
		Find(&refunds).Error; err != nil {
		return nil, apperrors.FromGORM(err, "billing_refund", fmt.Sprintf("billing_id=%d", billingID))
	}
	return refunds, nil
}

func (r *refundRepository) SumByBillingID(ctx context.Context, clinicID, billingID uint64) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&model.BillingRefund{}).
		Scopes(clinicScope(clinicID)).Where("billing_id = ?", billingID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error; err != nil {
		return 0, apperrors.FromGORM(err, "billing_refund", fmt.Sprintf("billing_id=%d", billingID))
	}
	return total, nil
}
