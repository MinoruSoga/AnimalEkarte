package model

import (
	"time"

	"github.com/google/uuid"
)

type VaccineSpecies string

const (
	VaccineSpeciesDog  VaccineSpecies = "dog"
	VaccineSpeciesCat  VaccineSpecies = "cat"
	VaccineSpeciesBoth VaccineSpecies = "both"
)

type Vaccine struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicID    uuid.UUID       `gorm:"type:uuid;not null"                             json:"clinic_id"`
	Code        string          `gorm:"default:''"                                     json:"code"`
	Name        string          `gorm:"not null"                                       json:"name"`
	Price       *float64        `gorm:"type:numeric(10,2)"                             json:"price,omitempty"`
	IsActive    bool            `gorm:"default:true"                                   json:"is_active"`
	Description string          `gorm:"default:''"                                     json:"description"`
	Species     *VaccineSpecies `gorm:"type:vaccine_species"                           json:"species,omitempty"`
	Interval    string          `gorm:"default:''"                                     json:"interval"`
	SortOrder   int             `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time       `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (Vaccine) TableName() string { return "vaccines" }
