package inventory

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// MerchandiseItemRepository is the merchandise-item persistence surface used
// by the domain service and type-safe application composition.
type MerchandiseItemRepository interface {
	FindAll(ctx context.Context, clinicID uint64, category string) ([]model.MerchandiseItem, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error)
	CountUsageByMerchandiseItemID(ctx context.Context, clinicID, merchandiseItemID uint64) (int64, error)
	Create(ctx context.Context, item *model.MerchandiseItem) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MerchandiseItem, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type merchandiseItemRepository struct{ db *gorm.DB }

// NewMerchandiseItemRepository constructs the merchandise-item repository.
func NewMerchandiseItemRepository(db *gorm.DB) MerchandiseItemRepository {
	return &merchandiseItemRepository{db: db}
}

func (r *merchandiseItemRepository) FindAll(ctx context.Context, clinicID uint64, category string) ([]model.MerchandiseItem, error) {
	items := make([]model.MerchandiseItem, 0)
	q := r.db.WithContext(ctx).Scopes(persistence.ClinicScope(clinicID))
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if err := q.Order("sort_order ASC, name ASC").Find(&items).Error; err != nil {
		return nil, apperrors.FromGORM(err, "merchandise_item", "")
	}
	return items, nil
}

func (r *merchandiseItemRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error) {
	return persistence.FindByIDScoped[model.MerchandiseItem](ctx, r.db, "merchandise_item", clinicID, id)
}

func (r *merchandiseItemRepository) Create(ctx context.Context, item *model.MerchandiseItem) error {
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return apperrors.FromGORM(err, "merchandise_item", "")
	}
	return nil
}

func (r *merchandiseItemRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MerchandiseItem, error) {
	if err := persistence.UpdateScopedByID(ctx, r.db, &model.MerchandiseItem{}, "merchandise_item", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *merchandiseItemRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.MerchandiseItem{}, "merchandise_item", clinicID, ids, "sort_order")
}

// CountUsageByMerchandiseItemID returns billing_items + estimate_items references (BUG-109).
// Child tables lack clinic_id; tenant isolation is via JOIN.
func (r *merchandiseItemRepository) CountUsageByMerchandiseItemID(ctx context.Context, clinicID, merchandiseItemID uint64) (int64, error) {
	var billingCount int64
	if err := r.db.WithContext(ctx).
		Model(&model.BillingItem{}).
		Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Where("billing_items.merchandise_item_id = ? AND billing_items.deleted_at IS NULL", merchandiseItemID).
		Count(&billingCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "billing_item", "")
	}

	var estimateCount int64
	if err := r.db.WithContext(ctx).
		Model(&model.EstimateItem{}).
		Joins("JOIN estimates ON estimates.id = estimate_items.estimate_id AND estimates.clinic_id = ? AND estimates.deleted_at IS NULL", clinicID).
		Where("estimate_items.merchandise_item_id = ? AND estimate_items.deleted_at IS NULL", merchandiseItemID).
		Count(&estimateCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "estimate_item", "")
	}

	return billingCount + estimateCount, nil
}

func (r *merchandiseItemRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return persistence.DeleteScopedByID(ctx, r.db, &model.MerchandiseItem{}, "merchandise_item", clinicID, id)
}
