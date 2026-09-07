package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// CampaignRepository は割引キャンペーンマスタのデータアクセスインターフェース (#81)
type CampaignRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.Campaign, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Campaign, error)
	Create(ctx context.Context, m *model.Campaign) (*model.Campaign, error)
	Update(ctx context.Context, clinicID, id uint64, cmd UpdateCampaignInput) (*model.Campaign, error)
	ReplaceTargets(ctx context.Context, campaignID uint64, categories []model.ItemCategory, itemIDs []uint64) error
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	// FindApplicableForItem は指定日・カテゴリ・商品に適用される有効なキャンペーンを返す（#81 段階2b）。
	// 該当が複数ある場合は discount_value が大きいものを優先。該当なしは (nil, nil)。
	FindApplicableForItem(ctx context.Context, clinicID uint64, date time.Time, category model.ItemCategory, merchandiseItemID *uint64) (*model.Campaign, error)
	// FindAllApplicableForItem は指定日・カテゴリ・商品に適用される全有効キャンペーンを返す（#81 Q-I スタッフ選択用）。
	// discount_value 降順。該当なしは (nil, nil)。
	FindAllApplicableForItem(ctx context.Context, clinicID uint64, date time.Time, category model.ItemCategory, merchandiseItemID *uint64) ([]*model.Campaign, error)
}

type campaignRepository struct{ db *gorm.DB }

// New は CampaignRepository を初期化して返す
func NewCampaignRepository(db *gorm.DB) CampaignRepository {
	return &campaignRepository{db: db}
}

func (r *campaignRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Campaign, error) {
	var ms []model.Campaign
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Preload("TargetCategories").
		Preload("TargetItems").
		Order("sort_order ASC, start_date DESC").
		Find(&ms).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "campaign", "")
	}
	return ms, nil
}

func (r *campaignRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Campaign, error) {
	var m model.Campaign
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Preload("TargetCategories").
		Preload("TargetItems").
		First(&m, id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "campaign", fmt.Sprintf("%d", id))
	}
	return &m, nil
}

func (r *campaignRepository) Create(ctx context.Context, m *model.Campaign) (*model.Campaign, error) {
	// GORM が TargetCategories / TargetItems も関連として同時作成する
	// Capture intent before Create: gorm default:true omits zero bools from
	// INSERT and may write the DB default back into the struct (BUG-455-S3).
	wantActive := m.IsActive
	db := persistence.DBOrTx(ctx, r.db)
	if err := db.Create(m).Error; err != nil {
		return nil, apperrors.FromGORM(err, "campaign", "")
	}
	if !wantActive {
		if err := db.Model(m).Update("is_active", false).Error; err != nil {
			return nil, apperrors.FromGORM(err, "campaign", fmt.Sprintf("%d", m.ID))
		}
		m.IsActive = false
	}
	return r.FindByID(ctx, m.ClinicID, m.ID)
}

func (r *campaignRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateCampaignInput) (*model.Campaign, error) {
	return r.update(ctx, clinicID, id, buildCampaignUpdate(&cmd))
}

func (r *campaignRepository) update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Campaign, error) {
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.Campaign{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "campaign", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("campaign", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

// ReplaceTargets は campaign の対象カテゴリ・対象商品を全削除→再作成で差し替える。
// 呼び出し側(service)で campaign の所有(clinic_id)を確認済みであること。
func (r *campaignRepository) ReplaceTargets(ctx context.Context, campaignID uint64, categories []model.ItemCategory, itemIDs []uint64) error {
	if err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("campaign_id = ?", campaignID).Delete(&model.CampaignTargetCategory{}).Error; err != nil {
			return apperrors.FromGORM(err, "campaign_target_category", "")
		}
		if err := tx.Where("campaign_id = ?", campaignID).Delete(&model.CampaignTargetItem{}).Error; err != nil {
			return apperrors.FromGORM(err, "campaign_target_item", "")
		}
		for _, c := range categories {
			if err := tx.Create(&model.CampaignTargetCategory{CampaignID: campaignID, Category: c}).Error; err != nil {
				return apperrors.FromGORM(err, "campaign_target_category", "")
			}
		}
		for _, id := range itemIDs {
			if err := tx.Create(&model.CampaignTargetItem{CampaignID: campaignID, MerchandiseItemID: id}).Error; err != nil {
				return apperrors.FromGORM(err, "campaign_target_item", "")
			}
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to replace campaign targets")
	}
	return nil
}

func (r *campaignRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Delete(&model.Campaign{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "campaign", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("campaign", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *campaignRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.Campaign{}, "campaign", clinicID, ids, "sort_order")
}

func (r *campaignRepository) buildApplicableCond(category model.ItemCategory, merchandiseItemID *uint64) (cond string, args []any) {
	cond = "EXISTS (SELECT 1 FROM campaign_target_categories ctc WHERE ctc.campaign_id = campaigns.id AND ctc.category = ?)"
	args = []any{category}
	if merchandiseItemID != nil {
		cond += " OR EXISTS (SELECT 1 FROM campaign_target_items cti WHERE cti.campaign_id = campaigns.id AND cti.merchandise_item_id = ?)"
		args = append(args, *merchandiseItemID)
	}
	return cond, args
}

func (r *campaignRepository) FindApplicableForItem(ctx context.Context, clinicID uint64, date time.Time, category model.ItemCategory, merchandiseItemID *uint64) (*model.Campaign, error) {
	cond, args := r.buildApplicableCond(category, merchandiseItemID)
	var campaign model.Campaign
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("is_active = ? AND start_date <= ? AND end_date >= ?", true, date, date).
		Where(cond, args...).
		Order("discount_value DESC").
		First(&campaign).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 該当キャンペーンなし
		}
		return nil, apperrors.FromGORM(err, "campaign", "")
	}
	return &campaign, nil
}

func (r *campaignRepository) FindAllApplicableForItem(ctx context.Context, clinicID uint64, date time.Time, category model.ItemCategory, merchandiseItemID *uint64) ([]*model.Campaign, error) {
	cond, args := r.buildApplicableCond(category, merchandiseItemID)
	var campaigns []*model.Campaign
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("is_active = ? AND start_date <= ? AND end_date >= ?", true, date, date).
		Where(cond, args...).
		Order("discount_value DESC").
		Find(&campaigns).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "campaign", "")
	}
	return campaigns, nil
}
