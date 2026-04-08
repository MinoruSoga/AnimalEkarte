// Package service provides business logic implementations for BillingItem entity.
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
	colBillingItemUnitPrice             = "unit_price"
	colBillingItemQuantity              = "quantity"
	colBillingItemTaxType               = "tax_type"
	colBillingItemTaxRate               = "tax_rate"
	colBillingItemIsInsuranceApplicable = "is_insurance_applicable"
)

// --- Input DTOs ---

// CreateBillingItemInput は billing_item 作成の入力DTO
type CreateBillingItemInput struct {
	BillingID             uint64
	Category              model.ItemCategory
	Name                  string
	UnitPrice             int64
	Quantity              float64
	TaxType               model.TaxType
	TaxRate               float64
	IsInsuranceApplicable bool
	Source                model.ItemSource
	SortOrder             int
}

// UpdateBillingItemInput は billing_item 更新の入力DTO（nil = 未指定）
type UpdateBillingItemInput struct {
	UnitPrice             *int64
	Quantity              *float64
	TaxType               *model.TaxType
	TaxRate               *float64
	IsInsuranceApplicable *bool
}

func buildBillingItemUpdateFields(input *UpdateBillingItemInput) map[string]any {
	fields := make(map[string]any)
	if input.UnitPrice != nil {
		fields[colBillingItemUnitPrice] = *input.UnitPrice
	}
	if input.Quantity != nil {
		fields[colBillingItemQuantity] = *input.Quantity
	}
	if input.TaxType != nil {
		fields[colBillingItemTaxType] = *input.TaxType
	}
	if input.TaxRate != nil {
		fields[colBillingItemTaxRate] = *input.TaxRate
	}
	if input.IsInsuranceApplicable != nil {
		fields[colBillingItemIsInsuranceApplicable] = *input.IsInsuranceApplicable
	}
	return fields
}

// ---- BillingItemService ----

// BillingItemService は billing_items の CRUD とトータル再計算を担うインターフェース
type BillingItemService interface {
	GetByID(ctx context.Context, clinicID uint64, id uint64) (*model.BillingItem, error)
	CreateItem(ctx context.Context, input *CreateBillingItemInput) (*model.BillingItem, error)
	UpdateItem(ctx context.Context, clinicID uint64, id uint64, input *UpdateBillingItemInput) (*model.BillingItem, error)
	DeleteItem(ctx context.Context, clinicID uint64, id uint64) error
}

type billingItemService struct {
	repo repository.BillingItemRepository
}

// NewBillingItemService は BillingItemService を初期化して返す
func NewBillingItemService(repo repository.BillingItemRepository) BillingItemService {
	return &billingItemService{repo: repo}
}

func (s *billingItemService) GetByID(ctx context.Context, clinicID uint64, id uint64) (*model.BillingItem, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *billingItemService) CreateItem(ctx context.Context, input *CreateBillingItemInput) (*model.BillingItem, error) {
	if input.Name == "" {
		return nil, apperrors.WrapInvalidInput("name is required")
	}
	if input.BillingID == 0 {
		return nil, apperrors.WrapInvalidInput("billing_id is required")
	}
	if input.UnitPrice < 0 {
		return nil, apperrors.WrapInvalidInput("金額は0以上を入力してください")
	}

	item := &model.BillingItem{
		BillingID:             input.BillingID,
		Category:              input.Category,
		Name:                  input.Name,
		UnitPrice:             input.UnitPrice,
		Quantity:              input.Quantity,
		TaxType:               input.TaxType,
		TaxRate:               input.TaxRate,
		IsInsuranceApplicable: input.IsInsuranceApplicable,
		Source:                input.Source,
		SortOrder:             input.SortOrder,
	}

	if err := s.repo.Create(ctx, item); err != nil {
		return nil, apperrors.Wrap(err, "failed to create billing item")
	}

	if err := s.recalculateTotals(ctx, input.BillingID); err != nil {
		slog.WarnContext(ctx, "failed to recalculate billing totals after create",
			slog.Uint64("billing_id", input.BillingID),
			slog.String("error", err.Error()),
		)
	}

	slog.InfoContext(ctx, "billing item created",
		slog.Uint64("billing_id", input.BillingID),
		slog.Uint64("item_id", item.ID),
	)
	return item, nil
}

func (s *billingItemService) UpdateItem(ctx context.Context, clinicID uint64, id uint64, input *UpdateBillingItemInput) (*model.BillingItem, error) {
	item, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get billing item")
	}
	if input.UnitPrice != nil && *input.UnitPrice < 0 {
		return nil, apperrors.WrapInvalidInput("金額は0以上を入力してください")
	}

	fields := buildBillingItemUpdateFields(input)
	if len(fields) == 0 {
		return item, nil
	}

	if err := s.repo.UpdateFields(ctx, id, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update billing item")
	}

	if err := s.recalculateTotals(ctx, item.BillingID); err != nil {
		slog.WarnContext(ctx, "failed to recalculate billing totals after update",
			slog.Uint64("billing_id", item.BillingID),
			slog.String("error", err.Error()),
		)
	}

	slog.InfoContext(ctx, "billing item updated",
		slog.Uint64("item_id", id),
		slog.Uint64("billing_id", item.BillingID),
	)
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *billingItemService) DeleteItem(ctx context.Context, clinicID uint64, id uint64) error {
	item, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to get billing item")
	}
	billingID := item.BillingID

	if err := s.repo.Delete(ctx, id); err != nil {
		return apperrors.Wrap(err, "failed to delete billing item")
	}

	if err := s.recalculateTotals(ctx, billingID); err != nil {
		slog.WarnContext(ctx, "failed to recalculate billing totals after delete",
			slog.Uint64("billing_id", billingID),
			slog.String("error", err.Error()),
		)
	}

	slog.InfoContext(ctx, "billing item deleted",
		slog.Uint64("item_id", id),
		slog.Uint64("billing_id", billingID),
	)
	return nil
}

// recalculateTotals は billing の全明細から subtotal/tax_total/total_amount を再計算して保存する
func (s *billingItemService) recalculateTotals(ctx context.Context, billingID uint64) error {
	items, err := s.repo.FindByBillingID(ctx, billingID)
	if err != nil {
		return apperrors.Wrap(err, "find billing items")
	}
	subtotal, taxTotal, totalAmount := CalculateBillingTotals(items)
	return s.repo.UpdateBillingTotals(ctx, billingID, subtotal, taxTotal, totalAmount)
}
