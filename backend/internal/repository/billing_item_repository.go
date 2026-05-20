// Package repository provides data access implementations for BillingItem entity.
package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// BillingItemRepository は billing_items テーブルの CRUD を担うインターフェース
type BillingItemRepository interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.BillingItem, error)
	FindByBillingID(ctx context.Context, clinicID, billingID uint64) ([]model.BillingItem, error)
	Create(ctx context.Context, item *model.BillingItem) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
	UpdateBillingTotals(ctx context.Context, clinicID, billingID uint64, subtotal, taxTotal, totalAmount int64) error
	// HasItemByOwnerSince は指定飼い主の請求アイテムに names いずれかが存在するか返す（FEAT-379）。
	HasItemByOwnerSince(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error)
	// HasFoodPurchaseByOwnerSince は names 指定時は名前で、未指定時は category=food で判定する（FEAT-379）。
	HasFoodPurchaseByOwnerSince(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error)
	// FindOwnersByCategoryPurchaseDate は指定カテゴリの最終購入日が purchaseDate と一致する飼い主IDリストを返す（FEAT-383）。
	FindOwnersByCategoryPurchaseDate(ctx context.Context, clinicID uint64, category string, purchaseDate time.Time) ([]uint64, error)
}

type billingItemRepository struct{ db *gorm.DB }

// NewBillingItemRepository は BillingItemRepository を初期化して返す
func NewBillingItemRepository(db *gorm.DB) BillingItemRepository {
	return &billingItemRepository{db: db}
}

func (r *billingItemRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.BillingItem, error) {
	var item model.BillingItem
	err := r.db.WithContext(ctx).
		Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
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
		Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Where("billing_items.billing_id = ?", billingID).
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

func (r *billingItemRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.BillingItem{}).
		Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Where("billing_items.id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "billing_item", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("billing_item", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *billingItemRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Delete(&model.BillingItem{}, "billing_items.id = ?", id)
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
		Scopes(clinicScope(clinicID)).Where("id = ?", billingID).
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

func (r *billingItemRepository) HasItemByOwnerSince(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error) {
	if len(names) == 0 {
		return false, nil
	}
	var count int64
	err := r.db.WithContext(ctx).Model(&model.BillingItem{}).
		Joins("JOIN billings ON billings.id = billing_items.billing_id").
		Where("billings.clinic_id = ? AND billings.owner_id = ? AND billings.issued_at >= ? AND billings.deleted_at IS NULL", clinicID, ownerID, since).
		Where("billing_items.name IN ? AND billing_items.deleted_at IS NULL", names).
		Count(&count).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "billing_item", fmt.Sprintf("clinic:%d owner:%d", clinicID, ownerID))
	}
	return count > 0, nil
}

// FindOwnersByCategoryPurchaseDate は指定カテゴリの最終購入日が purchaseDate と一致する飼い主IDリストを返す（FEAT-383）。
// billings.issued_at::date の MAX が purchaseDate と一致する飼い主を返す。
func (r *billingItemRepository) FindOwnersByCategoryPurchaseDate(ctx context.Context, clinicID uint64, category string, purchaseDate time.Time) ([]uint64, error) {
	target := purchaseDate.Format("2006-01-02")
	type row struct{ OwnerID uint64 }
	var rows []row
	err := r.db.WithContext(ctx).Raw(`
		SELECT billings.owner_id AS owner_id
		FROM billing_items
		JOIN billings ON billings.id = billing_items.billing_id
		WHERE billings.clinic_id = ?
		  AND billings.deleted_at IS NULL
		  AND billing_items.deleted_at IS NULL
		  AND billing_items.category = ?
		GROUP BY billings.owner_id
		HAVING MAX(billings.issued_at::date) = ?::date
	`, clinicID, category, target).Scan(&rows).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "billing_item", fmt.Sprintf("clinic=%d category=%s date=%s", clinicID, category, target))
	}
	ids := make([]uint64, len(rows))
	for i, r := range rows {
		ids[i] = r.OwnerID
	}
	return ids, nil
}

func (r *billingItemRepository) HasFoodPurchaseByOwnerSince(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.BillingItem{}).
		Joins("JOIN billings ON billings.id = billing_items.billing_id").
		Where("billings.clinic_id = ? AND billings.owner_id = ? AND billings.issued_at >= ? AND billings.deleted_at IS NULL", clinicID, ownerID, since).
		Where("billing_items.deleted_at IS NULL")
	if len(names) > 0 {
		q = q.Where("billing_items.name IN ?", names)
	} else {
		q = q.Where("billing_items.category = ?", string(model.ItemCategoryFood))
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, apperrors.FromGORM(err, "billing_item", fmt.Sprintf("clinic:%d owner:%d", clinicID, ownerID))
	}
	return count > 0, nil
}
