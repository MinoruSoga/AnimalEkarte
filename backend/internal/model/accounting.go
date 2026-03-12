package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BillingStatus string

const (
	BillingStatusWaiting   BillingStatus = "waiting"
	BillingStatusCompleted BillingStatus = "completed"
	BillingStatusCancelled BillingStatus = "cancelled"
	BillingStatusPending   BillingStatus = "pending"
)

type PaymentMethod string

const (
	PaymentMethodCash            PaymentMethod = "cash"
	PaymentMethodCreditCard      PaymentMethod = "credit_card"
	PaymentMethodElectronicMoney PaymentMethod = "electronic_money"
)

type ItemCategory string

const (
	ItemCategoryExamination ItemCategory = "examination"
	ItemCategoryTest        ItemCategory = "test"
	ItemCategoryProcedure   ItemCategory = "procedure"
	ItemCategorySurgery     ItemCategory = "surgery"
	ItemCategoryMedicine    ItemCategory = "medicine"
	ItemCategoryFood        ItemCategory = "food"
	ItemCategoryGoods       ItemCategory = "goods"
	ItemCategoryOther       ItemCategory = "other"
)

type ItemSource string

const (
	ItemSourceMedicalRecord   ItemSource = "medical_record"
	ItemSourceManual          ItemSource = "manual"
	ItemSourceHospitalization ItemSource = "hospitalization"
)

type Billing struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicID          uuid.UUID      `gorm:"type:uuid;not null"                             json:"clinic_id"`
	MedicalRecordID   *uuid.UUID     `gorm:"type:uuid"                                      json:"medical_record_id,omitempty"`
	HospitalizationID *uuid.UUID     `gorm:"type:uuid"                                      json:"hospitalization_id,omitempty"`
	OwnerID           *uuid.UUID     `gorm:"type:uuid"                                      json:"owner_id,omitempty"`
	PetID             *uuid.UUID     `gorm:"type:uuid"                                      json:"pet_id,omitempty"`
	Subtotal          int            `gorm:"default:0"                                      json:"subtotal"`
	TaxTotal          int            `gorm:"default:0"                                      json:"tax_total"`
	TotalAmount       int            `gorm:"default:0"                                      json:"total_amount"`
	HasInsurance      bool           `gorm:"default:false"                                  json:"has_insurance"`
	Status            BillingStatus  `gorm:"type:billing_status;default:'waiting'"          json:"status"`
	ScheduledDate     time.Time      `gorm:"type:date;not null"                             json:"scheduled_date"`
	CompletedAt       *time.Time     `                                                      json:"completed_at,omitempty"`
	Memo              string         `gorm:"default:''"                                     json:"memo"`
	DeletedAt         gorm.DeletedAt `                                                      json:"deleted_at"`
	CreatedAt         time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Owner         *Owner         `gorm:"foreignKey:OwnerID"          json:"owner,omitempty"`
	Pet           *Pet           `gorm:"foreignKey:PetID"            json:"pet,omitempty"`
	MedicalRecord *MedicalRecord `gorm:"foreignKey:MedicalRecordID"  json:"medical_record,omitempty"`
	Items         []BillingItem  `gorm:"foreignKey:BillingID"        json:"items,omitempty"`
	Payments      []Payment      `gorm:"foreignKey:BillingID"        json:"payments,omitempty"`
}

func (Billing) TableName() string { return "billings" }

type BillingItem struct {
	ID                    uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BillingID             uuid.UUID    `gorm:"type:uuid;not null"                             json:"billing_id"`
	Category              ItemCategory `gorm:"type:item_category;not null"                    json:"category"`
	Name                  string       `gorm:"not null;default:''"                            json:"name"`
	Code                  string       `gorm:"default:''"                                     json:"code"`
	UnitPrice             float64      `gorm:"type:numeric(10,2);not null;default:0"          json:"unit_price"`
	Quantity              int          `gorm:"not null;default:1"                             json:"quantity"`
	TaxRate               float64      `gorm:"type:numeric(3,2);default:0.10"                 json:"tax_rate"`
	IsInsuranceApplicable bool         `gorm:"default:false"                                  json:"is_insurance_applicable"`
	Source                ItemSource   `gorm:"type:item_source;default:'manual'"              json:"source"`
	SortOrder             int          `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt             time.Time    `gorm:"autoCreateTime"                                 json:"created_at"`
}

func (BillingItem) TableName() string { return "billing_items" }

type Payment struct {
	ID              uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BillingID       uuid.UUID     `gorm:"type:uuid;not null;uniqueIndex"                 json:"billing_id"`
	Subtotal        float64       `gorm:"type:numeric(10,2);not null;default:0"          json:"subtotal"`
	TaxTotal        float64       `gorm:"type:numeric(10,2);not null;default:0"          json:"tax_total"`
	TotalAmount     float64       `gorm:"type:numeric(10,2);not null;default:0"          json:"total_amount"`
	InsuranceName   string        `gorm:"default:''"                                     json:"insurance_name"`
	InsuranceRatio  float64       `gorm:"type:numeric(3,2);default:0"                    json:"insurance_ratio"`
	InsuranceAmount float64       `gorm:"type:numeric(10,2);default:0"                   json:"insurance_amount"`
	DiscountAmount  float64       `gorm:"type:numeric(10,2);default:0"                   json:"discount_amount"`
	BillingAmount   float64       `gorm:"type:numeric(10,2);not null;default:0"          json:"billing_amount"`
	ReceivedAmount  float64       `gorm:"type:numeric(10,2);default:0"                   json:"received_amount"`
	ChangeAmount    float64       `gorm:"type:numeric(10,2);default:0"                   json:"change_amount"`
	Method          PaymentMethod `gorm:"type:payment_method;default:'cash'"             json:"method"`
	CreatedAt       time.Time     `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time     `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (Payment) TableName() string { return "payments" }
