package model

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

// Current migration is authoritative. These are persistence values, not public DTOs or workflow implementations.

// BillingCalculationLine maps the current billing_calculation_lines storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type BillingCalculationLine struct {
	ClinicID                uint64          `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	CalculationID           uint64          `gorm:"column:calculation_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	BillingID               uint64          `gorm:"column:billing_id;type:bigint;not null" json:"-"`
	BillingItemID           uint64          `gorm:"column:billing_item_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	Ordinal                 int             `gorm:"column:ordinal;type:integer;not null" json:"-"`
	UnitPrice               ExactNumeric    `gorm:"column:unit_price;type:numeric;not null" json:"-"`
	Quantity                ExactNumeric    `gorm:"column:quantity;type:numeric;not null" json:"-"`
	SourceSnapshot          json.RawMessage `gorm:"column:source_snapshot;type:jsonb;not null" json:"-"`
	DiscountConditions      json.RawMessage `gorm:"column:discount_conditions;type:jsonb;not null" json:"-"`
	RequestedDiscountAmount ExactNumeric    `gorm:"column:requested_discount_amount;type:numeric;not null" json:"-"`
	AppliedDiscountAmount   int64           `gorm:"column:applied_discount_amount;type:bigint;not null" json:"-"`
	AllocatedBillDiscount   int64           `gorm:"column:allocated_bill_discount;type:bigint;not null" json:"-"`
	NetAmount               int64           `gorm:"column:net_amount;type:bigint;not null" json:"-"`
	TaxRate                 ExactNumeric    `gorm:"column:tax_rate;type:numeric;not null" json:"-"`
	TaxCategory             string          `gorm:"column:tax_category;type:text;not null" json:"-"`
	InsuranceCategory       string          `gorm:"column:insurance_category;type:text;not null" json:"-"`
}

func (BillingCalculationLine) TableName() string { return "billing_calculation_lines" }

// BillingCalculationSnapshot maps the current billing_calculation_snapshots storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type BillingCalculationSnapshot struct {
	ID                    uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID              uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	BillingID             uint64          `gorm:"column:billing_id;type:bigint;not null" json:"-"`
	Revision              int64           `gorm:"column:revision;type:bigint;not null" json:"-"`
	RuleVersion           string          `gorm:"column:rule_version;type:text;not null" json:"-"`
	SourceVersions        json.RawMessage `gorm:"column:source_versions;type:jsonb;not null" json:"-"`
	Conditions            json.RawMessage `gorm:"column:conditions;type:jsonb;not null" json:"-"`
	TaxRateTotals         json.RawMessage `gorm:"column:tax_rate_totals;type:jsonb;not null" json:"-"`
	InsuranceConfirmation json.RawMessage `gorm:"column:insurance_confirmation;type:jsonb;not null" json:"-"`
	Subtotal              int64           `gorm:"column:subtotal;type:bigint;not null" json:"-"`
	TaxTotal              int64           `gorm:"column:tax_total;type:bigint;not null" json:"-"`
	TotalAmount           int64           `gorm:"column:total_amount;type:bigint;not null" json:"-"`
	LineCount             int             `gorm:"column:line_count;type:integer;not null" json:"-"`
	OperationID           uint64          `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	RecordedAt            time.Time       `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (BillingCalculationSnapshot) TableName() string { return "billing_calculation_snapshots" }

// BillingChangeEvent maps the current billing_change_events storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type BillingChangeEvent struct {
	ID                  uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID            uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	BillingID           uint64          `gorm:"column:billing_id;type:bigint;not null" json:"-"`
	Revision            int64           `gorm:"column:revision;type:bigint;not null" json:"-"`
	Intent              string          `gorm:"column:intent;type:text;not null" json:"-"`
	SourceCalculationID *uint64         `gorm:"column:source_calculation_id;type:bigint" json:"-"`
	BeforeValues        json.RawMessage `gorm:"column:before_values;type:jsonb;not null" json:"-"`
	AfterValues         json.RawMessage `gorm:"column:after_values;type:jsonb;not null" json:"-"`
	Reason              *string         `gorm:"column:reason;type:text" json:"-"`
	OperationID         uint64          `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	RecordedBy          uint64          `gorm:"column:recorded_by;type:bigint;not null" json:"-"`
	RecordedAt          time.Time       `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (BillingChangeEvent) TableName() string { return "billing_change_events" }

// BillingFulfillmentEvent maps the current billing_fulfillment_events storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type BillingFulfillmentEvent struct {
	ID               uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID         uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	TreatmentID      *uint64   `gorm:"column:treatment_id;type:bigint" json:"-"`
	BillingItemID    *uint64   `gorm:"column:billing_item_id;type:bigint" json:"-"`
	SourceGeneration int64     `gorm:"column:source_generation;type:bigint;not null" json:"-"`
	EventVersion     int64     `gorm:"column:event_version;type:bigint;not null" json:"-"`
	EventKind        string    `gorm:"column:event_kind;type:text;not null" json:"-"`
	PrecedingEventID *uint64   `gorm:"column:preceding_event_id;type:bigint" json:"-"`
	RecordedBy       uint64    `gorm:"column:recorded_by;type:bigint;not null" json:"-"`
	OccurredAt       time.Time `gorm:"column:occurred_at;type:timestamptz;not null" json:"-"`
	RecordedAt       time.Time `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (BillingFulfillmentEvent) TableName() string { return "billing_fulfillment_events" }

