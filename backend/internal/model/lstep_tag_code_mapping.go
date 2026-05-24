package model

import (
	"time"

	"github.com/lib/pq"
)

const (
	CodeTypeCheckupType     = "checkup_type"
	CodeTypePrescription    = "prescription"
	CodeTypeMerchandiseItem = "merchandise_item"

	SpeciesScopeDog = "dog"
)

type LstepTagCodeMapping struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement"                        json:"id"`
	ClinicID     uint64         `gorm:"not null"                                        json:"clinic_id"`
	TagName      string         `gorm:"column:tag_name;not null"                        json:"tag_name"`
	CodeType     string         `gorm:"column:code_type;not null"                       json:"code_type"`
	Codes        pq.StringArray `gorm:"column:codes;type:text[];not null;default:'{}'"  json:"codes"`
	SpeciesScope *string        `gorm:"column:species_scope"                            json:"species_scope,omitempty"`
	AgeMin       *int           `gorm:"column:age_min"                                  json:"age_min,omitempty"`
	DeletedAt    *time.Time     `gorm:"column:deleted_at"                               json:"deleted_at,omitempty"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"                                  json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"                                  json:"updated_at"`
}

func (LstepTagCodeMapping) TableName() string { return "lstep_tag_code_mappings" }
