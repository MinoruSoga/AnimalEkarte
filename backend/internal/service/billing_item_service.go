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
	ClinicID              uint64
	BillingID             uint64
	Category              string // Service 内で model.ItemCategory に変換
	Name                  string
	UnitPrice             int64
	Quantity              float64
	TaxType               string // "" = デフォルト "excluded"
	TaxRate               float64
	IsInsuranceApplicable bool
	Source                string // "" = デフォルト "manual"
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

func buildBillingItemUpdate(input *UpdateBillingItemInput) map[string]any {
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
	CreateItem(ctx context.Context, input *CreateBillingItemInput) (*model.BillingItem, error)
	UpdateItem(ctx context.Context, clinicID, id uint64, input *UpdateBillingItemInput) (*model.BillingItem, error)
	DeleteItem(ctx context.Context, clinicID, id uint64) error
	GetUnbilledItems(ctx context.Context, clinicID, petID uint64) ([]model.Treatment, error)
}

type billingItemService struct {
	repo          repository.BillingItemRepository
	billingRepo   repository.AccountingRepository
	treatmentRepo repository.TreatmentRepository
}

// NewBillingItemService は BillingItemService を初期化して返す
func NewBillingItemService(repo repository.BillingItemRepository, billingRepo repository.AccountingRepository, treatmentRepo repository.TreatmentRepository) BillingItemService {
	return &billingItemService{repo: repo, billingRepo: billingRepo, treatmentRepo: treatmentRepo}
}

func (s *billingItemService) CreateItem(ctx context.Context, input *CreateBillingItemInput) (*model.BillingItem, error) {
	if input.BillingID == 0 {
		return nil, apperrors.WrapInvalidInput("請求IDは必須です")
	}
	if input.Name == "" {
		return nil, apperrors.WrapInvalidInput("商品名は必須です")
	}
	if input.UnitPrice < 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgPriceZeroOrMore)
	}

	// テナント所有権確認: billing が同一クリニックに属することを確認
	if _, err := s.billingRepo.FindByID(ctx, input.ClinicID, input.BillingID); err != nil {
		slog.ErrorContext(ctx, "billing not found or belongs to different clinic", "error", err)
		return nil, apperrors.Wrap(err, "billing not found or belongs to different clinic")
	}

	// カテゴリバリデーション
	if err := validateItemCategory(input.Category); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate item category")
	}

	// TaxType デフォルト設定とバリデーション
	taxType := model.TaxTypeExcluded
	if input.TaxType != "" {
		if err := validateTaxType(input.TaxType); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate tax type")
		}
		taxType = model.TaxType(input.TaxType)
	}

	// TaxRate デフォルト設定
	taxRate := 0.10
	if input.TaxRate > 0 {
		taxRate = input.TaxRate
	}

	// Source デフォルト設定とバリデーション
	source := model.ItemSourceManual
	if input.Source != "" {
		if err := validateItemSource(input.Source); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate item source")
		}
		source = model.ItemSource(input.Source)
	}

	item := &model.BillingItem{
		BillingID:             input.BillingID,
		Category:              model.ItemCategory(input.Category),
		Name:                  input.Name,
		UnitPrice:             input.UnitPrice,
		Quantity:              input.Quantity,
		TaxType:               taxType,
		TaxRate:               taxRate,
		IsInsuranceApplicable: input.IsInsuranceApplicable,
		Source:                source,
		SortOrder:             input.SortOrder,
	}

	if err := s.repo.Create(ctx, item); err != nil {
		slog.ErrorContext(ctx, "failed to create billing item", "error", err)
		return nil, apperrors.Wrap(err, "failed to create billing item")
	}

	if err := s.recalculateTotals(ctx, input.ClinicID, input.BillingID); err != nil {
		slog.WarnContext(ctx, "failed to recalculate billing totals after create",
			slog.Uint64("billing_id", input.BillingID),
			slog.String("error", err.Error()),
		)
	}

	slog.InfoContext(ctx, "billing item created",
		slog.Uint64("clinic_id", input.ClinicID),
		slog.Uint64("billing_id", input.BillingID),
		slog.Uint64("item_id", item.ID),
	)
	return item, nil
}

func (s *billingItemService) UpdateItem(ctx context.Context, clinicID, id uint64, input *UpdateBillingItemInput) (*model.BillingItem, error) {
	item, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get billing item", "error", err)
		return nil, apperrors.Wrap(err, "failed to get billing item")
	}
	if input.UnitPrice != nil && *input.UnitPrice < 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgPriceZeroOrMore)
	}

	fields := buildBillingItemUpdate(input)
	if len(fields) == 0 {
		return item, nil
	}

	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		slog.ErrorContext(ctx, "failed to update billing item", "error", err)
		return nil, apperrors.Wrap(err, "failed to update billing item")
	}

	if err := s.recalculateTotals(ctx, clinicID, item.BillingID); err != nil {
		slog.WarnContext(ctx, "failed to recalculate billing totals after update",
			slog.Uint64("billing_id", item.BillingID),
			slog.String("error", err.Error()),
		)
	}

	slog.InfoContext(ctx, "billing item updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("item_id", id),
		slog.Uint64("billing_id", item.BillingID),
	)
	updated, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get billing item after update", "error", err)
		return nil, apperrors.Wrap(err, "failed to get billing item after update")
	}
	return updated, nil
}

func (s *billingItemService) DeleteItem(ctx context.Context, clinicID, id uint64) error {
	item, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to get billing item")
	}
	billingID := item.BillingID

	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to delete billing item", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete billing item")
	}

	if err := s.recalculateTotals(ctx, clinicID, billingID); err != nil {
		slog.WarnContext(ctx, "failed to recalculate billing totals after delete",
			slog.Uint64("billing_id", billingID),
			slog.String("error", err.Error()),
		)
	}

	slog.InfoContext(ctx, "billing item deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("item_id", id),
		slog.Uint64("billing_id", billingID),
	)
	return nil
}

func (s *billingItemService) GetUnbilledItems(ctx context.Context, clinicID, petID uint64) ([]model.Treatment, error) {
	treatments, err := s.treatmentRepo.FindUnbilledByPetID(ctx, clinicID, petID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find unbilled treatments", "error", err)
		return nil, apperrors.Wrap(err, "failed to find unbilled treatments")
	}
	return treatments, nil
}

// recalculateTotals は billing の全明細から subtotal/tax_total/total_amount を再計算して保存する
func (s *billingItemService) recalculateTotals(ctx context.Context, clinicID, billingID uint64) error {
	items, err := s.repo.FindByBillingID(ctx, clinicID, billingID)
	if err != nil {
		return apperrors.Wrap(err, "failed to find billing items")
	}
	subtotal, taxTotal, totalAmount := CalculateBillingTotals(items)
	if err := s.repo.UpdateBillingTotals(ctx, clinicID, billingID, subtotal, taxTotal, totalAmount); err != nil {
		return apperrors.Wrap(err, "failed to update billing totals")
	}
	return nil
}
