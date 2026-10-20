package model

import (
	"gorm.io/gorm"
	"time"
)

type ExaminationType struct {
	ID             uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID       uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Name           string          `gorm:"column:name;type:text;not null" json:"name"`
	Price          *int64          `gorm:"column:price;type:bigint" json:"price,omitempty"`
	IsActive       bool            `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	Description    string          `gorm:"column:description;type:text;not null;default:''" json:"description"`
	ParentID       *uint64         `gorm:"column:parent_id;type:bigint" json:"parent_id,omitempty"`
	SortOrder      *int            `gorm:"column:sort_order;type:integer;default:0" json:"sort_order"`
	IsNonInsurance bool            `gorm:"column:is_non_insurance;type:boolean;not null;default:false" json:"is_non_insurance"`
	CreatedAt      time.Time       `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time       `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	Items          []ExamTypeField `gorm:"foreignKey:ExamTypeID,ClinicID;references:ID,ClinicID" json:"items,omitempty"`
}

func (ExaminationType) TableName() string { return "exam_types" }

type ExamTypeField struct {
	ID              uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ExamTypeID      uint64    `gorm:"column:exam_type_id;type:bigint;not null" json:"exam_type_id"`
	Name            string    `gorm:"column:name;type:text;not null" json:"name"`
	InspectionValue string    `gorm:"column:inspection_value;type:text;not null;default:''" json:"inspection_value"`
	NormalValue     string    `gorm:"column:normal_value;type:text;not null;default:''" json:"normal_value"`
	Unit            string    `gorm:"column:unit;type:text;not null;default:''" json:"unit"`
	SortOrder       *int      `gorm:"column:sort_order;type:integer;default:0" json:"sort_order"`
	CreatedAt       time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	ClinicID        uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
}

func (ExamTypeField) TableName() string { return "exam_type_fields" }
