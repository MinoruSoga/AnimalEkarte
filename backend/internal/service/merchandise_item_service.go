// Package service provides business logic implementations for MerchandiseItem entity.
package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// --- DB column constants ---

const (
	colMerchandiseItemName      = "name"
	colMerchandiseItemCategory  = "category"
	colMerchandiseItemUnitPrice = "unit_price"
	colMerchandiseItemTaxType   = "tax_type"
	colMerchandiseItemTaxRate   = "tax_rate"
	colMerchandiseItemIsActive  = "is_active"
	colMerchandiseItemSortOrder = "sort_order"
)

// --- Input DTOs ---

// CreateMerchandiseItemInput は物販品作成の入力DTO
type CreateMerchandiseItemInput struct {
	Name      string
	Category  string
	UnitPrice int64
	TaxType   string  // "" → "excluded" (default)
	TaxRate   float64 // 0 → 0.10 (default)
	IsActive  bool
	SortOrder int
}

// UpdateMerchandiseItemInput は物販品更新の入力DTO（nil = 未指定）
type UpdateMerchandiseItemInput struct {
	Name      *string
	Category  *string
	UnitPrice *int64
	TaxType   *string
	TaxRate   *float64
	IsActive  *bool
	SortOrder *int
}

// buildMerchandiseItemUpdateFields は UPDATE 用 map を構築する。
// GORM のゼロ値スキップ問題（bool false が無視される等）を回避するために使用する。
func buildMerchandiseItemUpdateFields(input *UpdateMerchandiseItemInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colMerchandiseItemName] = *input.Name
	}
	if input.Category != nil {
		fields[colMerchandiseItemCategory] = *input.Category
	}
	if input.UnitPrice != nil {
		fields[colMerchandiseItemUnitPrice] = *input.UnitPrice
	}
	if input.TaxType != nil {
		fields[colMerchandiseItemTaxType] = *input.TaxType
	}
	if input.TaxRate != nil {
		fields[colMerchandiseItemTaxRate] = *input.TaxRate
	}
	if input.IsActive != nil {
		fields[colMerchandiseItemIsActive] = *input.IsActive
	}
	if input.SortOrder != nil {
		fields[colMerchandiseItemSortOrder] = *input.SortOrder
	}
	return fields
}

// ---- MerchandiseItemService ----

// MerchandiseItemService は物販品マスタのビジネスロジック
type MerchandiseItemService interface {
	List(ctx context.Context, clinicID uint64, page, limit int, category string) ([]model.MerchandiseItem, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error)
	Create(ctx context.Context, clinicID uint64, input *CreateMerchandiseItemInput) (*model.MerchandiseItem, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateMerchandiseItemInput) (*model.MerchandiseItem, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type merchandiseItemService struct {
	repo repository.MerchandiseItemRepository
}

// NewMerchandiseItemService は物販品サービスを初期化する
func NewMerchandiseItemService(repo repository.MerchandiseItemRepository) MerchandiseItemService {
	return &merchandiseItemService{repo: repo}
}

func (s *merchandiseItemService) List(ctx context.Context, clinicID uint64, page, limit int, category string) ([]model.MerchandiseItem, int64, error) {
	return s.repo.FindAll(ctx, clinicID, page, limit, category)
}

func (s *merchandiseItemService) GetByID(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *merchandiseItemService) Create(ctx context.Context, clinicID uint64, input *CreateMerchandiseItemInput) (*model.MerchandiseItem, error) {
	if input.Name == "" {
		return nil, apperrors.WrapInvalidInput("name is required")
	}

	taxType := model.TaxTypeExcluded
	if input.TaxType != "" {
		taxType = model.TaxType(input.TaxType)
	}
	taxRate := 0.10
	if input.TaxRate != 0 {
		taxRate = input.TaxRate
	}
	item := &model.MerchandiseItem{
		ClinicID:  clinicID,
		Name:      input.Name,
		Category:  model.ItemCategory(input.Category),
		UnitPrice: input.UnitPrice,
		TaxType:   taxType,
		TaxRate:   taxRate,
		IsActive:  input.IsActive,
		SortOrder: input.SortOrder,
	}

	if err := s.repo.Create(ctx, item); err != nil {
		return nil, apperrors.Wrap(err, "failed to create merchandise item")
	}

	slog.InfoContext(ctx, "merchandise item created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("item_id", item.ID),
		slog.String("name", item.Name),
	)
	return item, nil
}

func (s *merchandiseItemService) Update(ctx context.Context, clinicID, id uint64, input *UpdateMerchandiseItemInput) (*model.MerchandiseItem, error) {
	fields := buildMerchandiseItemUpdateFields(input)
	if len(fields) == 0 {
		return s.repo.FindByID(ctx, clinicID, id)
	}

	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update merchandise item")
	}

	slog.InfoContext(ctx, "merchandise item updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("merchandise_item_id", id),
	)
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *merchandiseItemService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	return s.repo.Reorder(ctx, clinicID, ids)
}

func (s *merchandiseItemService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to get merchandise item")
	}

	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete merchandise item")
	}
	slog.InfoContext(ctx, "merchandise item deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("merchandise_item_id", id),
	)
	return nil
}
