package trimming

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// ListTrimmings はトリミング予約一覧を返す（BE-119: appointments ベース）
func (h *Handler) ListTrimmings(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceTrimming), "view") {
		return
	}

	page, limit, err := httpapi.ParsePagination(c)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	filters, err := newListTrimmingQuery(c.Request.URL.Query()).toServiceFilters()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	appts, total, err := h.svc.Trimming.List(
		c.Request.Context(),
		clinicID,
		filters.PetID,
		filters.OwnerID,
		filters.StartDate,
		filters.EndDate,
		page,
		limit,
	)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	responses := h.signedTrimmingResponses(c.Request.Context(), clinicID, appts)
	c.JSON(http.StatusOK, httpapi.NewPaginatedResponse(responses, total, page, limit))
}

// GetTrimming はトリミング予約詳細を返す
func (h *Handler) GetTrimming(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceTrimming), "view") {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	appt, err := h.svc.Trimming.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, h.signTrimmingResponse(c.Request.Context(), clinicID, appt))
}

// CreateTrimming はトリミング予約を作成する
func (h *Handler) CreateTrimming(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	actorID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}

	var req createTrimmingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	if req.ReservationTypeID == 0 {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("reservation_type_id は必須です"))
		return
	}

	input := req.toServiceInput()
	input.ActorID = &actorID
	if err := normalizeCreateTrimmingImages(clinicID, input); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	appt, err := h.svc.Trimming.Create(c.Request.Context(), clinicID, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	// 署名失敗は unsigned echo に degrade（commit 済みレコードを 500 にしない）。
	resp := h.signTrimmingResponse(c.Request.Context(), clinicID, appt)
	c.Header("Location", fmt.Sprintf("/api/v1/trimmings/%d", appt.ID))
	c.JSON(http.StatusCreated, resp)
}

// UpdateTrimming はトリミング予約を部分更新する
func (h *Handler) UpdateTrimming(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	actorID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req updateTrimmingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	input := req.toServiceInput()
	input.ActorID = &actorID
	if err := normalizeUpdateTrimmingImages(clinicID, input); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	appt, replaced, err := h.svc.Trimming.Update(c.Request.Context(), clinicID, id, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	h.deleteReplacedTrimmingImages(c.Request.Context(), clinicID, actorID, replaced, appt)
	// 署名失敗は unsigned echo に degrade（commit 済みレコードを 500 にしない）。
	resp := h.signTrimmingResponse(c.Request.Context(), clinicID, appt)
	c.JSON(http.StatusOK, resp)
}

// DeleteTrimming はトリミング予約を削除する（論理削除）
func (h *Handler) DeleteTrimming(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	actorID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	images, err := h.svc.Trimming.DeleteWithImages(c.Request.Context(), clinicID, id, &actorID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	if images != nil {
		h.deleteSupersededTrimmingImages(
			c.Request.Context(), clinicID, id, actorID,
			images.Style, images.Completed, "", "",
		)
	}
	c.Status(http.StatusNoContent)
}