// BillingPriceAcknowledgement maps the current billing_price_acknowledgements storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type BillingPriceAcknowledgement struct {
	ID                 uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID           uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	BillingID          uint64          `gorm:"column:billing_id;type:bigint;not null" json:"-"`
	BillingItemID      uint64          `gorm:"column:billing_item_id;type:bigint;not null" json:"-"`
	BillingVersion     int64           `gorm:"column:billing_version;type:bigint;not null" json:"-"`
	LineVersion        int64           `gorm:"column:line_version;type:bigint;not null" json:"-"`
	CalculationVersion int64           `gorm:"column:calculation_version;type:bigint;not null" json:"-"`
	OperationKey       uuid.UUID       `gorm:"column:operation_key;type:uuid;not null" json:"-"`
	RequestHash        []byte          `gorm:"column:request_hash;type:bytea;not null" json:"-"`
	BeforeValues       json.RawMessage `gorm:"column:before_values;type:jsonb;not null" json:"-"`
	AfterValues        json.RawMessage `gorm:"column:after_values;type:jsonb;not null" json:"-"`
	AcknowledgedBy     uint64          `gorm:"column:acknowledged_by;type:bigint;not null" json:"-"`
	AcknowledgedAt     time.Time       `gorm:"column:acknowledged_at;type:timestamptz;not null;default:clock_timestamp()" json:"-"`
}

func (BillingPriceAcknowledgement) TableName() string { return "billing_price_acknowledgements" }

// BillingReissueLink maps the current billing_reissue_links storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type BillingReissueLink struct {
	ClinicID              uint64          `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	OriginalBillingID     uint64          `gorm:"column:original_billing_id;type:bigint;not null" json:"-"`
	NewBillingID          uint64          `gorm:"column:new_billing_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	OriginalCalculationID uint64          `gorm:"column:original_calculation_id;type:bigint;not null" json:"-"`
	SelectedSourceLines   json.RawMessage `gorm:"column:selected_source_lines;type:jsonb;not null" json:"-"`
	Reason                string          `gorm:"column:reason;type:text;not null" json:"-"`
	OperationID           uint64          `gorm:"column:operation_id;type:bigint;not null" json:"-"`
}

func (BillingReissueLink) TableName() string { return "billing_reissue_links" }

