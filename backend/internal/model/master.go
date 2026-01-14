package model

import (
	"time"

	"github.com/google/uuid"
)

// MasterItem 診療項目マスタモデル
type MasterItem struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
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
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
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
