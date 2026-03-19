package model

import (
	"time"
)

type ExaminationType struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64    `gorm:"not null"                                       json:"clinic_id"`
	Name        string    `gorm:"not null"                                       json:"name"`
	Price       *int64    `gorm:"type:bigint"                                    json:"price,omitempty"`
	IsActive    bool      `gorm:"default:true"                                   json:"is_active"`
	Description string    `gorm:"default:''"                                     json:"description"`
	ParentID    *uint64   `gorm:"column:parent_id"                               json:"parent_id,omitempty"`
	SortOrder   int       `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Items []ExaminationTypeItem `gorm:"foreignKey:ExamTypeID" json:"items,omitempty"`
}

func (ExaminationType) TableName() string { return "exam_types" }

type ExaminationTypeItem struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ExamTypeID      uint64    `gorm:"not null"                                       json:"exam_type_id"`
	Name            string    `gorm:"not null"                                       json:"name"`
	InspectionValue string    `gorm:"default:''"                                     json:"inspection_value"`
	NormalValue     string    `gorm:"default:''"                                     json:"normal_value"`
	Unit            string    `gorm:"not null;default:''"                            json:"unit"`
	SortOrder       int       `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt       time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (ExaminationTypeItem) TableName() string { return "exam_type_items" }
