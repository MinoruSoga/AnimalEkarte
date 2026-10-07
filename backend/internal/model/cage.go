package model

import (
	"time"

	"gorm.io/gorm"
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
	ID          uint64   `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64   `gorm:"not null"                                       json:"clinic_id"`
	Name        string   `gorm:"not null"                                       json:"name"`
	Price       *int64   `gorm:"type:bigint"                                    json:"price,omitempty"`
	IsActive    bool     `gorm:"default:true"                                   json:"is_active"`
	Description string   `gorm:"default:''"                                     json:"description"`
	CageType    CageType `gorm:"type:cage_type;not null"                        json:"cage_type"`
	CageSize    CageSize `gorm:"type:cage_size;not null"                        json:"cage_size"`
	SortOrder   int      `gorm:"type:integer;default:0"                         json:"sort_order"`
	// MergedIntoCageID は EMR-306 ケージ一時連結の構成ケージが参照する代表ケージID。
	// NULL = 独立ケージまたは代表ケージ自身（連結なし）。
	MergedIntoCageID *uint64        `gorm:"column:merged_into_cage_id"                     json:"merged_into_cage_id,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt        gorm.DeletedAt `                                                      json:"-"`
}

func (Cage) TableName() string { return "cages" }
