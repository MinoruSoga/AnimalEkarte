package model

import (
	"time"

	"github.com/google/uuid"
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
)

// LabImportJob は外部検査結果のインポートジョブ。
// source_fingerprint には raw 接続文字列・認証情報を格納しない。
// error_message には安全なメッセージのみ格納し、スタックトレース・PHI 不可。
type LabImportJob struct {
	ID                uuid.UUID           `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicID          uint64              `gorm:"not null"                                        json:"clinic_id"`
	SourceType        LabImportSourceType `gorm:"type:lab_import_source_type;default:'fixture'"  json:"source_type"`
	SourceFingerprint string              `gorm:"default:''"                                      json:"source_fingerprint"`
	Status            LabImportJobStatus  `gorm:"type:lab_import_job_status;default:'received'"  json:"status"`
	RowCount          int                 `gorm:"default:0"                                       json:"row_count"`
	PersistedCount    int                 `gorm:"default:0"                                       json:"persisted_count"`
	DuplicateCount    int                 `gorm:"default:0"                                       json:"duplicate_count"`
	NeedsReviewCount  int                 `gorm:"default:0"                                       json:"needs_review_count"`
	FailedCount       int                 `gorm:"default:0"                                       json:"failed_count"`
	ErrorCode         *string             `                                                       json:"error_code,omitempty"`
	ErrorMessage      *string             `                                                       json:"error_message,omitempty"`
	StartedAt         *time.Time          `                                                       json:"started_at,omitempty"`
	FinishedAt        *time.Time          `                                                       json:"finished_at,omitempty"`
	CreatedAt         time.Time           `gorm:"autoCreateTime"                                  json:"created_at"`
	UpdatedAt         time.Time           `gorm:"autoUpdateTime"                                  json:"updated_at"`
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
	ID               uint64              `gorm:"primaryKey;autoIncrement"  json:"id"`
	ClinicID         uint64              `gorm:"not null"                  json:"clinic_id"`
	JobID            uuid.UUID           `gorm:"type:uuid;not null"        json:"job_id"`
	EventType        LabImportEventType  `gorm:"not null"                  json:"event_type"`
	FromStatus       *LabImportJobStatus `gorm:"type:lab_import_job_status" json:"from_status,omitempty"`
	ToStatus         *LabImportJobStatus `gorm:"type:lab_import_job_status" json:"to_status,omitempty"`
	RowCount         int                 `gorm:"default:0"                 json:"row_count"`
	PersistedCount   int                 `gorm:"default:0"                 json:"persisted_count"`
	DuplicateCount   int                 `gorm:"default:0"                 json:"duplicate_count"`
	NeedsReviewCount int                 `gorm:"default:0"                 json:"needs_review_count"`
	ErrorCode        *string             `                                 json:"error_code,omitempty"`
	CreatedAt        time.Time           `gorm:"autoCreateTime"            json:"created_at"`
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
	ID        uint64             `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID  uint64             `gorm:"not null"                 json:"clinic_id"`
	JobID     uuid.UUID          `gorm:"type:uuid;not null"       json:"job_id"`
	ExamID    uint64             `gorm:"not null"                 json:"exam_id"`
	UseKind   LabImportUsageKind `gorm:"not null"                 json:"use_kind"`
	ActorID   *uint64            `                                json:"actor_id,omitempty"`
	CreatedAt time.Time          `gorm:"autoCreateTime"           json:"created_at"`
}

func (LabImportUsageReceipt) TableName() string { return "lab_import_usage_receipts" }

// LabImportExamRetraction は compensating revert 時の parent 不変スナップショット。
type LabImportExamRetraction struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID       uint64    `gorm:"not null"                 json:"clinic_id"`
	JobID          uuid.UUID `gorm:"type:uuid;not null"       json:"job_id"`
	ExamID         uint64    `gorm:"not null"                 json:"exam_id"`
	ActorID        *uint64   `                                json:"actor_id,omitempty"`
	Reason         string    `gorm:"not null"                 json:"reason"`
	ParentSnapshot string    `gorm:"type:jsonb;not null"      json:"parent_snapshot"`
	CreatedAt      time.Time `gorm:"autoCreateTime"           json:"created_at"`
}

func (LabImportExamRetraction) TableName() string { return "lab_import_exam_retractions" }

// LabImportExamRetractionItem は retraction の item 不変スナップショット。
// exam_results 本体は hard delete しない。
type LabImportExamRetractionItem struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID     uint64    `gorm:"not null"                 json:"clinic_id"`
	RetractionID uint64    `gorm:"not null"                 json:"retraction_id"`
	JobID        uuid.UUID `gorm:"type:uuid;not null"       json:"job_id"`
	ExamID       uint64    `gorm:"not null"                 json:"exam_id"`
	ItemSnapshot string    `gorm:"type:jsonb;not null"      json:"item_snapshot"`
	SortOrder    int       `gorm:"not null;default:0"       json:"sort_order"`
	CreatedAt    time.Time `gorm:"autoCreateTime"           json:"created_at"`
}

func (LabImportExamRetractionItem) TableName() string { return "lab_import_exam_retraction_items" }

// LabImportRevertReceipt は POST /lab-imports/:id/revert の冪等レシート。
type LabImportRevertReceipt struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID         uint64    `gorm:"not null"                 json:"clinic_id"`
	JobID            uuid.UUID `gorm:"type:uuid;not null"       json:"job_id"`
	IdempotencyKey   uuid.UUID `gorm:"type:uuid;not null"       json:"idempotency_key"`
	RequestHash      string    `gorm:"not null"                 json:"request_hash"`
	Reason           string    `gorm:"not null"                 json:"reason"`
	ActorID          *uint64   `                                json:"actor_id,omitempty"`
	ResultStatus     string    `gorm:"not null"                 json:"result_status"`
	RetractedExamIDs string    `gorm:"type:jsonb;not null"      json:"retracted_exam_ids"`
	CreatedAt        time.Time `gorm:"autoCreateTime"           json:"created_at"`
}

func (LabImportRevertReceipt) TableName() string { return "lab_import_revert_receipts" }

// LabImportRevertResponse は revert エンドポイントのレスポンス。
type LabImportRevertResponse struct {
	JobID            uuid.UUID `json:"job_id"`
	Status           string    `json:"status"`
	RetractedExamIDs []uint64  `json:"retracted_exam_ids"`
	IdempotentReplay bool      `json:"idempotent_replay"`
}
