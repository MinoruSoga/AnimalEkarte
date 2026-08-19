package medicalrecord

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// LabImportHandler serves the lab-import saga HTTP boundary (preview / commit / job / events / revert).
// audit is nil-safe: a nil LabAuditLogger disables the best-effort audit trail without changing
// the import flow. revert may be nil only in legacy tests that do not exercise compensating revert.
type LabImportHandler struct {
	resultImport  LabResultImportService
	job           LabImportJobService
	audit         LabAuditLogger
	revert        LabImportRevertService
	deviceMasters LabDeviceItemMasterService
	deviceReceive LabDeviceReceiveService
}

// NewLabImportHandler initializes a LabImportHandler. audit may be nil (audit trail disabled).
// Optional revert service is accepted as a trailing argument for TASK-032 composition.
func NewLabImportHandler(
	resultImport LabResultImportService,
	job LabImportJobService,
	audit LabAuditLogger,
	revert ...LabImportRevertService,
) *LabImportHandler {
	var rev LabImportRevertService
	if len(revert) > 0 {
		rev = revert[0]
	}
	return &LabImportHandler{resultImport: resultImport, job: job, audit: audit, revert: rev}
}

// PreviewLabImport godoc
// POST /api/v1/lab-imports/preview — バッチ内容を検証してサマリを返す（DB 書き込みなし）。
func (h *LabImportHandler) PreviewLabImport(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	var req labImportPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	batch := model.LabInboundBatch{
		SourceType:        model.LabImportSourceType(req.SourceType),
		SourceFingerprint: req.SourceFingerprint,
		ReceivedAt:        time.Now(),
	}
	for i := range req.ResultRows {
		batch.ResultRows = append(batch.ResultRows, model.LabInboundResultRow{
			OldPetKey:      req.ResultRows[i].OldPetKey,
			OldChartKey:    req.ResultRows[i].OldChartKey,
			OldRowKey:      req.ResultRows[i].OldRowKey,
			ExamDate:       req.ResultRows[i].ExamDate,
			ExamCode:       req.ResultRows[i].ExamCode,
			ExamName:       req.ResultRows[i].ExamName,
			ItemName:       req.ResultRows[i].ItemName,
			DisplayValue:   req.ResultRows[i].DisplayValue,
			ReferenceValue: req.ResultRows[i].ReferenceValue,
		})
	}

	if la := h.labAudit(); la != nil {
		la.LogPreviewRequested(c.Request.Context(), clinicID, httpapi.OptionalStaffID(c), req.SourceType, len(batch.ResultRows))
	}

	preview, err := h.resultImport.Preview(c.Request.Context(), clinicID, batch)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	if la := h.labAudit(); la != nil && len(preview.BlockedReasons) > 0 {
		la.LogSourceBlocked(c.Request.Context(), clinicID, httpapi.OptionalStaffID(c),
			req.SourceType, "preview", model.LabBlockedReasonSourceTypeBlocked)
	}

	c.JSON(http.StatusOK, toLabImportPreviewResponse(preview))
}

// CommitLabImport godoc
// POST /api/v1/lab-imports — fixture バッチを commit して exams + job を永続化する。
// source_type=fixture 以外は 400 を返す（Phase 3 制約）。
func (h *LabImportHandler) CommitLabImport(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	var req labImportCommitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	batch, err := toBatch(req.Batch)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	inputs, err := toExamInputs(clinicID, req.Inputs)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	actorID := httpapi.OptionalStaffID(c)
	if la := h.labAudit(); la != nil {
		la.LogCommitRequested(c.Request.Context(), clinicID, actorID, req.Batch.SourceType, len(req.Inputs))
	}

	result, err := h.resultImport.Commit(c.Request.Context(), clinicID, batch, inputs)
	if err != nil {
		if la := h.labAudit(); la != nil {
			la.LogCommitFailed(c.Request.Context(), clinicID, actorID, errorCategory(err))
		}
		httpapi.RespondError(c, err)
		return
	}

	if la := h.labAudit(); la != nil {
		la.LogCommitSucceeded(c.Request.Context(), clinicID, actorID, result.JobID, CommitAuditCounts{
			RowCount:       result.PersistedCount + result.DuplicateCount + result.FailedCount,
			PersistedCount: result.PersistedCount,
			DuplicateCount: result.DuplicateCount,
			FailedCount:    result.FailedCount,
		})
	}

	c.Header("Location", fmt.Sprintf("/api/v1/lab-imports/%s", result.JobID.String()))
	c.JSON(http.StatusCreated, toLabImportCommitResponse(result))
}

