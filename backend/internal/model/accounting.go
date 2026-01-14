package model

import (
	"time"

	"github.com/google/uuid"
)

// Accounting 会計モデル
type Accounting struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	MedicalRecordID *uuid.UUID `json:"medical_record_id" gorm:"type:uuid"`
	PetID           uuid.UUID  `json:"pet_id" gorm:"type:uuid;not null;index:idx_acc_pet_id"`
	OwnerID         uuid.UUID  `json:"owner_id" gorm:"type:uuid;not null"`
	ScheduledDate   time.Time  `json:"scheduled_date" gorm:"type:date"`
	CompletedAt     *time.Time `json:"completed_at"`
	Status          string     `json:"status" gorm:"type:varchar(20);index:idx_acc_status;default:'未収'"` // 未収, 保留, 回収済, キャンセル
	Subtotal        *float64   `json:"subtotal" gorm:"type:decimal(10,2)"`
	TaxTotal        *float64   `json:"tax_total" gorm:"type:decimal(10,2)"`
	TotalAmount     *float64   `json:"total_amount" gorm:"type:decimal(10,2)"`
	InsuranceName   string     `json:"insurance_name" gorm:"type:varchar(100)"`
	InsuranceRatio  *float64   `json:"insurance_ratio" gorm:"type:decimal(3,2)"`
	InsuranceAmount *float64   `json:"insurance_amount" gorm:"type:decimal(10,2)"`
	DiscountAmount  *float64   `json:"discount_amount" gorm:"type:decimal(10,2)"`
	BillingAmount   *float64   `json:"billing_amount" gorm:"type:decimal(10,2)"`
	ReceivedAmount  *float64   `json:"received_amount" gorm:"type:decimal(10,2)"`
	ChangeAmount    *float64   `json:"change_amount" gorm:"type:decimal(10,2)"`
	PaymentMethod   string     `json:"payment_method" gorm:"type:varchar(30)"` // 現金, クレジットカード, 電子マネー
	Memo            string     `json:"memo" gorm:"type:text"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Relations
	Pet            *Pet             `json:"pet,omitempty" gorm:"foreignKey:PetID"`
	Owner          *Owner           `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
	MedicalRecord  *MedicalRecord   `json:"medical_record,omitempty" gorm:"foreignKey:MedicalRecordID"`
	AccountingItems []AccountingItem `json:"accounting_items,omitempty" gorm:"foreignKey:AccountingID"`
}

// TableName テーブル名を指定
func (Accounting) TableName() string {
	return "accountings"
}

// AccountingItem 会計明細モデル
type AccountingItem struct {
	ID                    uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	AccountingID          uuid.UUID  `json:"accounting_id" gorm:"type:uuid;not null"`
	MasterID              *uuid.UUID `json:"master_id" gorm:"type:uuid"`
	Code                  string     `json:"code" gorm:"type:varchar(20)"`
	Category              string     `json:"category" gorm:"type:varchar(50)"`
	Name                  string     `json:"name" gorm:"type:varchar(200)"`
	UnitPrice             *float64   `json:"unit_price" gorm:"type:decimal(10,2)"`
	Quantity              int        `json:"quantity" gorm:"default:1"`
	TaxRate               *float64   `json:"tax_rate" gorm:"type:decimal(3,2)"` // 0.1, 0.08
	IsInsuranceApplicable bool       `json:"is_insurance_applicable" gorm:"default:false"`
	Source                string     `json:"source" gorm:"type:varchar(20)"` // medical_record, manual
	CreatedAt             time.Time  `json:"created_at"`
}

// TableName テーブル名を指定
func (AccountingItem) TableName() string {
	return "accounting_items"
}
