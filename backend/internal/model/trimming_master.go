package model

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
	"time"
)

type TargetSize string

const (
	TargetSizeSmall  TargetSize = "small"
	TargetSizeMedium TargetSize = "medium"
	TargetSizeLarge  TargetSize = "large"
	TargetSizeCat    TargetSize = "cat"
)

type TrimmingCourse struct {
	ID              uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID        uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Name            string         `gorm:"column:name;type:text;not null" json:"name"`
	Price           *int64         `gorm:"column:price;type:bigint" json:"price,omitempty"`
	IsActive        bool           `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	Description     string         `gorm:"column:description;type:text;not null;default:''" json:"description"`
	TargetSize      *TargetSize    `gorm:"column:target_size;type:target_size" json:"target_size,omitempty"`
	Duration        *int           `gorm:"column:duration;type:integer" json:"duration,omitempty"`
	SortOrder       *int           `gorm:"column:sort_order;type:integer;default:0" json:"sort_order"`
	CourseTypeID    *uint64        `gorm:"column:course_type_id;type:bigint" json:"course_type_id,omitempty"`
	CreatedAt       time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	ShortName       *string        `gorm:"column:short_name;type:text" json:"-"`
	ImageObjectKey  *string        `gorm:"column:image_object_key;type:text" json:"-"`
	AllowedWeekdays *pq.Int64Array `gorm:"column:allowed_weekdays;type:smallint[]" json:"-"`
	ContractVersion int64          `gorm:"column:contract_version;type:bigint;not null;default:1" json:"-"`
}

func (TrimmingCourse) TableName() string { return "trimming_courses" }

type TrimmingOption struct {
	ID                uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID          uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Name              string         `gorm:"column:name;type:text;not null" json:"name"`
	Price             *int64         `gorm:"column:price;type:bigint" json:"price,omitempty"`
	IsActive          bool           `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	Description       string         `gorm:"column:description;type:text;not null;default:''" json:"description"`
	Duration          *int           `gorm:"column:duration;type:integer" json:"duration,omitempty"`
	IsCombinable      bool           `gorm:"column:is_combinable;type:boolean;not null;default:true" json:"is_combinable"`
	SortOrder         *int           `gorm:"column:sort_order;type:integer;default:0" json:"sort_order"`
	CreatedAt         time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	ContractVersion   int64          `gorm:"column:contract_version;type:bigint;not null;default:1" json:"-"`
	ServiceIdentityID *uint64        `gorm:"column:service_identity_id;type:bigint" json:"-"`
}

func (TrimmingOption) TableName() string { return "trimming_options" }
