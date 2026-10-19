package model

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

// Current migration is authoritative. These are persistence values, not public DTOs or workflow implementations.

// ActualVisitFact maps the current actual_visit_facts storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type ActualVisitFact struct {
	ID                       uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID                 uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	SourceKind               string          `gorm:"column:source_kind;type:text;not null" json:"-"`
	AppointmentChangeEventID *uint64         `gorm:"column:appointment_change_event_id;type:bigint" json:"-"`
	RecordCorrectionLinkID   *uint64         `gorm:"column:record_correction_link_id;type:bigint" json:"-"`
	OwnerID                  uint64          `gorm:"column:owner_id;type:bigint;not null" json:"-"`
	PetID                    uint64          `gorm:"column:pet_id;type:bigint;not null" json:"-"`
	VisitDateJST             time.Time       `gorm:"column:visit_date_jst;type:date;not null" json:"-"`
	SourceVersion            int64           `gorm:"column:source_version;type:bigint;not null" json:"-"`
	CorrectedFactID          *uint64         `gorm:"column:corrected_fact_id;type:bigint" json:"-"`
	Evidence                 json.RawMessage `gorm:"column:evidence;type:jsonb;not null" json:"-"`
	OperationID              uint64          `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	RecordedBy               uint64          `gorm:"column:recorded_by;type:bigint;not null" json:"-"`
	RecordedAt               time.Time       `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (ActualVisitFact) TableName() string { return "actual_visit_facts" }

// CheckupResultRevision maps the current checkup_result_revisions storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type CheckupResultRevision struct {
	ID              uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID        uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	CheckupID       uint64          `gorm:"column:checkup_id;type:bigint;not null" json:"-"`
	Revision        int64           `gorm:"column:revision;type:bigint;not null" json:"-"`
	PackageVersion  int64           `gorm:"column:package_version;type:bigint;not null" json:"-"`
	PackageSnapshot json.RawMessage `gorm:"column:package_snapshot;type:jsonb;not null" json:"-"`
	FieldsSnapshot  json.RawMessage `gorm:"column:fields_snapshot;type:jsonb;not null" json:"-"`
	RestoredFromID  *uint64         `gorm:"column:restored_from_id;type:bigint" json:"-"`
	Reason          *string         `gorm:"column:reason;type:text" json:"-"`
	OperationID     uint64          `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	RecordedBy      uint64          `gorm:"column:recorded_by;type:bigint;not null" json:"-"`
	RecordedAt      time.Time       `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (CheckupResultRevision) TableName() string { return "checkup_result_revisions" }

// CheckupSaveAttempt maps the current checkup_save_attempts storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type CheckupSaveAttempt struct {
	ID                 uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID           uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	PetID              uint64    `gorm:"column:pet_id;type:bigint;not null" json:"-"`
	PerformedDate      time.Time `gorm:"column:performed_date;type:date;not null" json:"-"`
	OperationID        uint64    `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	RequestFingerprint []byte    `gorm:"column:request_fingerprint;type:bytea;not null" json:"-"`
	ActorID            uint64    `gorm:"column:actor_id;type:bigint;not null" json:"-"`
	CreatedAt          time.Time `gorm:"column:created_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (CheckupSaveAttempt) TableName() string { return "checkup_save_attempts" }

// CheckupSaveStageResult maps the current checkup_save_stage_results storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type CheckupSaveStageResult struct {
	ClinicID        uint64    `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	AttemptID       uint64    `gorm:"column:attempt_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	Stage           string    `gorm:"column:stage;type:text;primaryKey;autoIncrement:false;not null" json:"-"`
	OperationID     uint64    `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	MedicalRecordID uint64    `gorm:"column:medical_record_id;type:bigint;not null" json:"-"`
	CheckupID       *uint64   `gorm:"column:checkup_id;type:bigint" json:"-"`
	ResultRevision  *int64    `gorm:"column:result_revision;type:bigint" json:"-"`
	RecordedAt      time.Time `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (CheckupSaveStageResult) TableName() string { return "checkup_save_stage_results" }

// ClinicalPlanDiagnose maps the current clinical_plan_diagnoses storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type ClinicalPlanDiagnose struct {
	ID                    uint64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID              uint64     `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	ClinicalPlanID        uint64     `gorm:"column:clinical_plan_id;type:bigint;not null" json:"-"`
	DiagnosisSlot         int16      `gorm:"column:diagnosis_slot;type:smallint;not null" json:"-"`
	DiagnosisTypeID       *uint64    `gorm:"column:diagnosis_type_id;type:bigint" json:"-"`
	DiagnosisNameID       *uint64    `gorm:"column:diagnosis_name_id;type:bigint" json:"-"`
	DiagnosisNameSnapshot *string    `gorm:"column:diagnosis_name_snapshot;type:text" json:"-"`
	CategoryNameSnapshot  *string    `gorm:"column:category_name_snapshot;type:text" json:"-"`
	SortOrder             int        `gorm:"column:sort_order;type:integer;not null;default:0" json:"-"`
	Version               int64      `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
	CreatedAt             time.Time  `gorm:"column:created_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
	UpdatedAt             time.Time  `gorm:"column:updated_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
	DeletedAt             *time.Time `gorm:"column:deleted_at;type:timestamptz" json:"-"`
}

