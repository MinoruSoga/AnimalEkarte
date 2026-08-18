package model

import (
	"math"
	"time"

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
	ID                   uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID             uint64         `gorm:"not null"                                       json:"clinic_id"`
	EstimateNo           string         `gorm:"not null;default:''"                            json:"estimate_no"`
	MedicalRecordID      *uint64        `                                                      json:"medical_record_id,omitempty"`
	Title                string         `gorm:"default:''"                                     json:"title"`
	OwnerID              *uint64        `                                                      json:"owner_id,omitempty"`
	PetID                *uint64        `                                                      json:"pet_id,omitempty"`
	Status               EstimateStatus `gorm:"type:estimate_status;default:'draft'"           json:"status"`
	Subtotal             int64          `gorm:"default:0"                                      json:"subtotal"`
	TaxTotal             int64          `gorm:"default:0"                                      json:"tax_total"`
	TotalAmount          int64          `gorm:"default:0"                                      json:"total_amount"`
	InsuranceAmount      int64          `gorm:"default:0"                                      json:"insurance_amount"`
	DiscountAmount       int64          `gorm:"default:0"                                      json:"discount_amount"`
	ValidUntil           *time.Time     `gorm:"type:date"                                      json:"valid_until,omitempty"`
	Comment              string         `gorm:"default:''"                                     json:"comment"`
	Notes                string         `gorm:"default:''"                                     json:"notes"`
	CreatedBy            *uint64        `                                                      json:"created_by,omitempty"`
	SupersedesEstimateID *uint64        `                                                      json:"supersedes_estimate_id,omitempty"`
	DeletedAt            gorm.DeletedAt `                                                      json:"-"`
	CreatedAt            time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt            time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	MedicalRecord *MedicalRecord `gorm:"foreignKey:MedicalRecordID" json:"medical_record,omitempty"`
	Owner         *Owner         `gorm:"foreignKey:OwnerID"         json:"owner,omitempty"`
	Pet           *Pet           `gorm:"foreignKey:PetID"           json:"pet,omitempty"`
	CreatedStaff  *Staff         `gorm:"foreignKey:CreatedBy"       json:"created_staff,omitempty"`
	Items         []EstimateItem `gorm:"foreignKey:EstimateID"      json:"items,omitempty"`
}

func (Estimate) TableName() string { return "estimates" }

// EstimateItem は見積書明細
type EstimateItem struct {
	ID                    uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	EstimateID            uint64         `gorm:"not null"                                       json:"estimate_id"`
	Name                  string         `gorm:"not null;default:''"                            json:"name"`
	Category              ItemCategory   `gorm:"type:item_category;not null"                    json:"category"`
	UnitPrice             int64          `gorm:"default:0"                                      json:"unit_price"`
	Quantity              float64        `gorm:"type:numeric(10,1);default:1"                   json:"quantity"`
	TaxType               TaxType        `gorm:"type:tax_type;not null;default:excluded"        json:"tax_type"`
	TaxRate               float64        `gorm:"type:numeric(3,2);default:0.10"                 json:"tax_rate"`
	DiscountRate          float64        `gorm:"type:numeric(5,2);default:0"                    json:"discount_rate"`
	DiscountAmount        int64          `gorm:"default:0"                                      json:"discount_amount"`
	IsInsuranceApplicable bool           `gorm:"default:false"                                  json:"is_insurance_applicable"`
	ConsultationID        *uint64        `                                                      json:"consultation_id,omitempty"`
	ProcedureID           *uint64        `                                                      json:"procedure_id,omitempty"`
	MedicineID            *uint64        `                                                      json:"medicine_id,omitempty"`
	MerchandiseItemID     *uint64        `                                                      json:"merchandise_item_id,omitempty"`
	SortOrder             int            `gorm:"type:integer;default:0"                         json:"sort_order"`
	CreatedAt             time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt             time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index"                                          json:"-"`

	// Relations
	Consultation *Consultation `gorm:"foreignKey:ConsultationID" json:"consultation,omitempty"`
	Procedure    *Procedure    `gorm:"foreignKey:ProcedureID"    json:"procedure,omitempty"`
	Medicine     *Medicine     `gorm:"foreignKey:MedicineID"     json:"medicine,omitempty"`
}

func (EstimateItem) TableName() string { return "estimate_items" }

// CalculateTaxAmount は課税区分に応じた税額（円）を計算する。
// 課税ベースは BillingItem と同じく max(単価×数量−割引額, 0)（#85 / MDL-01）。
func (item *EstimateItem) CalculateTaxAmount() int64 {
	subtotal := max(float64(item.UnitPrice)*item.Quantity-float64(item.DiscountAmount), 0)
	switch item.TaxType {
	case TaxTypeExcluded:
		return int64(math.Round(subtotal * item.TaxRate))
	case TaxTypeIncluded:
		return int64(math.Round(subtotal * item.TaxRate / (1 + item.TaxRate)))
	default:
		return 0
	}
}
