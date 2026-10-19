package model

import (
	"gorm.io/gorm"
	"time"
)

type BodySize string

const (
	BodySizeSmall  BodySize = "small"
	BodySizeMedium BodySize = "medium"
	BodySizeLarge  BodySize = "large"
)

type BillingUnit string

const (
	BillingUnitPerDay   BillingUnit = "per_day"
	BillingUnitPerNight BillingUnit = "per_night"
)

type HospitalizationPlan struct {
	ID          uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID    uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Name        string         `gorm:"column:name;type:text;not null" json:"name"`
	Price       *int64         `gorm:"column:price;type:bigint" json:"price,omitempty"`
	IsActive    bool           `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	Description string         `gorm:"column:description;type:text;not null;default:''" json:"description"`
	BodySize    *BodySize      `gorm:"column:body_size;type:body_size" json:"body_size,omitempty"`
	BillingUnit *BillingUnit   `gorm:"column:billing_unit;type:billing_unit;default:'per_day'" json:"billing_unit,omitempty"`
	TaxType     TaxType        `gorm:"column:tax_type;type:tax_type;not null;default:'excluded'" json:"tax_type"`
	TaxRate     ExactNumeric   `gorm:"column:tax_rate;type:numeric;not null;default:0.10" json:"tax_rate"`
	SortOrder   *int           `gorm:"column:sort_order;type:integer;default:0" json:"sort_order"`
	CreatedAt   time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
}

func (HospitalizationPlan) TableName() string { return "hospitalization_plans" }
