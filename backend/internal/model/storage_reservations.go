package model

import (
	"encoding/json"
	"github.com/lib/pq"
	"time"
)

// Current migration is authoritative. These are persistence values, not public DTOs or workflow implementations.

// AdmissionRecord maps the current admission_records storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type AdmissionRecord struct {
	ID                uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID          uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	HospitalizationID uint64    `gorm:"column:hospitalization_id;type:bigint;not null" json:"-"`
	StayCaseID        uint64    `gorm:"column:stay_case_id;type:bigint;not null" json:"-"`
	AdmissionEventID  uint64    `gorm:"column:admission_event_id;type:bigint;not null" json:"-"`
	RecordedAt        time.Time `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (AdmissionRecord) TableName() string { return "admission_records" }

// AppointmentChangeEvent maps the current appointment_change_events storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type AppointmentChangeEvent struct {
	ID               uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID         uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	AppointmentID    uint64          `gorm:"column:appointment_id;type:bigint;not null" json:"-"`
	ResultingVersion int64           `gorm:"column:resulting_version;type:bigint;not null" json:"-"`
	Intent           string          `gorm:"column:intent;type:text;not null" json:"-"`
	BeforeValues     json.RawMessage `gorm:"column:before_values;type:jsonb;not null" json:"-"`
	AfterValues      json.RawMessage `gorm:"column:after_values;type:jsonb;not null" json:"-"`
	Reason           *string         `gorm:"column:reason;type:text" json:"-"`
	OperationID      uint64          `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	RecordedBy       uint64          `gorm:"column:recorded_by;type:bigint;not null" json:"-"`
	RecordedAt       time.Time       `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (AppointmentChangeEvent) TableName() string { return "appointment_change_events" }

// HotelCapacitySetting maps the current hotel_capacity_settings storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type HotelCapacitySetting struct {
	ClinicID          uint64    `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	MaxConcurrentPets int       `gorm:"column:max_concurrent_pets;type:integer;not null" json:"-"`
	Version           int64     `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
	UpdatedBy         uint64    `gorm:"column:updated_by;type:bigint;not null" json:"-"`
	UpdatedAt         time.Time `gorm:"column:updated_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (HotelCapacitySetting) TableName() string { return "hotel_capacity_settings" }

// PatientOperationGate maps the current patient_operation_gates storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type PatientOperationGate struct {
	ClinicID   uint64 `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	PetID      uint64 `gorm:"column:pet_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	Generation int64  `gorm:"column:generation;type:bigint;not null;default:1" json:"-"`
}

func (PatientOperationGate) TableName() string { return "patient_operation_gates" }

