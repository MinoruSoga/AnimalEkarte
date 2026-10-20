package model

import (
	"gorm.io/gorm"
	"time"
)

// EstimateStatus は見積書ステータス
type EstimateStatus string

const (
	EstimateStatusDraft    EstimateStatus = "draft"
	EstimateStatusSent     EstimateStatus = "sent"
	EstimateStatusApproved EstimateStatus = "approved"
	EstimateStatusRejected EstimateStatus = "rejected"
)

// Estimate は見積書（v7.0追加）
type Estimate struct {
	ID                   uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID             uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	EstimateNo           string          `gorm:"column:estimate_no;type:text;not null" json:"estimate_no"`
	MedicalRecordID      *uint64         `gorm:"column:medical_record_id;type:bigint" json:"medical_record_id,omitempty"`
	Title                string          `gorm:"column:title;type:text;not null;default:''" json:"title"`
	OwnerID              *uint64         `gorm:"column:owner_id;type:bigint" json:"owner_id,omitempty"`
	Status               *EstimateStatus `gorm:"column:status;type:estimate_status;default:'draft'" json:"status"`
	ValidUntil           *time.Time      `gorm:"column:valid_until;type:date" json:"valid_until,omitempty"`
	Comment              string          `gorm:"column:comment;type:text;not null;default:''" json:"comment"`
	Notes                string          `gorm:"column:notes;type:text;not null;default:''" json:"notes"`
	CreatedBy            *uint64         `gorm:"column:created_by;type:bigint" json:"created_by,omitempty"`
	CreatedAt            time.Time       `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time       `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt            gorm.DeletedAt  `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	PetID                *uint64         `gorm:"column:pet_id;type:bigint" json:"pet_id,omitempty"`
	Version              int64           `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
	CurrentCalculationID *uint64         `gorm:"column:current_calculation_id;type:bigint" json:"-"`
	MedicalRecord        *MedicalRecord  `gorm:"foreignKey:MedicalRecordID;references:ID" json:"medical_record,omitempty"`
	Owner                *Owner          `gorm:"foreignKey:OwnerID;references:ID"         json:"owner,omitempty"`
	Pet                  *Pet            `gorm:"foreignKey:ClinicID,PetID;references:ClinicID,ID"           json:"pet,omitempty"`
	CreatedStaff         *Staff          `gorm:"foreignKey:CreatedBy;references:ID"       json:"created_staff,omitempty"`
	Items                []EstimateItem  `gorm:"foreignKey:EstimateID;references:ID"      json:"items,omitempty"`
}

func (Estimate) TableName() string { return "estimates" }

// EstimateItem は見積書明細
type EstimateItem struct {
	ID                    uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	EstimateID            uint64         `gorm:"column:estimate_id;type:bigint;not null" json:"estimate_id"`
	Name                  string         `gorm:"column:name;type:text;not null;default:''" json:"name"`
	Category              ItemCategory   `gorm:"column:category;type:item_category;not null" json:"category"`
	UnitPrice             ExactNumeric   `gorm:"column:unit_price;type:numeric;not null;default:0" json:"unit_price"`
	Quantity              ExactNumeric   `gorm:"column:quantity;type:numeric;not null;default:1" json:"quantity"`
	TaxType               TaxType        `gorm:"column:tax_type;type:tax_type;not null;default:'excluded'" json:"tax_type"`
	TaxRate               *ExactNumeric  `gorm:"column:tax_rate;type:numeric;default:0.10" json:"tax_rate"`
	DiscountRate          *ExactNumeric  `gorm:"column:discount_rate;type:numeric;default:0" json:"discount_rate"`
	DiscountAmount        *int64         `gorm:"column:discount_amount;type:bigint;default:0" json:"discount_amount"`
	IsInsuranceApplicable *bool          `gorm:"column:is_insurance_applicable;type:boolean;default:false" json:"is_insurance_applicable"`
	ConsultationID        *uint64        `gorm:"column:consultation_id;type:bigint" json:"consultation_id,omitempty"`
	ProcedureID           *uint64        `gorm:"column:procedure_id;type:bigint" json:"procedure_id,omitempty"`
	MedicineID            *uint64        `gorm:"column:medicine_id;type:bigint" json:"medicine_id,omitempty"`
	MerchandiseItemID     *uint64        `gorm:"column:merchandise_item_id;type:bigint" json:"merchandise_item_id,omitempty"`
	SortOrder             *int           `gorm:"column:sort_order;type:integer;default:0" json:"sort_order"`
	CreatedAt             time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	Consultation          *Consultation  `gorm:"foreignKey:ConsultationID;references:ID" json:"consultation,omitempty"`
	Procedure             *Procedure     `gorm:"foreignKey:ProcedureID;references:ID"    json:"procedure,omitempty"`
	Medicine              *Medicine      `gorm:"foreignKey:MedicineID;references:ID"     json:"medicine,omitempty"`
}

func (EstimateItem) TableName() string { return "estimate_items" }

// Tax calculation belongs to the approved accounting calculator; this persistence model does not select a rounding formula.