// BillingSourceBinding maps the current billing_source_bindings storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type BillingSourceBinding struct {
	ClinicID                uint64          `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	BillingItemID           uint64          `gorm:"column:billing_item_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	SourceKind              string          `gorm:"column:source_kind;type:text;not null" json:"-"`
	TreatmentID             *uint64         `gorm:"column:treatment_id;type:bigint" json:"-"`
	VaccinationID           *uint64         `gorm:"column:vaccination_id;type:bigint" json:"-"`
	ExaminationID           *uint64         `gorm:"column:examination_id;type:bigint" json:"-"`
	TrimmingSourceLineID    *uint64         `gorm:"column:trimming_source_line_id;type:bigint" json:"-"`
	AppliedSourceVersion    int64           `gorm:"column:applied_source_version;type:bigint;not null" json:"-"`
	ImportedValues          json.RawMessage `gorm:"column:imported_values;type:jsonb;not null" json:"-"`
	ManualOverride          *bool           `gorm:"column:manual_override;type:boolean" json:"-"`
	ContractVersion         int64           `gorm:"column:contract_version;type:bigint;not null;default:1" json:"-"`
	UpdatedAt               time.Time       `gorm:"column:updated_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
	ClaimReleasedAt         *time.Time      `gorm:"column:claim_released_at;type:timestamptz" json:"-"`
	ClaimReleaseOperationID *uint64         `gorm:"column:claim_release_operation_id;type:bigint" json:"-"`
	ClaimReleaseReason      *string         `gorm:"column:claim_release_reason;type:text" json:"-"`
}

func (BillingSourceBinding) TableName() string { return "billing_source_bindings" }

// BillingSourceDifferenceResolution maps the current billing_source_difference_resolutions storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type BillingSourceDifferenceResolution struct {
	ClinicID                uint64    `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	DifferenceID            uint64    `gorm:"column:difference_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	OperationKey            uuid.UUID `gorm:"column:operation_key;type:uuid;not null" json:"-"`
	ResultingBindingVersion int64     `gorm:"column:resulting_binding_version;type:bigint;not null" json:"-"`
	ResolvedBy              uint64    `gorm:"column:resolved_by;type:bigint;not null" json:"-"`
	ResolvedAt              time.Time `gorm:"column:resolved_at;type:timestamptz;not null;default:clock_timestamp()" json:"-"`
}

func (BillingSourceDifferenceResolution) TableName() string {
	return "billing_source_difference_resolutions"
}

// BillingSourceDifference maps the current billing_source_differences storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type BillingSourceDifference struct {
	ID                    uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID              uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	BillingItemID         uint64          `gorm:"column:billing_item_id;type:bigint;not null" json:"-"`
	ObservedSourceVersion int64           `gorm:"column:observed_source_version;type:bigint;not null" json:"-"`
	BindingVersion        int64           `gorm:"column:binding_version;type:bigint;not null" json:"-"`
	SourceDeleted         bool            `gorm:"column:source_deleted;type:boolean;not null" json:"-"`
	DifferenceKind        string          `gorm:"column:difference_kind;type:text;not null" json:"-"`
	ChangedValues         json.RawMessage `gorm:"column:changed_values;type:jsonb;not null" json:"-"`
	ObservedAt            time.Time       `gorm:"column:observed_at;type:timestamptz;not null;default:clock_timestamp()" json:"-"`
}

func (BillingSourceDifference) TableName() string { return "billing_source_differences" }

// CashActualCorrection maps the current cash_actual_corrections storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type CashActualCorrection struct {
	ID                     uint64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID               uint64     `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	CloseID                uint64     `gorm:"column:close_id;type:bigint;not null" json:"-"`
	Revision               int64      `gorm:"column:revision;type:bigint;not null" json:"-"`
	PreviousCorrectionID   *uint64    `gorm:"column:previous_correction_id;type:bigint" json:"-"`
	BeforeActualAmount     int64      `gorm:"column:before_actual_amount;type:bigint;not null" json:"-"`
	CorrectedActualAmount  int64      `gorm:"column:corrected_actual_amount;type:bigint;not null" json:"-"`
	TheoreticalBasisAmount int64      `gorm:"column:theoretical_basis_amount;type:bigint;not null" json:"-"`
	SourceGeneration       int64      `gorm:"column:source_generation;type:bigint;not null" json:"-"`
	OriginalCountedAt      *time.Time `gorm:"column:original_counted_at;type:timestamptz" json:"-"`
	Reason                 string     `gorm:"column:reason;type:text;not null" json:"-"`
	OperationID            uint64     `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	CorrectedBy            uint64     `gorm:"column:corrected_by;type:bigint;not null" json:"-"`
	CorrectedAt            time.Time  `gorm:"column:corrected_at;type:timestamptz;not null;default:clock_timestamp()" json:"-"`
}

func (CashActualCorrection) TableName() string { return "cash_actual_corrections" }

// CashAdjustmentEvent maps the current cash_adjustment_events storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type CashAdjustmentEvent struct {
	ID          uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID    uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	EventKind   string    `gorm:"column:event_kind;type:text;not null" json:"-"`
	Amount      int64     `gorm:"column:amount;type:bigint;not null" json:"-"`
	OccurredAt  time.Time `gorm:"column:occurred_at;type:timestamptz;not null" json:"-"`
	Reason      string    `gorm:"column:reason;type:text;not null" json:"-"`
	OperationID uint64    `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	RecordedBy  uint64    `gorm:"column:recorded_by;type:bigint;not null" json:"-"`
	RecordedAt  time.Time `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (CashAdjustmentEvent) TableName() string { return "cash_adjustment_events" }

// CashCloseContext maps the current cash_close_contexts storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type CashCloseContext struct {
	ClinicID              uint64          `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	CloseID               uint64          `gorm:"column:close_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	BoundaryRevision      int64           `gorm:"column:boundary_revision;type:bigint;not null" json:"-"`
	SourceGeneration      int64           `gorm:"column:source_generation;type:bigint;not null" json:"-"`
	SourceSetDigest       []byte          `gorm:"column:source_set_digest;type:bytea;not null" json:"-"`
	OpeningBalanceID      uint64          `gorm:"column:opening_balance_id;type:bigint;not null" json:"-"`
	CashAdjustmentVersion int64           `gorm:"column:cash_adjustment_version;type:bigint;not null" json:"-"`
	SourceEventCount      int             `gorm:"column:source_event_count;type:integer;not null" json:"-"`
	SourceSnapshot        json.RawMessage `gorm:"column:source_snapshot;type:jsonb;not null" json:"-"`
	OperationID           uint64          `gorm:"column:operation_id;type:bigint;not null" json:"-"`
}

func (CashCloseContext) TableName() string { return "cash_close_contexts" }

// CashCloseLateEvent maps the current cash_close_late_events storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type CashCloseLateEvent struct {
	ClinicID      uint64    `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	CloseID       uint64    `gorm:"column:close_id;type:bigint;not null" json:"-"`
	SourceEventID uint64    `gorm:"column:source_event_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	AssignedAt    time.Time `gorm:"column:assigned_at;type:timestamptz;not null;default:clock_timestamp()" json:"-"`
}

func (CashCloseLateEvent) TableName() string { return "cash_close_late_events" }

// CashCloseSourceEvent maps the current cash_close_source_events storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type CashCloseSourceEvent struct {
	ClinicID      uint64 `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	CloseID       uint64 `gorm:"column:close_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	SourceEventID uint64 `gorm:"column:source_event_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	SourceVersion int64  `gorm:"column:source_version;type:bigint;not null" json:"-"`
}

