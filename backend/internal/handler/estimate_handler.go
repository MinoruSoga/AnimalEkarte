package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// ListEstimates godoc
func (h *Handler) ListEstimates(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	filters, err := newListEstimateQuery(c.Request.URL.Query()).toServiceFilters()
	if err != nil {
		RespondError(c, err)
		return
	}

	estimates, total, err := h.svc.Estimate.List(c.Request.Context(), clinicID, filters.OwnerID, filters.MedicalRecordID, filters.Status, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(mapSlice(estimates, toEstimateResponse), total, page, limit))
}

// GetEstimate godoc
func (h *Handler) GetEstimate(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	estimate, err := h.svc.Estimate.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toEstimateResponse(estimate))
}

// CreateEstimate godoc
func (h *Handler) CreateEstimate(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := extractStaffID(c)
	if !ok {
		return
	}

	var req createEstimateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	// BUG-372: discount_amount にゼロ以外を指定する場合は権限要
	if err := h.requireDiscountCreateInt(c, req.DiscountAmount); err != nil {
		RespondError(c, err)
		return
	}

	ctx := c.Request.Context()
	estimate, err := h.svc.Estimate.Create(ctx, clinicID, req.toServiceInput(staffID))
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/estimates/%d", estimate.ID))
	c.JSON(http.StatusCreated, toEstimateResponse(estimate))
}

// UpdateEstimate godoc
func (h *Handler) UpdateEstimate(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := extractStaffID(c)
	if !ok {
		return
	}

	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req updateEstimateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	// BUG-372: discount_amount を変更する場合は既存値と比較し権限チェック
	if req.DiscountAmount != nil {
		existing, err := h.svc.Estimate.GetByID(c.Request.Context(), clinicID, id)
		if err != nil {
			RespondError(c, err)
			return
		}
		if err := h.requireDiscountEditInt(c, req.DiscountAmount, existing.DiscountAmount); err != nil {
			RespondError(c, err)
			return
		}
	}

	ctx := c.Request.Context()
	estimate, err := h.svc.Estimate.Update(ctx, clinicID, id, req.toServiceInput(staffID))
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toEstimateResponse(estimate))
}

// DeleteEstimate godoc
func (h *Handler) DeleteEstimate(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := extractStaffID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	actorID := staffID
	if err := h.svc.Estimate.Delete(c.Request.Context(), clinicID, id, &actorID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