// ReservationCapacityGate maps the current reservation_capacity_gates storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type ReservationCapacityGate struct {
	ClinicID     uint64 `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	CapacityKind string `gorm:"column:capacity_kind;type:text;primaryKey;autoIncrement:false;not null" json:"-"`
	ResourceKey  string `gorm:"column:resource_key;type:text;primaryKey;autoIncrement:false;not null" json:"-"`
	Generation   int64  `gorm:"column:generation;type:bigint;not null;default:1" json:"-"`
}

func (ReservationCapacityGate) TableName() string { return "reservation_capacity_gates" }

// ReservationContractState maps the current reservation_contract_states storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type ReservationContractState struct {
	ClinicID        uint64         `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	ContractKind    string         `gorm:"column:contract_kind;type:text;primaryKey;autoIncrement:false;not null" json:"-"`
	Revision        int64          `gorm:"column:revision;type:bigint;not null" json:"-"`
	CutoverState    string         `gorm:"column:cutover_state;type:text;not null" json:"-"`
	ApplicableKinds pq.StringArray `gorm:"column:applicable_kinds;type:text[];not null" json:"-"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (ReservationContractState) TableName() string { return "reservation_contract_states" }

// ReservationEligibilityEvidence maps the current reservation_eligibility_evidence storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type ReservationEligibilityEvidence struct {
	ID              uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID        uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	SourceKind      string          `gorm:"column:source_kind;type:text;not null" json:"-"`
	SourceReference string          `gorm:"column:source_reference;type:text;not null" json:"-"`
	OriginalValues  json.RawMessage `gorm:"column:original_values;type:jsonb;not null" json:"-"`
	CapturedAt      time.Time       `gorm:"column:captured_at;type:timestamptz;not null" json:"-"`
	CapturedBy      *uint64         `gorm:"column:captured_by;type:bigint" json:"-"`
}

func (ReservationEligibilityEvidence) TableName() string { return "reservation_eligibility_evidence" }

// ReservationNotificationEligibility maps the current reservation_notification_eligibility storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type ReservationNotificationEligibility struct {
	ClinicID           uint64     `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	AppointmentID      uint64     `gorm:"column:appointment_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	OwnerID            uint64     `gorm:"column:owner_id;type:bigint;not null" json:"-"`
	BindingRevision    int64      `gorm:"column:binding_revision;type:bigint;not null" json:"-"`
	AppointmentVersion int64      `gorm:"column:appointment_version;type:bigint;not null" json:"-"`
	State              string     `gorm:"column:state;type:text;not null" json:"-"`
	ConfirmedBy        *uint64    `gorm:"column:confirmed_by;type:bigint" json:"-"`
	ConfirmedAt        *time.Time `gorm:"column:confirmed_at;type:timestamptz" json:"-"`
	OperationID        uint64     `gorm:"column:operation_id;type:bigint;not null" json:"-"`
}

func (ReservationNotificationEligibility) TableName() string {
	return "reservation_notification_eligibility"
}

// ReservationTypeNameRevision maps the current reservation_type_name_revisions storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type ReservationTypeNameRevision struct {
	ClinicID          uint64    `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	ReservationTypeID uint64    `gorm:"column:reservation_type_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	Revision          int64     `gorm:"column:revision;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	PreviousName      *string   `gorm:"column:previous_name;type:text" json:"-"`
	DisplayName       string    `gorm:"column:display_name;type:text;not null" json:"-"`
	OperationID       uint64    `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	ChangedBy         uint64    `gorm:"column:changed_by;type:bigint;not null" json:"-"`
	ChangedAt         time.Time `gorm:"column:changed_at;type:timestamptz;not null;default:clock_timestamp()" json:"-"`
}

func (ReservationTypeNameRevision) TableName() string { return "reservation_type_name_revisions" }

// ShiftBulkOperation maps the current shift_bulk_operations storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type ShiftBulkOperation struct {
	ClinicID       uint64          `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	OperationID    uint64          `gorm:"column:operation_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	OperationKind  string          `gorm:"column:operation_kind;type:text;not null" json:"-"`
	SourceSnapshot json.RawMessage `gorm:"column:source_snapshot;type:jsonb;not null" json:"-"`
	ConfirmedRows  json.RawMessage `gorm:"column:confirmed_rows;type:jsonb;not null" json:"-"`
	PreviewDigest  []byte          `gorm:"column:preview_digest;type:bytea;not null" json:"-"`
}

func (ShiftBulkOperation) TableName() string { return "shift_bulk_operations" }

// StaffLeaveChange maps the current staff_leave_changes storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type StaffLeaveChange struct {
	ID               uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID         uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	LeavePeriodID    uint64          `gorm:"column:leave_period_id;type:bigint;not null" json:"-"`
	ResultingVersion int64           `gorm:"column:resulting_version;type:bigint;not null" json:"-"`
	BeforeValues     json.RawMessage `gorm:"column:before_values;type:jsonb;not null" json:"-"`
	AfterValues      json.RawMessage `gorm:"column:after_values;type:jsonb;not null" json:"-"`
	OperationID      uint64          `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	RecordedBy       uint64          `gorm:"column:recorded_by;type:bigint;not null" json:"-"`
	RecordedAt       time.Time       `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (StaffLeaveChange) TableName() string { return "staff_leave_changes" }

// StaffLeavePeriod maps the current staff_leave_periods storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type StaffLeavePeriod struct {
	ID          uint64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID    uint64     `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	StaffID     uint64     `gorm:"column:staff_id;type:bigint;not null" json:"-"`
	StartDate   time.Time  `gorm:"column:start_date;type:date;not null" json:"-"`
	EndDate     *time.Time `gorm:"column:end_date;type:date" json:"-"`
	EndBoundary *string    `gorm:"column:end_boundary;type:text" json:"-"`
	Version     int64      `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
	Provenance  *string    `gorm:"column:provenance;type:text" json:"-"`
	CreatedAt   time.Time  `gorm:"column:created_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;type:timestamptz" json:"-"`
}

func (StaffLeavePeriod) TableName() string { return "staff_leave_periods" }

// StayCase maps the current stay_cases storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type StayCase struct {
	ID             uint64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID       uint64     `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	OwnerID        uint64     `gorm:"column:owner_id;type:bigint;not null" json:"-"`
	PetID          uint64     `gorm:"column:pet_id;type:bigint;not null" json:"-"`
	StayKind       string     `gorm:"column:stay_kind;type:text;not null" json:"-"`
	PlannedEntryAt *time.Time `gorm:"column:planned_entry_at;type:timestamptz" json:"-"`
	PlannedExitAt  *time.Time `gorm:"column:planned_exit_at;type:timestamptz" json:"-"`
	Version        int64      `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
	CreatedAt      time.Time  `gorm:"column:created_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
	DeletedAt      *time.Time `gorm:"column:deleted_at;type:timestamptz" json:"-"`
}

