package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type InventoryService interface {
	List(ctx context.Context, clinicID uint64, category, status *string, page, limit int) ([]model.InventoryItem, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.InventoryItem, error)
	Create(ctx context.Context, clinicID uint64, item *model.InventoryItem) error
	Update(ctx context.Context, clinicID uint64, item *model.InventoryItem) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type inventoryService struct {
	repo repository.InventoryRepository
}

func NewInventoryService(repo repository.InventoryRepository) InventoryService {
	return &inventoryService{repo: repo}
}

func (s *inventoryService) List(ctx context.Context, clinicID uint64, category, status *string, page, limit int) ([]model.InventoryItem, int64, error) {
	return s.repo.FindAll(ctx, clinicID, category, status, page, limit)
}

func (s *inventoryService) GetByID(ctx context.Context, clinicID, id uint64) (*model.InventoryItem, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *inventoryService) Create(ctx context.Context, clinicID uint64, item *model.InventoryItem) error {
	return s.repo.Create(ctx, clinicID, item)
}

func (s *inventoryService) Update(ctx context.Context, clinicID uint64, item *model.InventoryItem) error {
	return s.repo.Update(ctx, clinicID, item)
}

func (s *inventoryService) Delete(ctx context.Context, clinicID, id uint64) error {
	return s.repo.Delete(ctx, clinicID, id)
}
