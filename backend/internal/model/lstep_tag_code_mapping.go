package model

import (
	"github.com/lib/pq"
	"time"
)

const (
	CodeTypeCheckupType     = "checkup_type"
	CodeTypePrescription    = "prescription"
	CodeTypeMerchandiseItem = "merchandise_item"

	SpeciesScopeDog = "dog"
)

type LstepTagCodeMapping struct {
	ID           uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID     uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	TagName      string         `gorm:"column:tag_name;type:text;not null" json:"tag_name"`
	CodeType     string         `gorm:"column:code_type;type:text;not null" json:"code_type"`
	Codes        pq.StringArray `gorm:"column:codes;type:text[];not null;default:'{}'" json:"codes"`
	SpeciesScope *string        `gorm:"column:species_scope;type:text" json:"species_scope,omitempty"`
	AgeMin       *int           `gorm:"column:age_min;type:int" json:"age_min,omitempty"`
	DeletedAt    *time.Time     `gorm:"column:deleted_at;type:timestamptz" json:"deleted_at,omitempty"`
	CreatedAt    time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
}

func (LstepTagCodeMapping) TableName() string { return "lstep_tag_code_mappings" }
