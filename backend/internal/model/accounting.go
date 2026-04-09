package model

import (
	"time"

	"gorm.io/gorm"
)

type TaxType string

const (
	TaxTypeIncluded TaxType = "included" // 内税
	TaxTypeExcluded TaxType = "excluded" // 外税
	TaxTypeExempt   TaxType = "exempt"   // 非課税
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
	ID                uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID          uint64         `gorm:"not null"                                       json:"clinic_id"`
	MedicalRecordID   *uint64        `                                                      json:"medical_record_id,omitempty"`
	HospitalizationID *uint64        `                                                      json:"hospitalization_id,omitempty"`
	OwnerID           *uint64        `                                                      json:"owner_id,omitempty"`
	PetID             *uint64        `                                                      json:"pet_id,omitempty"`
	Subtotal          int            `gorm:"default:0"                                      json:"subtotal"`
	TaxTotal          int            `gorm:"default:0"                                      json:"tax_total"`
	TotalAmount       int            `gorm:"default:0"                                      json:"total_amount"`
	HasInsurance      bool           `gorm:"default:false"                                  json:"has_insurance"`
	Status            BillingStatus  `gorm:"type:billing_status;default:'waiting'"          json:"status"`
	ScheduledDate     time.Time      `gorm:"type:date;not null"                             json:"scheduled_date"`
	CompletedAt       *time.Time     `                                                      json:"completed_at,omitempty"`
	Memo              string         `gorm:"default:''"                                     json:"memo"`
	DeletedAt         gorm.DeletedAt `                                                      json:"-"`
	CreatedAt         time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// 仮想フィールド（DB列なし）— FindAll のサブクエリで集計
	TotalRefundedAmount int64 `gorm:"-" json:"total_refunded_amount"`

	// Relations
	Owner         *Owner          `gorm:"foreignKey:OwnerID"          json:"owner,omitempty"`
	Pet           *Pet            `gorm:"foreignKey:PetID"            json:"pet,omitempty"`
	MedicalRecord *MedicalRecord  `gorm:"foreignKey:MedicalRecordID"  json:"medical_record,omitempty"`
	Items         []BillingItem   `gorm:"foreignKey:BillingID"        json:"items,omitempty"`
	Payments      []Payment       `gorm:"foreignKey:BillingID"        json:"payments,omitempty"`
	Refunds       []BillingRefund `gorm:"foreignKey:BillingID"        json:"refunds,omitempty"`
}

func (Billing) TableName() string { return "billings" }

type BillingItem struct {
	ID                    uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	BillingID             uint64         `gorm:"not null"                                       json:"billing_id"`
	Category              ItemCategory   `gorm:"type:item_category;not null"                    json:"category"`
	Name                  string         `gorm:"not null;default:''"                            json:"name"`
	UnitPrice             int64          `gorm:"not null;default:0"                             json:"unit_price"`
	Quantity              float64        `gorm:"type:numeric(10,1);not null;default:1"          json:"quantity"`
	TaxType               TaxType        `gorm:"type:tax_type;not null;default:excluded"        json:"tax_type"`
	TaxRate               float64        `gorm:"type:numeric(3,2);default:0.10"                 json:"tax_rate"`
	IsInsuranceApplicable bool           `gorm:"default:false"                                  json:"is_insurance_applicable"`
	Source                ItemSource     `gorm:"type:item_source;default:'manual'"              json:"source"`
	MerchandiseItemID     *uint64        `                                                      json:"merchandise_item_id,omitempty"`
	SortOrder             int            `gorm:"type:integer;default:0"                         json:"sort_order"`
	CreatedAt             time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt             time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt             gorm.DeletedAt `                                                      json:"-"`
}

func (BillingItem) TableName() string { return "billing_items" }

type Payment struct {
	ID              uint64        `gorm:"primaryKey;autoIncrement"                       json:"id"`
	BillingID       uint64        `gorm:"not null;uniqueIndex"                           json:"billing_id"`
	Subtotal        int64         `gorm:"not null;default:0"                             json:"subtotal"`
	TaxTotal        int64         `gorm:"not null;default:0"                             json:"tax_total"`
	TotalAmount     int64         `gorm:"not null;default:0"                             json:"total_amount"`
	InsuranceName   string        `gorm:"default:''"                                     json:"insurance_name"`
	InsuranceRatio  float64       `gorm:"type:numeric(3,2);default:0"                    json:"insurance_ratio"`
	InsuranceAmount int64         `gorm:"default:0"                                      json:"insurance_amount"`
	DiscountAmount  int64         `gorm:"default:0"                                      json:"discount_amount"`
	BillingAmount   int64         `gorm:"not null;default:0"                             json:"billing_amount"`
	ReceivedAmount  int64         `gorm:"default:0"                                      json:"received_amount"`
	ChangeAmount    int64         `gorm:"default:0"                                      json:"change_amount"`
	Method          PaymentMethod  `gorm:"type:payment_method;default:'cash'"             json:"method"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt       gorm.DeletedAt `                                                      json:"-"`
}

func (Payment) TableName() string { return "payments" }
