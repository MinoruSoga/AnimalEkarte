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
		return nil, 0, apperrors.Wrap(err, "count merchandise items")
	}
	if err := buildBase().
		Offset((page - 1) * limit).Limit(limit).
		Order("sort_order ASC, name ASC").
		Find(&items).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find merchandise items")
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
			return apperrors.WrapAlreadyExists("merchandise_item", item.Name)
		}
		return apperrors.Wrap(err, "create merchandise item")
	}
	return nil
}

func (r *merchandiseItemRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.MerchandiseItem{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update merchandise item")
	}
	if result.RowsAffected == 0 {
		var count int64
		r.db.WithContext(ctx).Model(&model.MerchandiseItem{}).
			Where("id = ? AND clinic_id = ?", id, clinicID).
			Count(&count)
		if count == 0 {
			return apperrors.WrapNotFound("merchandise_item", fmt.Sprintf("%d", id))
		}
	}
	return nil
}

func (r *merchandiseItemRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.MerchandiseItem{}).
				Where("id = ? AND clinic_id = ?", id, clinicID).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.Wrap(result.Error, "reorder merchandise item")
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("merchandise_item id %d not found in this clinic", id))
			}
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "reorder merchandise item")
	}
	return nil
}

func (r *merchandiseItemRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.MerchandiseItem{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete merchandise item")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("merchandise_item", fmt.Sprintf("%d", id))
	}
	return nil
}