func (CashCloseSourceEvent) TableName() string { return "cash_close_source_events" }

// CashOpeningBalance maps the current cash_opening_balances storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type CashOpeningBalance struct {
	ID            uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID      uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	BusinessDate  time.Time `gorm:"column:business_date;type:date;not null" json:"-"`
	Period        string    `gorm:"column:period;type:text;not null" json:"-"`
	Revision      int64     `gorm:"column:revision;type:bigint;not null" json:"-"`
	Amount        int64     `gorm:"column:amount;type:bigint;not null" json:"-"`
	SourceKind    string    `gorm:"column:source_kind;type:text;not null" json:"-"`
	SourceCloseID *uint64   `gorm:"column:source_close_id;type:bigint" json:"-"`
	OperationID   uint64    `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	ConfirmedBy   uint64    `gorm:"column:confirmed_by;type:bigint;not null" json:"-"`
	ConfirmedAt   time.Time `gorm:"column:confirmed_at;type:timestamptz;not null;default:clock_timestamp()" json:"-"`
}

func (CashOpeningBalance) TableName() string { return "cash_opening_balances" }

// ClosingBoundaryRevision maps the current closing_boundary_revisions storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type ClosingBoundaryRevision struct {
	ClinicID         uint64          `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	Revision         int64           `gorm:"column:revision;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	BoundarySnapshot json.RawMessage `gorm:"column:boundary_snapshot;type:jsonb;not null" json:"-"`
	OperationID      uint64          `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	RecordedAt       time.Time       `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (ClosingBoundaryRevision) TableName() string { return "closing_boundary_revisions" }

// DoctorReviewRecord maps the current doctor_review_records storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type DoctorReviewRecord struct {
	ClinicID           uint64          `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	ReviewSetID        uint64          `gorm:"column:review_set_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	BillingID          uint64          `gorm:"column:billing_id;type:bigint;not null" json:"-"`
	MedicalRecordID    uint64          `gorm:"column:medical_record_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	ClinicalGeneration int64           `gorm:"column:clinical_generation;type:bigint;not null" json:"-"`
	ClinicalDigest     []byte          `gorm:"column:clinical_digest;type:bytea;not null" json:"-"`
	ClinicalItems      json.RawMessage `gorm:"column:clinical_items;type:jsonb;not null" json:"-"`
}

func (DoctorReviewRecord) TableName() string { return "doctor_review_records" }

// DoctorReviewSet maps the current doctor_review_sets storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type DoctorReviewSet struct {
	ID                        uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID                  uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	BillingID                 uint64    `gorm:"column:billing_id;type:bigint;not null" json:"-"`
	BillingClinicalGeneration int64     `gorm:"column:billing_clinical_generation;type:bigint;not null" json:"-"`
	BillingClinicalDigest     []byte    `gorm:"column:billing_clinical_digest;type:bytea;not null" json:"-"`
	RecordCount               int       `gorm:"column:record_count;type:integer;not null" json:"-"`
	ReviewedBy                uint64    `gorm:"column:reviewed_by;type:bigint;not null" json:"-"`
	ReviewedAt                time.Time `gorm:"column:reviewed_at;type:timestamptz;not null;default:clock_timestamp()" json:"-"`
	ConfirmationNote          *string   `gorm:"column:confirmation_note;type:text" json:"-"`
}

func (DoctorReviewSet) TableName() string { return "doctor_review_sets" }

