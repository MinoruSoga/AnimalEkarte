// Package merchandiseitem owns merchandise_items master data access.
package merchandiseitem

import (
	"context"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Repository is the data access interface for merchandise items.
type Repository interface {
	FindAll(ctx context.Context, clinicID uint64, category string) ([]model.MerchandiseItem, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error)
	CountUsageByMerchandiseItemID(ctx context.Context, clinicID, merchandiseItemID uint64) (int64, error)
	Create(ctx context.Context, item *model.MerchandiseItem) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MerchandiseItem, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type repository struct{ db *gorm.DB }

// New constructs a Repository.
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context, clinicID uint64, category string) ([]model.MerchandiseItem, error) {
	items := make([]model.MerchandiseItem, 0)
	q := r.db.WithContext(ctx).Scopes(repohelpers.ClinicScope(clinicID))
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if err := q.Order("sort_order ASC, name ASC").Find(&items).Error; err != nil {
		return nil, apperrors.FromGORM(err, "merchandise_item", "")
	}
	return items, nil
}

func (r *repository) FindByID(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error) {
	return repohelpers.FindByIDScoped[model.MerchandiseItem](ctx, r.db, "merchandise_item", clinicID, id)
}

func (r *repository) Create(ctx context.Context, item *model.MerchandiseItem) error {
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return apperrors.FromGORM(err, "merchandise_item", "")
	}
	return nil
}

func (r *repository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MerchandiseItem, error) {
	if err := repohelpers.UpdateScopedByID(ctx, r.db, &model.MerchandiseItem{}, "merchandise_item", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *repository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return repohelpers.ReorderByClinicID(ctx, r.db, &model.MerchandiseItem{}, "merchandise_item", clinicID, ids, "sort_order")
}

// CountUsageByMerchandiseItemID returns billing_items + estimate_items references (BUG-109).
// Child tables lack clinic_id; tenant isolation is via JOIN.
func (r *repository) CountUsageByMerchandiseItemID(ctx context.Context, clinicID, merchandiseItemID uint64) (int64, error) {
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

func (r *repository) Delete(ctx context.Context, clinicID, id uint64) error {
	return repohelpers.DeleteScopedByID(ctx, r.db, &model.MerchandiseItem{}, "merchandise_item", clinicID, id)
}
