package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type InventoryRepository interface {
	FindAll(ctx context.Context, category *string, page, limit int) ([]model.InventoryItem, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.InventoryItem, error)
	Create(ctx context.Context, item *model.InventoryItem) error
	Update(ctx context.Context, item *model.InventoryItem) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type inventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) InventoryRepository {
	return &inventoryRepository{db: db}
}

func (r *inventoryRepository) FindAll(ctx context.Context, category *string, page, limit int) ([]model.InventoryItem, int64, error) {
	var items []model.InventoryItem
	var total int64

	q := r.db.WithContext(ctx).Model(&model.InventoryItem{})
	if category != nil {
		q = q.Where("category = ?", *category)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count inventory items")
	}
	if err := q.Offset((page - 1) * limit).Limit(limit).Order("name ASC").Find(&items).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find inventory items")
	}
	return items, total, nil
}

func (r *inventoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.InventoryItem, error) {
	var item model.InventoryItem
	if err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("inventory_item", id.String())
		}
		return nil, apperrors.Wrap(err, "find inventory item by id")
	}
	return &item, nil
}

func (r *inventoryRepository) Create(ctx context.Context, item *model.InventoryItem) error {
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("inventory_item", item.Name)
		}
		return apperrors.Wrap(err, "create inventory item")
	}
	return nil
}

func (r *inventoryRepository) Update(ctx context.Context, item *model.InventoryItem) error {
	if err := r.db.WithContext(ctx).Save(item).Error; err != nil {
		return apperrors.Wrap(err, "update inventory item")
	}
	return nil
}

func (r *inventoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.InventoryItem{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete inventory item")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("inventory_item", id.String())
	}
	return nil
}
