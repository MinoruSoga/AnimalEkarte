package medicalrecord

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// LabReportHandler serves the read-only lab-report HTTP boundary (job summaries / exam detail).
// PII-safe: the response DTOs omit owner/pet names and raw result_summary (Phase 4B.1 決定3).
type LabReportHandler struct {
	query LabReportQueryService
}

// NewLabReportHandler initializes a LabReportHandler.
func NewLabReportHandler(query LabReportQueryService) *LabReportHandler {
	return &LabReportHandler{query: query}
}

// GetLabJobReportSummaries godoc
// GET /api/v1/lab-reports/jobs/:job_id/summaries
// job_id に紐づく exam サマリ一覧を返す（clinic scope 必須）。
// Permission: ResourceLabImport "view"（Phase 4B.1 決定3）。
// PII-safe: owner 名・pet 名・result_summary を含まない。
func (h *LabReportHandler) GetLabJobReportSummaries(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceLabImport), "view") {
		return
	}
	jobID, ok := httpapi.ParseUUIDParam(c, "job_id")
	if !ok {
		return
	}

	summaries, err := h.query.ListJobReportSummaries(c.Request.Context(), clinicID, jobID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(summaries, toLabExamReportSummaryResponse))
}

// GetLabExamReport godoc
// GET /api/v1/lab-reports/exams/:exam_id
// exam_id の詳細 DTO を返す（clinic scope 必須）。
// Permission: ResourceLabImport "view"（Phase 4B.1 決定3）。
// PII-safe: owner 名・pet 名・result_summary を含まない。
func (h *LabReportHandler) GetLabExamReport(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceLabImport), "view") {
		return
	}
	examIDStr := c.Param("exam_id")
	examID, err := strconv.ParseUint(examIDStr, 10, 64)
	if err != nil || examID == 0 {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("exam_id must be a positive integer"))
		return
	}

	detail, err := h.query.GetExamReport(c.Request.Context(), clinicID, examID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLabExamReportDetailResponse(detail))
}
