package model

import (
	"gorm.io/gorm"
	"time"
)

// Occupation は職種マスタ（クリニックごとにカスタマイズ可能な職種定義）
type Occupation struct {
	ID          uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID    uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Name        string         `gorm:"column:name;type:text;not null;default:''" json:"name"`
	Description string         `gorm:"column:description;type:text;not null;default:''" json:"description"`
	SortOrder   int            `gorm:"column:sort_order;type:integer;not null;default:0" json:"sort_order"`
	IsActive    bool           `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	CreatedAt   time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
}

func (Occupation) TableName() string { return "occupations" }
