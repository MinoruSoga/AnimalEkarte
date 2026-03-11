package model

import (
	"time"

	"github.com/google/uuid"
)

// InventoryItem は在庫品目テーブルに対応するモデル
type InventoryItem struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name          string     `gorm:"column:name;not null" json:"name"`
	Category      string     `gorm:"column:category;type:inventory_category;not null" json:"category"`
	Quantity      int        `gorm:"column:quantity;not null;default:0" json:"quantity"`
	Unit          string     `gorm:"column:unit;not null" json:"unit"`
	MinStockLevel int        `gorm:"column:min_stock_level;not null;default:0" json:"min_stock_level"`
	Location      string     `gorm:"column:location;default:''" json:"location"`
	ExpiryDate    *time.Time `gorm:"column:expiry_date;type:date" json:"expiry_date,omitempty"`
	Supplier      string     `gorm:"column:supplier;default:''" json:"supplier"`
	LastRestocked *time.Time `gorm:"column:last_restocked;type:date" json:"last_restocked,omitempty"`
	Status        string     `gorm:"column:status;type:inventory_status;default:'sufficient'" json:"status"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// CreateInventoryItemRequest は在庫作成リクエスト
type CreateInventoryItemRequest struct {
	Name          string  `json:"name" binding:"required"`
	Category      string  `json:"category" binding:"required"`
	Quantity      int     `json:"quantity"`
	Unit          string  `json:"unit" binding:"required"`
	MinStockLevel int     `json:"min_stock_level"`
	Location      string  `json:"location"`
	ExpiryDate    *string `json:"expiry_date"`
	Supplier      string  `json:"supplier"`
	LastRestocked *string `json:"last_restocked"`
	Status        string  `json:"status"`
}

// UpdateInventoryItemRequest は在庫更新リクエスト
type UpdateInventoryItemRequest struct {
	Name          *string `json:"name"`
	Category      *string `json:"category"`
	Quantity      *int    `json:"quantity"`
	Unit          *string `json:"unit"`
	MinStockLevel *int    `json:"min_stock_level"`
	Location      *string `json:"location"`
	ExpiryDate    *string `json:"expiry_date"`
	Supplier      *string `json:"supplier"`
	LastRestocked *string `json:"last_restocked"`
	Status        *string `json:"status"`
}
