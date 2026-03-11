package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// InventoryItemRepository 在庫アイテムリポジトリインターフェース
type InventoryItemRepository interface {
	GetAllInventoryItems(ctx context.Context) ([]model.InventoryItem, error)
	GetInventoryItemByID(ctx context.Context, id string) (*model.InventoryItem, error)
	GetInventoryItemsByCategory(ctx context.Context, category string) ([]model.InventoryItem, error)
	GetInventoryItemsByStatus(ctx context.Context, status string) ([]model.InventoryItem, error)
	CreateInventoryItem(ctx context.Context, item *model.InventoryItem) error
	UpdateInventoryItem(ctx context.Context, item *model.InventoryItem) error
	DeleteInventoryItem(ctx context.Context, id string) error
}

// inventoryItemRepository 在庫アイテムリポジトリ実装
type inventoryItemRepository struct {
	db *gorm.DB
}

// NewInventoryItemRepository 新しい在庫アイテムリポジトリを作成
func NewInventoryItemRepository(db *gorm.DB) InventoryItemRepository {
	return &inventoryItemRepository{db: db}
}

// GetAllInventoryItems 全ての在庫アイテムを取得
func (r *inventoryItemRepository) GetAllInventoryItems(ctx context.Context) ([]model.InventoryItem, error) {
	var items []model.InventoryItem
	result := r.db.WithContext(ctx).
		Order("created_at DESC").
		Find(&items)

	if result.Error != nil {
		return nil, result.Error
	}

	return items, nil
}

// GetInventoryItemByID IDで在庫アイテムを取得
func (r *inventoryItemRepository) GetInventoryItemByID(ctx context.Context, id string) (*model.InventoryItem, error) {
	var item model.InventoryItem
	result := r.db.WithContext(ctx).
		First(&item, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("inventory item", id)
		}
		return nil, apperrors.WrapInternal(result.Error, "failed to get inventory item")
	}

	return &item, nil
}

// GetInventoryItemsByCategory カテゴリで在庫アイテムを取得
func (r *inventoryItemRepository) GetInventoryItemsByCategory(ctx context.Context, category string) ([]model.InventoryItem, error) {
	var items []model.InventoryItem
	result := r.db.WithContext(ctx).
		Where("category = ?", category).
		Order("created_at DESC").
		Find(&items)

	if result.Error != nil {
		return nil, result.Error
	}

	return items, nil
}

// GetInventoryItemsByStatus ステータスで在庫アイテムを取得
func (r *inventoryItemRepository) GetInventoryItemsByStatus(ctx context.Context, status string) ([]model.InventoryItem, error) {
	var items []model.InventoryItem
	result := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Find(&items)

	if result.Error != nil {
		return nil, result.Error
	}

	return items, nil
}

// CreateInventoryItem 在庫アイテムを作成
func (r *inventoryItemRepository) CreateInventoryItem(ctx context.Context, item *model.InventoryItem) error {
	// Generate UUID if not set
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	result := r.db.WithContext(ctx).Create(item)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdateInventoryItem 在庫アイテムを更新
func (r *inventoryItemRepository) UpdateInventoryItem(ctx context.Context, item *model.InventoryItem) error {
	result := r.db.WithContext(ctx).Save(item)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteInventoryItem 在庫アイテムを削除
func (r *inventoryItemRepository) DeleteInventoryItem(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.InventoryItem{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
