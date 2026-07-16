package model

import (
	"time"

	"gorm.io/gorm"
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
	ID            uint64            `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID      uint64            `gorm:"not null"                                       json:"clinic_id"`
	Name          string            `gorm:"not null"                                        json:"name"`
	Category      InventoryCategory `gorm:"type:inventory_category;not null"                json:"category"`
	Quantity      int               `gorm:"type:integer;default:0"                          json:"quantity"`
	Unit          string            `gorm:"not null;default:''"                             json:"unit"`
	MinStockLevel int               `gorm:"type:integer;default:0"                          json:"min_stock_level"`
	Location      string            `gorm:"default:''"                                      json:"location"`
	ExpiryDate    *time.Time        `gorm:"type:date"                                       json:"expiry_date,omitempty"`
	Supplier      string            `gorm:"default:''"                                      json:"supplier"`
	LastRestocked *time.Time        `gorm:"type:date"                                       json:"last_restocked,omitempty"`
	Status        InventoryStatus   `gorm:"type:inventory_status;default:'sufficient'"      json:"status"`
	CreatedAt     time.Time         `gorm:"autoCreateTime"                                  json:"created_at"`
	UpdatedAt     time.Time         `gorm:"autoUpdateTime"                                  json:"updated_at"`
	DeletedAt     gorm.DeletedAt    `                                                       json:"-"`
}

func (InventoryItem) TableName() string { return "inventory_items" }
