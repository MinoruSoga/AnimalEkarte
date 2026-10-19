package model

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"time"
)

// Current migration is authoritative. These are persistence values, not public DTOs or workflow implementations.

// FollowupConfirmation maps the current followup_confirmations storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type FollowupConfirmation struct {
	ID                   uint64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID             uint64     `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	OwnerID              uint64     `gorm:"column:owner_id;type:bigint;not null" json:"-"`
	PetID                uint64     `gorm:"column:pet_id;type:bigint;not null" json:"-"`
	MedicalRecordID      uint64     `gorm:"column:medical_record_id;type:bigint;not null" json:"-"`
	CheckupID            *uint64    `gorm:"column:checkup_id;type:bigint" json:"-"`
	Purpose              string     `gorm:"column:purpose;type:text;not null" json:"-"`
	SourceVersion        int64      `gorm:"column:source_version;type:bigint;not null" json:"-"`
	SourceDigest         []byte     `gorm:"column:source_digest;type:bytea;not null" json:"-"`
	PerformedConfirmedAt time.Time  `gorm:"column:performed_confirmed_at;type:timestamptz;not null" json:"-"`
	ConfirmedBy          uint64     `gorm:"column:confirmed_by;type:bigint;not null" json:"-"`
	DoctorDesignatedBy   *uint64    `gorm:"column:doctor_designated_by;type:bigint" json:"-"`
	DoctorDesignatedAt   *time.Time `gorm:"column:doctor_designated_at;type:timestamptz" json:"-"`
	OperationID          uint64     `gorm:"column:operation_id;type:bigint;not null" json:"-"`
}

func (FollowupConfirmation) TableName() string { return "followup_confirmations" }

// LineDispatchAttempt maps the current line_dispatch_attempts storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type LineDispatchAttempt struct {
	ID                      uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID                uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	IntentID                uint64    `gorm:"column:intent_id;type:bigint;not null" json:"-"`
	AttemptKey              uuid.UUID `gorm:"column:attempt_key;type:uuid;not null" json:"-"`
	AttemptNumber           int64     `gorm:"column:attempt_number;type:bigint;not null" json:"-"`
	PreviousAttemptID       *uint64   `gorm:"column:previous_attempt_id;type:bigint" json:"-"`
	AuthorizedRetryResultID *uint64   `gorm:"column:authorized_retry_result_id;type:bigint" json:"-"`
	ClaimedBy               string    `gorm:"column:claimed_by;type:text;not null" json:"-"`
	ClaimedAt               time.Time `gorm:"column:claimed_at;type:timestamptz;not null" json:"-"`
}

func (LineDispatchAttempt) TableName() string { return "line_dispatch_attempts" }

// LineDispatchIntent maps the current line_dispatch_intents storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type LineDispatchIntent struct {
	ID                     uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID               uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	OwnerID                uint64          `gorm:"column:owner_id;type:bigint;not null" json:"-"`
	BindingRevision        int64           `gorm:"column:binding_revision;type:bigint;not null" json:"-"`
	AppointmentID          *uint64         `gorm:"column:appointment_id;type:bigint" json:"-"`
	FollowupConfirmationID *uint64         `gorm:"column:followup_confirmation_id;type:bigint" json:"-"`
	Purpose                string          `gorm:"column:purpose;type:text;not null" json:"-"`
	OperationID            uint64          `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	DestinationSnapshot    json.RawMessage `gorm:"column:destination_snapshot;type:jsonb;not null" json:"-"`
	PayloadDigest          []byte          `gorm:"column:payload_digest;type:bytea;not null" json:"-"`
	ConfirmedAt            time.Time       `gorm:"column:confirmed_at;type:timestamptz;not null" json:"-"`
	ConfirmedPayload       json.RawMessage `gorm:"column:confirmed_payload;type:jsonb;not null" json:"-"`
	MessageType            *string         `gorm:"column:message_type;type:text" json:"-"`
	ContentSummary         *string         `gorm:"column:content_summary;type:text" json:"-"`
}

func (LineDispatchIntent) TableName() string { return "line_dispatch_intents" }

