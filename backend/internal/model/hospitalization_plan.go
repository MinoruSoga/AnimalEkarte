package model

import (
	"time"
)

// @name BodySize
type BodySize string

const (
	BodySizeSmall  BodySize = "small"
	BodySizeMedium BodySize = "medium"
	BodySizeLarge  BodySize = "large"
)

// @name BillingUnit
type BillingUnit string

const (
	BillingUnitPerDay   BillingUnit = "per_day"
	BillingUnitPerNight BillingUnit = "per_night"
)

// @name HospitalizationPlan
type HospitalizationPlan struct {
	ID          uint64       `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64       `gorm:"not null"                                       json:"clinic_id"`
	Name        string       `gorm:"not null"                                       json:"name"`
	Price       *float64     `gorm:"type:numeric(10,2)"                             json:"price,omitempty"`
	IsActive    bool         `gorm:"default:true"                                   json:"is_active"`
	Description string       `gorm:"default:''"                                     json:"description"`
	BodySize    *BodySize    `gorm:"type:body_size"                                 json:"body_size,omitempty"`
	BillingUnit *BillingUnit `gorm:"type:billing_unit"                              json:"billing_unit,omitempty"`
	SortOrder   int          `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time    `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (HospitalizationPlan) TableName() string { return "hospitalization_plans" }
