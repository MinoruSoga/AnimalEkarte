package model

import (
	"time"
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
	ID              uint64        `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID        uint64        `gorm:"not null"                                       json:"clinic_id"`
	Name            string        `gorm:"not null"                                       json:"name"`
	ParentID        *uint64       `gorm:"column:parent_id"                               json:"parent_id,omitempty"`
	Price           *int64        `gorm:"type:bigint"                                    json:"price,omitempty"`
	IsActive        bool          `gorm:"default:true"                                   json:"is_active"`
	Description     string        `gorm:"default:''"                                     json:"description"`
	DosageForm      *DosageForm   `gorm:"type:dosage_form"                               json:"dosage_form,omitempty"`
	MedicineUnit    *MedicineUnit `gorm:"type:medicine_unit"                             json:"medicine_unit,omitempty"`
	InventoryID     *uint64       `                                                      json:"inventory_id,omitempty"`
	DefaultQuantity float64       `gorm:"type:numeric(10,1);default:1"                   json:"default_quantity"`
	SortOrder       int           `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt       time.Time     `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time     `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Inventory *InventoryItem `gorm:"foreignKey:InventoryID" json:"inventory,omitempty"`
}

func (Medicine) TableName() string { return "medicines" }
