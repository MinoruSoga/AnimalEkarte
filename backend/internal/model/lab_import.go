package model

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

// LabImportJobStatus は lab_import_jobs のジョブ状態。
// 許可された遷移は medicalrecord.CanTransitionTo で強制される（BE9-2D sub-batch③ で internal/service から移動）。
type LabImportJobStatus string

const (
	LabImportJobStatusReceived    LabImportJobStatus = "received"
	LabImportJobStatusValidated   LabImportJobStatus = "validated"
	LabImportJobStatusMapped      LabImportJobStatus = "mapped"
	LabImportJobStatusPersisted   LabImportJobStatus = "persisted"
	LabImportJobStatusDuplicate   LabImportJobStatus = "duplicate"
	LabImportJobStatusNeedsReview LabImportJobStatus = "needs_review"
	LabImportJobStatusFailed      LabImportJobStatus = "failed"
	// LabImportJobStatusReverted is the terminal compensating-revert status (TASK-032).
	// Distinct from examination unconfirm; reached only via POST /lab-imports/:id/revert.
	LabImportJobStatusReverted LabImportJobStatus = "reverted"
)

// LabImportSourceType は入力元の種別。
// fixture: Phase 0 テスト用。drwan: Phase BLOCKED（MDB スキーマ未確認）。
type LabImportSourceType string

const (
	LabImportSourceTypeFixture LabImportSourceType = "fixture"
	LabImportSourceTypeDrWan   LabImportSourceType = "drwan"
	LabImportSourceTypeManual  LabImportSourceType = "manual"
	// Device source types are Go constants only until a later migration ADDs the enum values (ADR-007 / F9).
	LabImportSourceTypeFujiNX600    LabImportSourceType = "fuji_nx600"
	LabImportSourceTypeFujiAU10V    LabImportSourceType = "fuji_au10v"
	LabImportSourceTypeArkrayPU4010 LabImportSourceType = "arkray_pu4010"
	// LabImportSourceTypeIDEXXVetLab is the IDEXX VetLab Station PIMS serial path (COM5 / JOU-LAB-X).
	// 9600 8N1. Short frames (<= idexxShortFrameBodyMaxBytes) are discarded.
	LabImportSourceTypeIDEXXVetLab LabImportSourceType = "idexx_vetlab"
)

