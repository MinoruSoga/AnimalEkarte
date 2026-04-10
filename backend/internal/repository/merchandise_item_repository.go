// Package repository provides data access implementations for MerchandiseItem entity.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- MerchandiseItem ----

// MerchandiseItemRepository は物販品マスタのデータアクセス
type MerchandiseItemRepository interface {
	FindAll(ctx context.Context, clinicID uint64, page, limit int, category string) ([]model.MerchandiseItem, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error)
	CountUsageByMerchandiseItemID(ctx context.Context, merchandiseItemID uint64) (int64, error)
	Create(ctx context.Context, item *model.MerchandiseItem) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type merchandiseItemRepository struct{ db *gorm.DB }

// NewMerchandiseItemRepository は物販品リポジトリを初期化する
func NewMerchandiseItemRepository(db *gorm.DB) MerchandiseItemRepository {
	return &merchandiseItemRepository{db: db}
}

func (r *merchandiseItemRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int, category string) ([]model.MerchandiseItem, int64, error) {
	items := make([]model.MerchandiseItem, 0)
	var total int64

	buildBase := func() *gorm.DB {
		q := r.db.WithContext(ctx).Model(&model.MerchandiseItem{}).Where("clinic_id = ?", clinicID)
		if category != "" {
			q = q.Where("category = ?", category)
		}
		return q
	}

	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "merchandise_item", "")
	}
	if err := buildBase().
		Offset((page - 1) * limit).Limit(limit).
		Order("sort_order ASC, name ASC").
		Find(&items).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "merchandise_item", "")
	}
	return items, total, nil
}

func (r *merchandiseItemRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error) {
	var item model.MerchandiseItem
	err := r.db.WithContext(ctx).
		First(&item, "id = ? AND clinic_id = ?", id, clinicID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "merchandise_item", fmt.Sprintf("%d", id))
	}
	return &item, nil
}

func (r *merchandiseItemRepository) Create(ctx context.Context, item *model.MerchandiseItem) error {
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.FromGORM(err, "merchandise_item", "")
	}
	return nil
}

func (r *merchandiseItemRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.MerchandiseItem{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "merchandise_item", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		var count int64
		if err := r.db.WithContext(ctx).Model(&model.MerchandiseItem{}).
			Where("id = ? AND clinic_id = ?", id, clinicID).
			Count(&count).Error; err != nil {
			return apperrors.FromGORM(err, "merchandise_item", fmt.Sprintf("%d", id))
		}
		if count == 0 {
			return apperrors.WrapNotFound("merchandise_item", fmt.Sprintf("%d", id))
		}
	}
	return nil
}

func (r *merchandiseItemRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(r.db, ctx, &model.MerchandiseItem{}, "merchandise_item", clinicID, ids)
}

// CountUsageByMerchandiseItemID は物販品目を参照している billing_items と estimate_items の件数の合計を返す（BUG-109）
// Migration 002 でFK カラムが追加された後、このメソッドで依存チェックを実行する
func (r *merchandiseItemRepository) CountUsageByMerchandiseItemID(ctx context.Context, merchandiseItemID uint64) (int64, error) {
	var billingCount int64
	if err := r.db.WithContext(ctx).
		Model(&model.BillingItem{}).
		Where("merchandise_item_id = ? AND deleted_at IS NULL", merchandiseItemID).
		Count(&billingCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "billing_item", "")
	}

	var estimateCount int64
	// BUG-154: estimate_items に deleted_at カラムがないため条件を削除
	if err := r.db.WithContext(ctx).
		Model(&model.EstimateItem{}).
		Where("merchandise_item_id = ?", merchandiseItemID).
		Count(&estimateCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "estimate_item", "")
	}

	return billingCount + estimateCount, nil
}

func (r *merchandiseItemRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.MerchandiseItem{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "merchandise_item", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("merchandise_item", fmt.Sprintf("%d", id))
	}
	return nil
}
