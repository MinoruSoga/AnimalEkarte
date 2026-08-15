package inventory

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

const (
	colMerchandiseItemName      = "name"
	colMerchandiseItemCategory  = "category"
	colMerchandiseItemUnitPrice = "unit_price"
	colMerchandiseItemTaxType   = "tax_type"
	colMerchandiseItemTaxRate   = "tax_rate"
	colMerchandiseItemIsActive  = "is_active"
	colMerchandiseItemSortOrder = "sort_order"
)

func buildMerchandiseItemUpdate(input *UpdateMerchandiseItemInput) map[string]any {
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

// --- Input DTOs ---

// CreateMerchandiseItemInput は物販品作成の入力DTO
type CreateMerchandiseItemInput struct {
	Name      string
	Category  string
	UnitPrice int64
	TaxType   string   // "" → "excluded" (default)
	TaxRate   *float64 // nil → 0.10 (default)
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

// --- DB column constants ---

// buildMerchandiseItemUpdate は UPDATE 用 map を構築する。
// GORM のゼロ値スキップ問題（bool false が無視される等）を回避するために使用する。

// ---- MerchandiseItemService ----

// Transactor is the consumer-side ambient transaction port for atomic merchandise delete.
type Transactor interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// MerchandiseItemService は物販品マスタのビジネスロジック
type MerchandiseItemService interface {
	List(ctx context.Context, clinicID uint64, category string) ([]model.MerchandiseItem, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error)
	Create(ctx context.Context, clinicID uint64, input *CreateMerchandiseItemInput) (*model.MerchandiseItem, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateMerchandiseItemInput) (*model.MerchandiseItem, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

// merchandiseItemStore is the use case's consumer-side persistence view.
type merchandiseItemStore interface {
	FindAll(ctx context.Context, clinicID uint64, category string) ([]model.MerchandiseItem, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error)
	CountUsageByMerchandiseItemID(ctx context.Context, clinicID, merchandiseItemID uint64) (int64, error)
	Create(ctx context.Context, item *model.MerchandiseItem) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MerchandiseItem, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type merchandiseItemService struct {
	repo       merchandiseItemStore
	transactor Transactor
}

// NewMerchandiseItemService は物販品サービスを初期化する。
// transactor は Delete の soft-delete + usage 再確認を同一 transaction に載せるために必須
// （BE-ACT-MERCHANDISE-ATOMIC-DELETE）。
func NewMerchandiseItemService(repo merchandiseItemStore, transactor Transactor) MerchandiseItemService {
	return &merchandiseItemService{repo: repo, transactor: transactor}
}

func (s *merchandiseItemService) List(ctx context.Context, clinicID uint64, category string) ([]model.MerchandiseItem, error) {
	result, err := s.repo.FindAll(ctx, clinicID, category)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list merchandise items", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list merchandise items")
	}
	return result, nil
}

func (s *merchandiseItemService) GetByID(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get merchandise item", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get merchandise item")
	}
	return result, nil
}

func (s *merchandiseItemService) Create(ctx context.Context, clinicID uint64, input *CreateMerchandiseItemInput) (*model.MerchandiseItem, error) {
	if err := sharedkernel.ValidateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	if input.UnitPrice < 0 {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgPriceZeroOrMore)
	}

	taxType := model.TaxTypeExcluded
	if input.TaxType != "" {
		taxType = model.TaxType(input.TaxType)
	}
	taxRate := sharedkernel.DefaultTaxRate
	if input.TaxRate != nil {
		taxRate = *input.TaxRate
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
		slog.ErrorContext(ctx, "failed to create merchandise item", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to create merchandise item")
	}

	slog.InfoContext(ctx, "merchandise item created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("merchandise_item_id", item.ID),
	)
	return item, nil
}

func (s *merchandiseItemService) Update(ctx context.Context, clinicID, id uint64, input *UpdateMerchandiseItemInput) (*model.MerchandiseItem, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgInputNotNil)
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to get merchandise item", "error", err)
		return nil, apperrors.Wrap(err, "failed to get merchandise item")
	}
	if err := sharedkernel.ValidateOptionalName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	if input.UnitPrice != nil && *input.UnitPrice < 0 {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgPriceZeroOrMore)
	}
	fields := buildMerchandiseItemUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgAtLeastOneField)
	}

	result, err := s.repo.Update(ctx, clinicID, id, fields)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update merchandise item", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to update merchandise item")
	}
	slog.InfoContext(ctx, "merchandise item updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("merchandise_item_id", id),
	)
	return result, nil
}

func (s *merchandiseItemService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(sharedkernel.ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		slog.ErrorContext(ctx, "failed to reorder merchandise items", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to reorder merchandise items")
	}
	slog.InfoContext(ctx, "merchandise items reordered", slog.Uint64("clinic_id", clinicID), slog.Int("count", len(ids)))
	return nil
}

func (s *merchandiseItemService) Delete(ctx context.Context, clinicID, id uint64) error {
	// Soft-delete first acquires the exclusive row lock (serializes with campaign target
	// FOR SHARE validation). Usage is re-checked in the same ambient tx; any billing /
	// estimate / campaign-target reference rolls the soft-delete back as Conflict.
	// Pattern matches trimming master delete (soft-delete → count → conflict rollback).
	if s.transactor == nil {
		return apperrors.WrapInternalServerError("merchandise item transaction dependency is required")
	}
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
			// NotFound when the same-clinic non-deleted row is absent (already deleted / wrong clinic).
			return apperrors.Wrap(err, "failed to delete merchandise item")
		}
		count, err := s.repo.CountUsageByMerchandiseItemID(txCtx, clinicID, id)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to check merchandise item dependencies", "error", err, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to check merchandise item dependencies")
		}
		if count > 0 {
			return apperrors.WrapConflict("この物販品目は請求・見積・キャンペーンで使用中のため削除できません")
		}
		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "failed to delete merchandise item", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete merchandise item")
	}
	slog.InfoContext(ctx, "merchandise item deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("merchandise_item_id", id),
	)
	return nil
}
