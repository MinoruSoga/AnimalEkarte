package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type InventoryRepository interface {
	FindAll(ctx context.Context, clinicID uint64, category, status *string, page, limit int) ([]model.InventoryItem, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.InventoryItem, error)
	Create(ctx context.Context, clinicID uint64, item *model.InventoryItem) error
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.InventoryItem, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	DecreaseStock(ctx context.Context, id uint64, quantity float64) error
	CountUsageByInventoryID(ctx context.Context, inventoryID uint64) (int64, error)
}

type inventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) InventoryRepository {
	return &inventoryRepository{db: db}
}

func (r *inventoryRepository) FindAll(ctx context.Context, clinicID uint64, category, status *string, page, limit int) ([]model.InventoryItem, int64, error) {
	items := make([]model.InventoryItem, 0)
	var total int64

	q := r.db.WithContext(ctx).Model(&model.InventoryItem{}).Where("clinic_id = ?", clinicID)
	if category != nil {
		q = q.Where("category = ?", *category)
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "inventory_item", "")
	}
	if err := q.Offset((page - 1) * limit).Limit(limit).Order("name ASC").Find(&items).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "inventory_item", "")
	}
	return items, total, nil
}

func (r *inventoryRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.InventoryItem, error) {
	var item model.InventoryItem
	err := r.db.WithContext(ctx).First(&item, "id = ? AND clinic_id = ?", id, clinicID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "inventory_item", fmt.Sprintf("%d", id))
	}
	return &item, nil
}

func (r *inventoryRepository) Create(ctx context.Context, clinicID uint64, item *model.InventoryItem) error {
	item.ClinicID = clinicID
	err := r.db.WithContext(ctx).Create(item).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("inventory_item", item.Name)
		}
		return apperrors.FromGORM(err, "inventory_item", "")
	}
	return nil
}

func (r *inventoryRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.InventoryItem, error) {
	result := r.db.WithContext(ctx).
		Model(&model.InventoryItem{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "inventory_item", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("inventory_item", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *inventoryRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.InventoryItem{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "inventory_item", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("inventory_item", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *inventoryRepository) DecreaseStock(ctx context.Context, id uint64, quantity float64) error {
	result := r.db.WithContext(ctx).
		Model(&model.InventoryItem{}).
		Where("id = ?", id).
		UpdateColumn("quantity", gorm.Expr("quantity - ?", int(quantity)))
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "inventory_item", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("inventory_item", fmt.Sprintf("%d", id))
	}
	return nil
}

// CountUsageByInventoryID は在庫アイテムを参照している治療明細・ワクチン・薬剤の件数を返す（BUG-195）
func (r *inventoryRepository) CountUsageByInventoryID(ctx context.Context, inventoryID uint64) (int64, error) {
	var count int64
	// treatments, vaccines, medicines のいずれかから参照されていればカウント
	err := r.db.WithContext(ctx).
		Raw(`SELECT (
			SELECT COUNT(*) FROM treatments WHERE inventory_id = ? AND deleted_at IS NULL
		) + (
			SELECT COUNT(*) FROM vaccines WHERE inventory_id = ?
		) + (
			SELECT COUNT(*) FROM medicines WHERE inventory_id = ?
		) AS total`, inventoryID, inventoryID, inventoryID).
		Scan(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "inventory_item", fmt.Sprintf("%d", inventoryID))
	}
	return count, nil
}
