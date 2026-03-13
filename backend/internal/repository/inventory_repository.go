package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type InventoryRepository interface {
	FindAll(ctx context.Context, clinicID uint64, category, status *string, page, limit int) ([]model.InventoryItem, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.InventoryItem, error)
	Create(ctx context.Context, clinicID uint64, item *model.InventoryItem) error
	Update(ctx context.Context, clinicID uint64, item *model.InventoryItem) error
	Delete(ctx context.Context, clinicID, id uint64) error
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
		return nil, 0, apperrors.Wrap(err, "count inventory items")
	}
	if err := q.Offset((page - 1) * limit).Limit(limit).Order("name ASC").Find(&items).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find inventory items")
	}
	return items, total, nil
}

func (r *inventoryRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.InventoryItem, error) {
	var item model.InventoryItem
	if err := r.db.WithContext(ctx).First(&item, "id = ? AND clinic_id = ?", id, clinicID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("inventory_item", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find inventory item by id")
	}
	return &item, nil
}

func (r *inventoryRepository) Create(ctx context.Context, clinicID uint64, item *model.InventoryItem) error {
	item.ClinicID = clinicID
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("inventory_item", item.Name)
		}
		return apperrors.Wrap(err, "create inventory item")
	}
	return nil
}

func (r *inventoryRepository) Update(ctx context.Context, clinicID uint64, item *model.InventoryItem) error {
	item.ClinicID = clinicID
	result := r.db.WithContext(ctx).
		Model(&model.InventoryItem{}).
		Where("id = ? AND clinic_id = ?", item.ID, clinicID).
		Updates(item)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update inventory item")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("inventory_item", fmt.Sprintf("%d", item.ID))
	}
	return nil
}

func (r *inventoryRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.InventoryItem{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete inventory item")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("inventory_item", fmt.Sprintf("%d", id))
	}
	return nil
}
