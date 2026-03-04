package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// MasterItemRepository マスタアイテムリポジトリインターフェース
type MasterItemRepository interface {
	GetAllMasterItems(ctx context.Context) ([]model.MasterItem, error)
	GetMasterItemByID(ctx context.Context, id string) (*model.MasterItem, error)
	GetMasterItemsByCategory(ctx context.Context, category string) ([]model.MasterItem, error)
	CreateMasterItem(ctx context.Context, item *model.MasterItem) error
	UpdateMasterItem(ctx context.Context, item *model.MasterItem) error
	DeleteMasterItem(ctx context.Context, id string) error
}

// masterItemRepository マスタアイテムリポジトリ実装
type masterItemRepository struct {
	db *gorm.DB
}

// NewMasterItemRepository 新しいマスタアイテムリポジトリを作成
func NewMasterItemRepository(db *gorm.DB) MasterItemRepository {
	return &masterItemRepository{db: db}
}

// GetAllMasterItems 全てのマスタアイテムを取得
func (r *masterItemRepository) GetAllMasterItems(ctx context.Context) ([]model.MasterItem, error) {
	var items []model.MasterItem
	result := r.db.WithContext(ctx).
		Preload("InventoryItem").
		Order("category, code").
		Find(&items)

	if result.Error != nil {
		return nil, result.Error
	}

	return items, nil
}

// GetMasterItemByID IDでマスタアイテムを取得
func (r *masterItemRepository) GetMasterItemByID(ctx context.Context, id string) (*model.MasterItem, error) {
	var item model.MasterItem
	result := r.db.WithContext(ctx).
		Preload("InventoryItem").
		First(&item, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("master item with id %s not found", id)
		}
		return nil, apperrors.WrapInternal(result.Error, "failed to get master item")
	}

	return &item, nil
}

// GetMasterItemsByCategory カテゴリでマスタアイテムを取得
func (r *masterItemRepository) GetMasterItemsByCategory(ctx context.Context, category string) ([]model.MasterItem, error) {
	var items []model.MasterItem
	result := r.db.WithContext(ctx).
		Preload("InventoryItem").
		Where("category = ?", category).
		Order("code").
		Find(&items)

	if result.Error != nil {
		return nil, result.Error
	}

	return items, nil
}

// CreateMasterItem マスタアイテムを作成
func (r *masterItemRepository) CreateMasterItem(ctx context.Context, item *model.MasterItem) error {
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

// UpdateMasterItem マスタアイテムを更新
func (r *masterItemRepository) UpdateMasterItem(ctx context.Context, item *model.MasterItem) error {
	result := r.db.WithContext(ctx).Save(item)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteMasterItem マスタアイテムを削除
func (r *masterItemRepository) DeleteMasterItem(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.MasterItem{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
