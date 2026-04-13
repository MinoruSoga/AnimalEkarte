package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type AccountingRepository interface {
	FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error)
	Create(ctx context.Context, clinicID uint64, accounting *model.Billing) error
	UpdateFields(ctx context.Context, clinicID, billingID uint64, fields map[string]any) (*model.Billing, error)
	UpsertPayment(ctx context.Context, payment *model.Payment) error
	Delete(ctx context.Context, clinicID, id uint64) error
	CountItemsByBillingID(ctx context.Context, clinicID, billingID uint64) (int64, error)
}

type accountingRepository struct {
	db *gorm.DB
}

func NewAccountingRepository(db *gorm.DB) AccountingRepository {
	return &accountingRepository{db: db}
}

func (r *accountingRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error) {
	billings := make([]model.Billing, 0)
	var total int64

	q := r.db.WithContext(ctx).Model(&model.Billing{}).Scopes(clinicScope(clinicID))
	if petID != nil {
		q = q.Where("pet_id = ?", *petID)
	}
	if ownerID != nil {
		q = q.Where("owner_id = ?", *ownerID)
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if startDate != nil {
		q = q.Where("scheduled_date >= ?", *startDate)
	}
	if endDate != nil {
		q = q.Where("scheduled_date <= ?", *endDate)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "billing", "")
	}
	if err := q.Preload("Owner").Preload("Pet").Preload("Payments").Preload("Payments.PaidByStaff").Preload("Items").
		Offset((page - 1) * limit).Limit(limit).Order("scheduled_date DESC, created_at DESC").Find(&billings).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "billing", "")
	}

	// 返金合計をサブクエリで一括取得
	if len(billings) > 0 {
		ids := make([]uint64, 0, len(billings))
		for i := range billings {
			ids = append(ids, billings[i].ID)
		}
		type refundSum struct {
			BillingID uint64
			Total     int64
		}
		var sums []refundSum
		if err := r.db.WithContext(ctx).
			Model(&model.BillingRefund{}).
			Select("billing_id, COALESCE(SUM(amount), 0) AS total").
			Where("billing_id IN ?", ids).
			Group("billing_id").
			Scan(&sums).Error; err != nil {
			return nil, 0, apperrors.FromGORM(err, "billing_refund", "")
		}
		sumMap := make(map[uint64]int64, len(sums))
		for _, s := range sums {
			sumMap[s.BillingID] = s.Total
		}
		for i := range billings {
			billings[i].TotalRefundedAmount = sumMap[billings[i].ID]
		}
	}
	return billings, total, nil
}

func (r *accountingRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error) {
	var billing model.Billing
	err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Payments").
		Preload("Payments.PaidByStaff").
		Preload("Refunds").
		Preload("Refunds.RefundedByStaff").
		Preload("Owner").
		Preload("Pet").
		Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&billing).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "billing", fmt.Sprintf("%d", id))
	}
	// Preload した Refunds から TotalRefundedAmount を計算（FindAll と同じ算出ロジック）
	var total int64
	for i := range billing.Refunds {
		total += billing.Refunds[i].Amount
	}
	billing.TotalRefundedAmount = total
	return &billing, nil
}

func (r *accountingRepository) Create(ctx context.Context, clinicID uint64, accounting *model.Billing) error {
	accounting.ClinicID = clinicID
	if err := r.db.WithContext(ctx).Create(accounting).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("billing", accounting.ScheduledDate.String())
		}
		return apperrors.FromGORM(err, "billing", "")
	}
	return nil
}

// UpdateFields は指定フィールドのみを更新し、更新後のレコードを返す。
// map[string]any を使うことで GORM のゼロ値スキップ問題を回避する。
func (r *accountingRepository) UpdateFields(ctx context.Context, clinicID, billingID uint64, fields map[string]any) (*model.Billing, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Billing{}).
		Scopes(clinicScope(clinicID)).
		Where("id = ?", billingID).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "billing", fmt.Sprintf("%d", billingID))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("billing", fmt.Sprintf("%d", billingID))
	}
	var billing model.Billing
	if err := r.db.WithContext(ctx).
		Preload("Items").Preload("Payments").Preload("Payments.PaidByStaff").Preload("Refunds").Preload("Refunds.RefundedByStaff").Preload("Owner").Preload("Pet").
		Scopes(clinicScope(clinicID)).
		First(&billing, "id = ?", billingID).Error; err != nil {
		return nil, apperrors.FromGORM(err, "billing", fmt.Sprintf("%d", billingID))
	}
	return &billing, nil
}

func (r *accountingRepository) UpsertPayment(ctx context.Context, payment *model.Payment) error {
	result := r.db.WithContext(ctx).
		Where(model.Payment{BillingID: payment.BillingID}).
		Assign(model.Payment{
			Subtotal:        payment.Subtotal,
			TaxTotal:        payment.TaxTotal,
			TotalAmount:     payment.TotalAmount,
			InsuranceName:   payment.InsuranceName,
			InsuranceRatio:  payment.InsuranceRatio,
			InsuranceAmount: payment.InsuranceAmount,
			DiscountAmount:  payment.DiscountAmount,
			BillingAmount:   payment.BillingAmount,
			ReceivedAmount:  payment.ReceivedAmount,
			ChangeAmount:    payment.ChangeAmount,
			Method:          payment.Method,
			PaidBy:          payment.PaidBy,
		}).
		FirstOrCreate(payment)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "payment", fmt.Sprintf("billing_id=%d", payment.BillingID))
	}
	return nil
}

func (r *accountingRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(&model.Billing{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "billing", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("billing", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *accountingRepository) CountItemsByBillingID(ctx context.Context, clinicID, billingID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.BillingItem{}).
		Joins("JOIN billings ON billing_items.billing_id = billings.id").
		Where("billings.clinic_id = ? AND billing_items.billing_id = ?", clinicID, billingID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "billing_item", fmt.Sprintf("billing_id=%d", billingID))
	}
	return count, nil
}