// GetLabImportJob godoc
// GET /api/v1/lab-imports/:job_id — ジョブ詳細を返す。
func (h *LabImportHandler) GetLabImportJob(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	jobID, ok := httpapi.ParseUUIDParam(c, "job_id")
	if !ok {
		return
	}

	job, err := h.job.GetJob(c.Request.Context(), clinicID, jobID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLabImportJobResponse(job))
}

// ListLabImportEvents godoc
// GET /api/v1/lab-imports/:job_id/events — ジョブのイベント一覧を返す（作成昇順）。
func (h *LabImportHandler) ListLabImportEvents(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	jobID, ok := httpapi.ParseUUIDParam(c, "job_id")
	if !ok {
		return
	}

	events, err := h.job.ListEvents(c.Request.Context(), clinicID, jobID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	resp := make([]labImportEventResponse, len(events))
	for i, e := range events {
		resp[i] = toLabImportEventResponse(e)
	}
	c.JSON(http.StatusOK, resp)
}

// RevertLabImport godoc
// POST /api/v1/lab-imports/:job_id/revert — compensating terminal transition persisted → reverted.
// Distinct from examination unconfirm (DEC-57 / TASK-032). Requires lab-import:edit, reason, and
// Idempotency-Key header (UUID).
func (h *LabImportHandler) RevertLabImport(c *gin.Context) {
	if h.revert == nil {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("lab import revert service is not configured"))
		return
	}
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	jobID, ok := httpapi.ParseUUIDParam(c, "job_id")
	if !ok {
		return
	}
	actorID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}

	var req labImportRevertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	idempotencyKeyRaw := c.GetHeader("Idempotency-Key")
	idempotencyKey, err := uuid.Parse(idempotencyKeyRaw)
	if err != nil || idempotencyKey == uuid.Nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("Idempotency-Key must be a UUID"))
		return
	}

	actor := actorID
	result, err := h.revert.Revert(c.Request.Context(), RevertLabImportInput{
		ClinicID:       clinicID,
		JobID:          jobID,
		ActorID:        &actor,
		Reason:         req.Reason,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLabImportRevertResponse(result))
}

// labAudit は nil-safe な LabAuditLogger アクセサ。
// Handler が LabAudit を持たない場合（テスト等）は nil を返す。
func (h *LabImportHandler) labAudit() LabAuditLogger {
	return h.audit
}

// errorCategory は apperrors エラーから audit 用の LabAuditErrorCategory を返す。
// 生のエラーメッセージは返さない（PII 漏洩防止）。
func errorCategory(err error) model.LabAuditErrorCategory {
	if err == nil {
		return model.LabAuditErrorCategoryInternal
	}
	switch {
	case apperrors.IsInvalidInput(err):
		return model.LabAuditErrorCategoryInvalidInput
	case apperrors.IsNotFound(err):
		return model.LabAuditErrorCategoryNotFound
	case errors.Is(err, apperrors.ErrForbidden):
		return model.LabAuditErrorCategoryForbidden
	case errors.Is(err, apperrors.ErrUnauthorized):
		return model.LabAuditErrorCategoryUnauthorized
	default:
		return model.LabAuditErrorCategoryInternal
	}
}