func (ClinicalPlanDiagnose) TableName() string { return "clinical_plan_diagnoses" }

// ExaminationFieldSemanticMapping maps the current examination_field_semantic_mappings storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type ExaminationFieldSemanticMapping struct {
	ID              uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID        uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	ExamTypeFieldID uint64    `gorm:"column:exam_type_field_id;type:bigint;not null" json:"-"`
	AnimalSpeciesID uint64    `gorm:"column:animal_species_id;type:bigint;not null" json:"-"`
	Meaning         string    `gorm:"column:meaning;type:text;not null" json:"-"`
	Revision        int64     `gorm:"column:revision;type:bigint;not null" json:"-"`
	VerifiedBy      uint64    `gorm:"column:verified_by;type:bigint;not null" json:"-"`
	VerifiedAt      time.Time `gorm:"column:verified_at;type:timestamptz;not null" json:"-"`
	Evidence        string    `gorm:"column:evidence;type:text;not null" json:"-"`
	OperationID     uint64    `gorm:"column:operation_id;type:bigint;not null" json:"-"`
}

func (ExaminationFieldSemanticMapping) TableName() string {
	return "examination_field_semantic_mappings"
}

// InitialIntakeEntry maps the current initial_intake_entries storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type InitialIntakeEntry struct {
	ClinicID                uint64     `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	IntakeRevisionID        uint64     `gorm:"column:intake_revision_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	EntryKey                uuid.UUID  `gorm:"column:entry_key;type:uuid;primaryKey;autoIncrement:false;not null" json:"-"`
	FieldKey                string     `gorm:"column:field_key;type:text;not null" json:"-"`
	FieldLabel              string     `gorm:"column:field_label;type:text;not null" json:"-"`
	AnswerState             *string    `gorm:"column:answer_state;type:text" json:"-"`
	Detail                  *string    `gorm:"column:detail;type:text" json:"-"`
	DrugDescription         *string    `gorm:"column:drug_description;type:text" json:"-"`
	SymptomDescription      *string    `gorm:"column:symptom_description;type:text" json:"-"`
	ReportedResult          *string    `gorm:"column:reported_result;type:text" json:"-"`
	TimePrecision           *string    `gorm:"column:time_precision;type:text" json:"-"`
	EventYear               *int       `gorm:"column:event_year;type:integer" json:"-"`
	EventMonth              *int       `gorm:"column:event_month;type:integer" json:"-"`
	EventDate               *time.Time `gorm:"column:event_date;type:date" json:"-"`
	TimeDescription         *string    `gorm:"column:time_description;type:text" json:"-"`
	ReferenceExamRevisionID *uint64    `gorm:"column:reference_exam_revision_id;type:bigint" json:"-"`
}

func (InitialIntakeEntry) TableName() string { return "initial_intake_entries" }

