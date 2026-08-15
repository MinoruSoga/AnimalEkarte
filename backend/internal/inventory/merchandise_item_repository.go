package inventory

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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
	// G2F-11: vaccine/procedure/consultation と同型の master list safety Limit（unbounded Find 防止）。
	if err := q.Order("sort_order ASC, name ASC").Limit(persistence.MaxMasterListRows).Find(&items).Error; err != nil {
		return nil, apperrors.FromGORM(err, "merchandise_item", "")
	}
	return items, nil
}

// FindByID loads a clinic-scoped merchandise row. When called under an ambient
// transaction it takes FOR SHARE so concurrent soft-delete/update waits until
// the caller commits (campaign target attachment serialization).
func (r *merchandiseItemRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error) {
	var item model.MerchandiseItem
	db := persistence.DBOrTx(ctx, r.db)
	if persistence.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	if err := db.Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).First(&item).Error; err != nil {
		return nil, apperrors.FromGORM(err, "merchandise_item", fmt.Sprintf("%d", id))
	}
	return &item, nil
}

func (r *merchandiseItemRepository) Create(ctx context.Context, item *model.MerchandiseItem) error {
	db := persistence.DBOrTx(ctx, r.db)
	// Capture intent before Create: gorm default:true omits zero bools from
	// INSERT and may write the DB default back into the struct (BUG-455-S4).
	wantActive := item.IsActive
	if err := db.Create(item).Error; err != nil {
		return apperrors.FromGORM(err, "merchandise_item", "")
	}
	if !wantActive {
		if err := db.Model(item).Update("is_active", false).Error; err != nil {
			return apperrors.FromGORM(err, "merchandise_item", "")
		}
		item.IsActive = false
	}
	return nil
}

// Update updates fields and reloads the row in one transaction so a reload
// failure cannot invert a committed write into a failure response (BUG-465).
func (r *merchandiseItemRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MerchandiseItem, error) {
	var loaded *model.MerchandiseItem
	err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		if err := persistence.UpdateScopedByID(txCtx, tx, &model.MerchandiseItem{}, "merchandise_item", clinicID, id, fields); err != nil {
			return err
		}
		var err error
		loaded, err = persistence.FindByIDScoped[model.MerchandiseItem](txCtx, tx, "merchandise_item", clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "reload merchandise_item after update")
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update and reload merchandise_item")
	}
	return loaded, nil
}

func (r *merchandiseItemRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.MerchandiseItem{}, "merchandise_item", clinicID, ids, "sort_order")
}

// CountUsageByMerchandiseItemID returns active billing_items + estimate_items references
// (BUG-109) plus campaign_target_items joined to any same-clinic non-deleted campaign
// (regardless of is_active / date window). Child tables lack clinic_id; tenant isolation
// is via JOIN. Uses DBOrTx so delete's ambient transaction sees uncommitted attaches.
func (r *merchandiseItemRepository) CountUsageByMerchandiseItemID(ctx context.Context, clinicID, merchandiseItemID uint64) (int64, error) {
	db := persistence.DBOrTx(ctx, r.db)

	var billingCount int64
	if err := db.
		Model(&model.BillingItem{}).
		Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Where("billing_items.merchandise_item_id = ? AND billing_items.deleted_at IS NULL", merchandiseItemID).
		Count(&billingCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "billing_item", "")
	}

	var estimateCount int64
	if err := db.
		Model(&model.EstimateItem{}).
		Joins("JOIN estimates ON estimates.id = estimate_items.estimate_id AND estimates.clinic_id = ? AND estimates.deleted_at IS NULL", clinicID).
		Where("estimate_items.merchandise_item_id = ? AND estimate_items.deleted_at IS NULL", merchandiseItemID).
		Count(&estimateCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "estimate_item", "")
	}

	// Campaign targets block delete even when the campaign is inactive or outside its date
	// window: historical/target configuration still references the merchandise row.
	var campaignCount int64
	if err := db.
		Model(&model.CampaignTargetItem{}).
		Joins("JOIN campaigns ON campaigns.id = campaign_target_items.campaign_id AND campaigns.clinic_id = ? AND campaigns.deleted_at IS NULL", clinicID).
		Where("campaign_target_items.merchandise_item_id = ?", merchandiseItemID).
		Count(&campaignCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "campaign_target_item", "")
	}

	return billingCount + estimateCount + campaignCount, nil
}

func (r *merchandiseItemRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	// Ambient tx participation is required so soft-delete exclusive-locks serialize with
	// campaign target FOR SHARE validation and usage re-check in the same transaction.
	return persistence.DeleteScopedByID(ctx, persistence.DBOrTx(ctx, r.db), &model.MerchandiseItem{}, "merchandise_item", clinicID, id)
}
