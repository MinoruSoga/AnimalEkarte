package billing

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// EstimateHandler は EstimateService の HTTP handler。
type EstimateHandler struct {
	svc           EstimateService
	hasPermission httpapi.PermissionChecker
}

// NewEstimateHandler は EstimateHandler を構築する。
func NewEstimateHandler(svc EstimateService, hasPermission httpapi.PermissionChecker) *EstimateHandler {
	return &EstimateHandler{svc: svc, hasPermission: hasPermission}
}

// ListEstimates godoc
func (h *EstimateHandler) ListEstimates(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	page, limit, err := httpapi.ParsePagination(c)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	filters, err := newListEstimateQuery(c.Request.URL.Query()).toServiceFilters()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	estimates, total, err := h.svc.List(c.Request.Context(), clinicID, filters.OwnerID, filters.MedicalRecordID, filters.Status, page, limit)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.NewPaginatedResponse(httpapi.MapSlice(estimates, toEstimateResponse), total, page, limit))
}

// GetEstimate godoc
func (h *EstimateHandler) GetEstimate(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	estimate, err := h.svc.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toEstimateResponse(estimate))
}

// CreateEstimate godoc
func (h *EstimateHandler) CreateEstimate(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}

	var req createEstimateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	// BUG-372: discount_amount にゼロ以外を指定する場合は権限要
	if err := httpapi.RequireDiscountCreateInt(c, h.hasPermission, req.DiscountAmount); err != nil {
		httpapi.RespondError(c, err)
		return
	}

	ctx := c.Request.Context()
	estimate, err := h.svc.Create(ctx, clinicID, req.toServiceInput(staffID))
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/estimates/%d", estimate.ID))
	c.JSON(http.StatusCreated, toEstimateResponse(estimate))
}

// UpdateEstimate godoc
func (h *EstimateHandler) UpdateEstimate(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}

	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req updateEstimateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	// BUG-372: discount_amount を変更する場合は既存値と比較し権限チェック
	if req.DiscountAmount != nil {
		existing, err := h.svc.GetByID(c.Request.Context(), clinicID, id)
		if err != nil {
			httpapi.RespondError(c, err)
			return
		}
		if err := httpapi.RequireDiscountEditInt(c, h.hasPermission, req.DiscountAmount, existing.DiscountAmount); err != nil {
			httpapi.RespondError(c, err)
			return
		}
	}

	ctx := c.Request.Context()
	estimate, err := h.svc.Update(ctx, clinicID, id, req.toServiceInput(staffID))
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toEstimateResponse(estimate))
}

// DeleteEstimate godoc
func (h *EstimateHandler) DeleteEstimate(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	actorID := staffID
	if err := h.svc.Delete(c.Request.Context(), clinicID, id, &actorID); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// CreateEstimateSuccessor godoc
// POST /api/v1/estimates/:id/successors — 確定見積の後継ドラフト作成（TASK-012 FINAL B）。
// 権限: estimates:create。原見積は不変。unlock は存在しない。
func (h *EstimateHandler) CreateEstimateSuccessor(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req createEstimateSuccessorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	estimate, err := h.svc.CreateSuccessor(c.Request.Context(), clinicID, id, req.toServiceInput(staffID))
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/estimates/%d", estimate.ID))
	c.JSON(http.StatusCreated, toEstimateResponse(estimate))
}