// InitialIntakeRevision maps the current initial_intake_revisions storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type InitialIntakeRevision struct {
	ID                   uint64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID             uint64     `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	MedicalRecordID      uint64     `gorm:"column:medical_record_id;type:bigint;not null" json:"-"`
	PetID                uint64     `gorm:"column:pet_id;type:bigint;not null" json:"-"`
	Revision             int64      `gorm:"column:revision;type:bigint;not null" json:"-"`
	SourceKind           string     `gorm:"column:source_kind;type:text;not null" json:"-"`
	SourceReference      *string    `gorm:"column:source_reference;type:text" json:"-"`
	HeardAt              *time.Time `gorm:"column:heard_at;type:timestamptz" json:"-"`
	CopiedFromRevisionID *uint64    `gorm:"column:copied_from_revision_id;type:bigint" json:"-"`
	EntryCount           int        `gorm:"column:entry_count;type:integer;not null" json:"-"`
	OperationID          uint64     `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	RecordedBy           uint64     `gorm:"column:recorded_by;type:bigint;not null" json:"-"`
	RecordedAt           time.Time  `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (InitialIntakeRevision) TableName() string { return "initial_intake_revisions" }

// InventoryMovementEvent maps the current inventory_movement_events storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type InventoryMovementEvent struct {
	ID                  uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID            uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	InventoryItemID     uint64    `gorm:"column:inventory_item_id;type:bigint;not null" json:"-"`
	EventKind           string    `gorm:"column:event_kind;type:text;not null" json:"-"`
	Delta               int64     `gorm:"column:delta;type:bigint;not null" json:"-"`
	BeforeQuantity      int64     `gorm:"column:before_quantity;type:bigint;not null" json:"-"`
	AfterQuantity       int64     `gorm:"column:after_quantity;type:bigint;not null" json:"-"`
	PreviousVersion     int64     `gorm:"column:previous_version;type:bigint;not null" json:"-"`
	ResultingVersion    int64     `gorm:"column:resulting_version;type:bigint;not null" json:"-"`
	ManagementUnit      string    `gorm:"column:management_unit;type:text;not null" json:"-"`
	OriginalIssueID     *uint64   `gorm:"column:original_issue_id;type:bigint" json:"-"`
	CorrectedEventID    *uint64   `gorm:"column:corrected_event_id;type:bigint" json:"-"`
	SourceTreatmentID   *uint64   `gorm:"column:source_treatment_id;type:bigint" json:"-"`
	SourceBillingItemID *uint64   `gorm:"column:source_billing_item_id;type:bigint" json:"-"`
	Reason              string    `gorm:"column:reason;type:text;not null" json:"-"`
	OccurredAt          time.Time `gorm:"column:occurred_at;type:timestamptz;not null" json:"-"`
	OperationID         uint64    `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	Ordinal             int       `gorm:"column:ordinal;type:integer;not null" json:"-"`
	RecordedBy          uint64    `gorm:"column:recorded_by;type:bigint;not null" json:"-"`
	RecordedAt          time.Time `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (InventoryMovementEvent) TableName() string { return "inventory_movement_events" }

// InventoryQuantityBase maps the current inventory_quantity_bases storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type InventoryQuantityBase struct {
	ID                  uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID            uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	InventoryItemID     uint64    `gorm:"column:inventory_item_id;type:bigint;not null" json:"-"`
	ExpectedItemVersion int64     `gorm:"column:expected_item_version;type:bigint;not null" json:"-"`
	ConfirmedQuantity   int64     `gorm:"column:confirmed_quantity;type:bigint;not null" json:"-"`
	ManagementUnit      string    `gorm:"column:management_unit;type:text;not null" json:"-"`
	LegacyReference     *string   `gorm:"column:legacy_reference;type:text" json:"-"`
	ConfirmedBy         uint64    `gorm:"column:confirmed_by;type:bigint;not null" json:"-"`
	ConfirmedAt         time.Time `gorm:"column:confirmed_at;type:timestamptz;not null" json:"-"`
	OperationID         uint64    `gorm:"column:operation_id;type:bigint;not null" json:"-"`
}

func (InventoryQuantityBase) TableName() string { return "inventory_quantity_bases" }

// MedicalRecordConsultationEvent maps the current medical_record_consultation_events storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type MedicalRecordConsultationEvent struct {
	ID              uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID        uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	MedicalRecordID uint64    `gorm:"column:medical_record_id;type:bigint;not null" json:"-"`
	Revision        int64     `gorm:"column:revision;type:bigint;not null" json:"-"`
	EventKind       string    `gorm:"column:event_kind;type:text;not null" json:"-"`
	PreviousEventID *uint64   `gorm:"column:previous_event_id;type:bigint" json:"-"`
	SourceKind      string    `gorm:"column:source_kind;type:text;not null" json:"-"`
	OccurredAt      time.Time `gorm:"column:occurred_at;type:timestamptz;not null" json:"-"`
	Reason          *string   `gorm:"column:reason;type:text" json:"-"`
	OperationID     uint64    `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	RecordedBy      uint64    `gorm:"column:recorded_by;type:bigint;not null" json:"-"`
	RecordedAt      time.Time `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (MedicalRecordConsultationEvent) TableName() string { return "medical_record_consultation_events" }

// MedicalRecordCorrectionLink maps the current medical_record_correction_links storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type MedicalRecordCorrectionLink struct {
	ID                  uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID            uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	SourceRecordID      uint64          `gorm:"column:source_record_id;type:bigint;not null" json:"-"`
	SourceRevision      int64           `gorm:"column:source_revision;type:bigint;not null" json:"-"`
	DestinationRecordID *uint64         `gorm:"column:destination_record_id;type:bigint" json:"-"`
	Intent              string          `gorm:"column:intent;type:text;not null" json:"-"`
	ClinicalDate        *time.Time      `gorm:"column:clinical_date;type:date" json:"-"`
	Reason              string          `gorm:"column:reason;type:text;not null" json:"-"`
	ConfirmedBy         *uint64         `gorm:"column:confirmed_by;type:bigint" json:"-"`
	OperationID         uint64          `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	TransferredFields   json.RawMessage `gorm:"column:transferred_fields;type:jsonb;not null" json:"-"`
	RecordedBy          uint64          `gorm:"column:recorded_by;type:bigint;not null" json:"-"`
	RecordedAt          time.Time       `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (MedicalRecordCorrectionLink) TableName() string { return "medical_record_correction_links" }

// MedicalRecordDependencyVersion maps the current medical_record_dependency_versions storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type MedicalRecordDependencyVersion struct {
	ClinicID        uint64 `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	MedicalRecordID uint64 `gorm:"column:medical_record_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	OwnerKind       string `gorm:"column:owner_kind;type:text;primaryKey;autoIncrement:false;not null" json:"-"`
	Revision        int64  `gorm:"column:revision;type:bigint;not null" json:"-"`
}

func (MedicalRecordDependencyVersion) TableName() string { return "medical_record_dependency_versions" }

// MedicalRecordImageDeletion maps the current medical_record_image_deletions storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type MedicalRecordImageDeletion struct {
	ID                 uint64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID           uint64     `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	MedicalRecordID    uint64     `gorm:"column:medical_record_id;type:bigint;not null" json:"-"`
	ImageID            uint64     `gorm:"column:image_id;type:bigint;not null" json:"-"`
	ImageObjectKey     string     `gorm:"column:image_object_key;type:text;not null;default:''" json:"-"`
	ThumbnailObjectKey string     `gorm:"column:thumbnail_object_key;type:text;not null;default:''" json:"-"`
	StorageStatus      string     `gorm:"column:storage_status;type:text;not null" json:"-"`
	AttemptCount       int        `gorm:"column:attempt_count;type:integer;not null;default:0" json:"-"`
	LastError          string     `gorm:"column:last_error;type:text;not null;default:''" json:"-"`
	ActorStaffID       uint64     `gorm:"column:actor_staff_id;type:bigint;not null" json:"-"`
	CreatedAt          time.Time  `gorm:"column:created_at;type:timestamptz;not null;default:NOW();autoCreateTime:false;autoUpdateTime:false" json:"-"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;type:timestamptz;not null;default:NOW();autoCreateTime:false;autoUpdateTime:false" json:"-"`
	DeletedAt          *time.Time `gorm:"column:deleted_at;type:timestamptz" json:"-"`
}

