package model

import (
	"gorm.io/gorm"
	"time"
)

// CampaignDiscountType はキャンペーン割引の種別（率 or 額）。
type CampaignDiscountType string

const (
	CampaignDiscountTypeRate   CampaignDiscountType = "rate"
	CampaignDiscountTypeAmount CampaignDiscountType = "amount"
)

// Campaign は割引キャンペーンマスタ (issue #81)。
// payment_methods / trimming_course_types と同型の拡張可能マスタ。
// Q1=D: 対象はカテゴリ単位(TargetCategories)+個別商品(TargetItems)の併用。
type Campaign struct {
	ID               uint64                   `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID         uint64                   `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Name             string                   `gorm:"column:name;type:text;not null;default:''" json:"name"`
	StartDate        time.Time                `gorm:"column:start_date;type:date;not null" json:"start_date"`
	EndDate          time.Time                `gorm:"column:end_date;type:date;not null" json:"end_date"`
	DiscountType     CampaignDiscountType     `gorm:"column:discount_type;type:campaign_discount_type;not null;default:'rate'" json:"discount_type"`
	DiscountValue    ExactNumeric             `gorm:"column:discount_value;type:numeric(12,2);not null;default:0" json:"discount_value"`
	IsActive         bool                     `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	SortOrder        int                      `gorm:"column:sort_order;type:integer;not null;default:0" json:"sort_order"`
	CreatedAt        time.Time                `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time                `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt           `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	TargetCategories []CampaignTargetCategory `gorm:"foreignKey:CampaignID;references:ID" json:"target_categories,omitempty"`
	TargetItems      []CampaignTargetItem     `gorm:"foreignKey:CampaignID;references:ID" json:"target_items,omitempty"`
}

func (Campaign) TableName() string { return "campaigns" }

// CampaignTargetCategory はキャンペーン対象カテゴリ (Q1=D)。
type CampaignTargetCategory struct {
	ID         uint64       `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	CampaignID uint64       `gorm:"column:campaign_id;type:bigint;not null" json:"campaign_id"`
	Category   ItemCategory `gorm:"column:category;type:item_category;not null" json:"category"`
}

func (CampaignTargetCategory) TableName() string { return "campaign_target_categories" }

// CampaignTargetItem はキャンペーン対象の個別商品 (Q1=D)。
type CampaignTargetItem struct {
	ID                uint64 `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	CampaignID        uint64 `gorm:"column:campaign_id;type:bigint;not null" json:"campaign_id"`
	MerchandiseItemID uint64 `gorm:"column:merchandise_item_id;type:bigint;not null" json:"merchandise_item_id"`
}

func (CampaignTargetItem) TableName() string { return "campaign_target_items" }
