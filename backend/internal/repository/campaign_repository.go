package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// CampaignRepository は割引キャンペーンマスタのデータアクセスインターフェース (#81)
type CampaignRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.Campaign, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Campaign, error)
	Create(ctx context.Context, m *model.Campaign) (*model.Campaign, error)
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Campaign, error)
	ReplaceTargets(ctx context.Context, campaignID uint64, categories []model.ItemCategory, itemIDs []uint64) error
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type campaignRepository struct{ db *gorm.DB }

// NewCampaignRepository は CampaignRepository を初期化して返す
func NewCampaignRepository(db *gorm.DB) CampaignRepository {
	return &campaignRepository{db: db}
}

func (r *campaignRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Campaign, error) {
	var ms []model.Campaign
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
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
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
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
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, apperrors.FromGORM(err, "campaign", "")
	}
	return r.FindByID(ctx, m.ClinicID, m.ID)
}

func (r *campaignRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Campaign, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Campaign{}).
		Scopes(clinicScope(clinicID)).
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
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
	})
}

func (r *campaignRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
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
	return reorderByClinicID(ctx, r.db, &model.Campaign{}, "campaign", clinicID, ids)
}