func (MedicalRecordImageDeletion) TableName() string { return "medical_record_image_deletions" }

// MedicalRecordImageUploadQuota maps the current medical_record_image_upload_quota storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type MedicalRecordImageUploadQuota struct {
	ID            uint64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID      uint64     `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	StaffID       uint64     `gorm:"column:staff_id;type:bigint;not null" json:"-"`
	DeclaredBytes int64      `gorm:"column:declared_bytes;type:bigint;not null" json:"-"`
	AcquiredAt    time.Time  `gorm:"column:acquired_at;type:timestamptz;not null;default:now()" json:"-"`
	ReleasedAt    *time.Time `gorm:"column:released_at;type:timestamptz" json:"-"`
}

func (MedicalRecordImageUploadQuota) TableName() string { return "medical_record_image_upload_quota" }

// MedicineDosePatternRevision maps the current medicine_dose_pattern_revisions storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type MedicineDosePatternRevision struct {
	ID                  uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID            uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	PatternID           uint64          `gorm:"column:pattern_id;type:bigint;not null" json:"-"`
	Revision            int64           `gorm:"column:revision;type:bigint;not null" json:"-"`
	PatternName         *string         `gorm:"column:pattern_name;type:text" json:"-"`
	Purpose             *string         `gorm:"column:purpose;type:text" json:"-"`
	AdministrationRoute *string         `gorm:"column:administration_route;type:text" json:"-"`
	ParameterSnapshot   json.RawMessage `gorm:"column:parameter_snapshot;type:jsonb;not null" json:"-"`
	SourceReference     *string         `gorm:"column:source_reference;type:text" json:"-"`
	ApprovalReference   *string         `gorm:"column:approval_reference;type:text" json:"-"`
	ApprovedBy          *uint64         `gorm:"column:approved_by;type:bigint" json:"-"`
	ApprovedAt          *time.Time      `gorm:"column:approved_at;type:timestamptz" json:"-"`
	IsActive            bool            `gorm:"column:is_active;type:boolean;not null" json:"-"`
	OperationID         uint64          `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	RecordedAt          time.Time       `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (MedicineDosePatternRevision) TableName() string { return "medicine_dose_pattern_revisions" }

// OperationAttemptOutcome maps the current operation_attempt_outcomes storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type OperationAttemptOutcome struct {
	ID         uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID   uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	BindingID  uint64          `gorm:"column:binding_id;type:bigint;not null" json:"-"`
	AttemptKey uuid.UUID       `gorm:"column:attempt_key;type:uuid;not null" json:"-"`
	Outcome    string          `gorm:"column:outcome;type:text;not null" json:"-"`
	Details    json.RawMessage `gorm:"column:details;type:jsonb;not null" json:"-"`
	RecordedAt time.Time       `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (OperationAttemptOutcome) TableName() string { return "operation_attempt_outcomes" }

