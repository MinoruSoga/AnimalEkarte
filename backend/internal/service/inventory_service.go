package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type InventoryService interface {
	List(ctx context.Context, clinicID uint64, category, status *string, page, limit int) ([]model.InventoryItem, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.InventoryItem, error)
	Create(ctx context.Context, clinicID uint64, item *model.InventoryItem) error
	Update(ctx context.Context, clinicID, id uint64, input *UpdateInventoryInput) (*model.InventoryItem, error)
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

func (s *inventoryService) Update(ctx context.Context, clinicID, id uint64, input *UpdateInventoryInput) (*model.InventoryItem, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	fields := buildInventoryUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	item, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update inventory item")
	}
	slog.InfoContext(ctx, "inventory item updated", slog.Uint64("inventory_id", id))
	return item, nil
}

func (s *inventoryService) Delete(ctx context.Context, clinicID, id uint64) error {
	count, err := s.repo.CountUsageByInventoryID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check inventory item dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この項目は使用中のため削除できません")
	}
	return s.repo.Delete(ctx, clinicID, id)
}

// UpdateInventoryInput は在庫アイテム更新のサービス入力 DTO
type UpdateInventoryInput struct {
	Name          *string
	Category      *model.InventoryCategory
	Quantity      *int
	Unit          *string
	MinStockLevel *int
	Location      *string
	ExpiryDate    *time.Time
	Supplier      *string
	LastRestocked *time.Time
	Status        *model.InventoryStatus
}

func buildInventoryUpdateFields(input *UpdateInventoryInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.Category != nil {
		fields["category"] = *input.Category
	}
	if input.Quantity != nil {
		fields["quantity"] = *input.Quantity
	}
	if input.Unit != nil {
		fields["unit"] = *input.Unit
	}
	if input.MinStockLevel != nil {
		fields["min_stock_level"] = *input.MinStockLevel
	}
	if input.Location != nil {
		fields["location"] = *input.Location
	}
	if input.ExpiryDate != nil {
		fields["expiry_date"] = *input.ExpiryDate
	}
	if input.Supplier != nil {
		fields["supplier"] = *input.Supplier
	}
	if input.LastRestocked != nil {
		fields["last_restocked"] = *input.LastRestocked
	}
	if input.Status != nil {
		fields["status"] = *input.Status
	}
	return fields
}
