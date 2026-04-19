package model

import (
	"time"

	"gorm.io/gorm"
)

type TargetSize string

const (
	TargetSizeSmall  TargetSize = "small"
	TargetSizeMedium TargetSize = "medium"
	TargetSizeLarge  TargetSize = "large"
	TargetSizeCat    TargetSize = "cat"
)

type TrimmingCourse struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64         `gorm:"not null"                                       json:"clinic_id"`
	Name        string         `gorm:"not null"                                       json:"name"`
	Price       *int64         `gorm:"type:bigint"                                    json:"price,omitempty"`
	IsActive    bool           `gorm:"default:true"                                   json:"is_active"`
	Description string         `gorm:"default:''"                                     json:"description"`
	TargetSize  *TargetSize    `gorm:"type:target_size"                               json:"target_size,omitempty"`
	Duration    *int           `gorm:"type:integer"                                   json:"duration,omitempty"`
	SortOrder   int            `gorm:"type:integer;default:0"                         json:"sort_order"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt   gorm.DeletedAt `                                                      json:"-"`
}

func (TrimmingCourse) TableName() string { return "trimming_courses" }

type TrimmingOption struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID     uint64         `gorm:"not null"                                       json:"clinic_id"`
	Name         string         `gorm:"not null"                                       json:"name"`
	Price        *int64         `gorm:"type:bigint"                                    json:"price,omitempty"`
	IsActive     bool           `gorm:"default:true"                                   json:"is_active"`
	Description  string         `gorm:"default:''"                                     json:"description"`
	Duration     *int           `gorm:"type:integer"                                   json:"duration,omitempty"`
	IsCombinable bool           `gorm:"column:is_combinable;default:true"              json:"is_combinable"`
	SortOrder    int            `gorm:"type:integer;default:0"                         json:"sort_order"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt    gorm.DeletedAt `                                                      json:"-"`
}

func (TrimmingOption) TableName() string { return "trimming_options" }
