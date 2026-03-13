package model

import (
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
type InventoryItem struct {
	ID            uint64            `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID      uint64            `gorm:"not null"                                       json:"clinic_id"`
	Name          string            `gorm:"not null"                                        json:"name"`
	Category      InventoryCategory `gorm:"type:inventory_category;not null"                json:"category"`
	Quantity      int               `gorm:"default:0"                                       json:"quantity"`
	Unit          string            `gorm:"not null;default:''"                             json:"unit"`
	MinStockLevel int               `gorm:"default:0"                                       json:"min_stock_level"`
	Location      string            `gorm:"default:''"                                      json:"location"`
	ExpiryDate    *time.Time        `gorm:"type:date"                                       json:"expiry_date,omitempty"`
	Supplier      string            `gorm:"default:''"                                      json:"supplier"`
	LastRestocked *time.Time        `gorm:"type:date"                                       json:"last_restocked,omitempty"`
	Status        InventoryStatus   `gorm:"type:inventory_status;default:'sufficient'"      json:"status"`
	CreatedAt     time.Time         `gorm:"autoCreateTime"                                  json:"created_at"`
	UpdatedAt     time.Time         `gorm:"autoUpdateTime"                                  json:"updated_at"`
}

func (InventoryItem) TableName() string { return "inventory_items" }
