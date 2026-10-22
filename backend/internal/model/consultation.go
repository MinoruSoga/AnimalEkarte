package model

import (
	"gorm.io/gorm"
	"time"
)

type Consultation struct {
	Version        int64          `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
	ID             uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID       uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Name           string         `gorm:"column:name;type:text;not null" json:"name"`
	Price          *int64         `gorm:"column:price;type:bigint" json:"price,omitempty"`
	IsActive       bool           `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	Description    string         `gorm:"column:description;type:text;not null;default:''" json:"description"`
	TimeCondition  string         `gorm:"column:time_condition;type:text;not null;default:''" json:"time_condition"`
	Duration       *int           `gorm:"column:duration;type:integer" json:"duration,omitempty"`
	ParentID       *uint64        `gorm:"column:parent_id;type:bigint" json:"parent_id,omitempty"`
	TaxType        TaxType        `gorm:"column:tax_type;type:tax_type;not null;default:'excluded'" json:"tax_type"`
	TaxRate        ExactNumeric   `gorm:"column:tax_rate;type:numeric;not null;default:0.10" json:"tax_rate"`
	SortOrder      *int           `gorm:"column:sort_order;type:integer;default:0" json:"sort_order"`
	CreatedAt      time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	IsNonInsurance *bool          `gorm:"column:is_non_insurance;type:boolean" json:"-"`

	IsBillDiscountExcluded *bool `gorm:"column:is_bill_discount_excluded;type:boolean" json:"-"`
}

func (Consultation) TableName() string { return "consultations" }
