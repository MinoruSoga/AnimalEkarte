package medicalrecord

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// ExaminationHandler serves the examination HTTP boundary. Moved from internal/handler
// (BE9-2D ⑦ Batch C).
type ExaminationHandler struct {
	service ExaminationService
}

// NewExaminationHandler initializes a ExaminationHandler.
func NewExaminationHandler(service ExaminationService) *ExaminationHandler {
	return &ExaminationHandler{service: service}
}

// ListExaminations godoc
func (h *ExaminationHandler) ListExaminations(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceExaminations), "view") {
		return
	}

	page, limit, err := httpapi.ParsePagination(c)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	q := newListExaminationQuery(c.Request.URL.Query())
	filters, err := q.toServiceFilters()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	exams, total, err := h.service.List(
		c.Request.Context(),
		clinicID,
		filters.PetID,
		filters.OwnerID,
		filters.MedicalRecordID,
		filters.Status,
		filters.StartDate,
		filters.EndDate,
		page,
		limit,
		filters.IncludeItems,
	)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	responseMapper := toExaminationResponse
	if filters.IncludeItems {
		responseMapper = toExaminationResponseWithItems
	}
	c.JSON(http.StatusOK, httpapi.NewPaginatedResponse(httpapi.MapSlice(exams, responseMapper), total, page, limit))
}

// GetExamination godoc
func (h *ExaminationHandler) GetExamination(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceExaminations), "view") {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	exam, err := h.service.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toExaminationResponse(exam))
}

// GetExaminationPrintSnapshot godoc
// GET /examinations/:id/print-snapshot?version=
func (h *ExaminationHandler) GetExaminationPrintSnapshot(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceExaminations), "view") {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	version, ok := httpapi.ParseOptionalUint64Query(c, "version")
	if !ok {
		return
	}
	snapshot, err := h.service.GetPrintSnapshot(c.Request.Context(), clinicID, id, version)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toExaminationPrintSnapshotResponse(snapshot))
}

// CreateExamination godoc
func (h *ExaminationHandler) CreateExamination(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	var input createExaminationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	actorID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}

	serviceInput := input.toServiceInput()
	serviceInput.ActorID = &actorID
	exam, err := h.service.Create(c.Request.Context(), clinicID, serviceInput)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/examinations/%d", exam.ID))
	c.JSON(http.StatusCreated, toExaminationResponse(exam))
}

// UpdateExamination godoc
func (h *ExaminationHandler) UpdateExamination(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var input updateExaminationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	actorID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}

	serviceInput := input.toServiceInput()
	serviceInput.ActorID = &actorID
	exam, err := h.service.Update(c.Request.Context(), clinicID, id, serviceInput)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toExaminationResponse(exam))
}

// UnconfirmExamination reopens an immutable official examination as a new working revision.
func (h *ExaminationHandler) UnconfirmExamination(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var request unconfirmExaminationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	actorID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	exam, err := h.service.Unconfirm(c.Request.Context(), clinicID, id, UnconfirmExaminationInput{
		Reason:  request.Reason,
		ActorID: &actorID,
	})
	if err != nil {
		httpapi.RespondError(c, err)
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
func (h *ExaminationHandler) ListExaminationItems(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceExaminations), "view") {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	items, err := h.service.ListItems(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
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
func (h *ExaminationHandler) ReplaceExaminationItems(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req replaceExamItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	saved, err := h.service.ReplaceItems(c.Request.Context(), clinicID, id, httpapi.OptionalStaffID(c), req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toExamItemsResponse(saved))
}

// DeleteExamination godoc
func (h *ExaminationHandler) DeleteExamination(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	actorID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), clinicID, id, &actorID); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
