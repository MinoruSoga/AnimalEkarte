package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type InventoryService interface {
	List(ctx context.Context, category *string, page, limit int) ([]model.InventoryItem, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.InventoryItem, error)
	Create(ctx context.Context, item *model.InventoryItem) error
	Update(ctx context.Context, item *model.InventoryItem) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type inventoryService struct {
	repo repository.InventoryRepository
}

func NewInventoryService(repo repository.InventoryRepository) InventoryService {
	return &inventoryService{repo: repo}
}

func (s *inventoryService) List(ctx context.Context, category *string, page, limit int) ([]model.InventoryItem, int64, error) {
	return s.repo.FindAll(ctx, category, page, limit)
}

func (s *inventoryService) GetByID(ctx context.Context, id uuid.UUID) (*model.InventoryItem, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *inventoryService) Create(ctx context.Context, item *model.InventoryItem) error {
	return s.repo.Create(ctx, item)
}

func (s *inventoryService) Update(ctx context.Context, item *model.InventoryItem) error {
	return s.repo.Update(ctx, item)
}

func (s *inventoryService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
