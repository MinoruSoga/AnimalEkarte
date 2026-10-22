package model

import (
	"gorm.io/gorm"
	"time"
)

type VaccineSpecies string

const (
	VaccineSpeciesDog  VaccineSpecies = "dog"
	VaccineSpeciesCat  VaccineSpecies = "cat"
	VaccineSpeciesBoth VaccineSpecies = "both"
)

type Vaccine struct {
	Version        int64           `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
	ID             uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID       uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Name           string          `gorm:"column:name;type:text;not null" json:"name"`
	Price          *int64          `gorm:"column:price;type:bigint" json:"price,omitempty"`
	IsActive       bool            `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	Description    string          `gorm:"column:description;type:text;not null;default:''" json:"description"`
	Species        *VaccineSpecies `gorm:"column:species;type:vaccine_species" json:"species,omitempty"`
	Interval       string          `gorm:"column:interval;type:text;not null;default:''" json:"interval"`
	InventoryID    *uint64         `gorm:"column:inventory_id;type:bigint" json:"inventory_id,omitempty"`
	ParentID       *uint64         `gorm:"column:parent_id;type:bigint" json:"parent_id,omitempty"`
	SortOrder      *int            `gorm:"column:sort_order;type:integer;default:0" json:"sort_order"`
	CreatedAt      time.Time       `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time       `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	IsNonInsurance *bool           `gorm:"column:is_non_insurance;type:boolean" json:"-"`

	IsBillDiscountExcluded *bool `gorm:"column:is_bill_discount_excluded;type:boolean" json:"-"`
}

func (Vaccine) TableName() string { return "vaccines" }