func (StayCase) TableName() string { return "stay_cases" }

// StayExtensionProposal maps the current stay_extension_proposals storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type StayExtensionProposal struct {
	ID                  uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID            uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	StayCaseID          uint64    `gorm:"column:stay_case_id;type:bigint;not null" json:"-"`
	ExpectedStayVersion int64     `gorm:"column:expected_stay_version;type:bigint;not null" json:"-"`
	RequestedExitAt     time.Time `gorm:"column:requested_exit_at;type:timestamptz;not null" json:"-"`
	State               string    `gorm:"column:state;type:text;not null" json:"-"`
	Version             int64     `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
	RequestedBy         uint64    `gorm:"column:requested_by;type:bigint;not null" json:"-"`
	ConfirmedBy         *uint64   `gorm:"column:confirmed_by;type:bigint" json:"-"`
	OperationID         uint64    `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	RecordedAt          time.Time `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (StayExtensionProposal) TableName() string { return "stay_extension_proposals" }

// StayMovementEvent maps the current stay_movement_events storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type StayMovementEvent struct {
	ID                 uint64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID           uint64     `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	StayCaseID         uint64     `gorm:"column:stay_case_id;type:bigint;not null" json:"-"`
	HospitalizationID  *uint64    `gorm:"column:hospitalization_id;type:bigint" json:"-"`
	Revision           int64      `gorm:"column:revision;type:bigint;not null" json:"-"`
	EventKind          string     `gorm:"column:event_kind;type:text;not null" json:"-"`
	ActualAt           *time.Time `gorm:"column:actual_at;type:timestamptz" json:"-"`
	CorrectedEventID   *uint64    `gorm:"column:corrected_event_id;type:bigint" json:"-"`
	Reason             *string    `gorm:"column:reason;type:text" json:"-"`
	DoctorID           *uint64    `gorm:"column:doctor_id;type:bigint" json:"-"`
	AdmissionRecordRef *int64     `gorm:"column:admission_record_ref;type:bigint" json:"-"`
	OperationID        uint64     `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	RecordedBy         uint64     `gorm:"column:recorded_by;type:bigint;not null" json:"-"`
	RecordedAt         time.Time  `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (StayMovementEvent) TableName() string { return "stay_movement_events" }

// StayPlanChangeEvent maps the current stay_plan_change_events storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type StayPlanChangeEvent struct {
	ID                uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID          uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	HospitalizationID uint64          `gorm:"column:hospitalization_id;type:bigint;not null" json:"-"`
	CarePlanItemID    *uint64         `gorm:"column:care_plan_item_id;type:bigint" json:"-"`
	TreatmentPlanID   *uint64         `gorm:"column:treatment_plan_id;type:bigint" json:"-"`
	Revision          int64           `gorm:"column:revision;type:bigint;not null" json:"-"`
	Intent            string          `gorm:"column:intent;type:text;not null" json:"-"`
	BeforeValues      json.RawMessage `gorm:"column:before_values;type:jsonb;not null" json:"-"`
	AfterValues       json.RawMessage `gorm:"column:after_values;type:jsonb;not null" json:"-"`
	Reason            *string         `gorm:"column:reason;type:text" json:"-"`
	ConfirmedBy       *uint64         `gorm:"column:confirmed_by;type:bigint" json:"-"`
	PreviousEventID   *uint64         `gorm:"column:previous_event_id;type:bigint" json:"-"`
	OperationID       uint64          `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	RecordedBy        uint64          `gorm:"column:recorded_by;type:bigint;not null" json:"-"`
	RecordedAt        time.Time       `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (StayPlanChangeEvent) TableName() string { return "stay_plan_change_events" }

