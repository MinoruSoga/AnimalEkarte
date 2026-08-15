package model

import (
	"time"

	"gorm.io/gorm"
)

type ExaminationType struct {
	ID             uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID       uint64         `gorm:"not null"                                       json:"clinic_id"`
	Name           string         `gorm:"not null"                                       json:"name"`
	Price          *int64         `gorm:"type:bigint"                                    json:"price,omitempty"`
	IsActive       bool           `gorm:"default:true"                                   json:"is_active"`
	Description    string         `gorm:"default:''"                                     json:"description"`
	ParentID       *uint64        `gorm:"column:parent_id"                               json:"parent_id,omitempty"`
	SortOrder      int            `gorm:"type:integer;default:0"                         json:"sort_order"`
	IsNonInsurance bool           `gorm:"not null;default:false"                         json:"is_non_insurance"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt      gorm.DeletedAt `                                                      json:"-"`

	// Relations
	Items []ExamTypeField `gorm:"foreignKey:ExamTypeID" json:"items,omitempty"`
}

func (ExaminationType) TableName() string { return "exam_types" }

type ExamTypeField struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ExamTypeID      uint64    `gorm:"not null"                                       json:"exam_type_id"`
	ClinicID        uint64    `gorm:"not null"                                       json:"-"`
	Name            string    `gorm:"not null"                                       json:"name"`
	InspectionValue string    `gorm:"default:''"                                     json:"inspection_value"`
	NormalValue     string    `gorm:"default:''"                                     json:"normal_value"`
	Unit            string    `gorm:"not null;default:''"                            json:"unit"`
	SortOrder       int       `gorm:"type:integer;default:0"                         json:"sort_order"`
	CreatedAt       time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (ExamTypeField) TableName() string { return "exam_type_fields" }
