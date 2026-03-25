package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type AccountingRepository interface {
	FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status *string, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error)
	Create(ctx context.Context, clinicID uint64, accounting *model.Billing) error
	UpdateFields(ctx context.Context, clinicID, billingID uint64, fields map[string]any) (*model.Billing, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

type accountingRepository struct {
	db *gorm.DB
}

func NewAccountingRepository(db *gorm.DB) AccountingRepository {
	return &accountingRepository{db: db}
}

func (r *accountingRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status *string, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error) {
	billings := make([]model.Billing, 0)
	var total int64

	q := r.db.WithContext(ctx).Model(&model.Billing{}).Where("clinic_id = ?", clinicID)
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
		return nil, 0, apperrors.Wrap(err, "count billings")
	}
	if err := q.Preload("Owner").Preload("Pet").Preload("Payments").Preload("Items").
		Offset((page - 1) * limit).Limit(limit).Order("scheduled_date DESC, created_at DESC").Find(&billings).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find billings")
	}

	// 返金合計をサブクエリで一括取得
	if len(billings) > 0 {
		ids := make([]uint64, 0, len(billings))
		for _, b := range billings {
			ids = append(ids, b.ID)
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
			return nil, 0, apperrors.Wrap(err, "sum refunds")
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
	if err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Payments").
		Preload("Refunds").
		Preload("Owner").
		Preload("Pet").
		First(&billing, "id = ? AND clinic_id = ?", id, clinicID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("billing", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find billing by id")
	}
	return &billing, nil
}

func (r *accountingRepository) Create(ctx context.Context, clinicID uint64, accounting *model.Billing) error {
	accounting.ClinicID = clinicID
	if err := r.db.WithContext(ctx).Create(accounting).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("billing", accounting.ScheduledDate.String())
		}
		return apperrors.Wrap(err, "create billing")
	}
	return nil
}

// UpdateFields は指定フィールドのみを更新し、更新後のレコードを返す。
// map[string]any を使うことで GORM のゼロ値スキップ問題を回避する。
func (r *accountingRepository) UpdateFields(ctx context.Context, clinicID, billingID uint64, fields map[string]any) (*model.Billing, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Billing{}).
		Where("clinic_id = ? AND id = ?", clinicID, billingID).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.Wrap(result.Error, fmt.Sprintf("update billing id=%d", billingID))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("billing", fmt.Sprintf("%d", billingID))
	}
	var billing model.Billing
	if err := r.db.WithContext(ctx).
		Preload("Items").Preload("Payments").Preload("Refunds").Preload("Owner").Preload("Pet").
		First(&billing, "clinic_id = ? AND id = ?", clinicID, billingID).Error; err != nil {
		return nil, apperrors.Wrap(err, fmt.Sprintf("find billing after update id=%d", billingID))
	}
	return &billing, nil
}

func (r *accountingRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.Billing{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete billing")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("billing", fmt.Sprintf("%d", id))
	}
	return nil
}
