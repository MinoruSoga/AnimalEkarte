package model

import (
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

// TreatmentItemType は治療項目種別
type TreatmentItemType string

const (
	TreatmentItemTypeConsultation TreatmentItemType = "consultation"
	TreatmentItemTypeProcedure    TreatmentItemType = "procedure"
	TreatmentItemTypeMedicine     TreatmentItemType = "medicine"
	TreatmentItemTypeOther        TreatmentItemType = "other"
)

// TreatmentStatus は治療ステータス
type TreatmentStatus string

const (
	TreatmentStatusPending       TreatmentStatus = "pending"
	TreatmentStatusCompleted     TreatmentStatus = "completed"
	TreatmentStatusNotApplicable TreatmentStatus = "not_applicable"
)

// Treatment は治療項目（外来診療）
type Treatment struct {
	ID                         uint64            `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	MedicalRecordID            uint64            `gorm:"column:medical_record_id;type:bigint;not null" json:"medical_record_id"`
	ItemType                   TreatmentItemType `gorm:"column:item_type;type:treatment_item_type;not null;default:'other'" json:"item_type"`
	ConsultationID             *uint64           `gorm:"column:consultation_id;type:bigint" json:"consultation_id,omitempty"`
	ProcedureID                *uint64           `gorm:"column:procedure_id;type:bigint" json:"procedure_id,omitempty"`
	MedicineID                 *uint64           `gorm:"column:medicine_id;type:bigint" json:"medicine_id,omitempty"`
	IsSelected                 *bool             `gorm:"column:is_selected;type:boolean;default:false" json:"is_selected"`
	Status                     *TreatmentStatus  `gorm:"column:status;type:treatment_status;default:'pending'" json:"status"`
	Content                    string            `gorm:"column:content;type:text;not null;default:''" json:"content"`
	Memo                       string            `gorm:"column:memo;type:text;not null;default:''" json:"memo"`
	AdminRoute                 string            `gorm:"column:admin_route;type:varchar(50);not null;default:''" json:"admin_route"`
	IsInsurance                *bool             `gorm:"column:is_insurance;type:boolean;default:false" json:"is_insurance"`
	UnitPrice                  *int64            `gorm:"column:unit_price;type:bigint;default:0" json:"unit_price"`
	Quantity                   *ExactNumeric     `gorm:"column:quantity;type:numeric(10,2);default:1" json:"quantity"`
	DiscountRate               *ExactNumeric     `gorm:"column:discount_rate;type:numeric(5,2);default:0" json:"discount_rate"`
	DiscountAmount             *int64            `gorm:"column:discount_amount;type:bigint;default:0" json:"discount_amount"`
	InventoryID                *uint64           `gorm:"column:inventory_id;type:bigint" json:"inventory_id,omitempty"`
	SortOrder                  *int              `gorm:"column:sort_order;type:integer;default:0" json:"sort_order"`
	CreatedAt                  time.Time         `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt                  time.Time         `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt                  gorm.DeletedAt    `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	DoseWeightKg               *ExactNumeric     `gorm:"column:dose_weight_kg;type:numeric(6,2)" json:"dose_weight_kg,omitempty"`
	DoseWeightSource           *string           `gorm:"column:dose_weight_source;type:varchar(255)" json:"dose_weight_source,omitempty"`
	DoseAmountMg               *ExactNumeric     `gorm:"column:dose_amount_mg;type:numeric(12,6)" json:"dose_amount_mg,omitempty"`
	DoseAmountUnit             *string           `gorm:"column:dose_amount_unit;type:text" json:"dose_amount_unit,omitempty"`
	DoseParamSnapshot          *json.RawMessage  `gorm:"column:dose_param_snapshot;type:jsonb" json:"dose_param_snapshot,omitempty"`
	ClinicID                   uint64            `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	Version                    int               `gorm:"column:version;type:integer;not null;default:1" json:"version"`
	ContractVersion            int64             `gorm:"column:contract_version;type:bigint;not null;default:1" json:"-"`
	FulfillmentGeneration      int64             `gorm:"column:fulfillment_generation;type:bigint;not null;default:1" json:"-"`
	AdministrationDoseAmount   *ExactNumeric     `gorm:"column:administration_dose_amount;type:numeric(12,6)" json:"-"`
	AdministrationDoseUnit     *string           `gorm:"column:administration_dose_unit;type:text" json:"-"`
	AdministrationInstructions *string           `gorm:"column:administration_instructions;type:text" json:"-"`
	PrescriptionTotalAmount    *ExactNumeric     `gorm:"column:prescription_total_amount;type:numeric(12,6)" json:"-"`
	PrescriptionTotalUnit      *string           `gorm:"column:prescription_total_unit;type:text" json:"-"`
	DosePatternRevisionID      *uint64           `gorm:"column:dose_pattern_revision_id;type:bigint" json:"-"`
	MedicalRecord              *MedicalRecord    `gorm:"foreignKey:MedicalRecordID;references:ID" json:"medical_record,omitempty"`
	Consultation               *Consultation     `gorm:"foreignKey:ConsultationID,ClinicID;references:ID,ClinicID"  json:"consultation,omitempty"`
	Procedure                  *Procedure        `gorm:"foreignKey:ProcedureID,ClinicID;references:ID,ClinicID"     json:"procedure,omitempty"`
	Medicine                   *Medicine         `gorm:"foreignKey:MedicineID,ClinicID;references:ID,ClinicID"      json:"medicine,omitempty"`
	Inventory                  *InventoryItem    `gorm:"foreignKey:InventoryID,ClinicID;references:ID,ClinicID"     json:"inventory,omitempty"`
}

func (Treatment) TableName() string { return "treatments" }

// PetTreatmentHistoryFilter は #159 飼主レポート用の治療履歴絞り込み条件。
type PetTreatmentHistoryFilter struct {
	ItemType       *TreatmentItemType
	AnesthesiaOnly bool // true = procedures.anesthesia != 'none'
	IsSurgery      bool // true = procedures.is_surgery = true
}
