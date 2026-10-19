package model

import (
	"gorm.io/gorm"
	"time"
)

type CageType string

const (
	CageTypeICU     CageType = "icu"
	CageTypeDog     CageType = "dog"
	CageTypeCat     CageType = "cat"
	CageTypeGeneral CageType = "general"
)

type CageSize string

const (
	CageSizeSmall  CageSize = "small"
	CageSizeMedium CageSize = "medium"
	CageSizeLarge  CageSize = "large"
)

type Cage struct {
	ID               uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID         uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Name             string         `gorm:"column:name;type:text;not null" json:"name"`
	Price            *int64         `gorm:"column:price;type:bigint" json:"price,omitempty"`
	IsActive         bool           `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	Description      string         `gorm:"column:description;type:text;not null;default:''" json:"description"`
	CageType         CageType       `gorm:"column:cage_type;type:cage_type;not null" json:"cage_type"`
	CageSize         CageSize       `gorm:"column:cage_size;type:cage_size;not null" json:"cage_size"`
	SortOrder        *int           `gorm:"column:sort_order;type:integer;default:0" json:"sort_order"`
	CreatedAt        time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	MergedIntoCageID *uint64        `gorm:"column:merged_into_cage_id;type:bigint" json:"merged_into_cage_id,omitempty"`
}

func (Cage) TableName() string { return "cages" }