// EstimateBillingHandoff maps the current estimate_billing_handoffs storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type EstimateBillingHandoff struct {
	ClinicID       uint64 `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	PresentationID uint64 `gorm:"column:presentation_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	EstimateID     uint64 `gorm:"column:estimate_id;type:bigint;not null" json:"-"`
	EstimateItemID uint64 `gorm:"column:estimate_item_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	BillingID      uint64 `gorm:"column:billing_id;type:bigint;not null" json:"-"`
	BillingItemID  uint64 `gorm:"column:billing_item_id;type:bigint;not null" json:"-"`
	OperationID    uint64 `gorm:"column:operation_id;type:bigint;not null" json:"-"`
}

func (EstimateBillingHandoff) TableName() string { return "estimate_billing_handoffs" }

// EstimateCalculationRevision maps the current estimate_calculation_revisions storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type EstimateCalculationRevision struct {
	ID                    uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID              uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	EstimateID            uint64          `gorm:"column:estimate_id;type:bigint;not null" json:"-"`
	Revision              int64           `gorm:"column:revision;type:bigint;not null" json:"-"`
	RuleVersion           string          `gorm:"column:rule_version;type:text;not null" json:"-"`
	Conditions            json.RawMessage `gorm:"column:conditions;type:jsonb;not null" json:"-"`
	InsuranceConfirmation json.RawMessage `gorm:"column:insurance_confirmation;type:jsonb;not null" json:"-"`
	TaxRateTotals         json.RawMessage `gorm:"column:tax_rate_totals;type:jsonb;not null" json:"-"`
	LineValues            json.RawMessage `gorm:"column:line_values;type:jsonb;not null" json:"-"`
	Subtotal              int64           `gorm:"column:subtotal;type:bigint;not null" json:"-"`
	TaxTotal              int64           `gorm:"column:tax_total;type:bigint;not null" json:"-"`
	TotalAmount           int64           `gorm:"column:total_amount;type:bigint;not null" json:"-"`
	OperationID           uint64          `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	RecordedBy            uint64          `gorm:"column:recorded_by;type:bigint;not null" json:"-"`
	RecordedAt            time.Time       `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (EstimateCalculationRevision) TableName() string { return "estimate_calculation_revisions" }

// EstimatePresentationLine maps the current estimate_presentation_lines storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type EstimatePresentationLine struct {
	ClinicID       uint64          `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	PresentationID uint64          `gorm:"column:presentation_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	EstimateID     uint64          `gorm:"column:estimate_id;type:bigint;not null" json:"-"`
	EstimateItemID uint64          `gorm:"column:estimate_item_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	Ordinal        int             `gorm:"column:ordinal;type:integer;not null" json:"-"`
	SourceValues   json.RawMessage `gorm:"column:source_values;type:jsonb;not null" json:"-"`
	UnitPrice      ExactNumeric    `gorm:"column:unit_price;type:numeric;not null" json:"-"`
	Quantity       ExactNumeric    `gorm:"column:quantity;type:numeric;not null" json:"-"`
}

func (EstimatePresentationLine) TableName() string { return "estimate_presentation_lines" }

// EstimatePresentation maps the current estimate_presentations storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type EstimatePresentation struct {
	ID                uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID          uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	EstimateID        uint64          `gorm:"column:estimate_id;type:bigint;not null" json:"-"`
	Revision          int64           `gorm:"column:revision;type:bigint;not null" json:"-"`
	EstimateVersion   int64           `gorm:"column:estimate_version;type:bigint;not null" json:"-"`
	ContentDigest     []byte          `gorm:"column:content_digest;type:bytea;not null" json:"-"`
	RecipientSnapshot json.RawMessage `gorm:"column:recipient_snapshot;type:jsonb;not null" json:"-"`
	ValidUntil        *time.Time      `gorm:"column:valid_until;type:date" json:"-"`
	LineCount         int             `gorm:"column:line_count;type:integer;not null" json:"-"`
	PresentedBy       uint64          `gorm:"column:presented_by;type:bigint;not null" json:"-"`
	PresentedAt       time.Time       `gorm:"column:presented_at;type:timestamptz;not null;default:clock_timestamp()" json:"-"`
	OperationID       uint64          `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	CalculationID     uint64          `gorm:"column:calculation_id;type:bigint;not null" json:"-"`
}

func (EstimatePresentation) TableName() string { return "estimate_presentations" }

// EstimateReply maps the current estimate_replies storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type EstimateReply struct {
	ID             uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID       uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	PresentationID uint64    `gorm:"column:presentation_id;type:bigint;not null" json:"-"`
	EstimateID     uint64    `gorm:"column:estimate_id;type:bigint;not null" json:"-"`
	Response       string    `gorm:"column:response;type:text;not null" json:"-"`
	ResponseNote   *string   `gorm:"column:response_note;type:text" json:"-"`
	RecordedBy     uint64    `gorm:"column:recorded_by;type:bigint;not null" json:"-"`
	RecordedAt     time.Time `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
	OperationID    uint64    `gorm:"column:operation_id;type:bigint;not null" json:"-"`
}

