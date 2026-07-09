package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// GetLabJobReportSummaries godoc
// GET /api/v1/lab-reports/jobs/:job_id/summaries
// job_id に紐づく exam サマリ一覧を返す（clinic scope 必須）。
// Permission: ResourceLabImport "view"（Phase 4B.1 決定3）。
// PII-safe: owner 名・pet 名・result_summary を含まない。
func (h *Handler) GetLabJobReportSummaries(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	jobID, ok := parseUUIDParam(c, "job_id")
	if !ok {
		return
	}

	summaries, err := h.svc.LabReportQuery.ListJobReportSummaries(c.Request.Context(), clinicID, jobID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(summaries, toLabExamReportSummaryResponse))
}

// GetLabExamReport godoc
// GET /api/v1/lab-reports/exams/:exam_id
// exam_id の詳細 DTO を返す（clinic scope 必須）。
// Permission: ResourceLabImport "view"（Phase 4B.1 決定3）。
// PII-safe: owner 名・pet 名・result_summary を含まない。
func (h *Handler) GetLabExamReport(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	examIDStr := c.Param("exam_id")
	examID, err := strconv.ParseUint(examIDStr, 10, 64)
	if err != nil || examID == 0 {
		RespondError(c, apperrors.WrapInvalidInput("exam_id must be a positive integer"))
		return
	}

	detail, err := h.svc.LabReportQuery.GetExamReport(c.Request.Context(), clinicID, examID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLabExamReportDetailResponse(detail))
}

// RegisterLabReportRoutes は lab report エンドポイントのルートを登録する。
// Permission: ResourceLabImport "view"（Phase 4B.1 決定3）。
func (h *Handler) RegisterLabReportRoutes(rg *gin.RouterGroup) {
	perm := h.RequirePermission
	g := rg.Group("/lab-reports")
	g.GET("/jobs/:job_id/summaries", perm(string(model.ResourceLabImport), "view"), h.GetLabJobReportSummaries)
	g.GET("/exams/:exam_id", perm(string(model.ResourceLabImport), "view"), h.GetLabExamReport)
}
