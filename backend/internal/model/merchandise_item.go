package model

import (
	"gorm.io/gorm"
	"time"
)

// MerchandiseItem は物販・フード・その他マスタアイテム
type MerchandiseItem struct {
	ID        uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID  uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Name      string         `gorm:"column:name;type:text;not null;default:''" json:"name"`
	Category  ItemCategory   `gorm:"column:category;type:item_category;not null;default:'goods'" json:"category"`
	UnitPrice int64          `gorm:"column:unit_price;type:bigint;not null;default:0" json:"unit_price"`
	TaxType   TaxType        `gorm:"column:tax_type;type:tax_type;not null;default:'excluded'" json:"tax_type"`
	TaxRate   ExactNumeric   `gorm:"column:tax_rate;type:numeric;not null;default:0.10" json:"tax_rate"`
	IsActive  bool           `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	SortOrder int            `gorm:"column:sort_order;type:integer;not null;default:0" json:"sort_order"`
	CreatedAt time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
}

func (MerchandiseItem) TableName() string { return "merchandise_items" }