// StayPriceAgreement maps the current stay_price_agreements storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type StayPriceAgreement struct {
	ID                       uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID                 uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	StayCaseID               uint64    `gorm:"column:stay_case_id;type:bigint;not null" json:"-"`
	Revision                 int64     `gorm:"column:revision;type:bigint;not null" json:"-"`
	NormalChargeAmount       *int64    `gorm:"column:normal_charge_amount;type:bigint" json:"-"`
	RequestedDepositAmount   *int64    `gorm:"column:requested_deposit_amount;type:bigint" json:"-"`
	CancellationChargeAmount *int64    `gorm:"column:cancellation_charge_amount;type:bigint" json:"-"`
	Terms                    *string   `gorm:"column:terms;type:text" json:"-"`
	OperationID              uint64    `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	ConfirmedBy              uint64    `gorm:"column:confirmed_by;type:bigint;not null" json:"-"`
	ConfirmedAt              time.Time `gorm:"column:confirmed_at;type:timestamptz;not null;default:clock_timestamp()" json:"-"`
}

func (StayPriceAgreement) TableName() string { return "stay_price_agreements" }

// TrimmingBookingRevisionOption maps the current trimming_booking_revision_options storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type TrimmingBookingRevisionOption struct {
	ClinicID          uint64 `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	BookingRevisionID uint64 `gorm:"column:booking_revision_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	OptionID          uint64 `gorm:"column:option_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	OptionVersion     int64  `gorm:"column:option_version;type:bigint;not null" json:"-"`
	DurationMinutes   int    `gorm:"column:duration_minutes;type:integer;not null" json:"-"`
}

func (TrimmingBookingRevisionOption) TableName() string { return "trimming_booking_revision_options" }

// TrimmingBookingRevision maps the current trimming_booking_revisions storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type TrimmingBookingRevision struct {
	ID                    uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID              uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	AppointmentID         uint64    `gorm:"column:appointment_id;type:bigint;not null" json:"-"`
	PetID                 uint64    `gorm:"column:pet_id;type:bigint;not null" json:"-"`
	Revision              int64     `gorm:"column:revision;type:bigint;not null" json:"-"`
	CourseID              *uint64   `gorm:"column:course_id;type:bigint" json:"-"`
	CourseVersion         *int64    `gorm:"column:course_version;type:bigint" json:"-"`
	CourseDurationMinutes *int      `gorm:"column:course_duration_minutes;type:integer" json:"-"`
	RuleSetID             *uint64   `gorm:"column:rule_set_id;type:bigint" json:"-"`
	RuleSetVersion        *int64    `gorm:"column:rule_set_version;type:bigint" json:"-"`
	TotalDurationMinutes  int       `gorm:"column:total_duration_minutes;type:integer;not null" json:"-"`
	OptionCount           int       `gorm:"column:option_count;type:integer;not null" json:"-"`
	RecordedBy            uint64    `gorm:"column:recorded_by;type:bigint;not null" json:"-"`
	RecordedAt            time.Time `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (TrimmingBookingRevision) TableName() string { return "trimming_booking_revisions" }

// TrimmingCourseIncludedService maps the current trimming_course_included_services storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type TrimmingCourseIncludedService struct {
	ClinicID          uint64     `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	CourseID          uint64     `gorm:"column:course_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	ServiceIdentityID uint64     `gorm:"column:service_identity_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	CreatedAt         time.Time  `gorm:"column:created_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
	DeletedAt         *time.Time `gorm:"column:deleted_at;type:timestamptz" json:"-"`
}

func (TrimmingCourseIncludedService) TableName() string { return "trimming_course_included_services" }