// OperationBinding maps the current operation_bindings storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type OperationBinding struct {
	ID                    uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID              uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	OwnerKind             string    `gorm:"column:owner_kind;type:text;not null" json:"-"`
	ActorKind             string    `gorm:"column:actor_kind;type:text;not null" json:"-"`
	ActorRef              string    `gorm:"column:actor_ref;type:text;not null" json:"-"`
	Intent                string    `gorm:"column:intent;type:text;not null" json:"-"`
	OperationKey          uuid.UUID `gorm:"column:operation_key;type:uuid;not null" json:"-"`
	Fingerprint           []byte    `gorm:"column:fingerprint;type:bytea;not null" json:"-"`
	FingerprintKeyVersion string    `gorm:"column:fingerprint_key_version;type:text;not null" json:"-"`
	CreatedAt             time.Time `gorm:"column:created_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (OperationBinding) TableName() string { return "operation_bindings" }

// OperationResult maps the current operation_results storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type OperationResult struct {
	ClinicID          uint64          `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	BindingID         uint64          `gorm:"column:binding_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	ResultRefs        json.RawMessage `gorm:"column:result_refs;type:jsonb;not null" json:"-"`
	CommittedVersions json.RawMessage `gorm:"column:committed_versions;type:jsonb;not null" json:"-"`
	EffectSummary     json.RawMessage `gorm:"column:effect_summary;type:jsonb;not null" json:"-"`
	RecordedAt        time.Time       `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (OperationResult) TableName() string { return "operation_results" }

// PreventiveMedicationObservation maps the current preventive_medication_observations storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type PreventiveMedicationObservation struct {
	ID             uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID       uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	PrescriptionID uint64    `gorm:"column:prescription_id;type:bigint;not null" json:"-"`
	CreatedAt      time.Time `gorm:"column:created_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
	DisplayYear    int       `gorm:"column:display_year;type:integer;not null" json:"-"`
}

func (PreventiveMedicationObservation) TableName() string {
	return "preventive_medication_observations"
}

// PreventiveMedicationRevision maps the current preventive_medication_revisions storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type PreventiveMedicationRevision struct {
	ID                     uint64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID               uint64     `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	ObservationID          uint64     `gorm:"column:observation_id;type:bigint;not null" json:"-"`
	Revision               int64      `gorm:"column:revision;type:bigint;not null" json:"-"`
	State                  string     `gorm:"column:state;type:text;not null" json:"-"`
	SourceKind             string     `gorm:"column:source_kind;type:text;not null" json:"-"`
	SourceReference        *string    `gorm:"column:source_reference;type:text" json:"-"`
	ConfirmedAt            *time.Time `gorm:"column:confirmed_at;type:timestamptz" json:"-"`
	TimePrecision          *string    `gorm:"column:time_precision;type:text" json:"-"`
	EventYear              *int       `gorm:"column:event_year;type:integer" json:"-"`
	EventMonth             *int       `gorm:"column:event_month;type:integer" json:"-"`
	EventDate              *time.Time `gorm:"column:event_date;type:date" json:"-"`
	TimeDescription        *string    `gorm:"column:time_description;type:text" json:"-"`
	Notes                  *string    `gorm:"column:notes;type:text" json:"-"`
	RecordedBy             uint64     `gorm:"column:recorded_by;type:bigint;not null" json:"-"`
	RecordedAt             time.Time  `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
	OperationID            uint64     `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	PrescriptionID         uint64     `gorm:"column:prescription_id;type:bigint;not null" json:"-"`
	PrescriptionRevisionID uint64     `gorm:"column:prescription_revision_id;type:bigint;not null" json:"-"`
	ChangeReason           *string    `gorm:"column:change_reason;type:text" json:"-"`
}

func (PreventiveMedicationRevision) TableName() string { return "preventive_medication_revisions" }

