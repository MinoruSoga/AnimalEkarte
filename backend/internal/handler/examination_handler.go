package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// ListExaminations godoc
func (h *Handler) ListExaminations(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	q := newListExaminationQuery(c.Request.URL.Query())
	filters, err := q.toServiceFilters()
	if err != nil {
		RespondError(c, err)
		return
	}

	exams, total, err := h.svc.Examination.List(
		c.Request.Context(),
		clinicID,
		filters.PetID,
		filters.OwnerID,
		filters.Status,
		filters.StartDate,
		filters.EndDate,
		page,
		limit,
	)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(mapSlice(exams, toExaminationResponse), total, page, limit))
}

// GetExamination godoc
func (h *Handler) GetExamination(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	exam, err := h.svc.Examination.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toExaminationResponse(exam))
}

// CreateExamination godoc
func (h *Handler) CreateExamination(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input createExaminationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	exam, err := h.svc.Examination.Create(c.Request.Context(), clinicID, input.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/examinations/%d", exam.ID))
	c.JSON(http.StatusCreated, toExaminationResponse(exam))
}

// UpdateExamination godoc
func (h *Handler) UpdateExamination(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var input updateExaminationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	exam, err := h.svc.Examination.Update(c.Request.Context(), clinicID, id, input.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toExaminationResponse(exam))
}

// ListExaminationItems godoc
//
//	@Summary	検査項目一覧を取得
//	@Tags		Examinations
//	@Produce	json
//	@Param		id	path	int	true	"Examination ID"
//	@Success	200	{object}	examItemsResponse
func (h *Handler) ListExaminationItems(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	items, err := h.svc.Examination.ListItems(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toExamItemsResponse(items))
}

// ReplaceExaminationItems godoc
//
//	@Summary	検査項目を一括置換（PUT セマンティクス: 既存全削除→一括登録）
//	@Tags		Examinations
//	@Accept		json
//	@Produce	json
//	@Param		id		path	int						true	"Examination ID"
//	@Param		body	body	replaceExamItemsRequest	true	"items"
//	@Success	200	{object}	examItemsResponse
func (h *Handler) ReplaceExaminationItems(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req replaceExamItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	saved, err := h.svc.Examination.ReplaceItems(c.Request.Context(), clinicID, id, optionalStaffID(c), req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toExamItemsResponse(saved))
}

// DeleteExamination godoc
func (h *Handler) DeleteExamination(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Examination.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
