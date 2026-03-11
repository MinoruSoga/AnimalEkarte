package model

import (
	"time"

	"github.com/google/uuid"
)

type TargetSize string

const (
	TargetSizeSmall  TargetSize = "small"
	TargetSizeMedium TargetSize = "medium"
	TargetSizeLarge  TargetSize = "large"
	TargetSizeCat    TargetSize = "cat"
)

type Combinable string

const (
	CombinableYes Combinable = "yes"
	CombinableNo  Combinable = "no"
)

type TrimmingCourse struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Code        string       `gorm:"default:''"                                     json:"code"`
	Name        string       `gorm:"not null"                                       json:"name"`
	Price       *float64     `gorm:"type:numeric(10,2)"                             json:"price,omitempty"`
	Status      MasterStatus `gorm:"type:master_status;default:'active'"            json:"status"`
	Description string       `gorm:"default:''"                                     json:"description"`
	TargetSize  *TargetSize  `gorm:"type:target_size"                               json:"target_size,omitempty"`
	Duration    string       `gorm:"default:''"                                     json:"duration"`
	SortOrder   int          `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time    `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (TrimmingCourse) TableName() string { return "trimming_courses" }

type TrimmingOption struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Code        string       `gorm:"default:''"                                     json:"code"`
	Name        string       `gorm:"not null"                                       json:"name"`
	Price       *float64     `gorm:"type:numeric(10,2)"                             json:"price,omitempty"`
	Status      MasterStatus `gorm:"type:master_status;default:'active'"            json:"status"`
	Description string       `gorm:"default:''"                                     json:"description"`
	Duration    string       `gorm:"default:''"                                     json:"duration"`
	Combinable  Combinable   `gorm:"type:combinable;default:'yes'"                  json:"combinable"`
	SortOrder   int          `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time    `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (TrimmingOption) TableName() string { return "trimming_options" }
