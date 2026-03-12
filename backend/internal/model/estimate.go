package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicID        uuid.UUID      `gorm:"type:uuid;not null"                             json:"clinic_id"`
	EstimateNo      string         `gorm:"not null;default:''"                            json:"estimate_no"`
	MedicalRecordID *uuid.UUID     `gorm:"type:uuid"                                      json:"medical_record_id,omitempty"`
	Title           string         `gorm:"default:''"                                     json:"title"`
	OwnerID         *uuid.UUID     `gorm:"type:uuid"                                      json:"owner_id,omitempty"`
	Status          EstimateStatus `gorm:"type:estimate_status;default:'draft'"           json:"status"`
	Subtotal        float64        `gorm:"type:numeric(10,2);default:0"                   json:"subtotal"`
	TaxTotal        float64        `gorm:"type:numeric(10,2);default:0"                   json:"tax_total"`
	TotalAmount     float64        `gorm:"type:numeric(10,2);default:0"                   json:"total_amount"`
	InsuranceAmount float64        `gorm:"type:numeric(10,2);default:0"                   json:"insurance_amount"`
	DiscountAmount  float64        `gorm:"type:numeric(10,2);default:0"                   json:"discount_amount"`
	ValidUntil      *time.Time     `gorm:"type:date"                                      json:"valid_until,omitempty"`
	Comment         string         `gorm:"default:''"                                     json:"comment"`
	Notes           string         `gorm:"default:''"                                     json:"notes"`
	CreatedBy       *uuid.UUID     `gorm:"type:uuid"                                      json:"created_by,omitempty"`
	DeletedAt       gorm.DeletedAt `                                                      json:"deleted_at"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	MedicalRecord *MedicalRecord `gorm:"foreignKey:MedicalRecordID" json:"medical_record,omitempty"`
	Owner         *Owner         `gorm:"foreignKey:OwnerID"         json:"owner,omitempty"`
	CreatedStaff  *Staff         `gorm:"foreignKey:CreatedBy"       json:"created_staff,omitempty"`
	Items         []EstimateItem `gorm:"foreignKey:EstimateID"      json:"items,omitempty"`
}

func (Estimate) TableName() string { return "estimates" }

// EstimateItem は見積書明細
type EstimateItem struct {
	ID                    uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	EstimateID            uuid.UUID    `gorm:"type:uuid;not null"                             json:"estimate_id"`
	Name                  string       `gorm:"not null;default:''"                            json:"name"`
	Category              ItemCategory `gorm:"type:item_category;not null"                    json:"category"`
	UnitPrice             float64      `gorm:"type:numeric(10,2);default:0"                   json:"unit_price"`
	Quantity              int          `gorm:"default:1"                                      json:"quantity"`
	TaxRate               float64      `gorm:"type:numeric(3,2);default:0.10"                 json:"tax_rate"`
	DiscountRate          float64      `gorm:"type:numeric(5,2);default:0"                    json:"discount_rate"`
	DiscountAmount        float64      `gorm:"type:numeric(10,2);default:0"                   json:"discount_amount"`
	IsInsuranceApplicable bool         `gorm:"default:false"                                  json:"is_insurance_applicable"`
	ConsultationID        *uuid.UUID   `gorm:"type:uuid"                                      json:"consultation_id,omitempty"`
	ProcedureID           *uuid.UUID   `gorm:"type:uuid"                                      json:"procedure_id,omitempty"`
	MedicineID            *uuid.UUID   `gorm:"type:uuid"                                      json:"medicine_id,omitempty"`
	SortOrder             int          `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt             time.Time    `gorm:"autoCreateTime"                                 json:"created_at"`

	// Relations
	Consultation *Consultation `gorm:"foreignKey:ConsultationID" json:"consultation,omitempty"`
	Procedure    *Procedure    `gorm:"foreignKey:ProcedureID"    json:"procedure,omitempty"`
	Medicine     *Medicine     `gorm:"foreignKey:MedicineID"     json:"medicine,omitempty"`
}

func (EstimateItem) TableName() string { return "estimate_items" }