// LabImportJob は外部検査結果のインポートジョブ。
// source_fingerprint には raw 接続文字列・認証情報を格納しない。
// error_message には安全なメッセージのみ格納し、スタックトレース・PHI 不可。
type LabImportJob struct {
	ID                uuid.UUID           `gorm:"column:id;type:uuid;primaryKey;autoIncrement:false;not null;default:gen_random_uuid()" json:"id"`
	ClinicID          uint64              `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	SourceType        LabImportSourceType `gorm:"column:source_type;type:lab_import_source_type;not null;default:'fixture'" json:"source_type"`
	SourceFingerprint string              `gorm:"column:source_fingerprint;type:varchar(255);not null;default:''" json:"source_fingerprint"`
	Status            LabImportJobStatus  `gorm:"column:status;type:lab_import_job_status;not null;default:'received'" json:"status"`
	RowCount          int                 `gorm:"column:row_count;type:int;not null;default:0" json:"row_count"`
	PersistedCount    int                 `gorm:"column:persisted_count;type:int;not null;default:0" json:"persisted_count"`
	DuplicateCount    int                 `gorm:"column:duplicate_count;type:int;not null;default:0" json:"duplicate_count"`
	NeedsReviewCount  int                 `gorm:"column:needs_review_count;type:int;not null;default:0" json:"needs_review_count"`
	FailedCount       int                 `gorm:"column:failed_count;type:int;not null;default:0" json:"failed_count"`
	ErrorCode         *string             `gorm:"column:error_code;type:varchar(50)" json:"error_code,omitempty"`
	ErrorMessage      *string             `gorm:"column:error_message;type:varchar(1000)" json:"error_message,omitempty"`
	StartedAt         *time.Time          `gorm:"column:started_at;type:timestamptz" json:"started_at,omitempty"`
	FinishedAt        *time.Time          `gorm:"column:finished_at;type:timestamptz" json:"finished_at,omitempty"`
	CreatedAt         time.Time           `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time           `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	PetID             *uint64             `gorm:"column:pet_id;type:bigint" json:"pet_id,omitempty"`
	MeasuredAt        *time.Time          `gorm:"column:measured_at;type:timestamptz" json:"measured_at,omitempty"`
	ReceivedAt        *time.Time          `gorm:"column:received_at;type:timestamptz" json:"received_at,omitempty"`
	DeviceHint        string              `gorm:"column:device_hint;type:varchar(32);not null;default:''" json:"device_hint"`
	SpecimenIDRaw     string              `gorm:"column:specimen_id_raw;type:varchar(64);not null;default:''" json:"specimen_id_raw"`
	UnmappedItemCount int                 `gorm:"column:unmapped_item_count;type:int;not null;default:0" json:"unmapped_item_count"`
	Species           string              `gorm:"column:species;type:varchar(32);not null;default:''" json:"-"`
}

func (LabImportJob) TableName() string { return "lab_import_jobs" }

// LabImportEventType はイベントの種別ラベル。
type LabImportEventType string

const (
	LabImportEventTypeStatusTransition  LabImportEventType = "status_transition"
	LabImportEventTypeValidationResult  LabImportEventType = "validation_result"
	LabImportEventTypeMappingResult     LabImportEventType = "mapping_result"
	LabImportEventTypePersistenceResult LabImportEventType = "persistence_result"
	LabImportEventTypeRetryRequested    LabImportEventType = "retry_requested"
	// LabImportEventTypeUsageTrackingStarted marks that downstream usage receipts
	// are authoritative for this job (TASK-032). Recorded in the same transaction
	// as a successful import commit. Absence means usage_unknown (revert 409).
	LabImportEventTypeUsageTrackingStarted LabImportEventType = "usage_tracking_started"
	// LabImportEventTypeRevertRequested records a compensating revert transition.
	LabImportEventTypeRevertRequested LabImportEventType = "revert_requested"
)

// LabImportEvent は検査インポートジョブの監査イベント。
// PHI・raw デバイスペイロード・接続情報は格納しない。
// error_code は lab_error_taxonomy のコードのみ。スタックトレース不可。
type LabImportEvent struct {
	ID               uint64              `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID         uint64              `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	JobID            uuid.UUID           `gorm:"column:job_id;type:uuid;not null" json:"job_id"`
	EventType        string              `gorm:"column:event_type;type:varchar(50);not null" json:"event_type"`
	FromStatus       *LabImportJobStatus `gorm:"column:from_status;type:lab_import_job_status" json:"from_status,omitempty"`
	ToStatus         *LabImportJobStatus `gorm:"column:to_status;type:lab_import_job_status" json:"to_status,omitempty"`
	RowCount         int                 `gorm:"column:row_count;type:int;not null;default:0" json:"row_count"`
	PersistedCount   int                 `gorm:"column:persisted_count;type:int;not null;default:0" json:"persisted_count"`
	DuplicateCount   int                 `gorm:"column:duplicate_count;type:int;not null;default:0" json:"duplicate_count"`
	NeedsReviewCount int                 `gorm:"column:needs_review_count;type:int;not null;default:0" json:"needs_review_count"`
	ErrorCode        *string             `gorm:"column:error_code;type:varchar(50)" json:"error_code,omitempty"`
	CreatedAt        time.Time           `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
}

func (LabImportEvent) TableName() string { return "lab_import_events" }

// LabInboundBatch は外部入力バッチの DTO。
// raw 接続文字列・認証情報・臨床ナラティブ blob は含めない。
type LabInboundBatch struct {
	SourceType        LabImportSourceType   `json:"source_type"`
	SourceFingerprint string                `json:"source_fingerprint"`
	ReceivedAt        time.Time             `json:"received_at"`
	ResultRows        []LabInboundResultRow `json:"result_rows"`
}

// LabInboundResultRow は入力バッチの 1 行。
// フィクスチャ値はサニタイズ済みであること。raw PHI 不可。
type LabInboundResultRow struct {
	OldPetKey      string `json:"old_pet_key"`
	OldChartKey    string `json:"old_chart_key"`
	OldRowKey      string `json:"old_row_key"`
	ExamDate       string `json:"exam_date"`
	ExamCode       string `json:"exam_code"`
	ExamName       string `json:"exam_name"`
	ItemName       string `json:"item_name"`
	DisplayValue   string `json:"display_value"`
	ReferenceValue string `json:"reference_value"`
}

// LabImportPreviewResponse は preview エンドポイントのレスポンス。
// raw 検査値・PHI は含めない。
type LabImportPreviewResponse struct {
	RowCount        int      `json:"row_count"`
	MappingWarnings []string `json:"mapping_warnings"`
	BlockedReasons  []string `json:"blocked_reasons"`
}

// LabImportCommitResponse は commit エンドポイントのレスポンス。
type LabImportCommitResponse struct {
	JobID            uuid.UUID `json:"job_id"`
	PersistedCount   int       `json:"persisted_count"`
	DuplicateCount   int       `json:"duplicate_count"`
	NeedsReviewCount int       `json:"needs_review_count"`
	FailedCount      int       `json:"failed_count"`
}

