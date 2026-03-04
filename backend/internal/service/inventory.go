package service

import (
	"context"

	"github.com/google/uuid"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// InventoryItemService 在庫アイテムサービスインターフェース
type InventoryItemService interface {
	GetAllInventoryItems(ctx context.Context) ([]model.InventoryItem, error)
	GetInventoryItemByID(ctx context.Context, id string) (*model.InventoryItem, error)
	GetInventoryItemsByCategory(ctx context.Context, category string) ([]model.InventoryItem, error)
	GetInventoryItemsByStatus(ctx context.Context, status string) ([]model.InventoryItem, error)
	CreateInventoryItem(ctx context.Context, req *model.CreateInventoryItemRequest) (*model.InventoryItem, error)
	UpdateInventoryItem(ctx context.Context, id string, req *model.UpdateInventoryItemRequest) (*model.InventoryItem, error)
	DeleteInventoryItem(ctx context.Context, id string) error
}

// GetAllInventoryItems 全ての在庫アイテムを取得
func (s *Service) GetAllInventoryItems(ctx context.Context) ([]model.InventoryItem, error) {
	return s.inventoryItemRepo.GetAllInventoryItems(ctx)
}

// GetInventoryItemByID IDで在庫アイテムを取得
func (s *Service) GetInventoryItemByID(ctx context.Context, id string) (*model.InventoryItem, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid inventory item ID format")
	}

	item, err := s.inventoryItemRepo.GetInventoryItemByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if item == nil {
		return nil, apperrors.WrapNotFound("inventory item with id %s not found", id)
	}

	return item, nil
}

// GetInventoryItemsByCategory カテゴリで在庫アイテムを取得
func (s *Service) GetInventoryItemsByCategory(ctx context.Context, category string) ([]model.InventoryItem, error) {
	return s.inventoryItemRepo.GetInventoryItemsByCategory(ctx, category)
}

// GetInventoryItemsByStatus ステータスで在庫アイテムを取得
func (s *Service) GetInventoryItemsByStatus(ctx context.Context, status string) ([]model.InventoryItem, error) {
	return s.inventoryItemRepo.GetInventoryItemsByStatus(ctx, status)
}

// CreateInventoryItem 在庫アイテムを作成
func (s *Service) CreateInventoryItem(ctx context.Context, req *model.CreateInventoryItemRequest) (*model.InventoryItem, error) {
	// Determine initial status based on quantity
	status := "sufficient"
	if req.Quantity == 0 {
		status = "out_of_stock"
	} else if req.MinStockLevel > 0 && req.Quantity <= req.MinStockLevel {
		status = "low"
	}

	item := &model.InventoryItem{
		ID:            uuid.New(),
		Name:          req.Name,
		Category:      req.Category,
		Quantity:      req.Quantity,
		Unit:          req.Unit,
		MinStockLevel: req.MinStockLevel,
		Location:      req.Location,
		ExpiryDate:    req.ExpiryDate,
		Supplier:      req.Supplier,
		LastRestocked: req.LastRestocked,
		Status:        status,
	}

	if err := s.inventoryItemRepo.CreateInventoryItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

// UpdateInventoryItem 在庫アイテムを更新
func (s *Service) UpdateInventoryItem(ctx context.Context, id string, req *model.UpdateInventoryItemRequest) (*model.InventoryItem, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid inventory item ID format")
	}

	item, err := s.inventoryItemRepo.GetInventoryItemByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if item == nil {
		return nil, apperrors.WrapNotFound("inventory item with id %s not found", id)
	}

	// Update fields
	if req.Name != "" {
		item.Name = req.Name
	}
	if req.Category != "" {
		item.Category = req.Category
	}
	if req.Quantity != 0 {
		item.Quantity = req.Quantity
	}
	if req.Unit != "" {
		item.Unit = req.Unit
	}
	if req.MinStockLevel != 0 {
		item.MinStockLevel = req.MinStockLevel
	}
	if req.Location != "" {
		item.Location = req.Location
	}
	if req.ExpiryDate != nil {
		item.ExpiryDate = req.ExpiryDate
	}
	if req.Supplier != "" {
		item.Supplier = req.Supplier
	}
	if req.LastRestocked != nil {
		item.LastRestocked = req.LastRestocked
	}
	if req.Status != "" {
		item.Status = req.Status
	}

	if err := s.inventoryItemRepo.UpdateInventoryItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

// DeleteInventoryItem 在庫アイテムを削除
func (s *Service) DeleteInventoryItem(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return apperrors.WrapInvalidInput("invalid inventory item ID format")
	}

	return s.inventoryItemRepo.DeleteInventoryItem(ctx, uid.String())
}