// TrimmingDetailRevisionLine maps the current trimming_detail_revision_lines storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type TrimmingDetailRevisionLine struct {
	ClinicID               uint64       `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	DetailID               uint64       `gorm:"column:detail_id;type:bigint;not null" json:"-"`
	DetailRevisionID       uint64       `gorm:"column:detail_revision_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	SourceLineID           uint64       `gorm:"column:source_line_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	MasterVersion          int64        `gorm:"column:master_version;type:bigint;not null" json:"-"`
	AdoptedMasterUnitPrice *int64       `gorm:"column:adopted_master_unit_price;type:bigint" json:"-"`
	UnitPrice              *int64       `gorm:"column:unit_price;type:bigint" json:"-"`
	Quantity               ExactNumeric `gorm:"column:quantity;type:numeric;not null" json:"-"`
	DiscountRate           ExactNumeric `gorm:"column:discount_rate;type:numeric;not null" json:"-"`
	DiscountAmount         int64        `gorm:"column:discount_amount;type:bigint;not null" json:"-"`
	DiscountStage          string       `gorm:"column:discount_stage;type:text;not null" json:"-"`
}

func (TrimmingDetailRevisionLine) TableName() string { return "trimming_detail_revision_lines" }

// TrimmingDetailRevision maps the current trimming_detail_revisions storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type TrimmingDetailRevision struct {
	ID           uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID     uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	DetailID     uint64          `gorm:"column:detail_id;type:bigint;not null" json:"-"`
	Revision     int64           `gorm:"column:revision;type:bigint;not null" json:"-"`
	LineCount    int             `gorm:"column:line_count;type:integer;not null" json:"-"`
	DetailValues json.RawMessage `gorm:"column:detail_values;type:jsonb;not null" json:"-"`
	RecordedBy   uint64          `gorm:"column:recorded_by;type:bigint;not null" json:"-"`
	RecordedAt   time.Time       `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (TrimmingDetailRevision) TableName() string { return "trimming_detail_revisions" }

// TrimmingImageDeletion maps the current trimming_image_deletions storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type TrimmingImageDeletion struct {
	ClinicID         uint64     `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	ObjectKey        string     `gorm:"column:object_key;type:text;primaryKey;autoIncrement:false;not null" json:"-"`
	ActorStaffID     uint64     `gorm:"column:actor_staff_id;type:bigint;not null" json:"-"`
	CreatedAt        time.Time  `gorm:"column:created_at;type:timestamptz;not null;default:NOW();autoCreateTime:false;autoUpdateTime:false" json:"-"`
	StorageDeletedAt *time.Time `gorm:"column:storage_deleted_at;type:timestamptz" json:"-"`
	LastAttemptAt    *time.Time `gorm:"column:last_attempt_at;type:timestamptz" json:"-"`
}

func (TrimmingImageDeletion) TableName() string { return "trimming_image_deletions" }

// TrimmingOptionRuleSet maps the current trimming_option_rule_sets storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type TrimmingOptionRuleSet struct {
	ID              uint64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID        uint64     `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	CourseID        *uint64    `gorm:"column:course_id;type:bigint" json:"-"`
	ContractVersion int64      `gorm:"column:contract_version;type:bigint;not null;default:1" json:"-"`
	CreatedAt       time.Time  `gorm:"column:created_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
	DeletedAt       *time.Time `gorm:"column:deleted_at;type:timestamptz" json:"-"`
}

func (TrimmingOptionRuleSet) TableName() string { return "trimming_option_rule_sets" }

// TrimmingOptionRule maps the current trimming_option_rules storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type TrimmingOptionRule struct {
	ClinicID  uint64     `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	RuleSetID uint64     `gorm:"column:rule_set_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	OptionID  uint64     `gorm:"column:option_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	CreatedAt time.Time  `gorm:"column:created_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
	DeletedAt *time.Time `gorm:"column:deleted_at;type:timestamptz" json:"-"`
}

func (TrimmingOptionRule) TableName() string { return "trimming_option_rules" }

// TrimmingServiceIdentity maps the current trimming_service_identities storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type TrimmingServiceIdentity struct {
	ID          uint64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID    uint64     `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	DisplayName *string    `gorm:"column:display_name;type:text" json:"-"`
	CreatedAt   time.Time  `gorm:"column:created_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;type:timestamptz" json:"-"`
}

func (TrimmingServiceIdentity) TableName() string { return "trimming_service_identities" }

// TrimmingServiceLine maps the current trimming_service_lines storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type TrimmingServiceLine struct {
	ID        uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID  uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	DetailID  uint64    `gorm:"column:detail_id;type:bigint;not null" json:"-"`
	CourseID  *uint64   `gorm:"column:course_id;type:bigint" json:"-"`
	OptionID  *uint64   `gorm:"column:option_id;type:bigint" json:"-"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (TrimmingServiceLine) TableName() string { return "trimming_service_lines" }