// PreventivePrescriptionRevision maps the current preventive_prescription_revisions storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type PreventivePrescriptionRevision struct {
	ID                       uint64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID                 uint64     `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	PrescriptionID           uint64     `gorm:"column:prescription_id;type:bigint;not null" json:"-"`
	Revision                 int64      `gorm:"column:revision;type:bigint;not null" json:"-"`
	PrescriptionTitle        *string    `gorm:"column:prescription_title;type:text" json:"-"`
	PrescribedDate           *time.Time `gorm:"column:prescribed_date;type:date" json:"-"`
	MedicineID               *uint64    `gorm:"column:medicine_id;type:bigint" json:"-"`
	MedicineNameSnapshot     *string    `gorm:"column:medicine_name_snapshot;type:text" json:"-"`
	AdministrationConditions *string    `gorm:"column:administration_conditions;type:text" json:"-"`
	PeriodDescription        *string    `gorm:"column:period_description;type:text" json:"-"`
	PeriodStartPrecision     *string    `gorm:"column:period_start_precision;type:text" json:"-"`
	PeriodStartYear          *int       `gorm:"column:period_start_year;type:integer" json:"-"`
	PeriodStartMonth         *int       `gorm:"column:period_start_month;type:integer" json:"-"`
	PeriodStartDate          *time.Time `gorm:"column:period_start_date;type:date" json:"-"`
	PeriodEndPrecision       *string    `gorm:"column:period_end_precision;type:text" json:"-"`
	PeriodEndYear            *int       `gorm:"column:period_end_year;type:integer" json:"-"`
	PeriodEndMonth           *int       `gorm:"column:period_end_month;type:integer" json:"-"`
	PeriodEndDate            *time.Time `gorm:"column:period_end_date;type:date" json:"-"`
	CopiedHistoryRecordID    *uint64    `gorm:"column:copied_history_record_id;type:bigint" json:"-"`
	CopiedHistoryRevision    *int64     `gorm:"column:copied_history_revision;type:bigint" json:"-"`
	HistorySnapshot          *string    `gorm:"column:history_snapshot;type:text" json:"-"`
	Remarks                  *string    `gorm:"column:remarks;type:text" json:"-"`
	ResponsibleStaffID       *uint64    `gorm:"column:responsible_staff_id;type:bigint" json:"-"`
	RecordedBy               uint64     `gorm:"column:recorded_by;type:bigint;not null" json:"-"`
	RecordedAt               time.Time  `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
	OperationID              uint64     `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	HistoryCopiedAt          *time.Time `gorm:"column:history_copied_at;type:timestamptz" json:"-"`
}

func (PreventivePrescriptionRevision) TableName() string { return "preventive_prescription_revisions" }

// PreventivePrescription maps the current preventive_prescriptions storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type PreventivePrescription struct {
	ID                uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID          uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	PetID             uint64    `gorm:"column:pet_id;type:bigint;not null" json:"-"`
	MedicalRecordID   *uint64   `gorm:"column:medical_record_id;type:bigint" json:"-"`
	SourceTreatmentID *uint64   `gorm:"column:source_treatment_id;type:bigint" json:"-"`
	CreatedAt         time.Time `gorm:"column:created_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (PreventivePrescription) TableName() string { return "preventive_prescriptions" }

// RecordEditRightEvent maps the current record_edit_right_events storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type RecordEditRightEvent struct {
	ID                 uint64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID           uint64     `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	MedicalRecordID    uint64     `gorm:"column:medical_record_id;type:bigint;not null" json:"-"`
	PreviousEpoch      int64      `gorm:"column:previous_epoch;type:bigint;not null" json:"-"`
	ResultingEpoch     int64      `gorm:"column:resulting_epoch;type:bigint;not null" json:"-"`
	Intent             string     `gorm:"column:intent;type:text;not null" json:"-"`
	PriorHolderStaffID *uint64    `gorm:"column:prior_holder_staff_id;type:bigint" json:"-"`
	NewHolderStaffID   *uint64    `gorm:"column:new_holder_staff_id;type:bigint" json:"-"`
	SessionRef         *uuid.UUID `gorm:"column:session_ref;type:uuid" json:"-"`
	OperationID        uint64     `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	RecordedAt         time.Time  `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (RecordEditRightEvent) TableName() string { return "record_edit_right_events" }

// RecordEditRight maps the current record_edit_rights storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type RecordEditRight struct {
	ClinicID        uint64     `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	MedicalRecordID uint64     `gorm:"column:medical_record_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	Epoch           int64      `gorm:"column:epoch;type:bigint;not null;default:1" json:"-"`
	State           string     `gorm:"column:state;type:text;not null" json:"-"`
	HolderStaffID   *uint64    `gorm:"column:holder_staff_id;type:bigint" json:"-"`
	SessionRef      *uuid.UUID `gorm:"column:session_ref;type:uuid" json:"-"`
	AcquiredAt      *time.Time `gorm:"column:acquired_at;type:timestamptz" json:"-"`
	ReleasedAt      *time.Time `gorm:"column:released_at;type:timestamptz" json:"-"`
	LastSeenAt      *time.Time `gorm:"column:last_seen_at;type:timestamptz" json:"-"`
}

func (RecordEditRight) TableName() string { return "record_edit_rights" }

// RefundRequest maps the current refund_requests storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type RefundRequest struct {
	ID                  uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID            uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	OwnerID             uint64    `gorm:"column:owner_id;type:bigint;not null" json:"-"`
	FundID              uint64    `gorm:"column:fund_id;type:bigint;not null" json:"-"`
	BillingID           *uint64   `gorm:"column:billing_id;type:bigint" json:"-"`
	RequestedAmount     int64     `gorm:"column:requested_amount;type:bigint;not null" json:"-"`
	State               string    `gorm:"column:state;type:text;not null" json:"-"`
	ActualRefundEventID *uint64   `gorm:"column:actual_refund_event_id;type:bigint" json:"-"`
	Reason              string    `gorm:"column:reason;type:text;not null" json:"-"`
	Version             int64     `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
	OperationID         uint64    `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	RecordedAt          time.Time `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (RefundRequest) TableName() string { return "refund_requests" }

