package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// MasterItem はマスタデータテーブルに対応するモデル（STIパターン、15カテゴリ統合）
type MasterItem struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Code            string         `gorm:"column:code;not null" json:"code"`
	Name            string         `gorm:"column:name;not null" json:"name"`
	Category        string         `gorm:"column:category;type:master_category;not null" json:"category"`
	Price           *float64       `gorm:"column:price;type:numeric(10,2)" json:"price,omitempty"`
	Status          string         `gorm:"column:status;type:master_item_status;default:'active'" json:"status"`
	Description     string         `gorm:"column:description;default:''" json:"description"`
	InventoryID     *uuid.UUID     `gorm:"column:inventory_id;type:uuid" json:"inventory_id,omitempty"`
	DefaultQuantity *int           `gorm:"column:default_quantity" json:"default_quantity,omitempty"`
	Species         *string        `gorm:"column:species;type:vaccine_species" json:"species,omitempty"`
	Interval        *string        `gorm:"column:interval" json:"interval,omitempty"`
	ParentID        *uuid.UUID     `gorm:"column:parent_id;type:uuid" json:"parent_id,omitempty"`
	SortOrder       int            `gorm:"column:sort_order;default:0" json:"sort_order"`
	// カテゴリ固有フィールド
	Color          *string        `gorm:"column:color" json:"color,omitempty"`
	TimeCondition  *string        `gorm:"column:time_condition" json:"time_condition,omitempty"`
	Anesthesia     *string        `gorm:"column:anesthesia" json:"anesthesia,omitempty"`
	TargetAge      *string        `gorm:"column:target_age" json:"target_age,omitempty"`
	DosageForm     *string        `gorm:"column:dosage_form;type:dosage_form" json:"dosage_form,omitempty"`
	MedicineUnit   *string        `gorm:"column:medicine_unit;type:medicine_unit" json:"medicine_unit,omitempty"`
	StaffRole      *string        `gorm:"column:staff_role;type:staff_role" json:"staff_role,omitempty"`
	LicenseNumber  *string        `gorm:"column:license_number" json:"license_number,omitempty"`
	Email          *string        `gorm:"column:email" json:"email,omitempty"`
	UserType       *string        `gorm:"column:user_type" json:"user_type,omitempty"`
	Clinics        pq.StringArray `gorm:"column:clinics;type:text[]" json:"clinics,omitempty" swaggertype:"array,string"`
	LastLoginAt    *time.Time     `gorm:"column:last_login_at" json:"last_login_at,omitempty"`
	CageType       *string        `gorm:"column:cage_type;type:cage_type" json:"cage_type,omitempty"`
	CageSize       *string        `gorm:"column:cage_size;type:cage_size" json:"cage_size,omitempty"`
	CoverageRate   *string        `gorm:"column:coverage_rate;type:coverage_rate" json:"coverage_rate,omitempty"`
	ContactPhone   *string        `gorm:"column:contact_phone" json:"contact_phone,omitempty"`
	TargetSize     *string        `gorm:"column:target_size;type:target_size" json:"target_size,omitempty"`
	Duration       *string        `gorm:"column:duration" json:"duration,omitempty"`
	Combinable     *string        `gorm:"column:combinable;type:combinable" json:"combinable,omitempty"`
	BodySize       *string        `gorm:"column:body_size;type:body_size" json:"body_size,omitempty"`
	BillingUnit    *string        `gorm:"column:billing_unit;type:billing_unit" json:"billing_unit,omitempty"`
	CreatedAt      time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// Relations
	Children    []MasterItem           `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Inspections []MasterItemInspection `gorm:"foreignKey:MasterItemID" json:"inspections,omitempty"`
	Inventory   *InventoryItem         `gorm:"foreignKey:InventoryID;references:ID" json:"inventory,omitempty"`
}

// MasterItemInspection は検査マスタの検査項目定義テーブルに対応するモデル
type MasterItemInspection struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MasterItemID    uuid.UUID `gorm:"column:master_item_id;type:uuid;not null" json:"master_item_id"`
	Name            string    `gorm:"column:name;not null" json:"name"`
	InspectionValue string    `gorm:"column:inspection_value;default:''" json:"inspection_value"`
	NormalValue     string    `gorm:"column:normal_value;default:''" json:"normal_value"`
	SortOrder       int       `gorm:"column:sort_order;default:0" json:"sort_order"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// CreateMasterItemRequest はマスタ作成リクエスト
type CreateMasterItemRequest struct {
	Code            string   `json:"code" binding:"required"`
	Name            string   `json:"name" binding:"required"`
	Category        string   `json:"category" binding:"required"`
	Price           *float64 `json:"price"`
	Status          string   `json:"status"`
	Description     string   `json:"description"`
	InventoryID     *string  `json:"inventory_id"`
	DefaultQuantity *int     `json:"default_quantity"`
	Species         *string  `json:"species"`
	Interval        *string  `json:"interval"`
	ParentID        *string  `json:"parent_id"`
	SortOrder       int      `json:"sort_order"`
	// 拡張フィールド
	Color         *string `json:"color"`
	TimeCondition *string `json:"time_condition"`
	Anesthesia    *string `json:"anesthesia"`
	TargetAge     *string `json:"target_age"`
	DosageForm    *string `json:"dosage_form"`
	MedicineUnit  *string `json:"medicine_unit"`
	StaffRole     *string `json:"staff_role"`
	LicenseNumber *string `json:"license_number"`
	Email         *string `json:"email"`
	UserType      *string `json:"user_type"`
	CageType      *string `json:"cage_type"`
	CageSize      *string `json:"cage_size"`
	CoverageRate  *string `json:"coverage_rate"`
	ContactPhone  *string `json:"contact_phone"`
	TargetSize    *string `json:"target_size"`
	Duration      *string `json:"duration"`
	Combinable    *string `json:"combinable"`
	BodySize      *string `json:"body_size"`
	BillingUnit   *string `json:"billing_unit"`
}

// UpdateMasterItemRequest はマスタ更新リクエスト
type UpdateMasterItemRequest struct {
	Code            *string  `json:"code"`
	Name            *string  `json:"name"`
	Category        *string  `json:"category"`
	Price           *float64 `json:"price"`
	Status          *string  `json:"status"`
	Description     *string  `json:"description"`
	InventoryID     *string  `json:"inventory_id"`
	DefaultQuantity *int     `json:"default_quantity"`
	Species         *string  `json:"species"`
	Interval        *string  `json:"interval"`
	ParentID        *string  `json:"parent_id"`
	SortOrder       *int     `json:"sort_order"`
	// 拡張フィールド
	Color         *string `json:"color"`
	TimeCondition *string `json:"time_condition"`
	Anesthesia    *string `json:"anesthesia"`
	TargetAge     *string `json:"target_age"`
	DosageForm    *string `json:"dosage_form"`
	MedicineUnit  *string `json:"medicine_unit"`
	StaffRole     *string `json:"staff_role"`
	LicenseNumber *string `json:"license_number"`
	Email         *string `json:"email"`
	UserType      *string `json:"user_type"`
	CageType      *string `json:"cage_type"`
	CageSize      *string `json:"cage_size"`
	CoverageRate  *string `json:"coverage_rate"`
	ContactPhone  *string `json:"contact_phone"`
	TargetSize    *string `json:"target_size"`
	Duration      *string `json:"duration"`
	Combinable    *string `json:"combinable"`
	BodySize      *string `json:"body_size"`
	BillingUnit   *string `json:"billing_unit"`
}
