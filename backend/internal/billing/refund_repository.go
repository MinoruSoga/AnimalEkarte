package billing

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// RefundRepository is the data access interface for billing refunds.
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

// New constructs a RefundRepository.
func NewRefundRepository(db *gorm.DB) RefundRepository {
	return &refundRepository{db: db}
}

func (r *refundRepository) Create(ctx context.Context, refund *model.BillingRefund) error {
	if refund == nil || refund.BillingID == 0 || refund.ClinicID == 0 {
		return apperrors.WrapInvalidInput(
			"billing refund requires billing_id and clinic_id",
		)
	}
	if err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		clinicID, err := lockBillingClinic(tx, refund.BillingID)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return apperrors.WrapInvalidInput("invalid billing reference")
			}
			return err
		}
		if clinicID != refund.ClinicID {
			return apperrors.WrapInvalidInput("invalid billing reference")
		}
		if refund.RefundedBy != nil {
			if err := lockActiveBillingStaffs(
				tx,
				clinicID,
				[]uint64{*refund.RefundedBy},
			); err != nil {
				return err
			}
		}
		if err := tx.Create(refund).Error; err != nil {
			return apperrors.FromGORM(err, "billing_refund", "")
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to create billing refund")
	}
	return nil
}

func (r *refundRepository) FindByBillingID(ctx context.Context, clinicID, billingID uint64) ([]model.BillingRefund, error) {
	refunds := make([]model.BillingRefund, 0)
	if err := r.db.WithContext(ctx).
		Preload(
			"RefundedByStaff",
			scopedBillingStaffPreload([]uint64{clinicID}),
		).
		Preload(
			"RefundedByStaff.ClinicAssignments",
			scopedStaffAssignmentsPreload([]uint64{clinicID}),
		).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("billing_id = ?", billingID).
		Where(`EXISTS (
			SELECT 1
			FROM billings
			WHERE billings.id = billing_refunds.billing_id
			  AND billings.clinic_id = billing_refunds.clinic_id
			  AND billings.deleted_at IS NULL
		)`).
		Order("created_at DESC").
		Find(&refunds).Error; err != nil {
		return nil, apperrors.FromGORM(err, "billing_refund", fmt.Sprintf("billing_id=%d", billingID))
	}
	for i := range refunds {
		sanitizeBillingStaff(
			&refunds[i].RefundedByStaff,
			refunds[i].RefundedBy,
			clinicID,
		)
	}
	return refunds, nil
}

// BE-refactor.md R1-1 (D2): refund_service.Create が本メソッドを WithTx 内で txCtx 付きで呼び
// 「返金可能残額チェック（トランザクション内で再計算）」する。dbOrTx で ambient tx に参加しないと
// 同一 tx 内で直前に作成した返金の未コミット行を合計に含められず、TOCTOU チェックが機能しない。
func (r *refundRepository) SumByBillingID(ctx context.Context, clinicID, billingID uint64) (int64, error) {
	var total int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.BillingRefund{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("billing_id = ?", billingID).
		Where(`EXISTS (
			SELECT 1
			FROM billings
			WHERE billings.id = billing_refunds.billing_id
			  AND billings.clinic_id = billing_refunds.clinic_id
			  AND billings.deleted_at IS NULL
		)`).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error; err != nil {
		return 0, apperrors.FromGORM(err, "billing_refund", fmt.Sprintf("billing_id=%d", billingID))
	}
	return total, nil
}

func (r *refundRepository) SumByBillingIDAndPaymentMethod(ctx context.Context, clinicID, billingID uint64, method model.PaymentMethod) (int64, error) {
	var total int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.BillingRefund{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("billing_id = ? AND payment_method = ?", billingID, method).
		Where(`EXISTS (
			SELECT 1
			FROM billings
			WHERE billings.id = billing_refunds.billing_id
			  AND billings.clinic_id = billing_refunds.clinic_id
			  AND billings.deleted_at IS NULL
		)`).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error; err != nil {
		return 0, apperrors.FromGORM(err, "billing_refund", fmt.Sprintf("billing_id=%d method=%s", billingID, method))
	}
	return total, nil
}
