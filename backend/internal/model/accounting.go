package model

import (
	"gorm.io/gorm"
	"time"
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
	ItemCategoryHouseCall   ItemCategory = "house_call"
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
		ItemCategoryHouseCall,
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
	ID                    uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID              uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	MedicalRecordID       *uint64        `gorm:"column:medical_record_id;type:bigint" json:"medical_record_id,omitempty"`
	HospitalizationID     *uint64        `gorm:"column:hospitalization_id;type:bigint" json:"hospitalization_id,omitempty"`
	OwnerID               *uint64        `gorm:"column:owner_id;type:bigint" json:"owner_id,omitempty"`
	PetID                 *uint64        `gorm:"column:pet_id;type:bigint" json:"pet_id,omitempty"`
	ScheduledDate         time.Time      `gorm:"column:scheduled_date;type:date;not null" json:"scheduled_date"`
	Memo                  string         `gorm:"column:memo;type:text;not null;default:''" json:"memo"`
	CreatedAt             time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	DoctorReviewSetID     *uint64        `gorm:"column:doctor_review_set_id;type:bigint" json:"-"`
	ClinicalSetGeneration *int64         `gorm:"column:clinical_set_generation;type:bigint" json:"-"`
	ClinicalSetDigest     *[]byte        `gorm:"column:clinical_set_digest;type:bytea" json:"-"`
	ClaimState            *string        `gorm:"column:claim_state;type:text" json:"-"`
	Version               int64          `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
	CurrentCalculationID  *uint64        `gorm:"column:current_calculation_id;type:bigint" json:"-"`
	TotalRefundedAmount   int64          `gorm:"-" json:"total_refunded_amount"`
	OutstandingAmount     int64          `gorm:"-" json:"outstanding_amount"`
	Owner                 *Owner         `gorm:"foreignKey:ClinicID,OwnerID;references:ClinicID,ID"          json:"owner,omitempty"`
	Pet                   *Pet           `gorm:"foreignKey:ClinicID,PetID;references:ClinicID,ID"            json:"pet,omitempty"`
	MedicalRecord         *MedicalRecord `gorm:"foreignKey:MedicalRecordID;references:ID"  json:"medical_record,omitempty"`
	Items                 []BillingItem  `gorm:"foreignKey:BillingID,ClinicID;references:ID,ClinicID"        json:"items,omitempty"`
}

func (Billing) TableName() string { return "billings" }

type BillingItem struct {
	ID                    uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	BillingID             uint64         `gorm:"column:billing_id;type:bigint;not null" json:"billing_id"`
	Category              ItemCategory   `gorm:"column:category;type:item_category;not null" json:"category"`
	Name                  string         `gorm:"column:name;type:text;not null;default:''" json:"name"`
	UnitPrice             ExactNumeric   `gorm:"column:unit_price;type:numeric;not null;default:0" json:"unit_price"`
	Quantity              ExactNumeric   `gorm:"column:quantity;type:numeric;not null;default:1" json:"quantity"`
	TaxType               TaxType        `gorm:"column:tax_type;type:tax_type;not null;default:'excluded'" json:"tax_type"`
	TaxRate               *ExactNumeric  `gorm:"column:tax_rate;type:numeric;default:0.10" json:"tax_rate"`
	IsInsuranceApplicable *bool          `gorm:"column:is_insurance_applicable;type:boolean;default:false" json:"is_insurance_applicable"`
	Source                *ItemSource    `gorm:"column:source;type:item_source;default:'manual'" json:"source"`
	MerchandiseItemID     *uint64        `gorm:"column:merchandise_item_id;type:bigint" json:"merchandise_item_id,omitempty"`
	DiscountRate          ExactNumeric   `gorm:"column:discount_rate;type:numeric;not null;default:0" json:"discount_rate"`
	DiscountAmount        int64          `gorm:"column:discount_amount;type:bigint;not null;default:0" json:"discount_amount"`
	SortOrder             *int           `gorm:"column:sort_order;type:integer;default:0" json:"sort_order"`
	CreatedAt             time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	ClinicID              uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	OtherReason           *string        `gorm:"column:other_reason;type:text" json:"other_reason,omitempty"`
	CreatedBy             *uint64        `gorm:"column:created_by;type:bigint" json:"-"`
	ContractVersion       int64          `gorm:"column:contract_version;type:bigint;not null;default:1" json:"-"`
	FulfillmentGeneration int64          `gorm:"column:fulfillment_generation;type:bigint;not null;default:1" json:"-"`
	SourcePetID           *uint64        `gorm:"column:source_pet_id;type:bigint" json:"-"`
	SourceRecordID        *uint64        `gorm:"column:source_record_id;type:bigint" json:"-"`
	MedicalRecordID       *uint64        `gorm:"-"                                              json:"medical_record_id,omitempty"`
	ServiceDate           *time.Time     `gorm:"-"                                              json:"service_date,omitempty"`
}

func (BillingItem) TableName() string { return "billing_items" }

// Tax calculation belongs to the approved accounting calculator; this persistence model does not select a rounding formula.

// PaymentSplit は1会計に対する支払い手段ごとの内訳を表す。
// 混在会計では複数行存在する。delete-then-recreate パターンで管理する（soft-delete なし）。
// Method は各内訳の一次情報（source of truth）。PaymentMethodID と dual maintain する。