// LineDispatchResult maps the current line_dispatch_results storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type LineDispatchResult struct {
	ID         uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID   uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	AttemptID  uint64          `gorm:"column:attempt_id;type:bigint;not null" json:"-"`
	IntentID   uint64          `gorm:"column:intent_id;type:bigint;not null" json:"-"`
	Revision   int64           `gorm:"column:revision;type:bigint;not null" json:"-"`
	State      string          `gorm:"column:state;type:text;not null" json:"-"`
	ExternalID *string         `gorm:"column:external_id;type:text" json:"-"`
	Evidence   json.RawMessage `gorm:"column:evidence;type:jsonb;not null" json:"-"`
	RecordedAt time.Time       `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
	OccurredAt *time.Time      `gorm:"column:occurred_at;type:timestamptz" json:"-"`
}

func (LineDispatchResult) TableName() string { return "line_dispatch_results" }

// LstepMigrationProgress maps the current lstep_migration_progress storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type LstepMigrationProgress struct {
	ID           uint64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID     uint64     `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	OwnerID      uint64     `gorm:"column:owner_id;type:bigint;not null" json:"-"`
	Status       string     `gorm:"column:status;type:varchar(20);not null;default:'pending'" json:"-"`
	TagsAdded    int        `gorm:"column:tags_added;type:int;not null;default:0" json:"-"`
	TagsFailed   int        `gorm:"column:tags_failed;type:int;not null;default:0" json:"-"`
	ErrorMessage *string    `gorm:"column:error_message;type:text" json:"-"`
	StartedAt    *time.Time `gorm:"column:started_at;type:timestamptz" json:"-"`
	CompletedAt  *time.Time `gorm:"column:completed_at;type:timestamptz" json:"-"`
}

func (LstepMigrationProgress) TableName() string { return "lstep_migration_progress" }

// LstepNextVisitSyncRetry maps the current lstep_next_visit_sync_retries storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type LstepNextVisitSyncRetry struct {
	ClinicID          uint64    `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	OwnerID           uint64    `gorm:"column:owner_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	Revision          int64     `gorm:"column:revision;type:bigint;not null" json:"-"`
	CreatedAt         time.Time `gorm:"column:created_at;type:timestamptz;not null;default:NOW();autoCreateTime:false;autoUpdateTime:false" json:"-"`
	UpdatedAt         time.Time `gorm:"column:updated_at;type:timestamptz;not null;default:NOW();autoCreateTime:false;autoUpdateTime:false" json:"-"`
	CleanupLineUserID *string   `gorm:"column:cleanup_line_user_id;type:varchar(64)" json:"-"`
}

func (LstepNextVisitSyncRetry) TableName() string { return "lstep_next_visit_sync_retries" }

// ManualArticleOperation maps the current manual_article_operations storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type ManualArticleOperation struct {
	ID                 uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID           uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	ArticleID          uint64          `gorm:"column:article_id;type:bigint;not null" json:"-"`
	OperationID        uint64          `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	Intent             string          `gorm:"column:intent;type:text;not null" json:"-"`
	PreviousVersion    int64           `gorm:"column:previous_version;type:bigint;not null" json:"-"`
	ResultingVersion   int64           `gorm:"column:resulting_version;type:bigint;not null" json:"-"`
	ContentVersionID   *uint64         `gorm:"column:content_version_id;type:bigint" json:"-"`
	PublishedVersionID *uint64         `gorm:"column:published_version_id;type:bigint" json:"-"`
	RestoredFromID     *uint64         `gorm:"column:restored_from_id;type:bigint" json:"-"`
	BeforeState        json.RawMessage `gorm:"column:before_state;type:jsonb;not null" json:"-"`
	AfterState         json.RawMessage `gorm:"column:after_state;type:jsonb;not null" json:"-"`
	RecordedBy         uint64          `gorm:"column:recorded_by;type:bigint;not null" json:"-"`
	RecordedAt         time.Time       `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (ManualArticleOperation) TableName() string { return "manual_article_operations" }

// OwnerLineBindingEvent maps the current owner_line_binding_events storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type OwnerLineBindingEvent struct {
	ID          uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID    uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	OwnerID     uint64          `gorm:"column:owner_id;type:bigint;not null" json:"-"`
	Revision    int64           `gorm:"column:revision;type:bigint;not null" json:"-"`
	Intent      string          `gorm:"column:intent;type:text;not null" json:"-"`
	BeforeState json.RawMessage `gorm:"column:before_state;type:jsonb;not null" json:"-"`
	AfterState  json.RawMessage `gorm:"column:after_state;type:jsonb;not null" json:"-"`
	OperationID uint64          `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	RecordedAt  time.Time       `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (OwnerLineBindingEvent) TableName() string { return "owner_line_binding_events" }

// OwnerLineBindingState maps the current owner_line_binding_states storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type OwnerLineBindingState struct {
	ClinicID       uint64     `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	OwnerID        uint64     `gorm:"column:owner_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"-"`
	Revision       int64      `gorm:"column:revision;type:bigint;not null;default:1" json:"-"`
	State          string     `gorm:"column:state;type:text;not null" json:"-"`
	LineCustomerID *uint64    `gorm:"column:line_customer_id;type:bigint" json:"-"`
	VerifiedBy     *uint64    `gorm:"column:verified_by;type:bigint" json:"-"`
	VerifiedAt     *time.Time `gorm:"column:verified_at;type:timestamptz" json:"-"`
	OperationID    uint64     `gorm:"column:operation_id;type:bigint;not null" json:"-"`
}

