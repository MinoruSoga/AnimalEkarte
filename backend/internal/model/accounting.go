package model

import (
	"math"
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
	BillingStatusPending   BillingStatus = "pending"
	BillingStatusCompleted BillingStatus = "completed"
	BillingStatusCancelled BillingStatus = "cancelled"
)

type PaymentMethod string

const (
	PaymentMethodCash            PaymentMethod = "cash"
	PaymentMethodCreditCard      PaymentMethod = "credit_card"
	PaymentMethodElectronicMoney PaymentMethod = "electronic_money"
	PaymentMethodBankTransfer    PaymentMethod = "bank_transfer" // #127: 銀行振込
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
	ItemCategoryVaccine     ItemCategory = "vaccine"
	ItemCategoryTrimming    ItemCategory = "trimming"
	ItemCategoryHotel       ItemCategory = "hotel"
	ItemCategoryTraining    ItemCategory = "training"
)

// AllItemCategories returns every supported billing item category in declaration order.
func AllItemCategories() []ItemCategory {
	return []ItemCategory{
		ItemCategoryExamination,
		ItemCategoryTest,
		ItemCategoryProcedure,
		ItemCategorySurgery,
		ItemCategoryMedicine,
		ItemCategoryFood,
		ItemCategoryGoods,
		ItemCategoryOther,
		ItemCategoryVaccine,
		ItemCategoryTrimming,
		ItemCategoryHotel,
		ItemCategoryTraining,
	}
}

type ItemSource string

const (
	ItemSourceMedicalRecord   ItemSource = "medical_record"
	ItemSourceManual          ItemSource = "manual"
	ItemSourceHospitalization ItemSource = "hospitalization"
	ItemSourceTrimming        ItemSource = "trimming"
)

