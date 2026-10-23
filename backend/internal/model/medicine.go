package model

import (
	"gorm.io/gorm"
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
	Version             int64                   `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
	ID                  uint64                  `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID            uint64                  `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Name                string                  `gorm:"column:name;type:text;not null" json:"name"`
	Price               *int64                  `gorm:"column:price;type:bigint" json:"price,omitempty"`
	IsActive            bool                    `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	Description         string                  `gorm:"column:description;type:text;not null;default:''" json:"description"`
	ParentID            *uint64                 `gorm:"column:parent_id;type:bigint" json:"parent_id,omitempty"`
	DosageForm          *DosageForm             `gorm:"column:dosage_form;type:dosage_form" json:"dosage_form,omitempty"`
	MedicineUnit        *MedicineUnit           `gorm:"column:medicine_unit;type:medicine_unit" json:"medicine_unit,omitempty"`
	InventoryID         *uint64                 `gorm:"column:inventory_id;type:bigint" json:"inventory_id,omitempty"`
	DefaultQuantity     *ExactNumeric           `gorm:"column:default_quantity;type:numeric(10,2);default:1" json:"default_quantity"`
	TaxType             TaxType                 `gorm:"column:tax_type;type:tax_type;not null;default:'excluded'" json:"tax_type"`
	TaxRate             ExactNumeric            `gorm:"column:tax_rate;type:numeric;not null;default:0.10" json:"tax_rate"`
	SortOrder           *int                    `gorm:"column:sort_order;type:integer;default:0" json:"sort_order"`
	IsNonInsurance      bool                    `gorm:"column:is_non_insurance;type:boolean;not null;default:false" json:"is_non_insurance"`
	CreatedAt           time.Time               `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time               `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt           gorm.DeletedAt          `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	CalculationType     MedicineCalculationType `gorm:"column:calculation_type;type:medicine_calculation_type;not null;default:'none'" json:"calculation_type"`
	Strength            *ExactNumeric           `gorm:"column:strength;type:numeric(10,4)" json:"strength,omitempty"`
	FrequencyPerDay     *int                    `gorm:"column:frequency_per_day;type:integer" json:"frequency_per_day,omitempty"`
	DefaultDurationDays *int                    `gorm:"column:default_duration_days;type:integer" json:"default_duration_days,omitempty"`
	Inventory           *InventoryItem          `gorm:"foreignKey:InventoryID;references:ID" json:"inventory,omitempty"`
	DoseParams          []MedicineDoseParam     `gorm:"foreignKey:MedicineID,ClinicID;references:ID,ClinicID"  json:"dose_params,omitempty"`

	IsBillDiscountExcluded *bool `gorm:"column:is_bill_discount_excluded;type:boolean" json:"-"`
}

func (Medicine) TableName() string { return "medicines" }
