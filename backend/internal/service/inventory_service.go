package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// CreateInventoryInput は在庫アイテム作成の入力DTO
type CreateInventoryInput struct {
	Name          string
	Category      string
	Quantity      int
	Unit          string
	MinStockLevel int
	Location      string
	ExpiryDate    *time.Time
	Supplier      string
	LastRestocked *time.Time
	Status        string
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

func buildInventoryUpdate(input *UpdateInventoryInput) map[string]any {
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

type InventoryService interface {
	List(ctx context.Context, clinicID uint64, category, status *string, page, limit int) ([]model.InventoryItem, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.InventoryItem, error)
	Create(ctx context.Context, clinicID uint64, input *CreateInventoryInput) (*model.InventoryItem, error)
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
	items, total, err := s.repo.FindAll(ctx, clinicID, category, status, page, limit)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list inventory items")
	}
	return items, total, nil
}

func (s *inventoryService) GetByID(ctx context.Context, clinicID, id uint64) (*model.InventoryItem, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get inventory item")
	}
	return result, nil
}

func (s *inventoryService) Create(ctx context.Context, clinicID uint64, input *CreateInventoryInput) (*model.InventoryItem, error) {
	item := &model.InventoryItem{
		ClinicID:      clinicID,
		Name:          input.Name,
		Category:      model.InventoryCategory(input.Category),
		Quantity:      input.Quantity,
		Unit:          input.Unit,
		MinStockLevel: input.MinStockLevel,
		Location:      input.Location,
		ExpiryDate:    input.ExpiryDate,
		Supplier:      input.Supplier,
		LastRestocked: input.LastRestocked,
		Status:        model.InventoryStatusSufficient,
	}
	if input.Status != "" {
		item.Status = model.InventoryStatus(input.Status)
	}

	if err := s.repo.Create(ctx, clinicID, item); err != nil {
		return nil, apperrors.Wrap(err, "failed to create inventory item")
	}
	slog.InfoContext(ctx, "inventory item created", slog.Uint64("inventory_id", item.ID), slog.Uint64("clinic_id", clinicID))
	return item, nil
}

func (s *inventoryService) Update(ctx context.Context, clinicID, id uint64, input *UpdateInventoryInput) (*model.InventoryItem, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	fields := buildInventoryUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	item, err := s.repo.Update(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update inventory item")
	}
	slog.InfoContext(ctx, "inventory item updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("inventory_id", id))
	return item, nil
}

func (s *inventoryService) Delete(ctx context.Context, clinicID, id uint64) error {
	count, err := s.repo.CountUsageByInventoryID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check inventory item dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この項目は使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete inventory item")
	}
	slog.InfoContext(ctx, "inventory item deleted", slog.Uint64("inventory_id", id), slog.Uint64("clinic_id", clinicID))
	return nil
}