// LabImportUsageKind は clinical downstream use / manual mutation の種別。
type LabImportUsageKind string

const (
	LabImportUsageKindExaminationDetail LabImportUsageKind = "examination_detail"
	LabImportUsageKindExaminationItems  LabImportUsageKind = "examination_items"
	LabImportUsageKindLabReport         LabImportUsageKind = "lab_report"
	LabImportUsageKindPrintSnapshot     LabImportUsageKind = "print_snapshot"
	LabImportUsageKindManualMutation    LabImportUsageKind = "manual_mutation"
)

// LabImportUsageReceipt は import 由来 exam の clinical use / manual mutation を記録する append-only 台帳。
// audit_logs には相乗りしない（sink 分離）。
type LabImportUsageReceipt struct {
	ID        uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID  uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	JobID     uuid.UUID `gorm:"column:job_id;type:uuid;not null" json:"job_id"`
	ExamID    uint64    `gorm:"column:exam_id;type:bigint;not null" json:"exam_id"`
	UseKind   string    `gorm:"column:use_kind;type:text;not null" json:"use_kind"`
	ActorID   *uint64   `gorm:"column:actor_id;type:bigint" json:"actor_id,omitempty"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
}

func (LabImportUsageReceipt) TableName() string { return "lab_import_usage_receipts" }

// LabImportExamRetraction は compensating revert 時の parent 不変スナップショット。
type LabImportExamRetraction struct {
	ID             uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID       uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	JobID          uuid.UUID       `gorm:"column:job_id;type:uuid;not null" json:"job_id"`
	ExamID         uint64          `gorm:"column:exam_id;type:bigint;not null" json:"exam_id"`
	ActorID        *uint64         `gorm:"column:actor_id;type:bigint" json:"actor_id,omitempty"`
	Reason         string          `gorm:"column:reason;type:text;not null" json:"reason"`
	ParentSnapshot json.RawMessage `gorm:"column:parent_snapshot;type:jsonb;not null" json:"parent_snapshot"`
	CreatedAt      time.Time       `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
}

func (LabImportExamRetraction) TableName() string { return "lab_import_exam_retractions" }

// LabImportExamRetractionItem は retraction の item 不変スナップショット。
// exam_results 本体は hard delete しない。
type LabImportExamRetractionItem struct {
	ID           uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID     uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	RetractionID uint64          `gorm:"column:retraction_id;type:bigint;not null" json:"retraction_id"`
	JobID        uuid.UUID       `gorm:"column:job_id;type:uuid;not null" json:"job_id"`
	ExamID       uint64          `gorm:"column:exam_id;type:bigint;not null" json:"exam_id"`
	ItemSnapshot json.RawMessage `gorm:"column:item_snapshot;type:jsonb;not null" json:"item_snapshot"`
	SortOrder    int             `gorm:"column:sort_order;type:int;not null;default:0" json:"sort_order"`
	CreatedAt    time.Time       `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
}

func (LabImportExamRetractionItem) TableName() string { return "lab_import_exam_retraction_items" }

// LabImportRevertReceipt は POST /lab-imports/:id/revert の冪等レシート。
type LabImportRevertReceipt struct {
	ID               uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID         uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	JobID            uuid.UUID       `gorm:"column:job_id;type:uuid;not null" json:"job_id"`
	IdempotencyKey   uuid.UUID       `gorm:"column:idempotency_key;type:uuid;not null" json:"idempotency_key"`
	RequestHash      string          `gorm:"column:request_hash;type:text;not null" json:"request_hash"`
	Reason           string          `gorm:"column:reason;type:text;not null" json:"reason"`
	ActorID          *uint64         `gorm:"column:actor_id;type:bigint" json:"actor_id,omitempty"`
	ResultStatus     string          `gorm:"column:result_status;type:text;not null" json:"result_status"`
	RetractedExamIds json.RawMessage `gorm:"column:retracted_exam_ids;type:jsonb;not null;default:'[]'::jsonb" json:"-"`
	CreatedAt        time.Time       `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
}

func (LabImportRevertReceipt) TableName() string { return "lab_import_revert_receipts" }

// LabImportRevertResponse は revert エンドポイントのレスポンス。
type LabImportRevertResponse struct {
	JobID            uuid.UUID `json:"job_id"`
	Status           string    `json:"status"`
	RetractedExamIDs []uint64  `json:"retracted_exam_ids"`
	IdempotentReplay bool      `json:"idempotent_replay"`
}
