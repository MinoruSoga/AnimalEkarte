package model

import (
	"time"
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
	Price       *float64        `gorm:"type:numeric(10,2)"                             json:"price,omitempty"`
	IsActive    bool            `gorm:"default:true"                                   json:"is_active"`
	Description string          `gorm:"default:''"                                     json:"description"`
	Species     *VaccineSpecies `gorm:"type:vaccine_species"                           json:"species,omitempty"`
	Interval    string          `gorm:"default:''"                                     json:"interval"`
	InventoryID *uint64         `                                                      json:"inventory_id,omitempty"`
	SortOrder   int             `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time       `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (Vaccine) TableName() string { return "vaccines" }
