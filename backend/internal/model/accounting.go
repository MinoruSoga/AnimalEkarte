package model

import (
	"time"

	"github.com/google/uuid"
)

// Accounting は会計レコードテーブルに対応するモデル
type Accounting struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MedicalRecordID   *uuid.UUID `gorm:"column:medical_record_id;type:uuid;uniqueIndex" json:"medical_record_id,omitempty"`
	HospitalizationID *uuid.UUID `gorm:"column:hospitalization_id;type:uuid" json:"hospitalization_id,omitempty"`
	OwnerID           uuid.UUID  `gorm:"column:owner_id;type:uuid;not null" json:"owner_id"`
	OwnerName         string     `gorm:"column:owner_name;not null" json:"owner_name"`
	PetID             uuid.UUID  `gorm:"column:pet_id;type:uuid;not null" json:"pet_id"`
	PetName           string     `gorm:"column:pet_name;not null" json:"pet_name"`
	PetSpecies        *string    `gorm:"column:pet_species;type:pet_species" json:"pet_species,omitempty"`
	Status            string     `gorm:"column:status;type:accounting_status;default:'waiting'" json:"status"`
	ScheduledDate     time.Time  `gorm:"column:scheduled_date;type:date;not null" json:"scheduled_date"`
	CompletedAt       *time.Time `gorm:"column:completed_at" json:"completed_at,omitempty"`
	Memo              string     `gorm:"column:memo;default:''" json:"memo"`
	CreatedAt         time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// Relations
	Items       []AccountingItem `gorm:"foreignKey:AccountingID" json:"items,omitempty"`
	PaymentInfo *PaymentInfo     `gorm:"foreignKey:AccountingID" json:"payment_info,omitempty"`
}

// AccountingItem は会計明細行テーブルに対応するモデル
type AccountingItem struct {
	ID                    uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AccountingID          uuid.UUID `gorm:"column:accounting_id;type:uuid;not null" json:"accounting_id"`
	Code                  *string   `gorm:"column:code" json:"code,omitempty"`
	Category              string    `gorm:"column:category;type:item_category;not null" json:"category"`
	Name                  string    `gorm:"column:name;not null" json:"name"`
	UnitPrice             float64   `gorm:"column:unit_price;type:numeric(10,2);not null" json:"unit_price"`
	Quantity              int       `gorm:"column:quantity;not null;default:1" json:"quantity"`
	TaxRate               float64   `gorm:"column:tax_rate;type:numeric(3,2);not null;default:0.10" json:"tax_rate"`
	IsInsuranceApplicable bool      `gorm:"column:is_insurance_applicable;default:false" json:"is_insurance_applicable"`
	Source                string    `gorm:"column:source;type:item_source;default:'manual'" json:"source"`
	SortOrder             int       `gorm:"column:sort_order;default:0" json:"sort_order"`
	CreatedAt             time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// PaymentInfo は支払情報テーブルに対応するモデル（会計レコードと1:1）
type PaymentInfo struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AccountingID    uuid.UUID `gorm:"column:accounting_id;type:uuid;not null;uniqueIndex" json:"accounting_id"`
	Subtotal        float64   `gorm:"column:subtotal;type:numeric(10,2);not null;default:0" json:"subtotal"`
	TaxTotal        float64   `gorm:"column:tax_total;type:numeric(10,2);not null;default:0" json:"tax_total"`
	TotalAmount     float64   `gorm:"column:total_amount;type:numeric(10,2);not null;default:0" json:"total_amount"`
	InsuranceName   *string   `gorm:"column:insurance_name" json:"insurance_name,omitempty"`
	InsuranceRatio  *float64  `gorm:"column:insurance_ratio;type:numeric(3,2)" json:"insurance_ratio,omitempty"`
	InsuranceAmount float64   `gorm:"column:insurance_amount;type:numeric(10,2);default:0" json:"insurance_amount"`
	DiscountAmount  float64   `gorm:"column:discount_amount;type:numeric(10,2);default:0" json:"discount_amount"`
	BillingAmount   float64   `gorm:"column:billing_amount;type:numeric(10,2);not null;default:0" json:"billing_amount"`
	ReceivedAmount  float64   `gorm:"column:received_amount;type:numeric(10,2);default:0" json:"received_amount"`
	ChangeAmount    float64   `gorm:"column:change_amount;type:numeric(10,2);default:0" json:"change_amount"`
	Method          string    `gorm:"column:method;type:payment_method;default:'cash'" json:"method"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// CreateAccountingRequest は会計作成リクエスト
type CreateAccountingRequest struct {
	MedicalRecordID   *string `json:"medical_record_id"`
	HospitalizationID *string `json:"hospitalization_id"`
	OwnerID           string  `json:"owner_id" binding:"required"`
	OwnerName         string  `json:"owner_name" binding:"required"`
	PetID             string  `json:"pet_id" binding:"required"`
	PetName           string  `json:"pet_name" binding:"required"`
	PetSpecies        *string `json:"pet_species"`
	Status            string  `json:"status"`
	ScheduledDate     string  `json:"scheduled_date" binding:"required"`
	Memo              string  `json:"memo"`
}

// UpdateAccountingRequest は会計更新リクエスト
type UpdateAccountingRequest struct {
	MedicalRecordID   *string `json:"medical_record_id"`
	HospitalizationID *string `json:"hospitalization_id"`
	OwnerID           *string `json:"owner_id"`
	OwnerName         *string `json:"owner_name"`
	PetID             *string `json:"pet_id"`
	PetName           *string `json:"pet_name"`
	PetSpecies        *string `json:"pet_species"`
	Status            *string `json:"status"`
	ScheduledDate     *string `json:"scheduled_date"`
	Memo              *string `json:"memo"`
}
