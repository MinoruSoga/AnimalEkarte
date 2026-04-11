package model

import (
	"time"

	"gorm.io/gorm"
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
	ID              uint64            `gorm:"primaryKey;autoIncrement"                       json:"id"`
	MedicalRecordID uint64            `gorm:"not null"                                       json:"medical_record_id"`
	ItemType        TreatmentItemType `gorm:"type:treatment_item_type;not null;default:'other'" json:"item_type"`
	ConsultationID  *uint64           `                                                      json:"consultation_id,omitempty"`
	ProcedureID     *uint64           `                                                      json:"procedure_id,omitempty"`
	MedicineID      *uint64           `                                                      json:"medicine_id,omitempty"`
	InventoryID     *uint64           `                                                      json:"inventory_id,omitempty"`
	UnitPrice       int64             `gorm:"default:0"                                      json:"unit_price"`
	Quantity        float64           `gorm:"type:numeric(10,1);default:1"                   json:"quantity"`
	IsSelected      bool              `gorm:"column:is_selected;default:false"               json:"is_selected"`
	Status          TreatmentStatus   `gorm:"type:treatment_status;default:'pending'"          json:"status"`
	Content         string            `gorm:"not null;default:''"                            json:"content"`
	Memo            string            `gorm:"default:''"                                     json:"memo"`
	AdminRoute      string            `gorm:"column:admin_route;type:varchar(50);default:''" json:"admin_route"`
	IsInsurance     bool              `gorm:"column:is_insurance;default:false"              json:"is_insurance"`
	DiscountRate    float64           `gorm:"type:numeric(5,2);default:0"                    json:"discount_rate"`
	DiscountAmount  int64             `gorm:"default:0"                                      json:"discount_amount"`
	SortOrder       int               `gorm:"type:integer;default:0"                         json:"sort_order"`
	DeletedAt       gorm.DeletedAt    `                                                      json:"-"`
	CreatedAt       time.Time         `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time         `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	MedicalRecord *MedicalRecord `gorm:"foreignKey:MedicalRecordID" json:"medical_record,omitempty"`
	Consultation  *Consultation  `gorm:"foreignKey:ConsultationID"  json:"consultation,omitempty"`
	Procedure     *Procedure     `gorm:"foreignKey:ProcedureID"     json:"procedure,omitempty"`
	Medicine      *Medicine      `gorm:"foreignKey:MedicineID"      json:"medicine,omitempty"`
	Inventory     *InventoryItem `gorm:"foreignKey:InventoryID"     json:"inventory,omitempty"`
}

func (Treatment) TableName() string { return "treatments" }
