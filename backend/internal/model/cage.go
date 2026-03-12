package model

import (
	"time"

	"github.com/google/uuid"
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
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicID    uuid.UUID    `gorm:"type:uuid;not null"                             json:"clinic_id"`
	Code        string       `gorm:"default:''"                                     json:"code"`
	Name        string       `gorm:"not null"                                       json:"name"`
	Price       *float64     `gorm:"type:numeric(10,2)"                             json:"price,omitempty"`
	IsActive    bool         `gorm:"default:true"                                   json:"is_active"`
	Description string       `gorm:"default:''"                                     json:"description"`
	CageType    CageType     `gorm:"type:cage_type;not null"                        json:"cage_type"`
	CageSize    CageSize     `gorm:"type:cage_size;not null"                        json:"cage_size"`
	SortOrder   int          `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time    `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (Cage) TableName() string { return "cages" }
