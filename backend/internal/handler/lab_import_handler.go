package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ------------------------------------
// Handler methods
// ------------------------------------

// PreviewLabImport godoc
// POST /api/v1/lab-imports/preview — バッチ内容を検証してサマリを返す（DB 書き込みなし）。
func (h *Handler) PreviewLabImport(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req labImportPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
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
		la.LogPreviewRequested(c.Request.Context(), clinicID, optionalStaffID(c), req.SourceType, len(batch.ResultRows))
	}

	preview, err := h.svc.LabResultImport.Preview(c.Request.Context(), clinicID, batch)
	if err != nil {
		RespondError(c, err)
		return
	}

	if la := h.labAudit(); la != nil && len(preview.BlockedReasons) > 0 {
		la.LogSourceBlocked(c.Request.Context(), clinicID, optionalStaffID(c),
			req.SourceType, "preview", model.LabBlockedReasonSourceTypeBlocked)
	}

	c.JSON(http.StatusOK, toLabImportPreviewResponse(preview))
}

// CommitLabImport godoc
// POST /api/v1/lab-imports — fixture バッチを commit して exams + job を永続化する。
// source_type=fixture 以外は 400 を返す（Phase 3 制約）。
func (h *Handler) CommitLabImport(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req labImportCommitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	batch, err := toBatch(req.Batch)
	if err != nil {
		RespondError(c, err)
		return
	}

	inputs, err := toExamInputs(clinicID, req.Inputs)
	if err != nil {
		RespondError(c, err)
		return
	}

	actorID := optionalStaffID(c)
	if la := h.labAudit(); la != nil {
		la.LogCommitRequested(c.Request.Context(), clinicID, actorID, req.Batch.SourceType, len(req.Inputs))
	}

	result, err := h.svc.LabResultImport.Commit(c.Request.Context(), clinicID, batch, inputs)
	if err != nil {
		if la := h.labAudit(); la != nil {
			la.LogCommitFailed(c.Request.Context(), clinicID, actorID, errorCategory(err))
		}
		RespondError(c, err)
		return
	}

	if la := h.labAudit(); la != nil {
		la.LogCommitSucceeded(c.Request.Context(), clinicID, actorID, result.JobID, service.CommitAuditCounts{
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
func (h *Handler) GetLabImportJob(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	jobID, ok := parseUUIDParam(c, "job_id")
	if !ok {
		return
	}

	job, err := h.svc.LabImportJob.GetJob(c.Request.Context(), clinicID, jobID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLabImportJobResponse(job))
}

// ListLabImportEvents godoc
// GET /api/v1/lab-imports/:job_id/events — ジョブのイベント一覧を返す（作成昇順）。
func (h *Handler) ListLabImportEvents(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	jobID, ok := parseUUIDParam(c, "job_id")
	if !ok {
		return
	}

	events, err := h.svc.LabImportJob.ListEvents(c.Request.Context(), clinicID, jobID)
	if err != nil {
		RespondError(c, err)
		return
	}
	resp := make([]labImportEventResponse, len(events))
	for i, e := range events {
		resp[i] = toLabImportEventResponse(e)
	}
	c.JSON(http.StatusOK, resp)
}

// labAudit は nil-safe な LabAuditLogger アクセサ。
// Handler が LabAudit を持たない場合（テスト等）は nil を返す。
func (h *Handler) labAudit() service.LabAuditLogger {
	if h.svc == nil {
		return nil
	}
	return h.svc.LabAudit
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

// RegisterLabImportRoutes は lab import エンドポイントのルートを登録する。
// 全エンドポイントに RequirePermission を付与する（P5 準拠）。
func (h *Handler) RegisterLabImportRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/lab-imports")
	perm := h.RequirePermission
	g.POST("/preview", perm(string(model.ResourceLabImport), "create"), h.PreviewLabImport)
	g.POST("", perm(string(model.ResourceLabImport), "create"), h.CommitLabImport)
	g.GET("/:job_id", perm(string(model.ResourceLabImport), "view"), h.GetLabImportJob)
	g.GET("/:job_id/events", perm(string(model.ResourceLabImport), "view"), h.ListLabImportEvents)
}
