package medicalrecord

import (
	"strconv"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// ------------------------------------
// Response types
// ------------------------------------

type labImportJobResponse struct {
	ID                string  `json:"id"`
	ClinicID          string  `json:"clinic_id"`
	SourceType        string  `json:"source_type"`
	SourceFingerprint string  `json:"source_fingerprint"`
	Status            string  `json:"status"`
	RowCount          int     `json:"row_count"`
	PersistedCount    int     `json:"persisted_count"`
	DuplicateCount    int     `json:"duplicate_count"`
	NeedsReviewCount  int     `json:"needs_review_count"`
	FailedCount       int     `json:"failed_count"`
	ErrorCode         *string `json:"error_code,omitempty"`
	StartedAt         *string `json:"started_at,omitempty"`
	FinishedAt        *string `json:"finished_at,omitempty"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

type labImportEventResponse struct {
	ID               uint64  `json:"id"`
	ClinicID         string  `json:"clinic_id"`
	JobID            string  `json:"job_id"`
	EventType        string  `json:"event_type"`
	FromStatus       *string `json:"from_status,omitempty"`
	ToStatus         *string `json:"to_status,omitempty"`
	RowCount         int     `json:"row_count"`
	PersistedCount   int     `json:"persisted_count"`
	DuplicateCount   int     `json:"duplicate_count"`
	NeedsReviewCount int     `json:"needs_review_count"`
	ErrorCode        *string `json:"error_code,omitempty"`
	CreatedAt        string  `json:"created_at"`
}

// labImportPreviewResponse は preview エンドポイントのハンドラー境界 DTO（P7 準拠）。
type labImportPreviewResponse struct {
	RowCount        int      `json:"row_count"`
	MappingWarnings []string `json:"mapping_warnings"`
	BlockedReasons  []string `json:"blocked_reasons"`
}

// labImportCommitResponse は commit エンドポイントのハンドラー境界 DTO（P7 準拠）。
// JobID を string に変換してフロントエンドの uint64 精度問題を回避する。
type labImportCommitResponse struct {
	JobID            string `json:"job_id"`
	PersistedCount   int    `json:"persisted_count"`
	DuplicateCount   int    `json:"duplicate_count"`
	NeedsReviewCount int    `json:"needs_review_count"`
	FailedCount      int    `json:"failed_count"`
}

// labImportRevertResponse は compensating revert のハンドラー境界 DTO（TASK-032）。
type labImportRevertResponse struct {
	JobID            string   `json:"job_id"`
	Status           string   `json:"status"`
	RetractedExamIDs []string `json:"retracted_exam_ids"`
	IdempotentReplay bool     `json:"idempotent_replay"`
}

// ------------------------------------
// Conversion functions (P18)
// ------------------------------------

func toLabImportPreviewResponse(r *model.LabImportPreviewResponse) labImportPreviewResponse {
	warnings := r.MappingWarnings
	if warnings == nil {
		warnings = []string{}
	}
	blocked := r.BlockedReasons
	if blocked == nil {
		blocked = []string{}
	}
	return labImportPreviewResponse{
		RowCount:        r.RowCount,
		MappingWarnings: warnings,
		BlockedReasons:  blocked,
	}
}

func toLabImportCommitResponse(r *model.LabImportCommitResponse) labImportCommitResponse {
	return labImportCommitResponse{
		JobID:            r.JobID.String(),
		PersistedCount:   r.PersistedCount,
		DuplicateCount:   r.DuplicateCount,
		NeedsReviewCount: r.NeedsReviewCount,
		FailedCount:      r.FailedCount,
	}
}

func toLabImportJobResponse(j *model.LabImportJob) labImportJobResponse {
	r := labImportJobResponse{
		ID:                j.ID.String(),
		ClinicID:          strconv.FormatUint(j.ClinicID, 10),
		SourceType:        string(j.SourceType),
		SourceFingerprint: j.SourceFingerprint,
		Status:            string(j.Status),
		RowCount:          j.RowCount,
		PersistedCount:    j.PersistedCount,
		DuplicateCount:    j.DuplicateCount,
		NeedsReviewCount:  j.NeedsReviewCount,
		FailedCount:       j.FailedCount,
		ErrorCode:         j.ErrorCode,
		CreatedAt:         httpapi.LocalTimeRFC3339(j.CreatedAt),
		UpdatedAt:         httpapi.LocalTimeRFC3339(j.UpdatedAt),
	}
	if j.StartedAt != nil {
		s := httpapi.LocalTimeRFC3339(*j.StartedAt)
		r.StartedAt = &s
	}
	if j.FinishedAt != nil {
		s := httpapi.LocalTimeRFC3339(*j.FinishedAt)
		r.FinishedAt = &s
	}
	return r
}

func toLabImportEventResponse(e *model.LabImportEvent) labImportEventResponse {
	r := labImportEventResponse{
		ID:               e.ID,
		ClinicID:         strconv.FormatUint(e.ClinicID, 10),
		JobID:            e.JobID.String(),
		EventType:        string(e.EventType),
		RowCount:         e.RowCount,
		PersistedCount:   e.PersistedCount,
		DuplicateCount:   e.DuplicateCount,
		NeedsReviewCount: e.NeedsReviewCount,
		ErrorCode:        e.ErrorCode,
		CreatedAt:        httpapi.LocalTimeRFC3339(e.CreatedAt),
	}
	if e.FromStatus != nil {
		s := string(*e.FromStatus)
		r.FromStatus = &s
	}
	if e.ToStatus != nil {
		s := string(*e.ToStatus)
		r.ToStatus = &s
	}
	return r
}

func toLabImportRevertResponse(r *model.LabImportRevertResponse) labImportRevertResponse {
	ids := make([]string, len(r.RetractedExamIDs))
	for i, id := range r.RetractedExamIDs {
		ids[i] = strconv.FormatUint(id, 10)
	}
	return labImportRevertResponse{
		JobID:            r.JobID.String(),
		Status:           r.Status,
		RetractedExamIDs: ids,
		IdempotentReplay: r.IdempotentReplay,
	}
}
