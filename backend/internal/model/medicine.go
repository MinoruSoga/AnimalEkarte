package model

import (
	"time"

	"github.com/google/uuid"
)

type DosageForm string

const (
	DosageFormTablet    DosageForm = "tablet"
	DosageFormLiquid    DosageForm = "liquid"
	DosageFormInjection DosageForm = "injection"
	DosageFormTopical   DosageForm = "topical"
	DosageFormPowder    DosageForm = "powder"
)

type MedicineUnit string

const (
	MedicineUnitPerTablet MedicineUnit = "per_tablet"
	MedicineUnitPerML     MedicineUnit = "per_ml"
	MedicineUnitPerDose   MedicineUnit = "per_dose"
	MedicineUnitPerGram   MedicineUnit = "per_gram"
)

type Medicine struct {
	ID              uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicID        uuid.UUID     `gorm:"type:uuid;not null"                             json:"clinic_id"`
	Code            string        `gorm:"default:''"                                     json:"code"`
	Name            string        `gorm:"not null"                                       json:"name"`
	Price           *float64      `gorm:"type:numeric(10,2)"                             json:"price,omitempty"`
	IsActive        bool          `gorm:"default:true"                                   json:"is_active"`
	Description     string        `gorm:"default:''"                                     json:"description"`
	DosageForm      *DosageForm   `gorm:"type:dosage_form"                               json:"dosage_form,omitempty"`
	MedicineUnit    *MedicineUnit `gorm:"type:medicine_unit"                             json:"medicine_unit,omitempty"`
	InventoryID     *uuid.UUID    `gorm:"type:uuid"                                      json:"inventory_id,omitempty"`
	DefaultQuantity int           `gorm:"default:1"                                      json:"default_quantity"`
	SortOrder       int           `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt       time.Time     `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time     `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Inventory *InventoryItem `gorm:"foreignKey:InventoryID" json:"inventory,omitempty"`
}

func (Medicine) TableName() string { return "medicines" }
