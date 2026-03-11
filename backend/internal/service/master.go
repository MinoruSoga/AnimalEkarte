package service

import (
	"context"

	"github.com/google/uuid"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// MasterItemService マスタアイテムサービスインターフェース
type MasterItemService interface {
	GetAllMasterItems(ctx context.Context) ([]model.MasterItem, error)
	GetMasterItemByID(ctx context.Context, id string) (*model.MasterItem, error)
	GetMasterItemsByCategory(ctx context.Context, category string) ([]model.MasterItem, error)
	CreateMasterItem(ctx context.Context, req *model.CreateMasterItemRequest) (*model.MasterItem, error)
	UpdateMasterItem(ctx context.Context, id string, req *model.UpdateMasterItemRequest) (*model.MasterItem, error)
	DeleteMasterItem(ctx context.Context, id string) error
}

// GetAllMasterItems 全てのマスタアイテムを取得
func (s *Service) GetAllMasterItems(ctx context.Context) ([]model.MasterItem, error) {
	return s.masterItemRepo.GetAllMasterItems(ctx)
}

// GetMasterItemByID IDでマスタアイテムを取得
func (s *Service) GetMasterItemByID(ctx context.Context, id string) (*model.MasterItem, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid master item ID format")
	}

	item, err := s.masterItemRepo.GetMasterItemByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if item == nil {
		return nil, apperrors.WrapNotFound("master item with id %s not found", id)
	}

	return item, nil
}

// GetMasterItemsByCategory カテゴリでマスタアイテムを取得
func (s *Service) GetMasterItemsByCategory(ctx context.Context, category string) ([]model.MasterItem, error) {
	return s.masterItemRepo.GetMasterItemsByCategory(ctx, category)
}

// CreateMasterItem マスタアイテムを作成
func (s *Service) CreateMasterItem(ctx context.Context, req *model.CreateMasterItemRequest) (*model.MasterItem, error) {
	// InventoryIDの変換（オプショナル）
	var inventoryID *uuid.UUID
	if req.InventoryID != nil {
		uid, err := uuid.Parse(*req.InventoryID)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid inventory ID format")
		}
		inventoryID = &uid
	}

	item := &model.MasterItem{
		ID:              uuid.New(),
		Code:            req.Code,
		Name:            req.Name,
		Category:        req.Category,
		Price:           req.Price,
		Status:          "active",
		Description:     req.Description,
		InventoryID:     inventoryID,
		DefaultQuantity: req.DefaultQuantity,
	}

	if err := s.masterItemRepo.CreateMasterItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

// UpdateMasterItem マスタアイテムを更新
func (s *Service) UpdateMasterItem(ctx context.Context, id string, req *model.UpdateMasterItemRequest) (*model.MasterItem, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid master item ID format")
	}

	item, err := s.masterItemRepo.GetMasterItemByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if item == nil {
		return nil, apperrors.WrapNotFound("master item with id %s not found", id)
	}

	// Update fields
	if req.Code != nil {
		item.Code = *req.Code
	}
	if req.Name != nil {
		item.Name = *req.Name
	}
	if req.Category != nil {
		item.Category = *req.Category
	}
	if req.Price != nil {
		item.Price = req.Price
	}
	if req.Status != nil {
		item.Status = *req.Status
	}
	if req.Description != nil {
		item.Description = *req.Description
	}
	if req.DefaultQuantity != nil {
		item.DefaultQuantity = req.DefaultQuantity
	}

	// InventoryIDの更新
	if req.InventoryID != nil {
		uid, err := uuid.Parse(*req.InventoryID)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid inventory ID format")
		}
		item.InventoryID = &uid
	}

	if err := s.masterItemRepo.UpdateMasterItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

// DeleteMasterItem マスタアイテムを削除
func (s *Service) DeleteMasterItem(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return apperrors.WrapInvalidInput("invalid master item ID format")
	}

	return s.masterItemRepo.DeleteMasterItem(ctx, uid.String())
}