func (EstimateReply) TableName() string { return "estimate_replies" }

// EstimateSuccessorLink maps the current estimate_successor_links storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type EstimateSuccessorLink struct {
	ClinicID            uint64 `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	OriginalEstimateID  uint64 `gorm:"column:original_estimate_id;type:bigint;not null" json:"-"`
	PresentationID      uint64 `gorm:"column:presentation_id;type:bigint;not null" json:"-"`
	SuccessorEstimateID uint64 `gorm:"column:successor_estimate_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	Reason              string `gorm:"column:reason;type:text;not null" json:"-"`
	OperationID         uint64 `gorm:"column:operation_id;type:bigint;not null" json:"-"`
}

func (EstimateSuccessorLink) TableName() string { return "estimate_successor_links" }

// FinanceEventIngestion maps the current finance_event_ingestions storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type FinanceEventIngestion struct {
	ClinicID      uint64    `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	ConsumerKind  string    `gorm:"column:consumer_kind;type:text;primaryKey;autoIncrement:false;not null" json:"-"`
	SourceEventID uint64    `gorm:"column:source_event_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	SourceDigest  []byte    `gorm:"column:source_digest;type:bytea;not null" json:"-"`
	IngestedAt    time.Time `gorm:"column:ingested_at;type:timestamptz;not null;default:clock_timestamp()" json:"-"`
}

func (FinanceEventIngestion) TableName() string { return "finance_event_ingestions" }

// FinanceSourceEvent maps the current finance_source_events storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type FinanceSourceEvent struct {
	ID                     uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID               uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	SourceOwner            string          `gorm:"column:source_owner;type:text;not null" json:"-"`
	SourceEventKey         uuid.UUID       `gorm:"column:source_event_key;type:uuid;not null" json:"-"`
	FundRecordCorrectionID *uint64         `gorm:"column:fund_record_correction_id;type:bigint" json:"-"`
	SourceVersion          int64           `gorm:"column:source_version;type:bigint;not null" json:"-"`
	ContentDigest          []byte          `gorm:"column:content_digest;type:bytea;not null" json:"-"`
	BillingChangeEventID   *uint64         `gorm:"column:billing_change_event_id;type:bigint" json:"-"`
	FundID                 *uint64         `gorm:"column:fund_id;type:bigint" json:"-"`
	FundConsumptionEventID *uint64         `gorm:"column:fund_consumption_event_id;type:bigint" json:"-"`
	DepositEntryID         *uint64         `gorm:"column:deposit_entry_id;type:bigint" json:"-"`
	CashAdjustmentEventID  *uint64         `gorm:"column:cash_adjustment_event_id;type:bigint" json:"-"`
	CorrectedEventID       *uint64         `gorm:"column:corrected_event_id;type:bigint" json:"-"`
	OccurredAt             time.Time       `gorm:"column:occurred_at;type:timestamptz;not null" json:"-"`
	ActualMethodID         *uint64         `gorm:"column:actual_method_id;type:bigint" json:"-"`
	ActualMovementAmount   int64           `gorm:"column:actual_movement_amount;type:bigint;not null" json:"-"`
	CashMovementAmount     int64           `gorm:"column:cash_movement_amount;type:bigint;not null" json:"-"`
	ReceivableDelta        int64           `gorm:"column:receivable_delta;type:bigint;not null" json:"-"`
	ContextSnapshot        json.RawMessage `gorm:"column:context_snapshot;type:jsonb;not null" json:"-"`
	RecordedAt             time.Time       `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
	RecordedBy             *uint64         `gorm:"column:recorded_by;type:bigint" json:"-"`
	Reason                 *string         `gorm:"column:reason;type:text" json:"-"`
}

func (FinanceSourceEvent) TableName() string { return "finance_source_events" }

// FinancialDocumentArtifact maps the current financial_document_artifacts storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type FinancialDocumentArtifact struct {
	ClinicID        uint64    `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	EditionID       uint64    `gorm:"column:edition_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	RendererVersion string    `gorm:"column:renderer_version;type:text;not null" json:"-"`
	ObjectKey       string    `gorm:"column:object_key;type:text;not null" json:"-"`
	ContentSHA256   []byte    `gorm:"column:content_sha256;type:bytea;not null" json:"-"`
	CreatedAt       time.Time `gorm:"column:created_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (FinancialDocumentArtifact) TableName() string { return "financial_document_artifacts" }

