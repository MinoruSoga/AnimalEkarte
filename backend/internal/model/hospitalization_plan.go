package model

import (
	"time"

	"github.com/google/uuid"
)

type BodySize string

const (
	BodySizeSmall  BodySize = "small"
	BodySizeMedium BodySize = "medium"
	BodySizeLarge  BodySize = "large"
)

type BillingUnit string

const (
	BillingUnitPerDay   BillingUnit = "per_day"
	BillingUnitPerNight BillingUnit = "per_night"
)

type HospitalizationPlan struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicID    uuid.UUID    `gorm:"type:uuid;not null"                             json:"clinic_id"`
	Code        string       `gorm:"default:''"                                     json:"code"`
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