type Billing struct {
	ID                uint64        `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID          uint64        `gorm:"not null"                                       json:"clinic_id"`
	MedicalRecordID   *uint64       `                                                      json:"medical_record_id,omitempty"`
	HospitalizationID *uint64       `                                                      json:"hospitalization_id,omitempty"`
	OwnerID           *uint64       `                                                      json:"owner_id,omitempty"`
	PetID             *uint64       `                                                      json:"pet_id,omitempty"`
	Subtotal          int64         `gorm:"default:0"                                      json:"subtotal"`
	TaxTotal          int64         `gorm:"default:0"                                      json:"tax_total"`
	TotalAmount       int64         `gorm:"default:0"                                      json:"total_amount"`
	HasInsurance      bool          `gorm:"default:false"                                  json:"has_insurance"`
	Status            BillingStatus `gorm:"type:billing_status;default:'waiting'"          json:"status"`
	ScheduledDate     time.Time     `gorm:"type:date;not null"                             json:"scheduled_date"`
	CompletedAt       *time.Time    `                                                      json:"completed_at,omitempty"`
	Memo              string        `gorm:"default:''"                                     json:"memo"`
	// CompletionRequestID / CompletionRequestHash は BUG-018 complete command の冪等キー。
	// NULL は legacy POST /accountings 経路。soft-delete 後も key 再利用不可（full UNIQUE）。
	CompletionRequestID   *string        `gorm:"type:uuid"                                   json:"completion_request_id,omitempty"`
	CompletionRequestHash *string        `                                                   json:"-"`
	DeletedAt             gorm.DeletedAt `                                                   json:"-"`
	CreatedAt             time.Time      `gorm:"autoCreateTime"                              json:"created_at"`
	UpdatedAt             time.Time      `gorm:"autoUpdateTime"                              json:"updated_at"`

	// 仮想フィールド（DB列なし）— FindAll のサブクエリで集計
	TotalRefundedAmount int64 `gorm:"-" json:"total_refunded_amount"`
	// OutstandingAmount は未収残高（円）。waiting 全額 or クレジット訂正後の patient_due−支払額（BUG-007）。
	OutstandingAmount int64 `gorm:"-" json:"outstanding_amount"`

	// Relations
	Owner         *Owner          `gorm:"foreignKey:OwnerID"          json:"owner,omitempty"`
	Pet           *Pet            `gorm:"foreignKey:PetID"            json:"pet,omitempty"`
	MedicalRecord *MedicalRecord  `gorm:"foreignKey:MedicalRecordID"  json:"medical_record,omitempty"`
	Items         []BillingItem   `gorm:"foreignKey:BillingID"        json:"items,omitempty"`
	Payments      []Payment       `gorm:"foreignKey:BillingID"        json:"payments,omitempty"`
	Refunds       []BillingRefund `gorm:"foreignKey:BillingID"        json:"refunds,omitempty"`
	PaymentSplits []PaymentSplit  `gorm:"foreignKey:BillingID"        json:"payment_splits,omitempty"`
}

func (Billing) TableName() string { return "billings" }

type BillingItem struct {
	ID                    uint64       `gorm:"primaryKey;autoIncrement"                       json:"id"`
	BillingID             uint64       `gorm:"not null"                                       json:"billing_id"`
	ClinicID              *uint64      `                                                      json:"-"`
	Category              ItemCategory `gorm:"type:item_category;not null"                    json:"category"`
	Name                  string       `gorm:"not null;default:''"                            json:"name"`
	UnitPrice             int64        `gorm:"not null;default:0"                             json:"unit_price"`
	Quantity              float64      `gorm:"type:numeric(10,1);not null;default:1"          json:"quantity"`
	DiscountRate          float64      `gorm:"type:numeric(5,2);not null;default:0"           json:"discount_rate"`
	DiscountAmount        int64        `gorm:"not null;default:0"                             json:"discount_amount"`
	TaxType               TaxType      `gorm:"type:tax_type;not null;default:'excluded'"      json:"tax_type"`
	TaxRate               float64      `gorm:"type:numeric(3,2);default:0.10"                 json:"tax_rate"`
	IsInsuranceApplicable bool         `gorm:"default:false"                                  json:"is_insurance_applicable"`
	Source                ItemSource   `gorm:"type:item_source;default:'manual'"              json:"source"`
	OtherReason           *string      `                                                      json:"other_reason,omitempty"`
	CreatedBy             *uint64      `                                                      json:"-"`
	MerchandiseItemID     *uint64      `                                                      json:"merchandise_item_id,omitempty"`
	TreatmentID           *uint64      `                                                      json:"treatment_id,omitempty"`
	// MedicalRecordID は DB 列ではない。未請求候補（treatment 由来）など API 応答用の仮想フィールド。
	MedicalRecordID  *uint64        `gorm:"-"                                              json:"medical_record_id,omitempty"`
	VaccinationID    *uint64        `                                                      json:"vaccination_id,omitempty"`
	ExamID           *uint64        `                                                      json:"exam_id,omitempty"`
	AppointmentID    *uint64        `                                                      json:"appointment_id,omitempty"`
	TrimmingCourseID *uint64        `                                                      json:"trimming_course_id,omitempty"`
	TrimmingOptionID *uint64        `                                                      json:"trimming_option_id,omitempty"`
	SortOrder        int            `gorm:"type:integer;default:0"                         json:"sort_order"`
	CreatedAt        time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt        gorm.DeletedAt `                                                      json:"-"`
}

func (BillingItem) TableName() string { return "billing_items" }

// CalculateTaxAmount は課税区分に応じた税額（円）を計算する。
// #85: 課税ベースは割引後（単価 × 数量 − 割引額）。既存の課税ベース(単価×数量)から
// 割引額を引いた額に税率を適用する（外税なら実質「税抜額に割引」、内税なら税込額から割引）。
//
//	外税: 税額 = (単価 × 数量 − 割引額) × 税率
//	内税: 税額 = (単価 × 数量 − 割引額) × 税率 ÷ (1 + 税率)
//	非課税: 税額 = 0
func (item *BillingItem) CalculateTaxAmount() int64 {
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

type Payment struct {
	ID              uint64  `gorm:"primaryKey;autoIncrement"                       json:"id"`
	BillingID       uint64  `gorm:"not null;uniqueIndex"                           json:"billing_id"`
	ClinicID        uint64  `gorm:"not null"                                       json:"-"` // internal tenant key; not on FE wire
	Subtotal        int64   `gorm:"not null;default:0"                             json:"subtotal"`
	TaxTotal        int64   `gorm:"not null;default:0"                             json:"tax_total"`
	TotalAmount     int64   `gorm:"not null;default:0"                             json:"total_amount"`
	InsuranceName   string  `gorm:"default:''"                                     json:"insurance_name"`
	InsuranceRatio  float64 `gorm:"type:numeric(3,2);default:0"                    json:"insurance_ratio"`
	InsuranceAmount int64   `gorm:"default:0"                                      json:"insurance_amount"`
	DiscountAmount  int64   `gorm:"default:0"                                      json:"discount_amount"`
	BillingAmount   int64   `gorm:"not null;default:0"                             json:"billing_amount"`
	ReceivedAmount  int64   `gorm:"default:0"                                      json:"received_amount"`
	ChangeAmount    int64   `gorm:"default:0"                                      json:"change_amount"`
	// Method は代表支払い手段（PO判断B 2026-05-25: 正式フィールドとして長期維持）。
	// 混在会計では payment_splits の各内訳から representativeMethod() で導出する。
	// PaymentMethodID と dual maintain する（同期ルール: 書込み時は常に両フィールドをセット）。
	Method PaymentMethod `gorm:"type:payment_method;default:'cash'"             json:"method"`
	// PaymentMethodID は当面 nullable で Method と併存する。ID と method の整合は運用ルールで担保する。
	PaymentMethodID *uint64        `                                                      json:"payment_method_id,omitempty"`
	PaidBy          *uint64        `gorm:""                                               json:"paid_by"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt       gorm.DeletedAt `                                                      json:"-"`

	// Relations
	PaidByStaff *Staff `gorm:"foreignKey:PaidBy" json:"paid_by_staff,omitempty" tygo:"-"`
}

func (Payment) TableName() string { return "payments" }

// PaymentSplit は1会計に対する支払い手段ごとの内訳を表す。
// 混在会計では複数行存在する。delete-then-recreate パターンで管理する（soft-delete なし）。
// Method は各内訳の一次情報（source of truth）。PaymentMethodID と dual maintain する。
type PaymentSplit struct {
	ID              uint64        `gorm:"primaryKey;autoIncrement"        json:"id"`
	ClinicID        uint64        `gorm:"not null"                        json:"clinic_id"`
	BillingID       uint64        `gorm:"not null"                        json:"billing_id"`
	Method          PaymentMethod `gorm:"type:payment_method;not null"    json:"method"`
	PaymentMethodID *uint64       `                                       json:"payment_method_id,omitempty"`
	Amount          int64         `gorm:"not null;default:0"              json:"amount"`
	ReceivedAmount  int64         `gorm:"not null;default:0"              json:"received_amount"`
	ChangeAmount    int64         `gorm:"not null;default:0"              json:"change_amount"`
	PaidBy          *uint64       `                                       json:"paid_by,omitempty"`
	CreatedAt       time.Time     `gorm:"autoCreateTime"                 json:"created_at"`

	// Relations
	PaidByStaff *Staff `gorm:"foreignKey:PaidBy" json:"paid_by_staff,omitempty" tygo:"-"`
}

func (PaymentSplit) TableName() string { return "payment_splits" }