// FinancialDocumentEdition maps the current financial_document_editions storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type FinancialDocumentEdition struct {
	ID              uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID        uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	DocumentID      uint64          `gorm:"column:document_id;type:bigint;not null" json:"-"`
	Edition         int64           `gorm:"column:edition;type:bigint;not null" json:"-"`
	CalculationID   *uint64         `gorm:"column:calculation_id;type:bigint" json:"-"`
	BillingID       *uint64         `gorm:"column:billing_id;type:bigint" json:"-"`
	SourceSnapshot  json.RawMessage `gorm:"column:source_snapshot;type:jsonb;not null" json:"-"`
	DisplaySnapshot json.RawMessage `gorm:"column:display_snapshot;type:jsonb;not null" json:"-"`
	LayoutVersion   string          `gorm:"column:layout_version;type:text;not null" json:"-"`
	IssuedBy        uint64          `gorm:"column:issued_by;type:bigint;not null" json:"-"`
	IssuedAt        time.Time       `gorm:"column:issued_at;type:timestamptz;not null;default:clock_timestamp()" json:"-"`
	OperationID     uint64          `gorm:"column:operation_id;type:bigint;not null" json:"-"`
}

func (FinancialDocumentEdition) TableName() string { return "financial_document_editions" }

// FinancialDocumentIdentity maps the current financial_document_identities storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type FinancialDocumentIdentity struct {
	ID                        uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID                  uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	DocumentKind              string    `gorm:"column:document_kind;type:text;not null" json:"-"`
	BillingID                 *uint64   `gorm:"column:billing_id;type:bigint" json:"-"`
	FundID                    *uint64   `gorm:"column:fund_id;type:bigint" json:"-"`
	OwnerID                   uint64    `gorm:"column:owner_id;type:bigint;not null" json:"-"`
	RefundEventID             *uint64   `gorm:"column:refund_event_id;type:bigint" json:"-"`
	OriginalTransactionNumber string    `gorm:"column:original_transaction_number;type:text;not null" json:"-"`
	CreatedAt                 time.Time `gorm:"column:created_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (FinancialDocumentIdentity) TableName() string { return "financial_document_identities" }

// FundConsumptionEvent maps the current fund_consumption_events storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type FundConsumptionEvent struct {
	ID                   uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID             uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	OwnerID              uint64    `gorm:"column:owner_id;type:bigint;not null" json:"-"`
	FundID               uint64    `gorm:"column:fund_id;type:bigint;not null" json:"-"`
	EventKind            string    `gorm:"column:event_kind;type:text;not null" json:"-"`
	Amount               int64     `gorm:"column:amount;type:bigint;not null" json:"-"`
	BillingID            *uint64   `gorm:"column:billing_id;type:bigint" json:"-"`
	SourceAllocationID   *uint64   `gorm:"column:source_allocation_id;type:bigint" json:"-"`
	ActualRefundMethodID *uint64   `gorm:"column:actual_refund_method_id;type:bigint" json:"-"`
	SourceDepositEntryID *uint64   `gorm:"column:source_deposit_entry_id;type:bigint" json:"-"`
	OccurredAt           time.Time `gorm:"column:occurred_at;type:timestamptz;not null" json:"-"`
	OperationID          uint64    `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	Ordinal              int       `gorm:"column:ordinal;type:integer;not null" json:"-"`
	RecordedBy           uint64    `gorm:"column:recorded_by;type:bigint;not null" json:"-"`
	RecordedAt           time.Time `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
	Reason               *string   `gorm:"column:reason;type:text" json:"-"`
}

func (FundConsumptionEvent) TableName() string { return "fund_consumption_events" }

// FundRecordCorrection maps the current fund_record_corrections storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type FundRecordCorrection struct {
	ID                         uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID                   uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	OwnerID                    uint64          `gorm:"column:owner_id;type:bigint;not null" json:"-"`
	FundID                     uint64          `gorm:"column:fund_id;type:bigint;not null" json:"-"`
	Revision                   int64           `gorm:"column:revision;type:bigint;not null" json:"-"`
	CorrectedMethodID          uint64          `gorm:"column:corrected_method_id;type:bigint;not null" json:"-"`
	CorrectedReceivedAt        time.Time       `gorm:"column:corrected_received_at;type:timestamptz;not null" json:"-"`
	BeforeValues               json.RawMessage `gorm:"column:before_values;type:jsonb;not null" json:"-"`
	Reason                     string          `gorm:"column:reason;type:text;not null" json:"-"`
	OperationID                uint64          `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	RecordedBy                 uint64          `gorm:"column:recorded_by;type:bigint;not null" json:"-"`
	RecordedAt                 time.Time       `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
	CorrectedReceivedAmount    int64           `gorm:"column:corrected_received_amount;type:bigint;not null" json:"-"`
	CorrectedTenderedAmount    *int64          `gorm:"column:corrected_tendered_amount;type:bigint" json:"-"`
	CorrectedChangeGivenAmount *int64          `gorm:"column:corrected_change_given_amount;type:bigint" json:"-"`
}

func (FundRecordCorrection) TableName() string { return "fund_record_corrections" }

// OwnerDepositAccount maps the current owner_deposit_accounts storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type OwnerDepositAccount struct {
	ClinicID  uint64    `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	OwnerID   uint64    `gorm:"column:owner_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	Revision  int64     `gorm:"column:revision;type:bigint;not null;default:1" json:"-"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (OwnerDepositAccount) TableName() string { return "owner_deposit_accounts" }

