package model

import (
	"time"

	"gorm.io/gorm"
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
	ID            uint64               `gorm:"primaryKey;autoIncrement"                          json:"id"`
	ClinicID      uint64               `gorm:"not null"                                          json:"clinic_id"`
	Name          string               `gorm:"not null;default:''"                               json:"name"`
	StartDate     time.Time            `gorm:"type:date;not null"                                json:"start_date"`
	EndDate       time.Time            `gorm:"type:date;not null"                                json:"end_date"`
	DiscountType  CampaignDiscountType `gorm:"type:campaign_discount_type;not null;default:'rate'" json:"discount_type"`
	DiscountValue float64              `gorm:"type:numeric(12,2);not null;default:0"             json:"discount_value"`
	IsActive      bool                 `gorm:"not null;default:true"                             json:"is_active"`
	SortOrder     int                  `gorm:"not null;default:0"                                json:"sort_order"`
	CreatedAt     time.Time            `gorm:"autoCreateTime"                                    json:"created_at"`
	UpdatedAt     time.Time            `gorm:"autoUpdateTime"                                    json:"updated_at"`
	DeletedAt     gorm.DeletedAt       `                                                        json:"-"`

	TargetCategories []CampaignTargetCategory `gorm:"foreignKey:CampaignID" json:"target_categories,omitempty"`
	TargetItems      []CampaignTargetItem     `gorm:"foreignKey:CampaignID" json:"target_items,omitempty"`
}

func (Campaign) TableName() string { return "campaigns" }

// CampaignTargetCategory はキャンペーン対象カテゴリ (Q1=D)。
type CampaignTargetCategory struct {
	ID         uint64       `gorm:"primaryKey;autoIncrement"    json:"id"`
	CampaignID uint64       `gorm:"not null"                    json:"campaign_id"`
	Category   ItemCategory `gorm:"type:item_category;not null" json:"category"`
}

func (CampaignTargetCategory) TableName() string { return "campaign_target_categories" }

// CampaignTargetItem はキャンペーン対象の個別商品 (Q1=D)。
type CampaignTargetItem struct {
	ID                uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	CampaignID        uint64 `gorm:"not null"                 json:"campaign_id"`
	MerchandiseItemID uint64 `gorm:"not null"                 json:"merchandise_item_id"`
}

func (CampaignTargetItem) TableName() string { return "campaign_target_items" }
