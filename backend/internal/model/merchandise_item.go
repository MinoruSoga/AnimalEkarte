package model

import (
	"time"

	"gorm.io/gorm"
)

// MerchandiseItem は物販・フード・その他マスタアイテム
type MerchandiseItem struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID  uint64         `gorm:"not null"                                       json:"clinic_id"`
	Name      string         `gorm:"not null"                                       json:"name"`
	Category  ItemCategory   `gorm:"type:item_category;not null;default:goods"      json:"category"`
	UnitPrice int64          `gorm:"not null;default:0"                             json:"unit_price"`
	TaxRate   float64        `gorm:"type:numeric;not null;default:0.10"             json:"tax_rate"`
	IsActive  bool           `gorm:"default:true"                                   json:"is_active"`
	SortOrder int            `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                                          json:"-"`
}

func (MerchandiseItem) TableName() string { return "merchandise_items" }
