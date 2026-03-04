package model

import (
	"time"

	"github.com/google/uuid"
)

// MasterItem 診療項目マスタモデル
type MasterItem struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	Code            string     `json:"code" gorm:"type:varchar(20)"`
	Name            string     `json:"name" gorm:"type:varchar(200)"`
	Category        string     `json:"category" gorm:"type:varchar(50)"` // examination, vaccine, medicine, staff, insurance, cage, serviceType, trimming_course, trimming_option
	Price           *float64   `json:"price" gorm:"type:decimal(10,2)"`
	Status          string     `json:"status" gorm:"type:varchar(20);default:'active'"` // active, inactive
	Description     string     `json:"description" gorm:"type:text"`
	InventoryID     *uuid.UUID `json:"inventory_id" gorm:"type:uuid"`
	DefaultQuantity *int       `json:"default_quantity"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Relations
	InventoryItem *InventoryItem `json:"inventory_item,omitempty" gorm:"foreignKey:InventoryID"`
}

// TableName テーブル名を指定
func (MasterItem) TableName() string {
	return "master_items"
}

// InventoryItem 在庫管理モデル
type InventoryItem struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	Name          string     `json:"name" gorm:"type:varchar(200)"`
	Category      string     `json:"category" gorm:"type:varchar(30)"` // medicine, consumable, food, other
	Quantity      int        `json:"quantity" gorm:"default:0"`
	Unit          string     `json:"unit" gorm:"type:varchar(20)"`
	MinStockLevel int        `json:"min_stock_level" gorm:"default:0"`
	Location      string     `json:"location" gorm:"type:varchar(100)"`
	ExpiryDate    *time.Time `json:"expiry_date" gorm:"type:date"`
	Supplier      string     `json:"supplier" gorm:"type:varchar(200)"`
	LastRestocked *time.Time `json:"last_restocked" gorm:"type:date"`
	Status        string     `json:"status" gorm:"type:varchar(20);default:'sufficient'"` // sufficient, low, out_of_stock
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// TableName テーブル名を指定
func (InventoryItem) TableName() string {
	return "inventory_items"
}

// CreateMasterItemRequest マスタアイテム作成リクエスト
type CreateMasterItemRequest struct {
	Code            string   `json:"code" binding:"required"`
	Name            string   `json:"name" binding:"required"`
	Category        string   `json:"category" binding:"required"`
	Price           *float64 `json:"price"`
	Description     string   `json:"description"`
	InventoryID     *string  `json:"inventory_id"`
	DefaultQuantity *int     `json:"default_quantity"`
}

// UpdateMasterItemRequest マスタアイテム更新リクエスト
type UpdateMasterItemRequest struct {
	Code            string   `json:"code"`
	Name            string   `json:"name"`
	Category        string   `json:"category"`
	Price           *float64 `json:"price"`
	Status          string   `json:"status"`
	Description     string   `json:"description"`
	InventoryID     *string  `json:"inventory_id"`
	DefaultQuantity *int     `json:"default_quantity"`
}

// CreateInventoryItemRequest 在庫アイテム作成リクエスト
type CreateInventoryItemRequest struct {
	Name          string     `json:"name" binding:"required"`
	Category      string     `json:"category" binding:"required"`
	Quantity      int        `json:"quantity" binding:"required"`
	Unit          string     `json:"unit" binding:"required"`
	MinStockLevel int        `json:"min_stock_level"`
	Location      string     `json:"location"`
	ExpiryDate    *time.Time `json:"expiry_date"`
	Supplier      string     `json:"supplier"`
	LastRestocked *time.Time `json:"last_restocked"`
}

// UpdateInventoryItemRequest 在庫アイテム更新リクエスト
type UpdateInventoryItemRequest struct {
	Name          string     `json:"name"`
	Category      string     `json:"category"`
	Quantity      int        `json:"quantity"`
	Unit          string     `json:"unit"`
	MinStockLevel int        `json:"min_stock_level"`
	Location      string     `json:"location"`
	ExpiryDate    *time.Time `json:"expiry_date"`
	Supplier      string     `json:"supplier"`
	LastRestocked *time.Time `json:"last_restocked"`
	Status        string     `json:"status"`
}