// OwnerDepositEntry maps the current owner_deposit_entries storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type OwnerDepositEntry struct {
	ID                     uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID               uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	OwnerID                uint64    `gorm:"column:owner_id;type:bigint;not null" json:"-"`
	OperationKey           uuid.UUID `gorm:"column:operation_key;type:uuid;not null" json:"-"`
	Ordinal                int       `gorm:"column:ordinal;type:integer;not null" json:"-"`
	AccountRevision        int64     `gorm:"column:account_revision;type:bigint;not null" json:"-"`
	EntryKind              string    `gorm:"column:entry_kind;type:text;not null" json:"-"`
	Amount                 int64     `gorm:"column:amount;type:bigint;not null" json:"-"`
	SourceEntryID          *uint64   `gorm:"column:source_entry_id;type:bigint" json:"-"`
	BillingID              *uint64   `gorm:"column:billing_id;type:bigint" json:"-"`
	AppointmentID          *uint64   `gorm:"column:appointment_id;type:bigint" json:"-"`
	OccurredAt             time.Time `gorm:"column:occurred_at;type:timestamptz;not null" json:"-"`
	RecordedAt             time.Time `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
	FundRecordCorrectionID *uint64   `gorm:"column:fund_record_correction_id;type:bigint" json:"-"`
	CorrectionReason       *string   `gorm:"column:correction_reason;type:text" json:"-"`
}

func (OwnerDepositEntry) TableName() string { return "owner_deposit_entries" }

// OwnerDepositOperation maps the current owner_deposit_operations storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type OwnerDepositOperation struct {
	ClinicID                 uint64    `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	OwnerID                  uint64    `gorm:"column:owner_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	OperationKey             uuid.UUID `gorm:"column:operation_key;type:uuid;primaryKey;autoIncrement:false;not null" json:"-"`
	RequestHash              []byte    `gorm:"column:request_hash;type:bytea;not null" json:"-"`
	ExpectedAccountRevision  int64     `gorm:"column:expected_account_revision;type:bigint;not null" json:"-"`
	ResultingAccountRevision int64     `gorm:"column:resulting_account_revision;type:bigint;not null" json:"-"`
	ExpectedEntryCount       int       `gorm:"column:expected_entry_count;type:integer;not null" json:"-"`
	BalanceAfter             int64     `gorm:"column:balance_after;type:bigint;not null" json:"-"`
	PerformedBy              uint64    `gorm:"column:performed_by;type:bigint;not null" json:"-"`
	CommittedAt              time.Time `gorm:"column:committed_at;type:timestamptz;not null;default:clock_timestamp()" json:"-"`
}

func (OwnerDepositOperation) TableName() string { return "owner_deposit_operations" }

// PaymentFund maps the current payment_funds storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type PaymentFund struct {
	ID                      uint64           `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID                uint64           `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	OwnerID                 uint64           `gorm:"column:owner_id;type:bigint;not null" json:"-"`
	ReceivedAmount          int64            `gorm:"column:received_amount;type:bigint;not null" json:"-"`
	PaymentMethodID         uint64           `gorm:"column:payment_method_id;type:bigint;not null" json:"-"`
	ReceivedAt              time.Time        `gorm:"column:received_at;type:timestamptz;not null" json:"-"`
	OperationID             uint64           `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	RecordedBy              uint64           `gorm:"column:recorded_by;type:bigint;not null" json:"-"`
	RecordedAt              time.Time        `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
	TenderedAmount          *int64           `gorm:"column:tendered_amount;type:bigint" json:"-"`
	ChangeGivenAmount       *int64           `gorm:"column:change_given_amount;type:bigint" json:"-"`
	OriginalSourceReference *json.RawMessage `gorm:"column:original_source_reference;type:jsonb" json:"-"`
}

func (PaymentFund) TableName() string { return "payment_funds" }
