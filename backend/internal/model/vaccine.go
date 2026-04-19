package model

import (
	"time"

	"gorm.io/gorm"
)

type VaccineSpecies string

const (
	VaccineSpeciesDog  VaccineSpecies = "dog"
	VaccineSpeciesCat  VaccineSpecies = "cat"
	VaccineSpeciesBoth VaccineSpecies = "both"
)

type Vaccine struct {
	ID          uint64          `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64          `gorm:"not null"                                       json:"clinic_id"`
	Name        string          `gorm:"not null"                                       json:"name"`
	Price       *int64          `gorm:"type:bigint"                                    json:"price,omitempty"`
	IsActive    bool            `gorm:"default:true"                                   json:"is_active"`
	Description string          `gorm:"default:''"                                     json:"description"`
	Species     *VaccineSpecies `gorm:"type:vaccine_species"                           json:"species,omitempty"`
	Interval    string          `gorm:"default:''"                                     json:"interval"`
	InventoryID *uint64         `                                                      json:"inventory_id,omitempty"`
	ParentID    *uint64         `gorm:"column:parent_id"                               json:"parent_id,omitempty"`
	SortOrder   int             `gorm:"type:integer;default:0"                         json:"sort_order"`
	CreatedAt   time.Time       `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `                                                      json:"-"`
}

func (Vaccine) TableName() string { return "vaccines" }
