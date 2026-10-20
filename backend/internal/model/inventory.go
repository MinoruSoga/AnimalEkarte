package model

import (
	"gorm.io/gorm"
	"time"
)

type InventoryCategory string

const (
	InventoryCategoryMedicine   InventoryCategory = "medicine"
	InventoryCategoryConsumable InventoryCategory = "consumable"
	InventoryCategoryFood       InventoryCategory = "food"
	InventoryCategoryOther      InventoryCategory = "other"
)

type InventoryStatus string

const (
	InventoryStatusSufficient InventoryStatus = "sufficient"
	InventoryStatusLow        InventoryStatus = "low"
	InventoryStatusOutOfStock InventoryStatus = "out_of_stock"
)

// DeriveInventoryStatus は quantity と minStockLevel から在庫ステータスを導出する。
// minStockLevel <= 0 は「閾値未設定」を意味し、quantity <= 0 の在庫切れ判定以外では
// low 判定を行わない（SD-4: 従来 status は保存値のみで quantity/min_stock_level との
// 自動連動が存在しなかった）。
func DeriveInventoryStatus(quantity, minStockLevel int) InventoryStatus {
	if quantity <= 0 {
		return InventoryStatusOutOfStock
	}
	if minStockLevel > 0 && quantity <= minStockLevel {
		return InventoryStatusLow
	}
	return InventoryStatusSufficient
}

type InventoryItem struct {
	ID             uint64            `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID       uint64            `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Name           string            `gorm:"column:name;type:text;not null" json:"name"`
	Category       InventoryCategory `gorm:"column:category;type:inventory_category;not null" json:"category"`
	Quantity       *int64            `gorm:"column:quantity;type:bigint;default:0" json:"quantity"`
	Unit           string            `gorm:"column:unit;type:text;not null;default:''" json:"unit"`
	MinStockLevel  *int              `gorm:"column:min_stock_level;type:integer;default:0" json:"min_stock_level"`
	Location       string            `gorm:"column:location;type:text;not null;default:''" json:"location"`
	ExpiryDate     *time.Time        `gorm:"column:expiry_date;type:date" json:"expiry_date,omitempty"`
	Supplier       string            `gorm:"column:supplier;type:text;not null;default:''" json:"supplier"`
	CreatedAt      time.Time         `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time         `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt    `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	Version        int64             `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
	QuantityBaseID *uint64           `gorm:"column:quantity_base_id;type:bigint" json:"-"`
}

func (InventoryItem) TableName() string { return "inventory_items" }