// RevisitIntakeAnswer maps the current revisit_intake_answers storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type RevisitIntakeAnswer struct {
	ID                    uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID              uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	TargetID              uint64    `gorm:"column:target_id;type:bigint;not null" json:"-"`
	Revision              int64     `gorm:"column:revision;type:bigint;not null" json:"-"`
	FormDefinitionVersion int16     `gorm:"column:form_definition_version;type:smallint;not null" json:"-"`
	VisitPurpose          string    `gorm:"column:visit_purpose;type:text;not null" json:"-"`
	Onset                 *string   `gorm:"column:onset;type:text" json:"-"`
	Energy                *string   `gorm:"column:energy;type:text" json:"-"`
	Appetite              *string   `gorm:"column:appetite;type:text" json:"-"`
	CurrentSymptoms       *string   `gorm:"column:current_symptoms;type:text" json:"-"`
	Improvement           *string   `gorm:"column:improvement;type:text" json:"-"`
	ExaminationPreference *string   `gorm:"column:examination_preference;type:text" json:"-"`
	MainConcern           *string   `gorm:"column:main_concern;type:text" json:"-"`
	AuthorCustomerID      uint64    `gorm:"column:author_customer_id;type:bigint;not null" json:"-"`
	AuthorSubjectHash     []byte    `gorm:"column:author_subject_hash;type:bytea;not null" json:"-"`
	AuthorOwnerIDAtSubmit int64     `gorm:"column:author_owner_id_at_submit;type:bigint;not null" json:"-"`
	OperationKey          uuid.UUID `gorm:"column:operation_key;type:uuid;not null" json:"-"`
	ReceivedAt            time.Time `gorm:"column:received_at;type:timestamptz;not null;default:clock_timestamp()" json:"-"`
}

func (RevisitIntakeAnswer) TableName() string { return "revisit_intake_answers" }

// RevisitIntakeSubmitReceipt maps the current revisit_intake_submit_receipts storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type RevisitIntakeSubmitReceipt struct {
	ClinicID           uint64     `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	AuthorCustomerID   uint64     `gorm:"column:author_customer_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	OperationKey       uuid.UUID  `gorm:"column:operation_key;type:uuid;primaryKey;autoIncrement:false;not null" json:"-"`
	AuthorSubjectHash  []byte     `gorm:"column:author_subject_hash;type:bytea;not null" json:"-"`
	FingerprintVersion int16      `gorm:"column:fingerprint_version;type:smallint;not null;default:1" json:"-"`
	RequestHash        []byte     `gorm:"column:request_hash;type:bytea;not null" json:"-"`
	State              string     `gorm:"column:state;type:text;not null;default:'claimed'" json:"-"`
	ResultAnswerID     *uint64    `gorm:"column:result_answer_id;type:bigint" json:"-"`
	ClaimedAt          time.Time  `gorm:"column:claimed_at;type:timestamptz;not null;default:clock_timestamp()" json:"-"`
	CommittedAt        *time.Time `gorm:"column:committed_at;type:timestamptz" json:"-"`
}

func (RevisitIntakeSubmitReceipt) TableName() string { return "revisit_intake_submit_receipts" }

// RevisitIntakeTarget maps the current revisit_intake_targets storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type RevisitIntakeTarget struct {
	ID               uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID         uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	AppointmentID    uint64    `gorm:"column:appointment_id;type:bigint;not null" json:"-"`
	PetID            uint64    `gorm:"column:pet_id;type:bigint;not null" json:"-"`
	TargetGeneration int64     `gorm:"column:target_generation;type:bigint;not null" json:"-"`
	CreatedAt        time.Time `gorm:"column:created_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (RevisitIntakeTarget) TableName() string { return "revisit_intake_targets" }

// StaffRecordDraftOperation maps the current staff_record_draft_operations storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type StaffRecordDraftOperation struct {
	ClinicID        uint64 `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	DraftID         uint64 `gorm:"column:draft_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	OperationID     uint64 `gorm:"column:operation_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	InputGeneration int64  `gorm:"column:input_generation;type:bigint;not null" json:"-"`
}

