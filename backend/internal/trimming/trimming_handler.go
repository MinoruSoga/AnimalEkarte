package trimming

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// ListTrimmings はトリミング予約一覧を返す（BE-119: appointments ベース）
func (h *Handler) ListTrimmings(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
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
	responses := httpapi.MapSlice(appts, toTrimmingResponse)
	c.JSON(http.StatusOK, httpapi.NewPaginatedResponse(responses, total, page, limit))
}

// GetTrimming はトリミング予約詳細を返す
func (h *Handler) GetTrimming(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
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
	c.JSON(http.StatusOK, toTrimmingResponse(appt))
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
	appt, err := h.svc.Trimming.Create(c.Request.Context(), clinicID, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/trimmings/%d", appt.ID))
	c.JSON(http.StatusCreated, toTrimmingResponse(appt))
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
	appt, err := h.svc.Trimming.Update(c.Request.Context(), clinicID, id, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTrimmingResponse(appt))
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
	if err := h.svc.Trimming.Delete(c.Request.Context(), clinicID, id, &actorID); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
