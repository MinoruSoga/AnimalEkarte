// Package repository provides data access implementations for BillingItem entity.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// BillingItemRepository は billing_items テーブルの CRUD を担うインターフェース
type BillingItemRepository interface {
	FindByID(ctx context.Context, clinicID uint64, id uint64) (*model.BillingItem, error)
	FindByBillingID(ctx context.Context, clinicID, billingID uint64) ([]model.BillingItem, error)
	Create(ctx context.Context, item *model.BillingItem) error
	UpdateFields(ctx context.Context, clinicID uint64, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID uint64, id uint64) error
	UpdateBillingTotals(ctx context.Context, clinicID, billingID uint64, subtotal, taxTotal, totalAmount int64) error
}

type billingItemRepository struct{ db *gorm.DB }

// NewBillingItemRepository は BillingItemRepository を初期化して返す
func NewBillingItemRepository(db *gorm.DB) BillingItemRepository {
	return &billingItemRepository{db: db}
}

func (r *billingItemRepository) FindByID(ctx context.Context, clinicID uint64, id uint64) (*model.BillingItem, error) {
	var item model.BillingItem
	err := r.db.WithContext(ctx).
		Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.clinic_id = ?", clinicID).
		Where("billing_items.id = ?", id).
		First(&item).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "billing_item", fmt.Sprintf("%d", id))
	}
	return &item, nil
}

func (r *billingItemRepository) FindByBillingID(ctx context.Context, clinicID, billingID uint64) ([]model.BillingItem, error) {
	items := make([]model.BillingItem, 0)
	if err := r.db.WithContext(ctx).
		Joins("JOIN billings ON billings.id = billing_items.billing_id").
		Where("billings.clinic_id = ? AND billing_items.billing_id = ?", clinicID, billingID).
		Order("sort_order ASC, id ASC").
		Find(&items).Error; err != nil {
		return nil, apperrors.FromGORM(err, "billing_item", "")
	}
	return items, nil
}

func (r *billingItemRepository) Create(ctx context.Context, item *model.BillingItem) error {
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return apperrors.FromGORM(err, "billing_item", "")
	}
	return nil
}

func (r *billingItemRepository) UpdateFields(ctx context.Context, clinicID uint64, id uint64, fields map[string]any) error {
	// Verify clinic ownership before mutating
	if _, err := r.FindByID(ctx, clinicID, id); err != nil {
		return err
	}
	result := r.db.WithContext(ctx).
		Model(&model.BillingItem{}).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "billing_item", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("billing_item", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *billingItemRepository) Delete(ctx context.Context, clinicID uint64, id uint64) error {
	// Verify clinic ownership before mutating
	if _, err := r.FindByID(ctx, clinicID, id); err != nil {
		return err
	}
	result := r.db.WithContext(ctx).Delete(&model.BillingItem{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "billing_item", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("billing_item", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *billingItemRepository) UpdateBillingTotals(ctx context.Context, clinicID, billingID uint64, subtotal, taxTotal, totalAmount int64) error {
	result := r.db.WithContext(ctx).
		Model(&model.Billing{}).
		Where("clinic_id = ? AND id = ?", clinicID, billingID).
		Updates(map[string]any{
			"subtotal":     subtotal,
			"tax_total":    taxTotal,
			"total_amount": totalAmount,
		})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "billing", fmt.Sprintf("%d", billingID))
	}
	return nil
}