func (StaffRecordDraftOperation) TableName() string { return "staff_record_draft_operations" }

// StaffRecordDraft maps the current staff_record_drafts storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type StaffRecordDraft struct {
	ID                 uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID           uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	PrincipalAccountID uint64          `gorm:"column:principal_account_id;type:bigint;not null" json:"-"`
	MedicalRecordID    uint64          `gorm:"column:medical_record_id;type:bigint;not null" json:"-"`
	TabKey             string          `gorm:"column:tab_key;type:text;not null" json:"-"`
	InputGeneration    int64           `gorm:"column:input_generation;type:bigint;not null" json:"-"`
	BaseOwnerVersions  json.RawMessage `gorm:"column:base_owner_versions;type:jsonb;not null" json:"-"`
	Ciphertext         []byte          `gorm:"column:ciphertext;type:bytea;not null" json:"-"`
	EncryptionFormat   string          `gorm:"column:encryption_format;type:text;not null" json:"-"`
	KeyReference       string          `gorm:"column:key_reference;type:text;not null" json:"-"`
	CiphertextSHA256   []byte          `gorm:"column:ciphertext_sha256;type:bytea;not null" json:"-"`
	UpdatedAt          time.Time       `gorm:"column:updated_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (StaffRecordDraft) TableName() string { return "staff_record_drafts" }

// VaccinationCertificateArtifact maps the current vaccination_certificate_artifacts storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type VaccinationCertificateArtifact struct {
	ID              uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID        uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	EditionID       uint64    `gorm:"column:edition_id;type:bigint;not null" json:"-"`
	Layout          string    `gorm:"column:layout;type:text;not null" json:"-"`
	TemplateVersion string    `gorm:"column:template_version;type:text;not null" json:"-"`
	RendererVersion string    `gorm:"column:renderer_version;type:text;not null" json:"-"`
	ObjectKey       string    `gorm:"column:object_key;type:text;not null" json:"-"`
	ContentSHA256   []byte    `gorm:"column:content_sha256;type:bytea;not null" json:"-"`
	CreatedAt       time.Time `gorm:"column:created_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (VaccinationCertificateArtifact) TableName() string { return "vaccination_certificate_artifacts" }

// VaccinationCertificateEdition maps the current vaccination_certificate_editions storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type VaccinationCertificateEdition struct {
	ID                       uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID                 uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	SeriesID                 uint64          `gorm:"column:series_id;type:bigint;not null" json:"-"`
	Edition                  int64           `gorm:"column:edition;type:bigint;not null" json:"-"`
	SourceVaccinationVersion int64           `gorm:"column:source_vaccination_version;type:bigint;not null" json:"-"`
	Snapshot                 json.RawMessage `gorm:"column:snapshot;type:jsonb;not null" json:"-"`
	PreviousEditionID        *uint64         `gorm:"column:previous_edition_id;type:bigint" json:"-"`
	IssuedBy                 uint64          `gorm:"column:issued_by;type:bigint;not null" json:"-"`
	IssuedAt                 time.Time       `gorm:"column:issued_at;type:timestamptz;not null;default:clock_timestamp()" json:"-"`
}

func (VaccinationCertificateEdition) TableName() string { return "vaccination_certificate_editions" }

// VaccinationCertificateOperation maps the current vaccination_certificate_operations storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type VaccinationCertificateOperation struct {
	ClinicID      uint64    `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	OperationKey  uuid.UUID `gorm:"column:operation_key;type:uuid;primaryKey;autoIncrement:false;not null" json:"-"`
	RequestHash   []byte    `gorm:"column:request_hash;type:bytea;not null" json:"-"`
	OperationKind string    `gorm:"column:operation_kind;type:text;not null" json:"-"`
	EditionID     uint64    `gorm:"column:edition_id;type:bigint;not null" json:"-"`
	ArtifactID    uint64    `gorm:"column:artifact_id;type:bigint;not null" json:"-"`
	PerformedBy   uint64    `gorm:"column:performed_by;type:bigint;not null" json:"-"`
	CommittedAt   time.Time `gorm:"column:committed_at;type:timestamptz;not null;default:clock_timestamp()" json:"-"`
}

func (VaccinationCertificateOperation) TableName() string {
	return "vaccination_certificate_operations"
}

// VaccinationCertificateSery maps the current vaccination_certificate_series storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type VaccinationCertificateSery struct {
	ID              uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID        uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	VaccinationID   uint64    `gorm:"column:vaccination_id;type:bigint;not null" json:"-"`
	CertificateKind string    `gorm:"column:certificate_kind;type:text;not null" json:"-"`
	CreatedAt       time.Time `gorm:"column:created_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (VaccinationCertificateSery) TableName() string { return "vaccination_certificate_series" }