func (OwnerLineBindingState) TableName() string { return "owner_line_binding_states" }

// SupportDeliveryAttempt maps the current support_delivery_attempts storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type SupportDeliveryAttempt struct {
	ID                      uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID                uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	DestinationID           uint64    `gorm:"column:destination_id;type:bigint;not null" json:"-"`
	AttemptKey              uuid.UUID `gorm:"column:attempt_key;type:uuid;not null" json:"-"`
	AttemptNumber           int64     `gorm:"column:attempt_number;type:bigint;not null" json:"-"`
	PreviousAttemptID       *uint64   `gorm:"column:previous_attempt_id;type:bigint" json:"-"`
	AuthorizedRetryResultID *uint64   `gorm:"column:authorized_retry_result_id;type:bigint" json:"-"`
	ClaimedBy               string    `gorm:"column:claimed_by;type:text;not null" json:"-"`
	ClaimedAt               time.Time `gorm:"column:claimed_at;type:timestamptz;not null" json:"-"`
}

func (SupportDeliveryAttempt) TableName() string { return "support_delivery_attempts" }

// SupportDeliveryDestination maps the current support_delivery_destinations storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type SupportDeliveryDestination struct {
	ID             uint64 `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID       uint64 `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	IntentID       uint64 `gorm:"column:intent_id;type:bigint;not null" json:"-"`
	Provider       string `gorm:"column:provider;type:text;not null" json:"-"`
	DestinationRef string `gorm:"column:destination_ref;type:text;not null" json:"-"`
}

func (SupportDeliveryDestination) TableName() string { return "support_delivery_destinations" }

// SupportDeliveryIntent maps the current support_delivery_intents storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type SupportDeliveryIntent struct {
	ID               uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID         uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	ReportID         uint64          `gorm:"column:report_id;type:bigint;not null" json:"-"`
	ReportVersion    int64           `gorm:"column:report_version;type:bigint;not null" json:"-"`
	OperationID      uint64          `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	ConfirmedPayload json.RawMessage `gorm:"column:confirmed_payload;type:jsonb;not null" json:"-"`
	PayloadDigest    []byte          `gorm:"column:payload_digest;type:bytea;not null" json:"-"`
	DestinationCount int             `gorm:"column:destination_count;type:integer;not null" json:"-"`
	ConfirmedBy      uint64          `gorm:"column:confirmed_by;type:bigint;not null" json:"-"`
	ConfirmedAt      time.Time       `gorm:"column:confirmed_at;type:timestamptz;not null" json:"-"`
	CreatedAt        time.Time       `gorm:"column:created_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (SupportDeliveryIntent) TableName() string { return "support_delivery_intents" }

// SupportDeliveryResult maps the current support_delivery_results storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type SupportDeliveryResult struct {
	ID             uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID       uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	AttemptID      uint64          `gorm:"column:attempt_id;type:bigint;not null" json:"-"`
	DestinationID  uint64          `gorm:"column:destination_id;type:bigint;not null" json:"-"`
	ResultRevision int64           `gorm:"column:result_revision;type:bigint;not null" json:"-"`
	ResultState    string          `gorm:"column:result_state;type:text;not null" json:"-"`
	ExternalID     *string         `gorm:"column:external_id;type:text" json:"-"`
	Evidence       json.RawMessage `gorm:"column:evidence;type:jsonb;not null" json:"-"`
	RecordedAt     time.Time       `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (SupportDeliveryResult) TableName() string { return "support_delivery_results" }

// SupportLLMBudget maps the current support_llm_budgets storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type SupportLLMBudget struct {
	Scope         string    `gorm:"column:scope;type:text;primaryKey;autoIncrement:false;not null" json:"-"`
	Model         string    `gorm:"column:model;type:text;primaryKey;autoIncrement:false;not null" json:"-"`
	NextAllowedAt time.Time `gorm:"column:next_allowed_at;type:timestamptz;not null" json:"-"`
}

func (SupportLLMBudget) TableName() string { return "support_llm_budgets" }

// SupportManualEmbedding maps the current support_manual_embeddings storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type SupportManualEmbedding struct {
	CacheKey       string          `gorm:"column:cache_key;type:text;primaryKey;autoIncrement:false;not null" json:"-"`
	Model          string          `gorm:"column:model;type:text;not null" json:"-"`
	TaskType       string          `gorm:"column:task_type;type:text;not null" json:"-"`
	ChunkerVersion int             `gorm:"column:chunker_version;type:integer;not null" json:"-"`
	Dimensions     int             `gorm:"column:dimensions;type:integer;not null" json:"-"`
	Embedding      pq.Float32Array `gorm:"column:embedding;type:real[];not null" json:"-"`
	CreatedAt      time.Time       `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (SupportManualEmbedding) TableName() string { return "support_manual_embeddings" }
