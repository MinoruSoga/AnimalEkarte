package model

import (
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
	ID          uint64      `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64      `gorm:"not null"                                       json:"clinic_id"`
	Name        string      `gorm:"not null"                                       json:"name"`
	Price       *float64    `gorm:"type:numeric(10,2)"                             json:"price,omitempty"`
	IsActive    bool        `gorm:"default:true"                                   json:"is_active"`
	Description string      `gorm:"default:''"                                     json:"description"`
	TargetSize  *TargetSize `gorm:"type:target_size"                               json:"target_size,omitempty"`
	Duration    string      `gorm:"default:''"                                     json:"duration"`
	SortOrder   int         `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time   `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time   `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (TrimmingCourse) TableName() string { return "trimming_courses" }
type TrimmingOption struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64    `gorm:"not null"                                       json:"clinic_id"`
	Name        string    `gorm:"not null"                                       json:"name"`
	Price       *float64  `gorm:"type:numeric(10,2)"                             json:"price,omitempty"`
	IsActive    bool      `gorm:"default:true"                                   json:"is_active"`
	Description string    `gorm:"default:''"                                     json:"description"`
	Duration    string    `gorm:"default:''"                                     json:"duration"`
	Combinable  bool      `gorm:"default:true"                                   json:"combinable"`
	SortOrder   int       `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (TrimmingOption) TableName() string { return "trimming_options" }
