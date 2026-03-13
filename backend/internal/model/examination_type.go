package model

import (
	"time"
)

type ExamType struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64    `gorm:"not null"                                       json:"clinic_id"`
	Name        string    `gorm:"not null"                                       json:"name"`
	Price       *float64  `gorm:"type:numeric(10,2)"                             json:"price,omitempty"`
	IsActive    bool      `gorm:"default:true"                                   json:"is_active"`
	Description string    `gorm:"default:''"                                     json:"description"`
	SortOrder   int       `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Items []ExamTypeItem `gorm:"foreignKey:ExamTypeID" json:"items,omitempty"`
}

func (ExamType) TableName() string { return "exam_types" }

type ExamTypeItem struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ExamTypeID      uint64    `gorm:"not null"                                       json:"exam_type_id"`
	Name            string    `gorm:"not null"                                       json:"name"`
	InspectionValue string    `gorm:"default:''"                                     json:"inspection_value"`
	NormalValue     string    `gorm:"default:''"                                     json:"normal_value"`
	SortOrder       int       `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt       time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
}

func (ExamTypeItem) TableName() string { return "exam_type_items" }
