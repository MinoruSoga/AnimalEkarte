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
	Update(ctx context.Context, clinicID, id uint64, cmd UpdateMerchandiseItemInput) (*model.MerchandiseItem, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type merchandiseItemRepository struct{ db *gorm.DB }

// merchandiseItemSelectColumns は物販 reader の明示列リスト（EMR-449 A.3 read 証拠）。
// SELECT * + struct scan では列未接続が nil に黙って化けるため、新列
// is_bill_discount_excluded / version を含む明示取得で列欠落・読取失敗を error として表面化させる。
const merchandiseItemSelectColumns = "id, clinic_id, name, category, unit_price, tax_type, tax_rate, is_active, is_bill_discount_excluded, sort_order, version, created_at, updated_at, deleted_at"

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
	if err := q.Select(merchandiseItemSelectColumns).Order("sort_order ASC, name ASC").Limit(persistence.MaxMasterListRows).Find(&items).Error; err != nil {
		return nil, apperrors.FromGORM(err, "merchandise_item", "")
	}
	return items, nil
}

// FindByID loads a clinic-scoped merchandise row. When called under an ambient
// transaction it takes FOR SHARE so concurrent soft-delete/update waits until
// the caller commits (campaign target attachment serialization).
func (r *merchandiseItemRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error) {
	return r.findByID(ctx, clinicID, id, false)
}

// FindByIDForUpdate は FindByID と同じ query を ambient tx の行で FOR UPDATE
// ロックして読む。更新 TX の前値読取だけが使う（EMR-449 A-2）。
func (r *merchandiseItemRepository) FindByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error) {
	return r.findByID(ctx, clinicID, id, true)
}

func (r *merchandiseItemRepository) findByID(ctx context.Context, clinicID, id uint64, forUpdate bool) (*model.MerchandiseItem, error) {
	var item model.MerchandiseItem
	db := persistence.DBOrTx(ctx, r.db)
	if persistence.TxFromContext(ctx) != nil {
		// EMR-449 A-2: 既定は参照直列化の FOR SHARE。更新 TX の前値読取だけ
		// FOR UPDATE へ切替え、同版 PATCH が SHARE 保持後の UPDATE 昇格で
		// 相互待ち（40P01）しないようにする。
		strength := "SHARE"
		if forUpdate {
			strength = "UPDATE"
		}
		db = db.Clauses(clause.Locking{Strength: strength})
	}
	if err := db.Select(merchandiseItemSelectColumns).Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).First(&item).Error; err != nil {
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
// cmd.Version が送信された場合は編集開始版を WHERE で照合する（EMR-449 A.4）。
// version 列は書かない: app_private.entity_version_guard() トリガが実変更時だけ増分させる。
func (r *merchandiseItemRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateMerchandiseItemInput) (*model.MerchandiseItem, error) {
	return r.updateAndReload(ctx, clinicID, id, buildMerchandiseItemUpdate(&cmd), cmd.Version)
}

func (r *merchandiseItemRepository) update(ctx context.Context, clinicID, id uint64, fields map[string]any, expectedVersion *int64) error {
	db := persistence.DBOrTx(ctx, r.db).Model(&model.MerchandiseItem{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id)
	if expectedVersion != nil {
		db = db.Where("version = ?", *expectedVersion)
	}
	result := db.Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "merchandise_item", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		// 版不一致（行は存在）と行欠落を区別する: 既存の競合 response（409）と not found をそのまま使う。
		if _, err := r.FindByID(ctx, clinicID, id); err != nil {
			return err
		}
		return apperrors.WrapConflict("編集後に変更されています。再読み込みしてから保存してください")
	}
	return nil
}

func (r *merchandiseItemRepository) updateAndReload(ctx context.Context, clinicID, id uint64, fields map[string]any, expectedVersion *int64) (*model.MerchandiseItem, error) {
	var loaded *model.MerchandiseItem
	err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		if err := r.update(txCtx, clinicID, id, fields, expectedVersion); err != nil {
			return err
		}
		var err error
		loaded, err = persistence.FindByIDScoped[model.MerchandiseItem](txCtx, tx.Select(merchandiseItemSelectColumns), "merchandise_item", clinicID, id)
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
	return persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Scopes(persistence.ClinicScope(clinicID)).
			Where("id = ?", id).
			First(&model.MerchandiseItem{}).Error; err != nil {
			return apperrors.FromGORM(err, "merchandise_item", fmt.Sprintf("%d", id))
		}
		result := tx.
			Scopes(persistence.ClinicScope(clinicID)).
			Where("id = ?", id).
			Where(`NOT EXISTS (
				SELECT 1 FROM billing_items
				JOIN billings ON billings.id = billing_items.billing_id
				  AND billings.clinic_id = ?
				  AND billings.deleted_at IS NULL
				WHERE billing_items.merchandise_item_id = merchandise_items.id
				  AND billing_items.deleted_at IS NULL
			)`, clinicID).
			Where(`NOT EXISTS (
				SELECT 1 FROM estimate_items
				JOIN estimates ON estimates.id = estimate_items.estimate_id
				  AND estimates.clinic_id = ?
				  AND estimates.deleted_at IS NULL
				WHERE estimate_items.merchandise_item_id = merchandise_items.id
				  AND estimate_items.deleted_at IS NULL
			)`, clinicID).
			Where(`NOT EXISTS (
				SELECT 1 FROM campaign_target_items
				JOIN campaigns ON campaigns.id = campaign_target_items.campaign_id
				  AND campaigns.clinic_id = ?
				  AND campaigns.deleted_at IS NULL
				WHERE campaign_target_items.merchandise_item_id = merchandise_items.id
			)`, clinicID).
			Delete(&model.MerchandiseItem{})
		if result.Error != nil {
			return apperrors.FromGORM(result.Error, "merchandise_item", fmt.Sprintf("%d", id))
		}
		if result.RowsAffected == 0 {
			return r.normalizeMerchandiseDeleteMiss(persistence.WithTxValue(ctx, tx), clinicID, id)
		}
		return nil
	})
}

func (r *merchandiseItemRepository) normalizeMerchandiseDeleteMiss(ctx context.Context, clinicID, id uint64) error {
	if _, err := r.FindByID(ctx, clinicID, id); err != nil {
		return err
	}
	return apperrors.WrapConflict("この物販品目は請求・見積・キャンペーンで使用中のため削除できません")
}
